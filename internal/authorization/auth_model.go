package authorization

import (
	_ "embed"
	"encoding/json"
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/language/pkg/go/transformer"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:embed schema.openfga
var schema string

// Taken from
// https://github.com/openfga/cli/blob/d5bfb08cd540dc7c10737bcda12dbc292a649e22/internal/authorizationmodel/model.go#L156
var AuthModel = func() openfga.AuthorizationModel {
	var jsonAuthModel openfga.AuthorizationModel
	parsedAuthModel, err := transformer.TransformDSLToProto(schema)
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
	return jsonAuthModel
}()
