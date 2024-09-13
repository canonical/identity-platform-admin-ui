// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package resources

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/canonical/identity-platform-admin-ui/internal/http/types"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
	"github.com/canonical/identity-platform-admin-ui/pkg/authentication"
	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	"github.com/canonical/rebac-admin-ui-handlers/v1/interfaces"
	v1Resources "github.com/canonical/rebac-admin-ui-handlers/v1/resources"
	"github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package resources -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package resources -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package resources -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package resources -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package resources -destination ./mock_pool.go -source=../../internal/pool/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package resources -destination ./mock_authentication.go -source=../authentication/interfaces.go

func TestV1ServiceImplementsRebacServiceInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var svc interface{} = new(V1Service)

	if _, ok := svc.(interfaces.ResourcesService); !ok {
		t.Fatalf("V1Service doesnt implement interfaces.ResourcesService")
	}
}

func TestV1ServiceListResources(t *testing.T) {
	ctrl, mockStore, mockLogger, mockTracer, mockMonitor, principal := setupTest(t)
	defer ctrl.Finish()

	permissions := []ofga.Permission{
		{Relation: "can_view", Object: "client:grafana"},
		{Relation: "can_view", Object: "client:prometheus"},
		{Relation: "can_view", Object: "group:admin"},
		{Relation: "can_edit", Object: "group:random"},
	}

	currPageToken := map[string]string{
		"clients": "page-token",
	}

	nextPageToken := map[string]string{
		"clients": "new-page-token",
	}

	paginator := types.NewTokenPaginator(mockTracer, mockLogger)
	paginator.SetTokens(context.Background(), currPageToken)
	header, _ := paginator.PaginationHeader(context.Background())
	type testCase struct {
		name           string
		contextSetup   func() context.Context
		input          *v1Resources.GetResourcesParams
		expectedResult *v1Resources.PaginatedResponse[v1Resources.Resource]
		expectedError  error
	}
	clientType := "client"

	tests := []testCase{
		{
			name: "Successfully retrieves user resources with type",

			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			input: &v1Resources.GetResourcesParams{
				NextToken:  &header,
				EntityType: &clientType,
			},
			expectedResult: &v1Resources.PaginatedResponse[v1Resources.Resource]{
				Meta: v1Resources.ResponseMeta{Size: 2},
				Data: []v1Resources.Resource{
					{Entity: v1Resources.Entity{Id: "grafana", Name: "grafana", Type: "client"}},
					{Entity: v1Resources.Entity{Id: "prometheus", Name: "prometheus", Type: "client"}},
				},
			},
			expectedError: nil,
		},
		{
			name: "Error while retrieving permissions",
			contextSetup: func() context.Context {
				ctx := context.Background()
				return authentication.PrincipalContext(ctx, principal)
			},
			expectedResult: nil,
			expectedError:  v1.NewUnknownError("failed to get resources for user mock-subject: error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := test.contextSetup()

			mockStore.EXPECT().ListPermissionsWithFilters(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, ID string, opts ...ofga.ListPermissionsFiltersInterface) ([]ofga.Permission, map[string]string, error) {

					if ID != "mock-subject" {
						t.Errorf("expecting ID to be %s not %s", "mock-subject", ID)
					}

					ptypes := []string{"group", "role", "identity", "scheme", "provider", "client"}
					for _, opt := range opts {
						switch o := opt.(type) {
						case *ofga.TypesFilter:
							if test.input != nil && test.input.EntityType != nil {
								if !reflect.DeepEqual(o.WithFilter().([]string), []string{*test.input.EntityType}) {
									t.Errorf("expecting type filter to be %v not %s", test.input.EntityType, o.WithFilter())
								}
							}
							ptypes = o.WithFilter().([]string)
						case *ofga.RelationFilter:
							if !reflect.DeepEqual(o.WithFilter().(string), ofga.CAN_VIEW_RELATION) {
								t.Errorf("expecting relation filter to be %v not %s", ofga.CAN_VIEW_RELATION, o.WithFilter())
							}

						case *ofga.TokenMapFilter:
							if test.input != nil && test.input.NextToken != nil {
								p := types.NewTokenPaginator(mockTracer, mockLogger)
								p.SetTokens(context.Background(), o.WithFilter().(map[string]string))
								h, _ := paginator.PaginationHeader(ctx)
								if !reflect.DeepEqual(h, *test.input.NextToken) {
									t.Errorf("expecting token map filter to be %s not %s", *test.input.NextToken, h)
								}
							}

						}
					}

					if test.expectedError != nil {
						return nil, nextPageToken, fmt.Errorf("error")
					}

					ps := make([]ofga.Permission, 0)
					types := strings.Join(ptypes, ",")

					for _, p := range permissions {
						rc := p.Relation == ofga.CAN_VIEW_RELATION

						pType := strings.Split(p.Object, ":")[0]
						tc := strings.Contains(types, pType)

						if tc && rc {
							ps = append(ps, p)
						}
					}

					return ps, nextPageToken, nil
				},
			)

			s := NewV1Service(mockStore, mockTracer, mockMonitor, mockLogger)

			result, err := s.ListResources(ctx, test.input)

			if test.expectedError != nil && err == nil {
				t.Errorf("expected error to be %s not %s", test.expectedError, err)
			}

			if err != nil {
				return
			}

			if !reflect.DeepEqual(test.expectedResult.Meta, result.Meta) {
				t.Errorf("expected meta to be %v not %v", test.expectedResult.Meta, result.Meta)
			}
			if !reflect.DeepEqual(test.expectedResult.Data, result.Data) {
				t.Errorf("expected data to be %v not %v", test.expectedResult.Data, result.Data)
			}

			paginator.SetTokens(ctx, nextPageToken)
			expectedToken, _ := paginator.PaginationHeader(ctx)
			if expectedToken != *result.Next.PageToken {
				t.Errorf("expecting to be %s not %s", expectedToken, *result.Next.PageToken)
			}
		})
	}
}

func setupTest(t *testing.T) (
	*gomock.Controller,
	*MockOpenFGAStoreInterface,
	*MockLoggerInterface,
	*MockTracer,
	*monitoring.MockMonitorInterface,
	*authentication.ServicePrincipal,
) {
	ctrl := gomock.NewController(t)
	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockStore := NewMockOpenFGAStoreInterface(ctrl)
	mockProvider := NewMockProviderInterface(ctrl)

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
			return ctx, trace.SpanFromContext(ctx)
		},
	)
	mockProvider.EXPECT().Verifier(gomock.Any()).Return(
		oidc.NewVerifier("", nil, &oidc.Config{
			ClientID:                   "mock-client-id",
			SkipExpiryCheck:            true,
			SkipIssuerCheck:            true,
			InsecureSkipSignatureCheck: true,
		}),
	)

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtb2NrLXN1YmplY3QiLCJhdWQiOiJtb2NrLWNsaWVudC1pZCIsIm5hbWUiOiJKb2huIERvZSIsImlhdCI6MTUxNjIzOTAyMn0.BdspASNsnxeXnqZXZnFnkvv-ClMq0U6X1gCIUrh9V7c"
	principal, _ := authentication.NewJWKSTokenVerifier(mockProvider, "mock-client-id", mockTracer, mockLogger, mockMonitor).VerifyAccessToken(context.TODO(), token)

	return ctrl, mockStore, mockLogger, mockTracer, mockMonitor, principal
}
