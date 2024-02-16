// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package openfga

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/kelseyhightower/envconfig"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_client.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_openfga_client.go github.com/openfga/go-sdk/client SdkClientListObjectsRequestInterface,SdkClientReadRequestInterface,SdkClientWriteRequestInterface,SdkClientBatchCheckRequestInterface
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package openfga -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestNewClientAPIClientImplementsInterface(t *testing.T) {
	type EnvSpec struct {
		ApiScheme            string `envconfig:"openfga_api_scheme" default:"http"`
		ApiHost              string `envconfig:"openfga_api_host" default:"127.0.0.1:3000"`
		ApiToken             string `envconfig:"openfga_api_token" default:"42"`
		StoreID              string `envconfig:"openfga_store_id" default:"01HPSTD8C1V7Y35D7NMG2VRCXP"`
		AuthorizationModelID string `envconfig:"openfga_authorization_model_id" default:"01HPSTRTWY7SPT0W1357KRT4AE"`
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)

	specs := new(EnvSpec)

	if err := envconfig.Process("", specs); err != nil {
		t.Fatalf("issues with environment sourcing: %s", err)
	}

	cfg := NewConfig(
		specs.ApiScheme,
		specs.ApiHost,
		specs.StoreID,
		specs.ApiToken,
		specs.AuthorizationModelID,
		true,
		mockTracer,
		mockMonitor,
		mockLogger,
	)

	c := NewClient(cfg)
	c.c = mockOpenFGAClient

	if !reflect.TypeOf(c.APIClient()).Implements(
		reflect.TypeOf((*OpenFGAClientInterface)(nil)).Elem(),
	) {
		t.Fatal("APIClient doesn't implement interface OpenFGAClientInterface")
	}
}

func TestClientListObjectsSuccess(t *testing.T) {
	type input struct {
		user     string
		relation string
		object   string
	}

	tests := []struct {
		name     string
		input    input
		expected []string
		output   []string
	}{
		{
			name:     "empty result",
			input:    input{user: "user:me", relation: "member", object: "group"},
			expected: []string{},
			output:   []string{},
		},
		{
			name:     "full result",
			input:    input{user: "user:me", relation: "member", object: "group"},
			expected: []string{"group:test", "group:admin"},
			output:   []string{"test", "admin"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
			mockRequest := NewMockSdkClientListObjectsRequestInterface(ctrl)

			c := Client{
				c:       mockOpenFGAClient,
				tracer:  mockTracer,
				monitor: mockMonitor,
				logger:  mockLogger,
			}

			body := client.ClientListObjectsRequest{
				User:     test.input.user,
				Relation: test.input.relation,
				Type:     test.input.object,
			}
			expected := client.ClientListObjectsResponse{}
			expected.SetObjects(test.expected)

			mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.ListObjects").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGAClient.EXPECT().ListObjects(gomock.Any()).Return(mockRequest)
			mockRequest.EXPECT().Body(body).Return(mockRequest)
			mockOpenFGAClient.EXPECT().ListObjectsExecute(mockRequest).Times(1).Return(&expected, nil)

			r, err := c.ListObjects(context.TODO(), test.input.user, test.input.relation, test.input.object)

			if err != nil {
				t.Errorf("error while calling ListObjects %s", err)
			}

			if !reflect.DeepEqual(r, test.output) {
				t.Errorf("Objects returned %v, compared %v", r, test.output)
			}
		})
	}
}

func TestClientListObjectsFails(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
	mockRequest := NewMockSdkClientListObjectsRequestInterface(ctrl)

	c := Client{
		c:       mockOpenFGAClient,
		tracer:  mockTracer,
		monitor: mockMonitor,
		logger:  mockLogger,
	}

	body := client.ClientListObjectsRequest{
		User:     "user:me",
		Relation: "member",
		Type:     "group",
	}

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.ListObjects").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockOpenFGAClient.EXPECT().ListObjects(gomock.Any()).Return(mockRequest)
	mockRequest.EXPECT().Body(body).Return(mockRequest)
	mockOpenFGAClient.EXPECT().ListObjectsExecute(mockRequest).Times(1).Return(nil, fmt.Errorf("error"))

	r, err := c.ListObjects(context.TODO(), "user:me", "member", "group")

	if err == nil {
		t.Errorf("error expected while calling ListObjects")
	}

	if r != nil {
		t.Errorf("result expected to be nil ")
	}
}

func TestClientReadTuplesSuccess(t *testing.T) {
	type input struct {
		user     string
		relation string
		object   string
		cToken   string
	}

	tests := []struct {
		name   string
		input  input
		output []openfga.Tuple
	}{
		{
			name:   "empty result",
			input:  input{user: "user:me", relation: "member", object: "group", cToken: ""},
			output: []openfga.Tuple{},
		},
		{
			name:  "full result",
			input: input{user: "", relation: "", object: "", cToken: "xyz"},
			output: []openfga.Tuple{
				{Key: *openfga.NewTupleKey("user:*", "can_view", "group:global")},
				{Key: *openfga.NewTupleKey("user:*", "can_view", "client:global")},
				{Key: *openfga.NewTupleKey("user:*", "can_view", "identity:global")},
				{Key: *openfga.NewTupleKey("user:*", "can_view", "provider:global")},
				{Key: *openfga.NewTupleKey("role:administrator#assignee", "can_edit", "client:github-canonical")},
			},
		},
		{
			name:  "full result",
			input: input{user: "", relation: "assignee", object: "role:administrator", cToken: "abc"},
			output: []openfga.Tuple{
				{Key: *openfga.NewTupleKey("user:joe", "assignee", "role:administrator")},
				{Key: *openfga.NewTupleKey("group:c-level#member", "assignee", "role:administrator")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
			mockRequest := NewMockSdkClientReadRequestInterface(ctrl)

			c := Client{
				c:       mockOpenFGAClient,
				tracer:  mockTracer,
				monitor: mockMonitor,
				logger:  mockLogger,
			}

			body := client.ClientReadRequest{
				User:     &test.input.user,
				Relation: &test.input.relation,
				Object:   &test.input.object,
			}
			expected := client.ClientReadResponse{}
			expected.SetTuples(test.output)

			mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.ReadTuples").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGAClient.EXPECT().Read(gomock.Any()).Return(mockRequest)
			mockRequest.EXPECT().Body(body).Return(mockRequest)
			mockRequest.EXPECT().Options(client.ClientReadOptions{ContinuationToken: &test.input.cToken}).Return(mockRequest)
			mockOpenFGAClient.EXPECT().ReadExecute(mockRequest).Times(1).Return(&expected, nil)

			r, err := c.ReadTuples(context.TODO(), test.input.user, test.input.relation, test.input.object, test.input.cToken)

			if err != nil {
				t.Errorf("error while calling ReadTuples %s", err)
			}

			if !reflect.DeepEqual(r.GetTuples(), test.output) {
				t.Errorf("Objects returned %v, compared %v", r.GetTuples(), test.output)
			}
		})
	}
}

func TestClientReadTuplesFails(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
	mockRequest := NewMockSdkClientReadRequestInterface(ctrl)

	c := Client{
		c:       mockOpenFGAClient,
		tracer:  mockTracer,
		monitor: mockMonitor,
		logger:  mockLogger,
	}

	user := "user:me"
	relation := "member"
	oType := "group"
	cToken := "xyz"

	body := client.ClientReadRequest{
		User:     &user,
		Relation: &relation,
		Object:   &oType,
	}

	mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.ReadTuples").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockOpenFGAClient.EXPECT().Read(gomock.Any()).Return(mockRequest)
	mockRequest.EXPECT().Body(body).Return(mockRequest)
	mockRequest.EXPECT().Options(client.ClientReadOptions{ContinuationToken: &cToken}).Return(mockRequest)
	mockOpenFGAClient.EXPECT().ReadExecute(mockRequest).Times(1).Return(nil, fmt.Errorf("error"))

	r, err := c.ReadTuples(context.TODO(), user, relation, oType, cToken)

	if err == nil {
		t.Errorf("error expected while calling ReadTuples")
	}

	if r != nil {
		t.Errorf("result expected to be nil ")
	}
}

func TestClientWriteTuplesSuccess(t *testing.T) {
	tests := []struct {
		name  string
		input []Tuple
	}{
		{
			name: "one tuple",
			input: []Tuple{
				*NewTuple("user:me", "assignee", "role:administrator"),
			},
		},
		{
			name: "multiple tuples via variadic syntax",
			input: []Tuple{
				*NewTuple("user:me", "assignee", "role:administrator"),
				*NewTuple("user:you", "assignee", "role:administrator"),
				*NewTuple("role:administrator#assignee", "can_view", "client:xyz"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
			mockRequest := NewMockSdkClientWriteRequestInterface(ctrl)

			c := Client{
				c:       mockOpenFGAClient,
				tracer:  mockTracer,
				monitor: mockMonitor,
				logger:  mockLogger,
			}

			ts := make([]openfga.TupleKey, 0)

			for _, tuple := range test.input {
				ts = append(ts, *openfga.NewTupleKey(tuple.Values()))
			}

			body := client.ClientWriteRequest{
				Writes: ts,
			}

			mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.WriteTuples").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGAClient.EXPECT().Write(gomock.Any()).Return(mockRequest)
			mockRequest.EXPECT().Body(body).Return(mockRequest)
			mockOpenFGAClient.EXPECT().WriteExecute(mockRequest).Times(1).Return(nil, nil)

			if err := c.WriteTuples(context.TODO(), test.input...); err != nil {
				t.Errorf("error while calling WriteTuples %s", err)
			}

		})
	}
}

func TestClientWriteTuplesFails(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
	mockRequest := NewMockSdkClientWriteRequestInterface(ctrl)

	c := Client{
		c:       mockOpenFGAClient,
		tracer:  mockTracer,
		monitor: mockMonitor,
		logger:  mockLogger,
	}

	tuple := NewTuple("user:me", "assignee", "role:administrator")

	body := client.ClientWriteRequest{
		Writes: []openfga.TupleKey{*openfga.NewTupleKey(tuple.Values())},
	}

	mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.WriteTuples").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockOpenFGAClient.EXPECT().Write(gomock.Any()).Return(mockRequest)
	mockRequest.EXPECT().Body(body).Return(mockRequest)
	mockOpenFGAClient.EXPECT().WriteExecute(mockRequest).Times(1).Return(nil, fmt.Errorf("error"))

	if err := c.WriteTuples(context.TODO(), *tuple); err == nil {
		t.Errorf("expected error while calling WriteTuples")
	}
}

func TestClientDeleteTuplesSuccess(t *testing.T) {
	tests := []struct {
		name  string
		input []Tuple
	}{
		{
			name: "one tuple",
			input: []Tuple{
				*NewTuple("user:me", "assignee", "role:administrator"),
			},
		},
		{
			name: "multiple tuples via variadic syntax",
			input: []Tuple{
				*NewTuple("user:me", "assignee", "role:administrator"),
				*NewTuple("user:you", "assignee", "role:administrator"),
				*NewTuple("role:administrator#assignee", "can_view", "client:xyz"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
			mockRequest := NewMockSdkClientWriteRequestInterface(ctrl)

			c := Client{
				c:       mockOpenFGAClient,
				tracer:  mockTracer,
				monitor: mockMonitor,
				logger:  mockLogger,
			}

			ts := make([]openfga.TupleKeyWithoutCondition, 0)

			for _, tuple := range test.input {
				ts = append(ts, *openfga.NewTupleKeyWithoutCondition(tuple.Values()))
			}

			body := client.ClientWriteRequest{
				Deletes: ts,
			}

			mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.DeleteTuples").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGAClient.EXPECT().Write(gomock.Any()).Return(mockRequest)
			mockRequest.EXPECT().Body(body).Return(mockRequest)
			mockOpenFGAClient.EXPECT().WriteExecute(mockRequest).Times(1).Return(nil, nil)

			if err := c.DeleteTuples(context.TODO(), test.input...); err != nil {
				t.Errorf("error while calling DeleteTuples %s", err)
			}

		})
	}
}

func TestClientDeleteTuplesFails(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
	mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
	mockRequest := NewMockSdkClientWriteRequestInterface(ctrl)

	c := Client{
		c:       mockOpenFGAClient,
		tracer:  mockTracer,
		monitor: mockMonitor,
		logger:  mockLogger,
	}

	tuple := NewTuple("user:me", "assignee", "role:administrator")

	body := client.ClientWriteRequest{
		Deletes: []openfga.TupleKeyWithoutCondition{*openfga.NewTupleKeyWithoutCondition(tuple.Values())},
	}

	mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.DeleteTuples").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	mockOpenFGAClient.EXPECT().Write(gomock.Any()).Return(mockRequest)
	mockRequest.EXPECT().Body(body).Return(mockRequest)
	mockOpenFGAClient.EXPECT().WriteExecute(mockRequest).Times(1).Return(nil, fmt.Errorf("error"))

	if err := c.DeleteTuples(context.TODO(), *tuple); err == nil {
		t.Errorf("expected error while calling DeleteTuples")
	}
}

func TestClientWriteBatchCheckSuccess(t *testing.T) {

	allowedResponse := openfga.CheckResponse{}
	allowedResponse.SetAllowed(true)
	unallowedResponse := openfga.CheckResponse{}
	unallowedResponse.SetAllowed(false)

	tests := []struct {
		name     string
		input    []Tuple
		expected []client.ClientCheckResponse
		output   bool
	}{
		{
			name: "one tuple",
			input: []Tuple{
				*NewTuple("user:me", "assignee", "role:administrator"),
			},
			expected: []client.ClientCheckResponse{
				{CheckResponse: allowedResponse},
			},
			output: true,
		},
		{
			name: "multiple tuples all allowed",
			input: []Tuple{
				*NewTuple("user:me", "can_edit", "role:administrator"),
				*NewTuple("user:me", "can_view", "group:editor"),
			},
			expected: []client.ClientCheckResponse{
				{CheckResponse: allowedResponse},
				{CheckResponse: allowedResponse},
			},
			output: true,
		},
		{
			name: "multiple tuples with failing condition",
			input: []Tuple{
				*NewTuple("user:me", "can_edit", "role:administrator"),
				*NewTuple("user:me", "can_view", "group:editor"),
			},
			expected: []client.ClientCheckResponse{
				{CheckResponse: allowedResponse},
				{CheckResponse: unallowedResponse},
			},
			output: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockOpenFGAClient := NewMockOpenFGAClientInterface(ctrl)
			mockRequest := NewMockSdkClientBatchCheckRequestInterface(ctrl)

			c := Client{
				c:       mockOpenFGAClient,
				tracer:  mockTracer,
				monitor: mockMonitor,
				logger:  mockLogger,
			}
			modelID := "testModel12345"

			body := client.ClientBatchCheckBody{}

			for _, tuple := range test.input {
				body = append(
					body,
					client.ClientCheckRequest{
						User:     tuple.User,
						Relation: tuple.Relation,
						Object:   tuple.Object,
					},
				)
			}

			mockTracer.EXPECT().Start(gomock.Any(), "openfga.Client.BatchCheck").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockOpenFGAClient.EXPECT().GetAuthorizationModelId().Return(modelID, nil)
			mockOpenFGAClient.EXPECT().BatchCheck(gomock.Any()).Return(mockRequest)
			mockRequest.EXPECT().Options(client.ClientBatchCheckOptions{AuthorizationModelId: &modelID}).Return(mockRequest)
			mockRequest.EXPECT().Body(body).Return(mockRequest)
			mockOpenFGAClient.EXPECT().BatchCheckExecute(mockRequest).Times(1).DoAndReturn(
				func(client.SdkClientBatchCheckRequestInterface) (*client.ClientBatchCheckResponse, error) {
					res := client.ClientBatchCheckResponse{}

					for _, check := range test.expected {
						res = append(
							res,
							client.ClientBatchCheckSingleResponse{
								ClientCheckResponse: client.ClientCheckResponse{
									CheckResponse: openfga.CheckResponse{
										Allowed: check.Allowed,
									},
								},
								Error: nil,
							},
						)
					}

					return &res, nil
				},
			)

			r, err := c.BatchCheck(context.TODO(), test.input...)

			if r != test.output {
				t.Errorf("unexpected output while calling BatchCheck %v", r)
			}

			if test.output && err != nil {
				t.Errorf("error while calling BatchCheck %s", err)
			}
		})
	}
}
