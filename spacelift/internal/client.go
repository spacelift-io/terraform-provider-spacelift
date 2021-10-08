package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

const (
	maxRequestsPerSecond = 30
	maxRequestBurst      = 10
)

// Client represents a Spacelift client - in practice a thin wrapper over its
// (administrative) GraphQL API.
type Client struct {
	Endpoint string
	Token    string
	Version  string
	Commit   string
	limiter  *rate.Limiter
}

// NewClient returns a new Spacelift client for the specified endpoint and token.
func NewClient(endpoint string, token string) *Client {
	return &Client{
		Endpoint: endpoint,
		Token:    token,
		limiter:  rate.NewLimiter(rate.Every(time.Second/maxRequestsPerSecond), maxRequestBurst),
	}
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

	rateLimitingClient := &http.Client{
		Transport: newRateLimitingRoundTripper(oauthClient, c.limiter),
	}

	retryableClient := retryablehttp.NewClient()
	retryableClient.HTTPClient = rateLimitingClient
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
