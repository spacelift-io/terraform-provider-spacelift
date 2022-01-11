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

// Client represents a Spacelift client - in practice a thin wrapper over its
// (administrative) GraphQL API.
type Client struct {
	Endpoint          string
	Token             string
	Version           string
	Commit            string
	limiter           *rate.Limiter
	requestsPerSecond *int
	maxBurst          *int
}

// NewClient returns a new Spacelift client for the specified endpoint, token and limiter.
// If limiter is nil, no rate limit is imposed.
func NewClient(endpoint string, token string, requestsPerSecond, maxBurst *int) *Client {
	var limiter *rate.Limiter
	if requestsPerSecond != nil && maxBurst != nil {
		limiter = rate.NewLimiter(rate.Every(time.Second/time.Duration(*requestsPerSecond)), *maxBurst)
	}

	return &Client{
		Endpoint:          endpoint,
		Token:             token,
		limiter:           limiter,
		requestsPerSecond: requestsPerSecond,
		maxBurst:          maxBurst,
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
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token}))

	if c.limiter != nil {
		client = &http.Client{
			Transport: newRateLimitingRoundTripper(client, c.limiter),
		}
	}

	client.Timeout = time.Minute

	retryableClient := retryablehttp.NewClient()
	retryableClient.HTTPClient = client
	retryableClient.Logger = nil

	requestOptions := c.getRequestOptions()

	return graphql.NewClient(
		c.url(),
		retryableClient.StandardClient(),
		requestOptions...,
	)
}

func (c *Client) url() string {
	return fmt.Sprintf("%s/graphql", c.Endpoint)
}

func (c *Client) getRequestOptions() []graphql.RequestOption {
	options := []graphql.RequestOption{
		graphql.WithHeader("Spacelift-Client-Type", "provider"),
		graphql.WithHeader("Spacelift-Provider-Commit", c.Commit),
		graphql.WithHeader("Spacelift-Provider-Version", c.Version)}

	if c.requestsPerSecond != nil && c.maxBurst != nil {
		options = append(options, graphql.WithHeader("Spacelift-Provider-Max-RPS", fmt.Sprint(*c.requestsPerSecond)))
		options = append(options, graphql.WithHeader("Spacelift-Provider-Max-Request-Burst", fmt.Sprint(*c.maxBurst)))
	}

	return options
}
