// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package cmd

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/config"
	ih "github.com/canonical/identity-platform-admin-ui/internal/hydra"
	k8s "github.com/canonical/identity-platform-admin-ui/internal/k8s"
	ik "github.com/canonical/identity-platform-admin-ui/internal/kratos"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring/prometheus"
	io "github.com/canonical/identity-platform-admin-ui/internal/oathkeeper"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/rules"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/ui"
	"github.com/canonical/identity-platform-admin-ui/pkg/web"
)

//go:embed ui/dist
//go:embed ui/dist/assets
var jsFS embed.FS

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve starts the web server",
	Long:  `Launch the web application, list of environment variables is available in the README.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve() {

	specs := new(config.EnvSpec)

	if err := envconfig.Process("", specs); err != nil {
		panic(fmt.Errorf("issues with environment sourcing: %s", err))
	}

	logger := logging.NewLogger(specs.LogLevel, specs.LogFile)
	monitor := prometheus.NewMonitor("identity-admin-ui", logger)
	tracer := tracing.NewTracer(tracing.NewConfig(specs.TracingEnabled, specs.OtelGRPCEndpoint, specs.OtelHTTPEndpoint, logger))

	distFS, err := fs.Sub(jsFS, "ui/dist")
	if err != nil {
		logger.Fatalf("issue with ui files %s", err)
	}

	hydraAdminClient := ih.NewClient(specs.HydraAdminURL, specs.Debug)
	externalConfig := web.NewExternalClientsConfig(
		hydraAdminClient,
		ik.NewClient(specs.KratosAdminURL, specs.Debug),
		ik.NewClient(specs.KratosPublicURL, specs.Debug),
		io.NewClient(specs.OathkeeperPublicURL, specs.Debug),
		openfga.NewClient(
			openfga.NewConfig(
				specs.ApiScheme,
				specs.ApiHost,
				specs.StoreId,
				specs.ApiToken,
				specs.ModelId,
				specs.Debug,
				tracer,
				monitor,
				logger,
			),
		),
		// default to noop client for authorization
		openfga.NewNoopClient(tracer, monitor, logger),
	)

	if specs.AuthorizationEnabled {
		logger.Info("Authorization is enabled")
		externalConfig.SetAuthorizer(externalConfig.OpenFGA())
	}

	k8sCoreV1, err := k8s.NewCoreV1Client(specs.KubeconfigFile)

	if err != nil {
		panic(err)
	}

	// TODO @shipperizer standardize idp, schemas and rules configs
	idpConfig := &idp.Config{
		K8s:       k8sCoreV1,
		Name:      specs.IDPConfigMapName,
		Namespace: specs.IDPConfigMapNamespace,
	}

	schemasConfig := &schemas.Config{
		K8s:       k8sCoreV1,
		Kratos:    externalConfig.KratosPublic().IdentityAPI(),
		Name:      specs.SchemasConfigMapName,
		Namespace: specs.SchemasConfigMapNamespace,
	}

	rulesConfig := rules.NewConfig(specs.RulesConfigMapName, specs.RulesConfigFileName, specs.RulesConfigMapNamespace, k8sCoreV1, externalConfig.OathkeeperPublic().ApiApi())

	uiConfig := &ui.Config{
		DistFS: distFS,
	}

	if specs.AuthorizationEnabled {
		authorizer := authorization.NewAuthorizer(
			externalConfig.OpenFGA(),
			tracer,
			monitor,
			logger,
		)

		if authorizer.ValidateModel(context.Background()) != nil {
			panic("Invalid authorization model provided")
		}
	}

	oauth2Config := authentication.NewAuthenticationConfig(
		specs.AuthenticationEnabled,
		specs.OIDCIssuer,
		specs.OAuth2ClientId,
		specs.OAuth2ClientSecret,
		specs.OAuth2RedirectURI,
		specs.AccessTokenVerificationStrategy,
		specs.OAuth2AuthCookiesTTLSeconds,
		specs.OAuth2UserSessionTTLSeconds,
		specs.OAuth2AuthCookiesEncryptionKey,
		specs.OAuth2CodeGrantScopes,
		ih.NewClient(specs.OIDCIssuer, specs.Debug),
		hydraAdminClient,
	)

	ollyConfig := web.NewO11yConfig(tracer, monitor, logger)

	routerConfig := web.NewRouterConfig(specs.PayloadValidationEnabled, idpConfig, schemasConfig, rulesConfig, uiConfig, externalConfig, oauth2Config, ollyConfig)

	wpool := pool.NewWorkerPool(specs.OpenFGAWorkersTotal, tracer, monitor, logger)
	defer wpool.Stop()

	router := web.NewRouter(routerConfig, wpool)

	logger.Infof("Starting server on port %v", specs.Port)

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", specs.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)

	logger.Desugar().Sync()

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	logger.Info("Shutting down")
	os.Exit(0)

}
