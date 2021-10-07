package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

// Client represents a Spacelift client - in practice a thin wrapper over its
// (administrative) GraphQL API.
type Client struct {
	Endpoint string
	Token    string
	Version  string
	Commit   string
}

// Mutate runs a GraphQL mutation.
func (c *Client) Mutate(ctx context.Context, mutationName string, m interface{}, variables map[string]interface{}) error {
	client := c.client(ctx)

	return client.Mutate(ctx, m, variables, graphql.WithHeader("Spacelift-GraphQL-Mutation", mutationName))
}

// Query runs a GraphQL query.
func (c *Client) Query(ctx context.Context, queryName string, q interface{}, variables map[string]interface{}) error {
	client := c.client(ctx)

	return client.Query(ctx, q, variables, graphql.WithHeader("Spacelift-GraphQL-Query", queryName))
}

func (c *Client) client(ctx context.Context) *graphql.Client {
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token}))

	retryableClient := retryablehttp.NewClient()
	retryableClient.HTTPClient = oauthClient
	retryableClient.Logger = nil

	return graphql.NewClient(
		c.url(),
		retryableClient.StandardClient(),
		graphql.WithHeader("Spacelift-Client-Type", "provider"),
		graphql.WithHeader("Spacelift-Provider-Commit", c.Commit),
		graphql.WithHeader("Spacelift-Provider-Version", c.Version),
	)
}

func (c *Client) url() string {
	return fmt.Sprintf("%s/graphql", c.Endpoint)
}
