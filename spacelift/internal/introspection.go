package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		EnumValues  []EnumValue `json:"enumValues"`
	} `json:"types"`
}

type IntrospectionQuery struct {
	Schema Schema `graphql:"__schema"`
}

func (c *IntrospectionClient) GetEnumValues(ctx context.Context, enumName string) ([]string, error) {
	resp, err := c.Introspect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect schema: %w", err)
	}

	for _, t := range resp.Schema.Types {
		if t.Name == enumName && t.Kind == enumKind {
			var values []string
			for _, enumValue := range t.EnumValues {
				values = append(values, enumValue.Name)
			}
			tflog.Debug(ctx, "Found enum values", map[string]interface{}{
				"enumName": enumName,
				"values":   values,
			})
			return values, nil
		}
	}

	return nil, fmt.Errorf("enum type %s not found in schema", enumName)
}

func (c *IntrospectionClient) Introspect(ctx context.Context) (*IntrospectionQuery, error) {
	tflog.Debug(ctx, "Introspecting GraphQL endpoint")

	var query IntrospectionQuery
	if err := c.client.Query(ctx, "Introspection", &query, nil); err != nil {
		return nil, fmt.Errorf("failed to perform GraphQL introspection: %w", err)
	}

	return &query, nil
}
