package structs

// AzureDevOpsRepoIntegration represents the Azure DevOps integration data relevant to the provider.
type AzureDevOpsRepoIntegration struct {
	ID          string `graphql:"id"`
	Name        string `graphql:"name"`
	Description string `graphql:"description"`
	IsDefault   bool   `graphql:"isDefault"`
	Space       struct {
		ID string `graphql:"id"`
	} `graphql:"space"`
	Labels             []string `graphql:"labels"`
	OrganizationURL    string   `graphql:"organizationURL"`
	UserFacingHost     string   `graphql:"userFacingHost"`
	WebhookPassword    string   `graphql:"webhookPassword"`
	WebhookURL         string   `graphql:"webhookUrl"`
	VCSChecks          string   `graphql:"vcsChecks"`
	UseGitCheckout     bool     `graphql:"useGitCheckout"`
	AccessibleProjects []string `graphql:"accessibleProjects"`
}
