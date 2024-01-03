package structs

import "github.com/shurcooL/graphql"

// SavedFilterType represents a saved filter type.
type SavedFilterType string

// SavedFilter represents SavedFilter data relevant to the provider.
type SavedFilter struct {
	ID        string `graphql:"id"`
	Name      string `graphql:"name"`
	Data      string `graphql:"data"`
	Type      string `graphql:"type"`
	IsPublic  bool   `graphql:"isPublic"`
	CreatedBy string `graphql:"createdBy"`
}

// SavedFilterInput represents the input required to create a saved filter.
type SavedFilterInput struct {
	Name     graphql.String  `json:"name"`
	Data     graphql.String  `json:"data"`
	Type     graphql.String  `json:"type"`
	IsPublic graphql.Boolean `json:"isPublic"`
}

var SavedFilterTypes = []string{
	"stacks",
	"blueprints",
	"contexts",
	"webhooks",
}
