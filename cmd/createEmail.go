package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/canonical/identity-platform-admin-ui/internal/authorization"
	"github.com/canonical/identity-platform-admin-ui/internal/config"
	"github.com/canonical/identity-platform-admin-ui/internal/kratos"
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/mail"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/pkg/identities"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var sendInvitationCmd = &cobra.Command{
	Use:   "send-invite [identity-id]",
	Short: "Send an invitation email for an identity",
	Example: `Send an invitation email:

identity-platform-admin-ui send-invite d0bebe60-f4fa-4bdc-b6d6-dff9eb42d287`,
	Long: `Send an invitation email for an identity if previous email sending efforts failed.`,
	Args: cobra.ExactArgs(1),
	Run:  sendInvitationHandler,
}

func init() {
	rootCmd.AddCommand(sendInvitationCmd)
}

func sendInvitationHandler(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewNoopTracer()
	monitor := monitoring.NewNoopMonitor("", logger)
	wpool := pool.NewWorkerPool(1, tracer, monitor, logger)
	defer wpool.Stop()

	// Initialize environment variables
	specs, err := initializeEnv()
	if err != nil {
		panic(err)
	}

	// Initialize identity service
	service := initializeIdentityService(specs, logger, tracer, monitor, wpool)

	// Get the identity
	identityID := args[0]
	identityData, err := service.GetIdentity(ctx, identityID)
	if err != nil {
		panic(fmt.Errorf("failed to get the identity: %v", err))
	}

	ids := identityData.Identities
	if len(ids) == 0 {
		cmd.PrintErrln("No identity found for identity:", identityID)
		os.Exit(1)
	}

	identity := &ids[0]

	// Send invitation email
	err = service.SendUserCreationEmail(ctx, identity)
	if err != nil {
		panic(fmt.Errorf("failed to send the invitation email: %v", err))
	}

	cmd.Printf("Invitation email sent for identity: %s\n", identity.GetId())
}

func initializeEnv() (*config.EnvSpec, error) {
	specs := new(config.EnvSpec)
	if err := envconfig.Process("", specs); err != nil {
		return nil, fmt.Errorf("failed to populate environment variables: %v", err)
	}

	return specs, nil
}

func initializeIdentityService(specs *config.EnvSpec, logger logging.LoggerInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, wpool pool.WorkerPoolInterface) *identities.Service {
	// Set up Kratos client
	kratosClient := kratos.NewClient(specs.KratosAdminURL, specs.Debug)

	// Set up OpenFGA authorization
	openfgaConfig := openfga.NewConfig(
		specs.ApiScheme,
		specs.ApiHost,
		specs.StoreId,
		specs.ApiToken,
		specs.ModelId,
		specs.Debug,
		tracer,
		monitor,
		logger,
	)
	openfgaClient := openfga.NewClient(openfgaConfig)
	authorizer := authorization.NewAuthorizer(openfgaClient, wpool, tracer, monitor, logger)

	// Set up mail service
	mailConfig := mail.NewConfig(
		specs.MailHost,
		specs.MailPort,
		specs.MailUsername,
		specs.MailPassword,
		specs.MailFromAddress,
		specs.MailSendTimeoutSeconds,
	)
	mailService := mail.NewEmailService(mailConfig, tracer, monitor, logger)

	return identities.NewService(kratosClient.IdentityAPI(), authorizer, mailService, tracer, monitor, logger)
}
