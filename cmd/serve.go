// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package cmd

import (
	"context"
	"fmt"
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
	"github.com/canonical/identity-platform-admin-ui/pkg/idp"
	"github.com/canonical/identity-platform-admin-ui/pkg/rules"
	"github.com/canonical/identity-platform-admin-ui/pkg/schemas"
	"github.com/canonical/identity-platform-admin-ui/pkg/web"
)

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

	extCfg := web.NewExternalClientsConfig(
		ih.NewClient(specs.HydraAdminURL, specs.Debug),
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
		extCfg.SetAuthorizer(extCfg.OpenFGA())
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
		Kratos:    extCfg.KratosPublic().IdentityAPI(),
		Name:      specs.SchemasConfigMapName,
		Namespace: specs.SchemasConfigMapNamespace,
	}

	rulesConfig := rules.NewConfig(specs.RulesConfigMapName, specs.RulesConfigFileName, specs.RulesConfigMapNamespace, k8sCoreV1, extCfg.OathkeeperPublic().ApiApi())

	if specs.AuthorizationEnabled {
		authorizer := authorization.NewAuthorizer(
			extCfg.OpenFGA(),
			tracer,
			monitor,
			logger,
		)

		if authorizer.ValidateModel(context.Background()) != nil {
			panic("Invalid authorization model provided")
		}
	}

	wpool := pool.NewWorkerPool(specs.OpenFGAWorkersTotal, tracer, monitor, logger)
	defer wpool.Stop()

	router := web.NewRouter(idpConfig, schemasConfig, rulesConfig, extCfg, wpool, web.NewO11yConfig(tracer, monitor, logger))

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
