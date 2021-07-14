package structs

// BitbucketDatacenterIntegration represents bitbucket datacenter integration.
type BitbucketDatacenterIntegration struct {
	APIHost        string `graphql:"apiHost"`
	WebhookSecret  string `graphql:"webhookSecret"`
	UserFacingHost string `graphql:"userFacingHost"`
}
