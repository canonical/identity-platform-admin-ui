// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package hydra

import (
	"net/http"

	client "github.com/ory/hydra-client-go/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	c *client.APIClient
}

func (c *Client) OAuth2Api() client.OAuth2Api {
	return c.c.OAuth2Api
}

func NewClient(url string, debug bool) *Client {
	c := new(Client)

	configuration := client.NewConfiguration()

	configuration.Debug = debug
	configuration.Servers = []client.ServerConfiguration{
		{
			URL: url,
		},
	}

	configuration.HTTPClient = new(http.Client)
	configuration.HTTPClient.Transport = otelhttp.NewTransport(http.DefaultTransport)

	c.c = client.NewAPIClient(configuration)

	return c
}
