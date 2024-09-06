// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package identities

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	reflect "reflect"
	"testing"

	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	"github.com/canonical/rebac-admin-ui-handlers/v1/interfaces"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"
	"github.com/google/uuid"
	kClient "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/otel/trace"
	gomock "go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"

	"github.com/canonical/identity-platform-admin-ui/internal/mail"
	ofga "github.com/canonical/identity-platform-admin-ui/internal/openfga"
)

//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_corev1.go k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,ConfigMapInterface
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_kratos.go github.com/ory/kratos-client-go IdentityAPI

func TestListIdentitiesSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPIListIdentitiesRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identities := make([]kClient.Identity, 0)

	for i := 0; i < 10; i++ {
		identities = append(identities, *kClient.NewIdentity(fmt.Sprintf("test-%v", i), "test.json", "https://test.com/test.json", map[string]string{"name": "name"}))
	}

	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().ListIdentities(ctx).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().ListIdentitiesExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.IdentityAPIListIdentitiesRequest) ([]kClient.Identity, *http.Response, error) {

			// use reflect as attributes are private, also are pointers so need to cast it multiple times
			if pageToken := (*string)(reflect.ValueOf(r).FieldByName("pageToken").UnsafePointer()); *pageToken != "eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ" {
				t.Fatalf("expected pageToken as eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ, got %v", *pageToken)
			}

			if pageSize := (*int64)(reflect.ValueOf(r).FieldByName("pageSize").UnsafePointer()); *pageSize != 10 {
				t.Fatalf("expected page size as 10, got %v", *pageSize)
			}

			if credID := (*string)(reflect.ValueOf(r).FieldByName("credentialsIdentifier").UnsafePointer()); credID != nil {
				t.Fatalf("expected credential id to be empty, got %v", *credID)
			}

			rr := new(http.Response)
			rr.Header = make(http.Header)
			rr.Header.Set("Link", `<http://kratos-admin.default.svc.cluster.local/identities?page_size=250&page_token=eyJvZmZzZXQiOiIwIiwidiI6Mn0&per_page=250>; rel="first",<http://kratos-admin.default.svc.cluster.local/identities?page_size=250&page_token=eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ&per_page=250>; rel="next",<http://kratos-admin.default.svc.cluster.local/identities?page_size=250&page_token=eyJvZmZzZXQiOiItMjUwIiwidiI6Mn0&per_page=250>; rel="prev`)

			return identities, rr, nil
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).ListIdentities(ctx, 10, "eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ", "")

	if !reflect.DeepEqual(ids.Identities, identities) {
		t.Fatalf("expected identities to be %v not  %v", identities, ids.Identities)
	}

	if !reflect.DeepEqual(
		[]string{ids.Tokens.Next, ids.Tokens.Prev},
		[]string{"eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ", "eyJvZmZzZXQiOiItMjUwIiwidiI6Mn0"},
	) {
		t.Fatalf("expected tokens to be set, not %v", ids.Tokens)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestListIdentitiesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPIListIdentitiesRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identities := make([]kClient.Identity, 0)

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().ListIdentities(ctx).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().ListIdentitiesExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.IdentityAPIListIdentitiesRequest) ([]kClient.Identity, *http.Response, error) {

			// use reflect as attributes are private, also are pointers so need to cast it multiple times
			if pageToken := (*string)(reflect.ValueOf(r).FieldByName("pageToken").UnsafePointer()); *pageToken != "eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ" {
				t.Fatalf("expected pageToken as eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ, got %v", *pageToken)
			}

			if pageSize := (*int64)(reflect.ValueOf(r).FieldByName("pageSize").UnsafePointer()); *pageSize != 10 {
				t.Fatalf("expected page size as 10, got %v", *pageSize)
			}

			if credID := (*string)(reflect.ValueOf(r).FieldByName("credentialsIdentifier").UnsafePointer()); *credID != "test" {
				t.Fatalf("expected credential id to be test, got %v", *credID)
			}

			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")
			rr.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(rr).Encode(
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusInternalServerError,
						"debug":   "--------",
						"details": map[string]interface{}{},
						"id":      "string",
						"message": "error",
						"reason":  "error",
						"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
						"status":  "Not Found",
					},
				},
			)

			return identities, rr.Result(), fmt.Errorf("error")
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).ListIdentities(ctx, 10, "eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ", "test")

	if !reflect.DeepEqual(ids.Identities, identities) {
		t.Fatalf("expected identities to be empty not  %v", ids.Identities)
	}

	if ids.Error == nil {
		t.Fatal("expected ids.Error to be not nil")
	}

	if *ids.Error.Code != http.StatusInternalServerError {
		t.Fatalf("expected code to be %v not  %v", http.StatusInternalServerError, *ids.Error.Code)
	}

	if err == nil {
		t.Fatal("expected error to be not nil")
	}
}

func TestGetIdentitySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()
	credID := "test-1"

	identityRequest := kClient.IdentityAPIGetIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identity := kClient.NewIdentity(credID, "test.json", "https://test.com/test.json", map[string]string{"name": "name"})

	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().GetIdentity(ctx, credID).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().GetIdentityExecute(gomock.Any()).Times(1).Return(identity, new(http.Response), nil)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).GetIdentity(ctx, credID)

	if !reflect.DeepEqual(ids.Identities, []kClient.Identity{*identity}) {
		t.Fatalf("expected identities to be %v not  %v", *identity, ids.Identities)
	}
	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestGetIdentityFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()
	credID := "test"

	identityRequest := kClient.IdentityAPIGetIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().GetIdentity(ctx, credID).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().GetIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.IdentityAPIGetIdentityRequest) (*kClient.Identity, *http.Response, error) {
			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")
			rr.WriteHeader(http.StatusNotFound)

			json.NewEncoder(rr).Encode(
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusNotFound,
						"debug":   "--------",
						"details": map[string]interface{}{},
						"id":      "string",
						"message": "error",
						"reason":  "error",
						"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
						"status":  "Not Found",
					},
				},
			)

			return nil, rr.Result(), fmt.Errorf("error")
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).GetIdentity(ctx, credID)

	if !reflect.DeepEqual(ids.Identities, make([]kClient.Identity, 0)) {
		t.Fatalf("expected identities to be empty not  %v", ids.Identities)
	}

	if ids.Error == nil {
		t.Fatal("expected ids.Error to be not nil")
	}

	if *ids.Error.Code != int64(http.StatusNotFound) {
		t.Fatalf("expected code to be %v not  %v", http.StatusNotFound, *ids.Error.Code)
	}

	if err == nil {
		t.Fatal("expected error to be not nil")
	}
}

func TestCreateIdentitySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPICreateIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identity := kClient.NewIdentity("test", "test.json", "https://test.com/test.json", map[string]interface{}{"name": "name", "email": "test@example.com"})
	credentials := kClient.NewIdentityWithCredentialsWithDefaults()
	identityBody := kClient.NewCreateIdentityBody("test.json", map[string]interface{}{"name": "name", "email": "test@example.com"})
	identityBody.SetCredentials(*credentials)

	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockAuthz.EXPECT().SetCreateIdentityEntitlements(gomock.Any(), identity.Id)
	mockKratosIdentityAPI.EXPECT().CreateIdentity(ctx).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().CreateIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.IdentityAPICreateIdentityRequest) (*kClient.Identity, *http.Response, error) {

			// use reflect as attributes are private, also are pointers so need to cast it multiple times
			if IDBody := (*kClient.CreateIdentityBody)(reflect.ValueOf(r).FieldByName("createIdentityBody").UnsafePointer()); !reflect.DeepEqual(*IDBody, *identityBody) {
				t.Fatalf("expected body to be %v, got %v", identityBody, IDBody)
			}

			return identity, new(http.Response), nil
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).CreateIdentity(ctx, identityBody)

	if !reflect.DeepEqual(ids.Identities, []kClient.Identity{*identity}) {
		t.Fatalf("expected identities to be %v not  %v", *identity, ids.Identities)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestCreateIdentityFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPICreateIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	credentials := kClient.NewIdentityWithCredentialsWithDefaults()
	identityBody := kClient.NewCreateIdentityBody("test.json", map[string]interface{}{"name": "name"})
	identityBody.SetCredentials(*credentials)

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().CreateIdentity(ctx).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().CreateIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.IdentityAPICreateIdentityRequest) (*kClient.Identity, *http.Response, error) {
			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")
			rr.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(rr).Encode(
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusInternalServerError,
						"debug":   "--------",
						"details": map[string]interface{}{},
						"id":      "string",
						"message": "error",
						"reason":  "error",
						"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
						"status":  "Internal Server Error",
					},
				},
			)

			return nil, rr.Result(), fmt.Errorf("error")
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).CreateIdentity(ctx, identityBody)

	if !reflect.DeepEqual(ids.Identities, make([]kClient.Identity, 0)) {
		t.Fatalf("expected identities to be empty not  %v", ids.Identities)
	}

	if ids.Error == nil {
		t.Fatal("expected ids.Error to be not nil")
	}

	if *ids.Error.Code != int64(http.StatusInternalServerError) {
		t.Fatalf("expected code to be %v not  %v", http.StatusInternalServerError, *ids.Error.Code)
	}

	if err == nil {
		t.Fatal("expected error to be not nil")
	}
}

func TestUpdateIdentitySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPIUpdateIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identity := kClient.NewIdentity("test", "test.json", "https://test.com/test.json", map[string]string{"name": "name"})
	credentials := kClient.NewIdentityWithCredentialsWithDefaults()
	identityBody := kClient.NewUpdateIdentityBodyWithDefaults()
	identityBody.SetTraits(map[string]interface{}{"name": "name"})
	identityBody.SetCredentials(*credentials)

	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().UpdateIdentity(ctx, identity.Id).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().UpdateIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.IdentityAPIUpdateIdentityRequest) (*kClient.Identity, *http.Response, error) {

			// use reflect as attributes are private, also are pointers so need to cast it multiple times
			if IDBody := (*kClient.UpdateIdentityBody)(reflect.ValueOf(r).FieldByName("updateIdentityBody").UnsafePointer()); !reflect.DeepEqual(*IDBody, *identityBody) {
				t.Fatalf("expected body to be %v, got %v", identityBody, IDBody)
			}

			return identity, new(http.Response), nil
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).UpdateIdentity(ctx, identity.Id, identityBody)

	if !reflect.DeepEqual(ids.Identities, []kClient.Identity{*identity}) {
		t.Fatalf("expected identities to be %v not  %v", *identity, ids.Identities)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestUpdateIdentityFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()

	credID := "test"

	identityRequest := kClient.IdentityAPIUpdateIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	credentials := kClient.NewIdentityWithCredentialsWithDefaults()
	identityBody := kClient.NewUpdateIdentityBodyWithDefaults()
	identityBody.SetTraits(map[string]interface{}{"name": "name"})
	identityBody.SetCredentials(*credentials)

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().UpdateIdentity(ctx, credID).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().UpdateIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.IdentityAPIUpdateIdentityRequest) (*kClient.Identity, *http.Response, error) {
			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")
			rr.WriteHeader(http.StatusConflict)

			json.NewEncoder(rr).Encode(
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusConflict,
						"debug":   "--------",
						"details": map[string]interface{}{},
						"id":      "string",
						"message": "error",
						"reason":  "error",
						"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
						"status":  "Conflict",
					},
				},
			)

			return nil, rr.Result(), fmt.Errorf("error")
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).UpdateIdentity(ctx, credID, identityBody)

	if !reflect.DeepEqual(ids.Identities, make([]kClient.Identity, 0)) {
		t.Fatalf("expected identities to be empty not  %v", ids.Identities)
	}

	if ids.Error == nil {
		t.Fatal("expected ids.Error to be not nil")
	}

	if *ids.Error.Code != int64(http.StatusConflict) {
		t.Fatalf("expected code to be %v not  %v", http.StatusConflict, *ids.Error.Code)
	}

	if err == nil {
		t.Fatal("expected error to be not nil")
	}
}

func TestDeleteIdentitySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()
	credID := "test-1"

	identityRequest := kClient.IdentityAPIDeleteIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockAuthz.EXPECT().SetDeleteIdentityEntitlements(gomock.Any(), credID)
	mockKratosIdentityAPI.EXPECT().DeleteIdentity(ctx, credID).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().DeleteIdentityExecute(gomock.Any()).Times(1).Return(new(http.Response), nil)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).DeleteIdentity(ctx, credID)

	if len(ids.Identities) > 0 {
		t.Fatalf("invalid result, expected no identities, got %v", ids.Identities)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestDeleteIdentityFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
	mockEmail := mail.NewMockEmailServiceInterface(ctrl)

	ctx := context.Background()
	credID := "test-1"

	identityRequest := kClient.IdentityAPIDeleteIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().DeleteIdentity(ctx, credID).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().DeleteIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r kClient.IdentityAPIDeleteIdentityRequest) (*http.Response, error) {
			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")
			rr.WriteHeader(http.StatusNotFound)

			json.NewEncoder(rr).Encode(
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusNotFound,
						"debug":   "--------",
						"details": map[string]interface{}{},
						"id":      "string",
						"message": "error",
						"reason":  "error",
						"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
						"status":  "Not Found",
					},
				},
			)

			return rr.Result(), fmt.Errorf("error")
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger).DeleteIdentity(ctx, credID)

	if !reflect.DeepEqual(ids.Identities, make([]kClient.Identity, 0)) {
		t.Fatalf("expected identities to be empty not  %v", ids.Identities)
	}

	if ids.Error == nil {
		t.Fatal("expected ids.Error to be not nil")
	}

	if *ids.Error.Code != int64(http.StatusNotFound) {
		t.Fatalf("expected code to be %v not  %v", http.StatusNotFound, *ids.Error.Code)
	}

	if err == nil {
		t.Fatal("expected error to be not nil")
	}
}

func TestV1ServiceImplementsRebacServiceInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var svc interface{} = new(V1Service)

	if _, ok := svc.(interfaces.IdentitiesService); !ok {
		t.Fatalf("V1Service doesnt implement interfaces.IdentitiesService")
	}
}

func TestV1ServiceListIdentities(t *testing.T) {
	type input struct {
		size  int
		token string
	}

	type expected struct {
		err        error
		identities []resources.Identity
	}

	kIdentities := make([]kClient.Identity, 0)
	identities := make([]resources.Identity, 0)

	for i := 0; i < 10; i++ {
		id := uuid.NewString()
		name := "Test User"
		surname := fmt.Sprintf("%v", i)
		email := fmt.Sprintf("test%v@gmail.com", i)
		identities = append(
			identities,
			resources.Identity{
				Id:        &id,
				Email:     email,
				FirstName: &name,
				LastName:  &surname,
			},
		)
		kIdentities = append(
			kIdentities,
			*kClient.NewIdentity(
				id,
				"test.json",
				"https://test.com/test.json",
				map[string]string{
					"name":  fmt.Sprintf("%s %s", name, surname),
					"email": email,
				},
			),
		)
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "empty result",
			expected: expected{
				identities: []resources.Identity{},
				err:        nil,
			},
		},
		{
			name: "error",
			expected: expected{
				identities: nil,
				err:        fmt.Errorf("Internal Server Error: error"),
			},
		},
		{
			name: "full result",
			input: input{
				size:  1000,
				token: "eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ",
			},
			expected: expected{
				identities: identities,
				err:        nil,
			},
		},
		{
			name: "paginated result",
			input: input{
				size: 2,
			},
			expected: expected{
				identities: identities[:2],
				err:        nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			identityRequest := kClient.IdentityAPIListIdentitiesRequest{
				ApiService: mockKratosIdentityAPI,
			}

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockKratosIdentityAPI.EXPECT().ListIdentities(ctx).Times(1).Return(identityRequest)
			mockKratosIdentityAPI.EXPECT().ListIdentitiesExecute(gomock.Any()).Times(1).DoAndReturn(
				func(r kClient.IdentityAPIListIdentitiesRequest) ([]kClient.Identity, *http.Response, error) {

					// use reflect as attributes are private, also are pointers so need to cast it multiple times
					if pageToken := (*string)(reflect.ValueOf(r).FieldByName("pageToken").UnsafePointer()); *pageToken != test.input.token {
						t.Errorf("expected pageToken as %s, got %v", test.input.token, *pageToken)
					}

					pageSize := (*int64)(reflect.ValueOf(r).FieldByName("pageSize").UnsafePointer())
					if *pageSize != int64(test.input.size) {
						t.Errorf("expected page size as %v, got %v", test.input.size, *pageSize)
					}

					if credID := (*string)(reflect.ValueOf(r).FieldByName("credentialsIdentifier").UnsafePointer()); credID != nil {
						t.Errorf("expected credential id to be empty, got %v", *credID)
					}

					if test.expected.err != nil {
						rr := httptest.NewRecorder()
						rr.Header().Set("Content-Type", "application/json")
						rr.WriteHeader(http.StatusInternalServerError)

						json.NewEncoder(rr).Encode(
							map[string]interface{}{
								"error": map[string]interface{}{
									"code":    http.StatusInternalServerError,
									"debug":   "--------",
									"details": map[string]interface{}{},
									"id":      "string",
									"message": "error",
									"reason":  "error",
									"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
									"status":  "Not Found",
								},
							},
						)

						return []kClient.Identity{}, rr.Result(), fmt.Errorf("error")
					}

					rr := new(http.Response)
					rr.Header = make(http.Header)
					rr.Header.Set("Link", `<http://kratos-admin.default.svc.cluster.local/identities?page_size=250&page_token=eyJvZmZzZXQiOiIwIiwidiI6Mn0&per_page=250>; rel="first",<http://kratos-admin.default.svc.cluster.local/identities?page_size=250&page_token=eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ&per_page=250>; rel="next",<http://kratos-admin.default.svc.cluster.local/identities?page_size=250&page_token=eyJvZmZzZXQiOiItMjUwIiwidiI6Mn0&per_page=250>; rel="prev`)

					if int64(len(kIdentities)) > *pageSize {
						return kIdentities[:*pageSize], rr, nil
					}

					return kIdentities, rr, nil

				},
			)

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			r, err := svc.ListIdentities(
				ctx,
				&resources.GetIdentitiesParams{
					Size:      &test.input.size,
					NextToken: &test.input.token,
				},
			)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			for n, i := range r.Data {
				if i.Email != test.expected.identities[n].Email {
					t.Errorf("expected identities to be %s not  %s", test.expected.identities[n].Email, i.Email)
				}

				if *i.FirstName != *test.expected.identities[n].FirstName {
					t.Errorf("expected name to be %s not %s", *test.expected.identities[n].FirstName, *i.FirstName)
				}

				if *i.LastName != *test.expected.identities[n].LastName {
					t.Errorf("expected surname to be %s not %s", *test.expected.identities[n].LastName, *i.LastName)
				}
			}

			if len(r.Data) > 0 && test.input.size > 0 && *r.Next.PageToken != "eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ" {
				t.Errorf("expected token to be eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ, not %s", *r.Next.PageToken)
			}

		})
	}
}

func TestV1ServiceCreateIdentity(t *testing.T) {
	type input struct {
		identity *resources.Identity
	}

	type expected struct {
		err      error
		identity *resources.Identity
	}

	id := uuid.NewString()
	name := "Test"
	surname := "User"
	email := "test@gmail.com"
	kIdentity := kClient.NewIdentity(
		id,
		"test",
		"https://test.com/test.json",
		map[string]interface{}{
			"name":  fmt.Sprintf("%s %s", name, surname),
			"email": email,
		},
	)
	identity := resources.Identity{
		Email:     email,
		FirstName: &name,
		LastName:  &surname,
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "error",
			input: input{
				identity: &identity,
			},
			expected: expected{

				err: v1.NewRequestBodyValidationError("bad identity payload"),
			},
		},
		{
			name: "success",
			input: input{
				identity: &identity,
			},
			expected: expected{
				identity: &identity,
				err:      nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			ctx := context.Background()

			identityRequest := kClient.IdentityAPICreateIdentityRequest{
				ApiService: mockKratosIdentityAPI,
			}

			identityBody := kClient.NewCreateIdentityBody(
				kIdentity.SchemaId,
				map[string]interface{}{
					"name":  fmt.Sprintf("%s %s", name, surname),
					"email": email,
				},
			)
			identityBody.SetState("StateActive")

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockAuthz.EXPECT().SetCreateIdentityEntitlements(gomock.Any(), id).MinTimes(0).MaxTimes(1)
			mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).MinTimes(0).MaxTimes(1).Return(mockConfigMapV1)
			mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).MinTimes(0).MaxTimes(1).Return(cm, nil)

			mockKratosIdentityAPI.EXPECT().CreateIdentity(gomock.Any()).Times(1).Return(identityRequest)
			mockKratosIdentityAPI.EXPECT().CreateIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
				func(r kClient.IdentityAPICreateIdentityRequest) (*kClient.Identity, *http.Response, error) {

					// use reflect as attributes are private, also are pointers so need to cast it multiple times
					if IDBody := (*kClient.CreateIdentityBody)(reflect.ValueOf(r).FieldByName("createIdentityBody").UnsafePointer()); !reflect.DeepEqual(*IDBody, *identityBody) {
						t.Errorf("expected body to be %v, got %v", identityBody, IDBody)
					}

					if test.expected.err != nil {
						rr := httptest.NewRecorder()
						rr.Header().Set("Content-Type", "application/json")
						rr.WriteHeader(http.StatusInternalServerError)

						json.NewEncoder(rr).Encode(
							map[string]interface{}{
								"error": map[string]interface{}{
									"code":    http.StatusInternalServerError,
									"debug":   "--------",
									"details": map[string]interface{}{},
									"id":      "string",
									"message": "error",
									"reason":  "error",
									"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
									"status":  "Internal Server Error",
								},
							},
						)

						return nil, rr.Result(), fmt.Errorf("error")
					}

					return kIdentity, new(http.Response), nil
				},
			)

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			newIdentity, err := svc.CreateIdentity(ctx, test.input.identity)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not  %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if newIdentity.Id != nil && *newIdentity.Id != id {
				t.Errorf("expected ID to be %s, not %s", id, *newIdentity.Id)
			}

			if newIdentity.Email != identity.Email {
				t.Errorf("expected email to be %s, not %s", identity.Email, newIdentity.Email)
			}

			if newIdentity.FirstName != nil && *newIdentity.FirstName != *identity.FirstName {
				t.Errorf("expected name to be %s, not %s", *identity.FirstName, *newIdentity.FirstName)
			}

			if newIdentity.LastName != nil && *newIdentity.LastName != *identity.LastName {
				t.Errorf("expected surname to be %s, not %s", *identity.LastName, *newIdentity.LastName)
			}

		})
	}
}

func TestV1ServiceGetIdentity(t *testing.T) {
	type expected struct {
		err      error
		identity *resources.Identity
	}

	id := uuid.NewString()
	name := "Test"
	surname := "User"
	email := "test@gmail.com"
	kIdentity := kClient.NewIdentity(
		id,
		"test",
		"https://test.com/test.json",
		map[string]string{
			"name":  fmt.Sprintf("%s %s", name, surname),
			"email": email,
		},
	)

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "error",
			input: uuid.NewString(),
			expected: expected{
				err: fmt.Errorf("error"),
			},
		},
		{
			name:  "success",
			input: id,
			expected: expected{
				identity: &resources.Identity{
					Id:        &id,
					Email:     email,
					FirstName: &name,
					LastName:  &surname,
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			identityRequest := kClient.IdentityAPIGetIdentityRequest{
				ApiService: mockKratosIdentityAPI,
			}

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockKratosIdentityAPI.EXPECT().GetIdentity(ctx, test.input).Times(1).Return(identityRequest)
			mockKratosIdentityAPI.EXPECT().GetIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
				func(r kClient.IdentityAPIGetIdentityRequest) (*kClient.Identity, *http.Response, error) {
					if test.expected.err != nil {
						rr := httptest.NewRecorder()
						rr.Header().Set("Content-Type", "application/json")
						rr.WriteHeader(http.StatusNotFound)

						json.NewEncoder(rr).Encode(
							map[string]interface{}{
								"error": map[string]interface{}{
									"code":    http.StatusNotFound,
									"debug":   "--------",
									"details": map[string]interface{}{},
									"id":      "string",
									"message": "error",
									"reason":  "error",
									"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
									"status":  "Not Found",
								},
							},
						)

						return nil, rr.Result(), fmt.Errorf("error")
					}

					return kIdentity, new(http.Response), nil
				},
			)

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			identity, err := svc.GetIdentity(ctx, test.input)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not  %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if identity.Id != nil && *identity.Id != id {
				t.Errorf("expected ID to be %s, not %s", id, *identity.Id)
			}

			if identity.Email != test.expected.identity.Email {
				t.Errorf("expected email to be %s, not %s", test.expected.identity.Email, identity.Email)
			}

			if identity.FirstName != nil && *identity.FirstName != *test.expected.identity.FirstName {
				t.Errorf("expected name to be %s, not %s", *test.expected.identity.FirstName, *identity.FirstName)
			}

			if identity.LastName != nil && *identity.LastName != *test.expected.identity.LastName {
				t.Errorf("expected surname to be %s, not %s", *test.expected.identity.LastName, *identity.LastName)
			}
		},
		)
	}
}

func TestV1ServiceUpdateIdentity(t *testing.T) {
	type expected struct {
		err      error
		identity *resources.Identity
	}

	id := uuid.NewString()
	name := "Test"
	surname := "User"
	email := "test@gmail.com"
	kIdentity := kClient.NewIdentity(
		id,
		"test",
		"https://test.com/test.json",
		map[string]string{
			"name":  fmt.Sprintf("%s %s", name, surname),
			"email": email,
		},
	)

	tests := []struct {
		name     string
		input    *resources.Identity
		expected expected
	}{
		{
			name: "error",
			input: &resources.Identity{
				Id:        &id,
				Email:     email,
				FirstName: &name,
				LastName:  &surname,
			},
			expected: expected{
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "success",
			input: &resources.Identity{
				Id:        &id,
				Email:     email,
				FirstName: &name,
				LastName:  &surname,
			},
			expected: expected{
				identity: &resources.Identity{
					Id:        &id,
					Email:     email,
					FirstName: &name,
					LastName:  &surname,
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			identityRequest := kClient.IdentityAPIUpdateIdentityRequest{
				ApiService: mockKratosIdentityAPI,
			}

			identityBody := kClient.NewUpdateIdentityBodyWithDefaults()
			// identityBody.SetSchemaId(kIdentity.SchemaId)
			identityBody.SetTraits(map[string]interface{}{
				"name":  fmt.Sprintf("%s %s", name, surname),
				"email": email,
			},
			)

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockKratosIdentityAPI.EXPECT().UpdateIdentity(gomock.Any(), *test.input.Id).Times(1).Return(identityRequest)
			mockKratosIdentityAPI.EXPECT().UpdateIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
				func(r kClient.IdentityAPIUpdateIdentityRequest) (*kClient.Identity, *http.Response, error) {

					// use reflect as attributes are private, also are pointers so need to cast it multiple times
					if IDBody := (*kClient.UpdateIdentityBody)(reflect.ValueOf(r).FieldByName("updateIdentityBody").UnsafePointer()); !reflect.DeepEqual(*IDBody, *identityBody) {
						t.Errorf("expected body to be %v, got %v", identityBody, IDBody)
					}

					if test.expected.err != nil {
						rr := httptest.NewRecorder()
						rr.Header().Set("Content-Type", "application/json")
						rr.WriteHeader(http.StatusNotFound)

						json.NewEncoder(rr).Encode(
							map[string]interface{}{
								"error": map[string]interface{}{
									"code":    http.StatusConflict,
									"debug":   "--------",
									"details": map[string]interface{}{},
									"id":      "string",
									"message": "error",
									"reason":  "error",
									"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
									"status":  "Conflict",
								},
							},
						)

						return nil, rr.Result(), fmt.Errorf("error")
					}

					return kIdentity, new(http.Response), nil
				},
			)

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			identity, err := svc.UpdateIdentity(ctx, test.input)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not  %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if identity.Id != nil && *identity.Id != id {
				t.Errorf("expected ID to be %s, not %s", id, *identity.Id)
			}

			if identity.Email != test.expected.identity.Email {
				t.Errorf("expected email to be %s, not %s", test.expected.identity.Email, identity.Email)
			}

			if identity.FirstName != nil && *identity.FirstName != *test.expected.identity.FirstName {
				t.Errorf("expected name to be %s, not %s", *test.expected.identity.FirstName, *identity.FirstName)
			}

			if identity.LastName != nil && *identity.LastName != *test.expected.identity.LastName {
				t.Errorf("expected surname to be %s, not %s", *test.expected.identity.LastName, *identity.LastName)
			}
		},
		)
	}
}

func TestV1ServiceDeleteIdentity(t *testing.T) {
	type expected struct {
		err error
		ok  bool
	}

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "error",
			input: uuid.NewString(),
			expected: expected{
				err: fmt.Errorf("error"),
				ok:  false,
			},
		},
		{
			name:  "success",
			input: uuid.NewString(),
			expected: expected{
				ok:  true,
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			identityRequest := kClient.IdentityAPIDeleteIdentityRequest{
				ApiService: mockKratosIdentityAPI,
			}

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockAuthz.EXPECT().SetDeleteIdentityEntitlements(gomock.Any(), test.input).MinTimes(0).MaxTimes(1)
			mockKratosIdentityAPI.EXPECT().DeleteIdentity(ctx, test.input).Times(1).Return(identityRequest)
			mockKratosIdentityAPI.EXPECT().DeleteIdentityExecute(gomock.Any()).Times(1).DoAndReturn(
				func(r kClient.IdentityAPIDeleteIdentityRequest) (*http.Response, error) {
					if test.expected.err != nil {
						rr := httptest.NewRecorder()
						rr.Header().Set("Content-Type", "application/json")
						rr.WriteHeader(http.StatusNotFound)

						json.NewEncoder(rr).Encode(
							map[string]interface{}{
								"error": map[string]interface{}{
									"code":    http.StatusNotFound,
									"debug":   "--------",
									"details": map[string]interface{}{},
									"id":      "string",
									"message": "error",
									"reason":  "error",
									"request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
									"status":  "Not Found",
								},
							},
						)

						return rr.Result(), fmt.Errorf("error")
					}

					return new(http.Response), nil
				},
			)

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			ok, err := svc.DeleteIdentity(ctx, test.input)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not  %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if ok != test.expected.ok {
				t.Errorf("expected result to be %v, not %v", test.expected.ok, ok)
			}
		},
		)
	}
}

func TestV1ServiceGetIdentityGroups(t *testing.T) {
	type expected struct {
		groups []resources.Group
		err    error
	}

	cLevel := "c-level"
	itAdmin := "it-admin"
	devops := "devops"

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "empty result",
			input: uuid.NewString(),
			expected: expected{
				groups: []resources.Group{},
				err:    nil,
			},
		},
		{
			name:  "error",
			input: uuid.NewString(),
			expected: expected{
				err: fmt.Errorf("error"),
			},
		},
		{
			name:  "full result",
			input: uuid.NewString(),
			expected: expected{
				groups: []resources.Group{
					{Id: &cLevel, Name: cLevel},
					{Id: &itAdmin, Name: itAdmin},
					{Id: &devops, Name: devops},
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockOpenFGAStore.EXPECT().ListAssignedGroups(gomock.Any(), fmt.Sprintf("user:%s", test.input)).DoAndReturn(
				func(ctx context.Context, ID string) ([]string, error) {
					if test.expected.err != nil {
						return nil, fmt.Errorf("error")
					}

					groups := make([]string, 0)

					for _, g := range test.expected.groups {
						groups = append(groups, g.Name)
					}

					return groups, nil
				},
			)

			r, err := svc.GetIdentityGroups(context.Background(), test.input, nil)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			for i, group := range r.Data {
				if group.Name != test.expected.groups[i].Name {
					t.Errorf("invalid result, expected: %v, got: %v", test.expected.groups[i].Name, group.Name)
				}
			}

		})
	}
}

func TestV1ServiceGetIdentityRoles(t *testing.T) {
	type expected struct {
		roles []resources.Role
		err   error
	}

	cLevel := "c-level"
	itAdmin := "it-admin"
	devops := "devops"

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "empty result",
			input: uuid.NewString(),
			expected: expected{
				roles: []resources.Role{},
				err:   nil,
			},
		},
		{
			name:  "error",
			input: uuid.NewString(),
			expected: expected{
				err: fmt.Errorf("error"),
			},
		},
		{
			name:  "full result",
			input: uuid.NewString(),
			expected: expected{
				roles: []resources.Role{
					{Id: &cLevel, Name: cLevel},
					{Id: &itAdmin, Name: itAdmin},
					{Id: &devops, Name: devops},
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockOpenFGAStore.EXPECT().ListAssignedRoles(gomock.Any(), fmt.Sprintf("user:%s", test.input)).DoAndReturn(
				func(ctx context.Context, ID string) ([]string, error) {
					if test.expected.err != nil {
						return nil, fmt.Errorf("error")
					}

					roles := make([]string, 0)

					for _, r := range test.expected.roles {
						roles = append(roles, r.Name)
					}

					return roles, nil
				},
			)

			r, err := svc.GetIdentityRoles(context.Background(), test.input, nil)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			for i, role := range r.Data {
				if role.Name != test.expected.roles[i].Name {
					t.Errorf("invalid result, expected: %v, got: %v", test.expected.roles[i].Name, role.Name)
				}
			}
		})
	}
}

func TestV1ServicePatchIdentityRoles(t *testing.T) {
	type input struct {
		patches []resources.IdentityRolesPatchItem
		id      string
	}
	type expected struct {
		ok  bool
		err error
	}

	additions := []resources.IdentityRolesPatchItem{
		{Op: "add", Role: "test1"},
		{Op: "add", Role: "test2"},
	}
	removals := []resources.IdentityRolesPatchItem{
		{Op: "remove", Role: "test1"},
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "empty payload",
			input: input{
				id:      uuid.NewString(),
				patches: []resources.IdentityRolesPatchItem{},
			},
			expected: expected{
				ok:  true,
				err: nil,
			},
		},
		{
			name: "error assign",
			input: input{
				id:      uuid.NewString(),
				patches: additions,
			},
			expected: expected{
				err: fmt.Errorf("error"),
				ok:  false,
			},
		},
		{
			name: "error unassign",
			input: input{
				id:      uuid.NewString(),
				patches: removals,
			},
			expected: expected{
				err: fmt.Errorf("error"),
				ok:  false,
			},
		},
		{
			name: "success",
			input: input{
				id:      uuid.NewString(),
				patches: append(removals, additions...),
			},
			expected: expected{
				ok:  true,
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			// AssignRoles(context.Context, string, ...string) error
			// UnassignRoles(context.Context, string, ...string) error
			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockOpenFGAStore.EXPECT().AssignRoles(gomock.Any(), fmt.Sprintf("user:%s", test.input.id), gomock.Any()).MinTimes(0).MaxTimes(1).DoAndReturn(
				func(ctx context.Context, ID string, roles ...string) error {
					if ID != fmt.Sprintf("user:%s", test.input.id) {
						t.Errorf("expected ID to be user:%s got %s", test.input.id, ID)
					}

					if test.expected.err != nil {
						return fmt.Errorf("error")
					}

					rs := make([]string, 0)

					for _, r := range test.input.patches {
						if r.Op == "add" {
							rs = append(rs, fmt.Sprintf("role:%s", r.Role))
						}
					}

					if !reflect.DeepEqual(rs, roles) {
						t.Errorf("expected roles to be %v got %v", rs, roles)
					}

					return nil
				},
			)

			mockOpenFGAStore.EXPECT().UnassignRoles(gomock.Any(), fmt.Sprintf("user:%s", test.input.id), gomock.Any()).MinTimes(0).MaxTimes(1).DoAndReturn(
				func(ctx context.Context, ID string, roles ...string) error {
					if ID != fmt.Sprintf("user:%s", test.input.id) {
						t.Errorf("expected ID to be user:%s got %s", test.input.id, ID)
					}

					if test.expected.err != nil {
						return fmt.Errorf("error")
					}

					rs := make([]string, 0)

					for _, r := range test.input.patches {
						if r.Op == "remove" {
							rs = append(rs, fmt.Sprintf("role:%s", r.Role))
						}
					}

					if !reflect.DeepEqual(rs, roles) {
						t.Errorf("expected roles to be %v got %v", rs, roles)
					}

					return nil
				},
			)

			ok, err := svc.PatchIdentityRoles(context.Background(), test.input.id, test.input.patches)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if ok != test.expected.ok {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.ok, ok)
			}
		})
	}
}

func TestV1ServicePatchIdentityGroups(t *testing.T) {
	type input struct {
		patches []resources.IdentityGroupsPatchItem
		id      string
	}
	type expected struct {
		ok  bool
		err error
	}

	additions := []resources.IdentityGroupsPatchItem{
		{Op: "add", Group: "test1"},
		{Op: "add", Group: "test2"},
	}
	removals := []resources.IdentityGroupsPatchItem{
		{Op: "remove", Group: "test1"},
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "empty payload",
			input: input{
				id:      uuid.NewString(),
				patches: []resources.IdentityGroupsPatchItem{},
			},
			expected: expected{
				ok:  true,
				err: nil,
			},
		},
		{
			name: "error assign",
			input: input{
				id:      uuid.NewString(),
				patches: additions,
			},
			expected: expected{
				err: fmt.Errorf("error"),
				ok:  false,
			},
		},
		{
			name: "error unassign",
			input: input{
				id:      uuid.NewString(),
				patches: removals,
			},
			expected: expected{
				err: fmt.Errorf("error"),
				ok:  false,
			},
		},
		{
			name: "success",
			input: input{
				id:      uuid.NewString(),
				patches: append(removals, additions...),
			},
			expected: expected{
				ok:  true,
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			// AssignGroups(context.Context, string, ...string) error
			// UnassignGroups(context.Context, string, ...string) error
			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockOpenFGAStore.EXPECT().AssignGroups(gomock.Any(), fmt.Sprintf("user:%s", test.input.id), gomock.Any()).MinTimes(0).MaxTimes(1).DoAndReturn(
				func(ctx context.Context, ID string, groups ...string) error {
					if ID != fmt.Sprintf("user:%s", test.input.id) {
						t.Errorf("expected ID to be user:%s got %s", test.input.id, ID)
					}

					if test.expected.err != nil {
						return fmt.Errorf("error")
					}

					gs := make([]string, 0)

					for _, g := range test.input.patches {
						if g.Op == "add" {
							gs = append(gs, fmt.Sprintf("group:%s", g.Group))
						}
					}

					if !reflect.DeepEqual(gs, groups) {
						t.Errorf("expected groups to be %v got %v", gs, groups)
					}

					return nil
				},
			)

			mockOpenFGAStore.EXPECT().UnassignGroups(gomock.Any(), fmt.Sprintf("user:%s", test.input.id), gomock.Any()).MinTimes(0).MaxTimes(1).DoAndReturn(
				func(ctx context.Context, ID string, groups ...string) error {
					if ID != fmt.Sprintf("user:%s", test.input.id) {
						t.Errorf("expected ID to be user:%s got %s", test.input.id, ID)
					}

					if test.expected.err != nil {
						return fmt.Errorf("error")
					}

					gs := make([]string, 0)

					for _, g := range test.input.patches {
						if g.Op == "remove" {
							gs = append(gs, fmt.Sprintf("group:%s", g.Group))
						}
					}

					if !reflect.DeepEqual(gs, groups) {
						t.Errorf("expected groups to be %v got %v", gs, groups)
					}

					return nil
				},
			)

			ok, err := svc.PatchIdentityGroups(context.Background(), test.input.id, test.input.patches)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if ok != test.expected.ok {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.ok, ok)
			}
		})
	}
}

func TestV1ServiceGetIdentityEntitlements(t *testing.T) {
	type input struct {
		params *resources.GetIdentitiesItemEntitlementsParams
		id     string
	}
	type expected struct {
		permissions []resources.EntityEntitlement
		err         error
	}

	permissions := []resources.EntityEntitlement{
		{
			Entitlement: "can_view",
			EntityId:    "okta",
			EntityType:  "client",
		},
		{
			Entitlement: "can_delete",
			EntityId:    "github",
			EntityType:  "client",
		},
		{
			Entitlement: "can_create",
			EntityId:    "github",
			EntityType:  "client",
		},
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "empty payload",
			input: input{
				id: uuid.NewString(),
			},
			expected: expected{
				permissions: []resources.EntityEntitlement{},
				err:         nil,
			},
		},
		{
			name: "error",
			input: input{
				id: uuid.NewString(),
			},
			expected: expected{
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "success",
			input: input{
				id: uuid.NewString(),
			},
			expected: expected{
				permissions: permissions,
				err:         nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockOpenFGAStore.EXPECT().ListPermissions(gomock.Any(), fmt.Sprintf("user:%s", test.input.id), gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, ID string, tokens map[string]string) ([]ofga.Permission, map[string]string, error) {
					if ID != fmt.Sprintf("user:%s", test.input.id) {
						t.Errorf("expected ID to be user:%s got %s", test.input.id, ID)
					}

					if test.expected.err != nil {
						return nil, nil, fmt.Errorf("error")
					}

					ps := make([]ofga.Permission, 0)

					for _, p := range test.expected.permissions {
						ps = append(
							ps,
							ofga.Permission{
								Relation: p.Entitlement,
								Object:   fmt.Sprintf("%s:%s", p.EntityType, p.EntityId),
							},
						)
					}
					return ps, map[string]string{}, nil
				},
			)

			r, err := svc.GetIdentityEntitlements(context.Background(), test.input.id, test.input.params)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if !reflect.DeepEqual(r.Data, test.expected.permissions) {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.permissions, r.Data)
			}
		})
	}
}

func TestV1ServicePatchIdentityEntitlements(t *testing.T) {
	type input struct {
		patches []resources.IdentityEntitlementsPatchItem
		id      string
	}
	type expected struct {
		ok  bool
		err error
	}

	additions := []resources.IdentityEntitlementsPatchItem{
		{
			Op: "add",
			Entitlement: resources.EntityEntitlement{
				Entitlement: "can_view",
				EntityId:    "okta",
				EntityType:  "client",
			},
		},
		{
			Op: "add",
			Entitlement: resources.EntityEntitlement{
				Entitlement: "can_delete",
				EntityId:    "github",
				EntityType:  "client",
			},
		},
	}

	removals := []resources.IdentityEntitlementsPatchItem{
		{
			Op: "remove",
			Entitlement: resources.EntityEntitlement{
				Entitlement: "can_create",
				EntityId:    "github",
				EntityType:  "client",
			},
		},
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "empty payload",
			input: input{
				id:      uuid.NewString(),
				patches: []resources.IdentityEntitlementsPatchItem{},
			},
			expected: expected{
				ok:  true,
				err: nil,
			},
		},
		{
			name: "error assign",
			input: input{
				id:      uuid.NewString(),
				patches: additions,
			},
			expected: expected{
				err: fmt.Errorf("error"),
				ok:  false,
			},
		},
		{
			name: "error unassign",
			input: input{
				id:      uuid.NewString(),
				patches: removals,
			},
			expected: expected{
				err: fmt.Errorf("error"),
				ok:  false,
			},
		},
		{
			name: "success",
			input: input{
				id:      uuid.NewString(),
				patches: append(removals, additions...),
			},
			expected: expected{
				ok:  true,
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)
			mockOpenFGAStore := NewMockOpenFGAStoreInterface(ctrl)
			mockEmail := mail.NewMockEmailServiceInterface(ctrl)

			ctx := context.Background()

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.Name = "schemas"
			cfg.Namespace = "default"
			cfg.OpenFGAStore = mockOpenFGAStore

			cm := new(corev1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[DEFAULT_SCHEMA] = "test"

			svc := NewV1Service(
				cfg,
				NewService(mockKratosIdentityAPI, mockAuthz, mockEmail, mockTracer, mockMonitor, mockLogger),
			)

			// AssignGroups(context.Context, string, ...string) error
			// UnassignGroups(context.Context, string, ...string) error
			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockOpenFGAStore.EXPECT().AssignPermissions(gomock.Any(), fmt.Sprintf("user:%s", test.input.id), gomock.Any()).MinTimes(0).MaxTimes(1).DoAndReturn(
				func(ctx context.Context, ID string, permissions ...ofga.Permission) error {
					if ID != fmt.Sprintf("user:%s", test.input.id) {
						t.Errorf("expected ID to be user:%s got %s", test.input.id, ID)
					}

					if test.expected.err != nil {
						return fmt.Errorf("error")
					}

					ps := make([]ofga.Permission, 0)

					for _, p := range test.input.patches {
						if p.Op == "add" {
							ps = append(
								ps,
								ofga.Permission{
									Relation: p.Entitlement.Entitlement,
									Object:   fmt.Sprintf("%s:%s", p.Entitlement.EntityType, p.Entitlement.EntityId),
								},
							)
						}
					}

					if !reflect.DeepEqual(ps, permissions) {
						t.Errorf("expected permissions to be %v got %v", ps, permissions)
					}

					return nil
				},
			)

			mockOpenFGAStore.EXPECT().UnassignPermissions(gomock.Any(), fmt.Sprintf("user:%s", test.input.id), gomock.Any()).MinTimes(0).MaxTimes(1).DoAndReturn(
				func(ctx context.Context, ID string, permissions ...ofga.Permission) error {
					if ID != fmt.Sprintf("user:%s", test.input.id) {
						t.Errorf("expected ID to be user:%s got %s", test.input.id, ID)
					}

					if test.expected.err != nil {
						return fmt.Errorf("error")
					}

					ps := make([]ofga.Permission, 0)

					for _, p := range test.input.patches {
						if p.Op == "remove" {
							ps = append(
								ps,
								ofga.Permission{
									Relation: p.Entitlement.Entitlement,
									Object:   fmt.Sprintf("%s:%s", p.Entitlement.EntityType, p.Entitlement.EntityId),
								},
							)
						}
					}

					if !reflect.DeepEqual(ps, permissions) {
						t.Errorf("expected permissions to be %v got %v", ps, permissions)
					}

					return nil
				},
			)

			ok, err := svc.PatchIdentityEntitlements(context.Background(), test.input.id, test.input.patches)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v got %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if ok != test.expected.ok {
				t.Errorf("invalid result, expected: %v, got: %v", test.expected.ok, ok)
			}
		})
	}
}
