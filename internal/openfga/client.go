// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package openfga

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type Client struct {
	c OpenFGACoreClientInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (c *Client) APIClient() OpenFGACoreClientInterface {
	return c.c
}

func (c *Client) SetStoreID(ctx context.Context, storeID string) error {
	client, ok := c.c.(*client.OpenFgaClient)

	if !ok {
		return fmt.Errorf("error casting to *openfga.Client")
	}

	client.SetStoreId(storeID)

	c.c = client

	return nil
}

func (c *Client) SetAuthorizationModelID(ctx context.Context, modelID string) error {
	client, ok := c.c.(*client.OpenFgaClient)

	if !ok {
		return fmt.Errorf("error casting to *openfga.Client")
	}

	client.SetAuthorizationModelId(modelID)

	c.c = client

	return nil
}

// ########################## Store Operations #######################################
func (c *Client) CreateStore(ctx context.Context, name string) (string, error) {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.CreateStore")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r, err := c.c.CreateStoreExecute(c.c.CreateStore(ctx).Body(client.ClientCreateStoreRequest{Name: name}))

	if err != nil {
		return "", err
	}

	return r.GetId(), nil
}

// ########################## Store Operations #######################################
// ########################## Model Operations #######################################
func (c *Client) ReadModel(ctx context.Context) (*openfga.AuthorizationModel, error) {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.ReadModel")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	authModel, err := c.c.ReadAuthorizationModelExecute(c.c.ReadAuthorizationModel(ctx))

	if err != nil {
		return nil, err
	}

	return authModel.AuthorizationModel, nil
}

func (c *Client) WriteModel(ctx context.Context, authModel *client.ClientWriteAuthorizationModelRequest) (string, error) {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.WriteModel")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data, err := c.c.WriteAuthorizationModelExecute(
		c.c.WriteAuthorizationModel(ctx).Body(*authModel),
	)

	if err != nil {
		return "", err
	}

	return data.GetAuthorizationModelId(), nil
}

func (c *Client) CompareModel(ctx context.Context, model openfga.AuthorizationModel) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.ReadModel")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	authModel, err := c.ReadModel(ctx)
	if err != nil {
		return false, err
	}

	if authModel.SchemaVersion != model.SchemaVersion {
		c.logger.Errorf("invalid authorization model schema version")
		return false, nil
	}
	if reflect.DeepEqual(authModel.TypeDefinitions, model.TypeDefinitions) {
		c.logger.Errorf("invalid authorization model type definitions")
		return false, nil
	}

	return true, nil
}

// ########################## Model Operations #######################################

// ########################## Write Operations #######################################
func (c *Client) WriteTuple(ctx context.Context, user, relation, object string) error {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.WriteTuple")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r := c.c.Write(ctx)
	body := client.ClientWriteRequest{
		Writes: []openfga.TupleKey{
			*openfga.NewTupleKey(user, relation, object),
		},
	}

	r = r.Body(body)
	_, err := c.c.WriteExecute(r)

	return err
}

func (c *Client) DeleteTuple(ctx context.Context, user, relation, object string) error {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.DeleteTuple")
	defer span.End()

	r := c.c.Write(ctx)
	body := client.ClientWriteRequest{
		Deletes: []openfga.TupleKeyWithoutCondition{
			*openfga.NewTupleKeyWithoutCondition(user, relation, object),
		},
	}
	r = r.Body(body)
	_, err := c.c.WriteExecute(r)

	return err
}

func (c *Client) WriteTuples(ctx context.Context, tuples ...Tuple) error {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.WriteTuples")
	defer span.End()

	ts := make([]openfga.TupleKey, 0)

	for _, tuple := range tuples {
		ts = append(ts, *openfga.NewTupleKey(tuple.Values()))
	}

	r := c.c.Write(ctx)
	body := client.ClientWriteRequest{
		Writes: ts,
	}

	r = r.Body(body)
	_, err := c.c.WriteExecute(r)

	return err
}

func (c *Client) DeleteTuples(ctx context.Context, tuples ...Tuple) error {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.DeleteTuples")
	defer span.End()

	ts := make([]openfga.TupleKeyWithoutCondition, 0)

	for _, tuple := range tuples {
		ts = append(ts, *openfga.NewTupleKeyWithoutCondition(tuple.Values()))
	}

	r := c.c.Write(ctx)
	body := client.ClientWriteRequest{
		Deletes: ts,
	}

	r = r.Body(body)
	_, err := c.c.WriteExecute(r)

	return err
}

// ########################## Write Operations #######################################

// ########################## Check Operations #######################################
func (c *Client) Check(ctx context.Context, user, relation, object string, tuples ...Tuple) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.Check")
	defer span.End()

	contextualTuples := make([]client.ClientContextualTupleKey, len(tuples))
	for i, t := range tuples {
		contextualTuples[i] = client.ClientContextualTupleKey{
			User:     t.User,
			Relation: t.Relation,
			Object:   t.Object,
		}
	}
	r := c.c.Check(ctx)
	body := client.ClientCheckRequest{
		User:             user,
		Relation:         relation,
		Object:           object,
		ContextualTuples: contextualTuples,
	}

	r = r.Body(body)

	check, err := c.c.CheckExecute(r)
	if err != nil {
		c.logger.Infof("body args: %s %s %s", user, relation, object)
		c.logger.Errorf("issues performing check operation: %s", err)
		return false, err
	}

	return check.GetAllowed(), nil
}
func (c *Client) BatchCheck(ctx context.Context, tuples ...Tuple) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.BatchCheck")
	defer span.End()

	modelID, err := c.c.GetAuthorizationModelId()

	if err != nil {
		return false, err
	}

	body := client.ClientBatchCheckBody{}

	for _, t := range tuples {
		body = append(
			body,
			client.ClientCheckRequest{
				User:     t.User,
				Relation: t.Relation,
				Object:   t.Object,
			},
		)
	}

	// should be already set, worth checking and in case removing
	options := client.ClientBatchCheckOptions{
		// You can rely on the model id set in the configuration or override it for this specific request
		AuthorizationModelId: &modelID,
	}

	r := c.c.BatchCheck(ctx).Options(options).Body(body)

	data, err := c.c.BatchCheckExecute(r)

	if err != nil {
		return false, err
	}

	allowed := true
	errString := make([]string, 0)
	errString = append(errString, "error while performing Check operation:")

	for _, check := range *data {
		allowed = allowed && *check.Allowed
		if check.Error != nil {
			errString = append(errString, fmt.Sprintf("* %s", check.Error))
		}
	}

	if !allowed {
		return false, fmt.Errorf(strings.Join(errString, "\n"))
	}

	return allowed, nil
}

// ########################## Check Operations #######################################

// ########################## Read Operations #######################################
func (c *Client) ReadTuples(ctx context.Context, user, relation, object, continuationToken string) (*client.ClientReadResponse, error) {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.ReadTuples")
	defer span.End()

	r := c.c.Read(ctx)

	body := client.ClientReadRequest{
		User:     &user,
		Relation: &relation,
		Object:   &object,
	}

	r = r.Body(body).Options(client.ClientReadOptions{ContinuationToken: &continuationToken})
	res, err := c.c.ReadExecute(r)

	// TODO @shipperizer do we want to log in here or simply return the error?

	return res, err
}

func (c *Client) ListObjects(ctx context.Context, user, relation, objectType string) ([]string, error) {
	ctx, span := c.tracer.Start(ctx, "openfga.Client.ListObjects")
	defer span.End()

	r := c.c.ListObjects(ctx)

	body := client.ClientListObjectsRequest{
		User:     user,
		Relation: relation,
		Type:     objectType,
	}
	r = r.Body(body)
	objectsResponse, err := c.c.ListObjectsExecute(r)
	if err != nil {
		c.logger.Errorf("issues performing list operation: %s", err)
		return nil, err
	}

	allowedObjs := make([]string, len(objectsResponse.GetObjects()))
	// TODO @shipperizer evaluate if this needs removing
	for i, p := range objectsResponse.GetObjects() {
		// remove the "{objectType}:" prefix from the response
		allowedObjs[i] = p[len(fmt.Sprintf("%s:", objectType)):]
	}

	return allowedObjs, nil
}

// ########################## Read Operations #######################################

func NewClient(cfg *Config) *Client {
	c := new(Client)

	if cfg == nil {
		panic("OpenFGA config missing")
	}

	fga, err := client.NewSdkClient(
		&client.ClientConfiguration{
			ApiScheme: cfg.ApiScheme,
			ApiHost:   cfg.ApiHost,
			StoreId:   cfg.StoreID,
			Credentials: &credentials.Credentials{
				Method: credentials.CredentialsMethodApiToken,
				Config: &credentials.Config{
					ApiToken: cfg.ApiToken,
				},
			},
			AuthorizationModelId: cfg.AuthModelID,
			Debug:                cfg.Debug,
			HTTPClient:           &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
		},
	)
	if err != nil {
		panic(fmt.Sprintf("issues setting up OpenFGA client %s", err))
	}

	c.c = fga
	c.tracer = cfg.Tracer
	c.monitor = cfg.Monitor
	c.logger = cfg.Logger

	return c
}
