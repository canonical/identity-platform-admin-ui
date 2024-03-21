// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package openfga

import (
	"context"

	"github.com/openfga/go-sdk/client"
)

type OpenFGAClientInterface interface {
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
}
