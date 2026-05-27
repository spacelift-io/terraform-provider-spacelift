package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shurcooL/graphql"
)

const enumKind string = "ENUM"

type IntrospectionClient struct {
	client *Client
}

func NewIntrospectionClient(client *Client) *IntrospectionClient {
	return &IntrospectionClient{
		client: client,
	}
}

type EnumValue struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Schema struct {
	Types []struct {
		Name        string      `json:"name"`
		Kind        string      `json:"kind"`
		Description string      `json:"description"`
		EnumValues  []EnumValue `graphql:"enumValues(includeDeprecated: $includeDeprecated)" json:"enumValues"`
	} `json:"types"`
}

type IntrospectionQuery struct {
	Schema Schema `graphql:"__schema"`
}

func (c *IntrospectionClient) GetEnumValues(ctx context.Context, enumName string, introspectionOpts ...IntrospectionOption) ([]string, error) {
	resp, err := c.Introspect(ctx, introspectionOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect schema: %w", err)
	}

	for _, t := range resp.Schema.Types {
		if t.Name == enumName && t.Kind == enumKind {
			var values []string
			for _, enumValue := range t.EnumValues {
				values = append(values, enumValue.Name)
			}
			tflog.Debug(ctx, "Found enum values", map[string]any{
				"enumName": enumName,
				"values":   values,
			})
			return values, nil
		}
	}

	return nil, fmt.Errorf("enum type %s not found in schema", enumName)
}

type introspectOpts struct {
	includeDeprecated bool
}

type IntrospectionOption func(*introspectOpts)

func WithIncludeDeprecated(include bool) IntrospectionOption {
	return func(io *introspectOpts) {
		io.includeDeprecated = include
	}
}

func (c *IntrospectionClient) Introspect(ctx context.Context, opts ...IntrospectionOption) (*IntrospectionQuery, error) {
	var query IntrospectionQuery
	introOpts := &introspectOpts{}
	for i := range opts {
		opts[i](introOpts)
	}

	tflog.Debug(ctx, "Introspecting GraphQL endpoint", map[string]any{
		"includeDeprecated": introOpts.includeDeprecated,
	})
	if err := c.client.Query(ctx, "Introspection", &query, map[string]any{
		// https://github.com/graphql/graphql-spec/blob/September2025/spec/Section%204%20--%20Introspection.md
		"includeDeprecated": graphql.Boolean(introOpts.includeDeprecated),
	}); err != nil {
		return nil, fmt.Errorf("failed to perform GraphQL introspection: %w", err)
	}

	return &query, nil
}
