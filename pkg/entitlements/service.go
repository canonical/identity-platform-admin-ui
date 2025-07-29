// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package entitlements

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"
	openfga "github.com/openfga/go-sdk"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type V1Service struct {
	ofga ofga.OpenFGAClientInterface

	authModel *openfga.AuthorizationModel

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (s *V1Service) ListEntitlements(ctx context.Context, params *resources.GetEntitlementsParams) ([]resources.EntitlementSchema, error) {
	ctx, span := s.tracer.Start(ctx, "entitlements.V1Service.ListEntitlements")
	defer span.End()

	entitlementSchemas := make([]resources.EntitlementSchema, 0)
	for _, typeDef := range s.authModel.TypeDefinitions {
		if typeDef.Metadata == nil {
			continue
		}

		if params.Filter != nil && !strings.Contains(typeDef.Type, *params.Filter) {
			continue
		}

		relations := *typeDef.GetMetadata().Relations
		for relation, relationMetadata := range relations {
			relationReferences := relationMetadata.GetDirectlyRelatedUserTypes()
			receivers := buildReceivers(relationReferences)
			entitlementSchema := resources.EntitlementSchema{
				Entitlement:  relation,
				EntityType:   typeDef.Type,
				ReceiverType: receivers,
			}

			entitlementSchemas = append(entitlementSchemas, entitlementSchema)
		}
	}

	return entitlementSchemas, nil
}

func (s *V1Service) RawEntitlements(ctx context.Context) (string, error) {
	ctx, span := s.tracer.Start(ctx, "entitlements.V1Service.RawEntitlements")
	defer span.End()

	rawAuthModel, err := json.Marshal(s.authModel)
	if err != nil {
		s.logger.Errorf("failed to serialize the authorization model: %v", err)
		return "", v1.NewUnknownError(fmt.Sprintf("failed to serialize the authorization model: %v", err))
	}

	return string(rawAuthModel), nil
}

func buildReceivers(relationReferences []openfga.RelationReference) string {
	var builder strings.Builder
	for i, ref := range relationReferences {
		if i > 0 {
			builder.WriteString(",")
		}

		builder.WriteString(ref.Type)

		if ref.Relation != nil {
			builder.WriteString("#")
			builder.WriteString(*ref.Relation)
		}

		if ref.Wildcard != nil {
			builder.WriteString(":*")
		}
	}
	return builder.String()
}

func NewV1Service(ofga ofga.OpenFGAClientInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *V1Service {
	authModel, err := ofga.ReadModel(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to read the authorization model: %v", err))
	}

	return &V1Service{
		ofga:      ofga,
		authModel: authModel,
		tracer:    tracer,
		monitor:   monitor,
		logger:    logger,
	}
}
