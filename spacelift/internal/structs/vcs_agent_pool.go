package structs

// VCSAgentPool represents the VCS Agent Pool data relevant to the provider.
type VCSAgentPool struct {
	ID          string  `graphql:"id"`
	Config      *string `graphql:"config"`
	Description *string `graphql:"description"`
	Name        string  `graphql:"name"`
}
