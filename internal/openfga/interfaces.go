// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package openfga

import (
	"context"

	"github.com/openfga/go-sdk/client"
)

type OpenFGACoreClientInterface interface {
	GetAuthorizationModelId() (string, error)
	CreateStore(context.Context) client.SdkClientCreateStoreRequestInterface
	CreateStoreExecute(client.SdkClientCreateStoreRequestInterface) (*client.ClientCreateStoreResponse, error)
	ReadAuthorizationModel(context.Context) client.SdkClientReadAuthorizationModelRequestInterface
	ReadAuthorizationModelExecute(client.SdkClientReadAuthorizationModelRequestInterface) (*client.ClientReadAuthorizationModelResponse, error)
	ReadAuthorizationModels(context.Context) client.SdkClientReadAuthorizationModelsRequestInterface
	ReadAuthorizationModelsExecute(client.SdkClientReadAuthorizationModelsRequestInterface) (*client.ClientReadAuthorizationModelsResponse, error)
	WriteAuthorizationModel(context.Context) client.SdkClientWriteAuthorizationModelRequestInterface
	WriteAuthorizationModelExecute(client.SdkClientWriteAuthorizationModelRequestInterface) (*client.ClientWriteAuthorizationModelResponse, error)
	Read(context.Context) client.SdkClientReadRequestInterface
	ReadExecute(client.SdkClientReadRequestInterface) (*client.ClientReadResponse, error)
	Check(context.Context) client.SdkClientCheckRequestInterface
	CheckExecute(client.SdkClientCheckRequestInterface) (*client.ClientCheckResponse, error)
	BatchCheck(context.Context) client.SdkClientBatchCheckRequestInterface
	BatchCheckExecute(client.SdkClientBatchCheckRequestInterface) (*client.ClientBatchCheckResponse, error)
	Write(context.Context) client.SdkClientWriteRequestInterface
	WriteExecute(client.SdkClientWriteRequestInterface) (*client.ClientWriteResponse, error)
	ListObjects(context.Context) client.SdkClientListObjectsRequestInterface
	ListObjectsExecute(client.SdkClientListObjectsRequestInterface) (*client.ClientListObjectsResponse, error)
	ListUsers(context.Context) client.SdkClientListUsersRequestInterface
	ListUsersExecute(client.SdkClientListUsersRequestInterface) (*client.ClientListUsersResponse, error)
}

// OpenFGAClientInterface is the interface used to decouple the OpenFGA store implementation
type OpenFGAClientInterface interface {
	ListObjects(context.Context, string, string, string) ([]string, error)
	ReadTuples(context.Context, string, string, string, string) (*client.ClientReadResponse, error)
	WriteTuples(context.Context, ...Tuple) error
	DeleteTuples(context.Context, ...Tuple) error
	Check(context.Context, string, string, string, ...Tuple) (bool, error)
}

type ListPermissionsFiltersInterface interface {
	WithFilter() any
}
