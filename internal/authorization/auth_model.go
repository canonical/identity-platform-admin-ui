// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authorization

import (
	_ "embed"
	"encoding/json"
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/language/pkg/go/transformer"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:embed authorization_model.v0.openfga
var v0Schema string

//go:embed authorization_model.v2.openfga
var v2Schema string

type AuthorizationModelProvider struct {
	apiVersion string
	model      *openfga.AuthorizationModel
}

func readAuthzModelFromDSLString(dslString string) *openfga.AuthorizationModel {
	var jsonAuthModel openfga.AuthorizationModel
	parsedAuthModel, err := transformer.TransformDSLToProto(dslString)
	if err != nil {
		panic(fmt.Errorf("failed to transform due to %w", err))
	}

	bytes, err := protojson.Marshal(parsedAuthModel)
	if err != nil {
		panic(fmt.Errorf("failed to transform due to %w", err))
	}

	err = json.Unmarshal(bytes, &jsonAuthModel)
	if err != nil {
		panic(fmt.Errorf("failed to transform due to %w", err))
	}

	return &jsonAuthModel
}

func (a *AuthorizationModelProvider) prepareModel() *openfga.AuthorizationModel {
	var model string
	switch a.apiVersion {
	case "v2":
		model = v2Schema
	default:
		model = v0Schema
	}

	return readAuthzModelFromDSLString(model)
}

func (a *AuthorizationModelProvider) GetModel() *openfga.AuthorizationModel {
	if a.model == nil {
		a.model = a.prepareModel()
	}

	return a.model
}

func NewAuthorizationModelProvider(apiVersion string) *AuthorizationModelProvider {
	a := new(AuthorizationModelProvider)
	a.apiVersion = apiVersion

	return a
}
