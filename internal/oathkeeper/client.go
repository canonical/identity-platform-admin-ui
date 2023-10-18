package oathkeeper

import (
	"net/http"

	client "github.com/ory/oathkeeper-client-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	c *client.APIClient
}

func (c *Client) ApiApi() client.ApiApi {
	return c.c.ApiApi
}

func NewClient(url string, debug bool) *Client {
	c := new(Client)

	configuration := client.NewConfiguration()
	configuration.Debug = debug
	configuration.Servers = []client.ServerConfiguration{
		{
			URL:         url,
			Description: "Oathkeeper endpoint",
		},
	}

	configuration.HTTPClient = new(http.Client)
	configuration.HTTPClient.Transport = otelhttp.NewTransport(http.DefaultTransport)

	c.c = client.NewAPIClient(configuration)

	return c
}
