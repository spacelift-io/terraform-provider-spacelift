package structs

// PluginTemplate represents a Spacelift Plugin Template.
type PluginTemplate struct {
	ID          string                    `graphql:"id"`
	IsGlobal    bool                      `graphql:"isGlobal"`
	Name        string                    `graphql:"name"`
	Description *string                   `graphql:"description"`
	Manifest    string                    `graphql:"manifest"`
	Parameters  []PluginTemplateParameter `graphql:"parameters"`
	Labels      []string                  `graphql:"labels"`
	CreatedAt   int                       `graphql:"createdAt"`
	UpdatedAt   int                       `graphql:"updatedAt"`
}

// PluginTemplateParameter represents a parameter for a plugin template.
type PluginTemplateParameter struct {
	ID          string  `graphql:"id"`
	Name        string  `graphql:"name"`
	Type        string  `graphql:"type"`
	Description *string `graphql:"description"`
	Sensitive   bool    `graphql:"sensitive"`
	Required    bool    `graphql:"required"`
	Default     *string `graphql:"default"`
}
