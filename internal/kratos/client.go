package kratos

import (
	client "github.com/ory/kratos-client-go"
)

type Client struct {
	c *client.APIClient
}

func (c *Client) IdentityApi() client.IdentityApi {
	return c.c.IdentityApi
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

	c.c = client.NewAPIClient(configuration)

	return c
}
