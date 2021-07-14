package structs

// GitlabIntegration represents gitlab integration.
type GitlabIntegration struct {
	APIHost       string `graphql:"apiHost"`
	WebhookSecret string `graphql:"webhookSecret"`
}
