package structs

// BitbucketDatacenterIntegration represents the bitbucket datacenter integration data relevant to the provider.
type BitbucketDatacenterIntegration struct {
	Id             string `graphql:"id"`
	Name           string `graphql:"name"`
	APIHost        string `graphql:"apiHost"`
	Username       string `graphql:"username"`
	UserFacingHost string `graphql:"userFacingHost"`
	WebhookSecret  string `graphql:"webhookSecret"`
	WebhookURL     string `graphql:"webhookURL"`
}
