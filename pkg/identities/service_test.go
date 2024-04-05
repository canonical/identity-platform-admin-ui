// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package identities

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	reflect "reflect"
	"testing"

	kClient "github.com/ory/kratos-client-go"
	"go.opentelemetry.io/otel/trace"
	gomock "go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package identities -destination ./mock_kratos.go github.com/ory/kratos-client-go IdentityAPI

func TestListIdentitiesSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPIListIdentitiesRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identities := make([]kClient.Identity, 0)

	for i := 0; i < 10; i++ {
		identities = append(identities, *kClient.NewIdentity(fmt.Sprintf("test-%v", i), "test.json", "https://test.com/test.json", map[string]string{"name": "name"}))
	}

	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.ListIdentities").Times(1).Return(ctx, trace.SpanFromContext(ctx))
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
			rr.Header.Set("Link", `<http://kratos-admin.default.svc.cluster.local/identities?page=0&page_size=250&page_token=eyJvZmZzZXQiOiIwIiwidiI6Mn0&per_page=250>; rel="first",<http://kratos-admin.default.svc.cluster.local/identities?page=1&page_size=250&page_token=eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ&per_page=250>; rel="next",<http://kratos-admin.default.svc.cluster.local/identities?page=-1&page_size=250&page_token=eyJvZmZzZXQiOiItMjUwIiwidiI6Mn0&per_page=250>; rel="prev`)

			return identities, rr, nil
		},
	)

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).ListIdentities(ctx, 10, "eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ", "")

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPIListIdentitiesRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identities := make([]kClient.Identity, 0)

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.ListIdentities").Times(1).Return(ctx, trace.SpanFromContext(ctx))
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

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).ListIdentities(ctx, 10, "eyJvZmZzZXQiOiIyNTAiLCJ2IjoyfQ", "test")

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()
	credID := "test-1"

	identityRequest := kClient.IdentityAPIGetIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identity := kClient.NewIdentity(credID, "test.json", "https://test.com/test.json", map[string]string{"name": "name"})

	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.GetIdentity").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().GetIdentity(ctx, credID).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().GetIdentityExecute(gomock.Any()).Times(1).Return(identity, new(http.Response), nil)

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).GetIdentity(ctx, credID)

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()
	credID := "test"

	identityRequest := kClient.IdentityAPIGetIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.GetIdentity").Times(1).Return(ctx, trace.SpanFromContext(ctx))
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

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).GetIdentity(ctx, credID)

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPICreateIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identity := kClient.NewIdentity("test", "test.json", "https://test.com/test.json", map[string]string{"name": "name"})
	credentials := kClient.NewIdentityWithCredentialsWithDefaults()
	identityBody := kClient.NewCreateIdentityBody("test.json", map[string]interface{}{"name": "name"})
	identityBody.SetCredentials(*credentials)

	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.CreateIdentity").Times(1).Return(ctx, trace.SpanFromContext(ctx))
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

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).CreateIdentity(ctx, identityBody)

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPICreateIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	credentials := kClient.NewIdentityWithCredentialsWithDefaults()
	identityBody := kClient.NewCreateIdentityBody("test.json", map[string]interface{}{"name": "name"})
	identityBody.SetCredentials(*credentials)

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.CreateIdentity").Times(1).Return(ctx, trace.SpanFromContext(ctx))
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

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).CreateIdentity(ctx, identityBody)

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()

	identityRequest := kClient.IdentityAPIUpdateIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	identity := kClient.NewIdentity("test", "test.json", "https://test.com/test.json", map[string]string{"name": "name"})
	credentials := kClient.NewIdentityWithCredentialsWithDefaults()
	identityBody := kClient.NewUpdateIdentityBodyWithDefaults()
	identityBody.SetTraits(map[string]interface{}{"name": "name"})
	identityBody.SetCredentials(*credentials)

	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.UpdateIdentity").Times(1).Return(ctx, trace.SpanFromContext(ctx))
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

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).UpdateIdentity(ctx, identity.Id, identityBody)

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

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
	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.UpdateIdentity").Times(1).Return(ctx, trace.SpanFromContext(ctx))
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

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).UpdateIdentity(ctx, credID, identityBody)

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()
	credID := "test-1"

	identityRequest := kClient.IdentityAPIDeleteIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.DeleteIdentity").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockKratosIdentityAPI.EXPECT().DeleteIdentity(ctx, credID).Times(1).Return(identityRequest)
	mockKratosIdentityAPI.EXPECT().DeleteIdentityExecute(gomock.Any()).Times(1).Return(new(http.Response), nil)

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).DeleteIdentity(ctx, credID)

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
	mockKratosIdentityAPI := NewMockIdentityAPI(ctrl)

	ctx := context.Background()
	credID := "test-1"

	identityRequest := kClient.IdentityAPIDeleteIdentityRequest{
		ApiService: mockKratosIdentityAPI,
	}

	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	mockTracer.EXPECT().Start(ctx, "kratos.IdentityAPI.DeleteIdentity").Times(1).Return(ctx, trace.SpanFromContext(ctx))
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

	ids, err := NewService(mockKratosIdentityAPI, mockTracer, mockMonitor, mockLogger).DeleteIdentity(ctx, credID)

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
