package structs

import "github.com/shurcooL/graphql"

// PluginInstallParameterInput represents a parameter for plugin installation.
type PluginInstallParameterInput struct {
	ID    graphql.String `json:"id"`
	Value graphql.String `json:"value"`
}

// PluginInstallInput represents the input for creating a plugin.
type PluginInstallInput struct {
	Name             graphql.String                 `json:"name"`
	Parameters       *[]PluginInstallParameterInput `json:"parameters,omitempty"`
	PluginTemplateID graphql.ID                     `json:"pluginTemplateID"`
	Labels           *[]graphql.String              `json:"labels,omitempty"`
	Space            graphql.ID                     `json:"space"`
	LabelIdentifier  graphql.String                 `json:"labelIdentifier"`
}
