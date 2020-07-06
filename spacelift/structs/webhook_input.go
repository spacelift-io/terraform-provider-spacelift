package structs

import "github.com/shurcooL/graphql"

// WebhooksIntegrationInput represents the input required to create or update a webhook.
type WebhooksIntegrationInput struct {
	Enabled  graphql.Boolean `graphql:"enabled"`
	Endpoint graphql.String  `graphql:"endpoint"`
	Secret   graphql.String  `graphql:"secret"`
}
