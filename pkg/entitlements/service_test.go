package entitlements

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"
	openfga "github.com/openfga/go-sdk"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package entitlements -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package entitlements -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package entitlements -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package entitlements -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestV1Service_ListEntitlements(t *testing.T) {
	ctrl, mockOpenFGA, mockLogger, mockTracer, mockMonitor, authModel := setupTest(t)
	defer ctrl.Finish()

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		expectedResult []resources.EntitlementSchema
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "List entitlements successfully",
			setupMocks: func() {
				mockOpenFGA.EXPECT().ReadModel(gomock.Any()).Return(authModel, nil).Times(1)
			},
			contextSetup: func() context.Context {
				return context.Background()
			},
			expectedResult: []resources.EntitlementSchema{
				{
					Entitlement:  "can_create",
					EntityType:   "role",
					ReceiverType: "user,role#assignee,group#member",
				},
				{
					Entitlement:  "can_view",
					EntityType:   "role",
					ReceiverType: "user,user:*,role#assignee,group#member",
				},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(ctx, mockOpenFGA, mockTracer, mockMonitor, mockLogger)

			entitlements, err := s.ListEntitlements(ctx, &resources.GetEntitlementsParams{Filter: openfga.PtrString("role")})

			assert.Equal(t, tc.expectedResult, entitlements)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestV1Service_RawEntitlements(t *testing.T) {
	ctrl, mockOpenFGA, mockLogger, mockTracer, mockMonitor, authModel := setupTest(t)
	defer ctrl.Finish()

	rawAuthModel, _ := json.Marshal(authModel)

	type testCase struct {
		name           string
		setupMocks     func()
		contextSetup   func() context.Context
		expectedResult string
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "List raw entitlements successfully",
			setupMocks: func() {
				mockOpenFGA.EXPECT().ReadModel(gomock.Any()).Return(authModel, nil).Times(1)
			},
			contextSetup: func() context.Context {
				return context.Background()
			},
			expectedResult: string(rawAuthModel),
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			ctx := tc.contextSetup()

			s := NewV1Service(ctx, mockOpenFGA, mockTracer, mockMonitor, mockLogger)

			rawEntitlements, err := s.RawEntitlements(ctx)

			assert.Equal(t, tc.expectedResult, rawEntitlements)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func setupTest(t *testing.T) (
	*gomock.Controller,
	*MockOpenFGAClientInterface,
	*MockLoggerInterface,
	*MockTracer,
	*monitoring.MockMonitorInterface,
	*openfga.AuthorizationModel,
) {
	ctrl := gomock.NewController(t)
	mockOpenFGA := NewMockOpenFGAClientInterface(ctrl)
	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)

	mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
			return ctx, trace.SpanFromContext(ctx)
		},
	)

	metadata := openfga.Metadata{
		Relations: &map[string]openfga.RelationMetadata{
			"can_create": {
				DirectlyRelatedUserTypes: &[]openfga.RelationReference{
					{
						Type: "user",
					},
					{
						Type:     "role",
						Relation: openfga.PtrString("assignee"),
					},
					{
						Type:     "group",
						Relation: openfga.PtrString("member"),
					},
				},
			},
			"can_view": {
				DirectlyRelatedUserTypes: &[]openfga.RelationReference{
					{
						Type: "user",
					},
					{
						Type:     "user",
						Wildcard: &map[string]interface{}{},
					},
					{
						Type:     "role",
						Relation: openfga.PtrString("assignee"),
					},
					{
						Type:     "group",
						Relation: openfga.PtrString("member"),
					},
				},
			},
		},
	}

	authModel := &openfga.AuthorizationModel{
		Id:            "id",
		SchemaVersion: "1.1",
		TypeDefinitions: []openfga.TypeDefinition{
			{
				Type:     "user",
				Metadata: nil,
			},
			{
				Type:     "role",
				Metadata: &metadata,
			},
			{
				Type:     "group",
				Metadata: &metadata,
			},
		},
	}

	return ctrl, mockOpenFGA, mockLogger, mockTracer, mockMonitor, authModel
}
