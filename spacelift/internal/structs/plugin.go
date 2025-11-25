package structs

// Plugin represents a Spacelift Plugin.
type Plugin struct {
	ID           string `graphql:"id"`
	Name         string `graphql:"name"`
	SpaceDetails struct {
		ID string `graphql:"id"`
	} `graphql:"spaceDetails"`
	PluginTemplateDetails *struct {
		ID string `graphql:"id"`
	} `graphql:"pluginTemplateDetails"`
	Labels          []string `graphql:"labels"`
	LabelIdentifier string   `graphql:"labelIdentifier"`
	CreatedAt       int      `graphql:"createdAt"`
	UpdatedAt       int      `graphql:"updatedAt"`
}
