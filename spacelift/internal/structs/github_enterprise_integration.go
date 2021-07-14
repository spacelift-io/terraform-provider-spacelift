package structs

// GithubEnterpriseIntegration represents github enterprise integration.
type GithubEnterpriseIntegration struct {
	AppID         string `graphql:"appID"`
	APIHost       string `graphql:"apiHost"`
	WebhookSecret string `graphql:"webhookSecret"`
}
