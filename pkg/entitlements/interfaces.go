package entitlements

import (
	"context"

	openfga "github.com/openfga/go-sdk"
)

// OpenFGAClientInterface is the interface used to decouple the OpenFGA store implementation.
type OpenFGAClientInterface interface {
	ReadModel(ctx context.Context) (*openfga.AuthorizationModel, error)
}
