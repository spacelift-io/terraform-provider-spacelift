package internal

import (
	"context"
	"fmt"
	"time"

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
func (c *Client) Mutate(ctx context.Context, mutationName string, m interface{}, variables map[string]interface{}) (err error) {
	client := c.client(ctx)

	for i := 1; i <= 3; i++ {
		err = client.Mutate(ctx, m, variables, graphql.WithHeader("Spacelift-GraphQL-Mutation", mutationName))

		if err == nil || !isRetriable(err) {
			break
		}

		time.Sleep(time.Duration(i) * time.Second)
	}

	return
}

// Query runs a GraphQL query.
func (c *Client) Query(ctx context.Context, queryName string, q interface{}, variables map[string]interface{}) (err error) {
	client := c.client(ctx)

	for i := 1; i <= 3; i++ {
		err = client.Query(ctx, q, variables, graphql.WithHeader("Spacelift-GraphQL-Query", queryName))

		if err == nil || !isRetriable(err) {
			break
		}

		time.Sleep(time.Duration(i) * time.Second)
	}

	return
}

func (c *Client) client(ctx context.Context) *graphql.Client {
	return graphql.NewClient(
		c.url(),
		oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token})),
		graphql.WithHeader("Spacelift-Client-Type", "provider"),
		graphql.WithHeader("Spacelift-Provider-Commit", c.Commit),
		graphql.WithHeader("Spacelift-Provider-Version", c.Version),
	)
}

func (c *Client) url() string {
	return fmt.Sprintf("%s/graphql", c.Endpoint)
}

func isRetriable(err error) bool {
	switch typedErr := err.(type) {
	case *graphql.ResponseError:
		return true
	case *graphql.ServerError:
		return typedErr.StatusCode/100 == 5
	default:
		return false
	}
}
