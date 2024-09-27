package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/pool"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	kClient "github.com/ory/kratos-client-go"
	"github.com/spf13/cobra"
)

var createIdentityCmd = &cobra.Command{
	Use:   "create-identity [/path/to/file.json]",
	Short: "Create an identity",
	Example: `Create an example identity:

cat > ./example.json <<EOF
{
	"schema_id": "default",
	"traits": {
		"email": "foo@example.com"
	},
	"credentials"": {
		"password": {
			"config": {
				"password": "bar"
			}
		}
	}
}
EOF

identity-platform-admin-ui create-identity example.json`,
	Long: `Create an identity using JSON input either from a file or piped through stdin.`,
	Args: cobra.MaximumNArgs(1),
	Run:  createIdentityHandler,
}

func init() {
	rootCmd.AddCommand(createIdentityCmd)
}

func createIdentityHandler(cmd *cobra.Command, args []string) {
	var byteValue []byte
	var err error

	if len(args) == 0 {
		byteValue, err = io.ReadAll(cmd.InOrStdin())
		if err != nil {
			panic(fmt.Errorf("failed to read from STDIN: %v", err))
		}
	} else {
		byteValue, err = os.ReadFile(args[0])
		if err != nil {
			panic(fmt.Errorf("failed to read from the JSON file: %v", err))
		}
	}

	var identityBody kClient.CreateIdentityBody
	if err = json.Unmarshal(byteValue, &identityBody); err != nil {
		panic(fmt.Errorf("failed to parse the identity: %v", err))
	}

	createdIdentity, err := createIdentity(identityBody)
	if err != nil {
		panic(err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Identity created: %s\n", createdIdentity.GetId())
}

func createIdentity(identity kClient.CreateIdentityBody) (*kClient.Identity, error) {
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

	// Create the identity
	identityData, err := service.CreateIdentity(ctx, &identity)
	if err != nil {
		return nil, fmt.Errorf("failed to create the identity: %v", err)
	}

	createdIdentity := &identityData.Identities[0]

	// Send identity creation email
	err = service.SendUserCreationEmail(ctx, createdIdentity)
	if err != nil {
		return nil, fmt.Errorf("failed to send the identity creation email: %v", err)
	}

	return createdIdentity, nil
}
