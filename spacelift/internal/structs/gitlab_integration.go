package structs

// GitLabIntegration represents an GitLab identity provided by the Spacelift
// integration.
type GitLabIntegration struct {
	ID    string `graphql:"id"`
	Name  string `graphql:"name"`
	Space struct {
		ID string `graphql:"id"`
	} `graphql:"space"`
	IsDefault      bool     `graphql:"isDefault"`
	Labels         []string `graphql:"labels"`
	Description    *string  `graphql:"description"`
	APIHost        string   `graphql:"apiHost"`
	UserFacingHost string   `graphql:"userFacingHost"`
	Token          string   `graphql:"privateToken"`
	WebhookSecret  string   `graphql:"webhookSecret"`
	WebhookURL     string   `graphql:"webhookURL"`
}
