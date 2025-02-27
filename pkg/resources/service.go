// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package resources

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	v1Resources "github.com/canonical/rebac-admin-ui-handlers/v1/resources"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

// V1Service contains the business logic to deal with resoruces on the Admin UI OpenFGA model
type V1Service struct {
	store OpenFGAStoreInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

// ListResources returns a page of Resource objects of at least `size` elements if available.
func (s *V1Service) ListResources(ctx context.Context, params *v1Resources.GetResourcesParams) (*v1Resources.PaginatedResponse[v1Resources.Resource], error) {
	ctx, span := s.tracer.Start(ctx, "resources.V1Service.ListResources")
	defer span.End()

	principal := authentication.PrincipalFromContext(ctx)
	if principal == nil {
		return nil, v1.NewAuthorizationError("unauthorized")
	}

	paginator := types.NewTokenPaginator(s.tracer, s.logger)
	filters := make([]ofga.ListPermissionsFiltersInterface, 0)

	if params != nil {
		if eType := params.EntityType; eType != nil {
			filters = append(filters,
				ofga.NewTypesFilter(*eType),
			)
		}

		if token := params.NextToken; token != nil {
			err := paginator.LoadFromString(ctx, *token)

			if err == nil {
				filters = append(
					filters,
					ofga.NewTokenMapFilter(paginator.GetAllTokens(ctx)),
				)
			}

		}

	}

	filters = append(filters, ofga.NewRelationFilter(ofga.CAN_VIEW_RELATION))

	// TODO using params.EntityId requires a different OpenFGA operation (namely the Check)
	// not implementing that for now
	resources, pageTokens, err := s.store.ListPermissionsWithFilters(
		ctx,
		principal.Identifier(),
		filters...,
	)

	if err != nil {
		return nil, v1.NewUnknownError(fmt.Sprintf("failed to get resources for user %s: %v", principal.Identifier(), err))
	}

	paginator.SetTokens(ctx, pageTokens)
	metaParam, err := paginator.PaginationHeader(ctx)

	if err != nil {
		s.logger.Errorf("error producing pagination meta param: %s", err)
		metaParam = ""
	}

	r := new(v1Resources.PaginatedResponse[v1Resources.Resource])
	r.Meta = v1Resources.ResponseMeta{Size: len(resources)}
	r.Data = make([]v1Resources.Resource, 0)
	r.Next.PageToken = &metaParam

	for _, resource := range resources {
		res := strings.Split(resource.Object, ":")

		if len(res) != 2 {
			s.logger.Warnf("invalid permission object %v", resource)
			continue

		}
		r.Data = append(
			r.Data,
			v1Resources.Resource{
				Entity: v1Resources.Entity{
					Id:   res[1],
					Name: res[1],
					Type: res[0],
				},
			},
		)
	}

	return r, nil
}

func NewV1Service(store OpenFGAStoreInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *V1Service {
	s := new(V1Service)

	s.store = store
	s.tracer = tracer
	s.monitor = monitor
	s.logger = logger

	return s
}
