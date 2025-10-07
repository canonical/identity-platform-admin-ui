// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package cmd

import (
	"context"
	"embed"
	"errors"
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
	"github.com/canonical/identity-platform-admin-ui/internal/mail"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring/prometheus"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/storage"
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
		main()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve() error {

	specs := new(config.EnvSpec)

	if err := envconfig.Process("", specs); err != nil {
		panic(fmt.Errorf("issues with environment sourcing: %s", err))
	}

	logger := logging.NewLogger(specs.LogLevel)
	defer logger.Sync()
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
		nil,
	)

	k8sCoreV1, err := k8s.NewCoreV1Client(specs.KubeconfigFile)

	if err != nil {
		panic(err)
	}

	// TODO @shipperizer standardize idp and schemas configs
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

	wpool := pool.NewWorkerPool(specs.OpenFGAWorkersTotal, tracer, monitor, logger)
	defer wpool.Stop()

	dbClient := storage.NewDBClient(specs.DSN, true, specs.TracingEnabled, tracer, monitor, logger)
	defer dbClient.Close()

	if specs.AuthorizationEnabled {
		authorizer := authorization.NewAuthorizer(
			externalConfig.OpenFGA(),
			wpool,
			tracer,
			monitor,
			logger,
		)
		logger.Info("Authorization is enabled")
		externalConfig.SetAuthorizer(authorizer)

		if authorizer.ValidateModel(context.Background()) != nil {
			panic("Invalid authorization model provided")
		}
	} else {
		authorizer := authorization.NewAuthorizer(
			openfga.NewNoopClient(tracer, monitor, logger),
			wpool,
			tracer,
			monitor,
			logger,
		)
		logger.Info("Using noop authorizer")
		externalConfig.SetAuthorizer(authorizer)
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

	mailConfig := mail.NewConfig(specs.MailHost, specs.MailPort, specs.MailUsername, specs.MailPassword, specs.MailFromAddress, specs.MailSendTimeoutSeconds)

	ollyConfig := web.NewO11yConfig(tracer, monitor, logger)

	router := web.NewRouter(
		wpool,
		dbClient,
		web.WithIDPConfig(idpConfig),
		web.WithSchemasConfig(schemasConfig),
		web.WithUIDistFS(&distFS),
		web.WithExternalClients(externalConfig),
		web.WithOAuth2Config(oauth2Config),
		web.WithMailConfig(mailConfig),
		web.WithO11y(ollyConfig),
		web.WithContextPath(specs.ContextPath),
		web.WithPayloadValidationEnabled(specs.PayloadValidationEnabled),
	)

	logger.Infof("Starting server on port %v", specs.Port)

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", specs.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	var serverError error
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Security().SystemStartup()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError = fmt.Errorf("server error: %w", err)
			c <- os.Interrupt
		}
	}()

	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	logger.Security().SystemShutdown()
	if err := srv.Shutdown(ctx); err != nil {
		serverError = fmt.Errorf("server shutdown error: %w", err)
	}

	return serverError
}

func main() {
	if err := serve(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}
