package structs

import "github.com/shurcooL/graphql"

// WebhooksIntegrationInput represents the input required to create or update a webhook.
type WebhooksIntegrationInput struct {
	Enabled  graphql.Boolean `json:"enabled"`
	Endpoint graphql.String  `json:"endpoint"`
	Secret   graphql.String  `json:"secret"`
}
