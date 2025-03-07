// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package types

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	v0Types "github.com/canonical/identity-platform-api/v0/http"
	rpcStatus "google.golang.org/genproto/googleapis/rpc/status"
)

// SetHeaderFromMetadataFilter is the filter function to allow headers from the handlers to be set
// on the HTTP response by the gRPC gateway (to be registered with WithForwardResponseOption)
// DO NOT use WriteHeader func in filters like this one
//
// usage example:
//
// mux := runtime.NewServeMux(
//
//	runtime.WithForwardResponseOption(SetHeaderFromMetadataFilter),
//
// )
func SetHeaderFromMetadataFilter(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Internal, "error getting incoming context metadata")
	}

	for key, values := range md {
		// filter out grpcgateway headers
		if strings.HasPrefix(key, "grpcgateway-") {
			continue
		}

		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	return nil
}

// ForwardErrorResponseRewriter rewrites error message to comply with Admin UI
// standard json response for errors. It doesn't do anything on other messages
// usage example:
//
// mux := runtime.NewServeMux(
//
//	runtime.WithForwardResponseRewriter(ForwardErrorResponseRewriter),
//
// )
func ForwardErrorResponseRewriter(_ context.Context, response proto.Message) (any, error) {
	codeError, ok := response.(*rpcStatus.Status)
	if !ok {
		return response, nil
	}

	httpStatus := runtime.HTTPStatusFromCode(
		codes.Code(codeError.Code),
	)

	return &v0Types.ErrorResponse{
		Status:  int32(httpStatus),
		Message: codeError.GetMessage(),
	}, nil
}
