package spacelift

import (
	"context"
	"fmt"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

// Client represents a Spacelift client - in practice a thin wrapper over its
// (administrative) GraphQL API.
type Client struct {
	Endpoint string
	Token    string
}

// Mutate runs a GraphQL mutation.
func (c *Client) Mutate(m interface{}, variables map[string]interface{}) error {
	ctx := context.Background()
	return c.client(ctx).Mutate(ctx, m, variables)
}

// Query runs a GraphQL query.
func (c *Client) Query(q interface{}, variables map[string]interface{}) error {
	ctx := context.Background()
	return c.client(ctx).Query(ctx, q, variables)
}

func (c *Client) client(ctx context.Context) *graphql.Client {
	return graphql.NewClient(c.url(), oauth2.NewClient(
		ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: c.Token},
		),
	))
}

func (c *Client) url() string {
	return fmt.Sprintf("%s/graphql", c.Endpoint)
}
