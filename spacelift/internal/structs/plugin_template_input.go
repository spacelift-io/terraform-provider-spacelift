package structs

import "github.com/shurcooL/graphql"

// PluginTemplateCreateInput represents the input for creating a plugin template.
type PluginTemplateCreateInput struct {
	Name        graphql.String    `json:"name"`
	Description *graphql.String   `json:"description,omitempty"`
	Manifest    graphql.String    `json:"manifest"`
	Labels      *[]graphql.String `json:"labels,omitempty"`
}
