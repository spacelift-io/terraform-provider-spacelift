package structs

// BitbucketDatacenterIntegration represents the bitbucket datacenter integration data relevant to the provider.
type BitbucketDatacenterIntegration struct {
	ID        string `graphql:"id"`
	Name      string `graphql:"name"`
	IsDefault bool   `graphql:"isDefault"`
	Space     struct {
		ID string `graphql:"id"`
	} `graphql:"space"`
	Labels         []string `graphql:"labels"`
	Description    *string  `graphql:"description"`
	APIHost        string   `graphql:"apiHost"`
	Username       string   `graphql:"username"`
	UserFacingHost string   `graphql:"userFacingHost"`
	WebhookSecret  string   `graphql:"webhookSecret"`
	WebhookURL     string   `graphql:"webhookURL"`
	VCSChecks      string   `graphql:"vcsChecks"`
}
