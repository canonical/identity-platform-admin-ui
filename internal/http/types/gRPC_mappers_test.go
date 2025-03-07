// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package types

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	v0Types "github.com/canonical/identity-platform-api/v0/http"
	v0Roles "github.com/canonical/identity-platform-api/v0/roles"
	rpcStatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestSetHeaderFromMetadataFilter(t *testing.T) {
	tests := []struct {
		name     string
		context  context.Context
		expected map[string]string
		err      error
	}{
		{
			name: "Valid metadata",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"key1":            "value1",
				"key2":            "value2",
				"grpcgateway-foo": "should-be-filtered",
			})),
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			err: nil,
		},
		{
			name:     "No metadata in context",
			context:  context.Background(),
			expected: nil,
			err:      status.Errorf(codes.Internal, "error getting incoming context metadata"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := SetHeaderFromMetadataFilter(test.context, w, nil)

			if (err != nil) != (test.err != nil) {
				t.Errorf("expected error: %v, got: %v", test.err, err)
			}

			if test.expected != nil {
				for key, expectedValue := range test.expected {
					if got := w.Header().Get(key); got != expectedValue {
						t.Errorf("expected header %s: %s, got: %s", key, expectedValue, got)
					}
				}
				if w.Header().Get("grpcgateway-foo") != "" {
					t.Errorf("grpcgateway-foo header should be filtered out")
				}
			}
		})
	}
}

func TestForwardErrorResponseRewriter(t *testing.T) {
	untouchedResponse := &v0Roles.ListRolesResp{}

	tests := []struct {
		name     string
		response proto.Message
		expected any
	}{
		{
			name:     "Valid grpc status",
			response: &rpcStatus.Status{Code: int32(codes.NotFound), Message: "Resource not found"},
			expected: &v0Types.ErrorResponse{
				Status:  int32(http.StatusNotFound),
				Message: "Resource not found",
			},
		},
		{
			name:     "Invalid response type",
			response: untouchedResponse,
			expected: untouchedResponse,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, _ := ForwardErrorResponseRewriter(context.Background(), test.response)

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected result: %v, got: %v", test.expected, result)
			}
		})
	}
}
