package structs

import "github.com/shurcooL/graphql"

// WebhooksIntegration represents the input required to create or update a webhook.
type NamedWebhooksIntegration struct {
	ID       string   `graphql:"id" json:"id"`
	Enabled  bool     `graphql:"enabled" json:"enabled"`
	Endpoint string   `graphql:"endpoint" json:"endpoint"`
	Space    Space    `graphql:"space" json:"space"`
	Name     string   `graphql:"name" json:"name"`
	Secret   string   `graphql:"secret" json:"secret"`
	Labels   []string `graphql:"labels" json:"labels"`
}

// WebhooksIntegrationInput represents the input required to create or update a webhook.
type NamedWebhooksIntegrationInput struct {
	Enabled  graphql.Boolean  `json:"enabled"`
	Endpoint graphql.String   `json:"endpoint"`
	Space    graphql.ID       `json:"space"`
	Name     graphql.String   `json:"name"`
	Secret   graphql.String   `json:"secret"`
	Labels   []graphql.String `json:"labels"`
}
