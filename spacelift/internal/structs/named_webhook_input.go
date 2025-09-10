package structs

import "github.com/shurcooL/graphql"

// WebhooksIntegration represents the input required to create or update a webhook.
type NamedWebhooksIntegration struct {
	ID             string   `graphql:"id" json:"id"`
	Enabled        bool     `graphql:"enabled" json:"enabled"`
	Endpoint       string   `graphql:"endpoint" json:"endpoint"`
	Space          Space    `graphql:"space" json:"space"`
	Name           string   `graphql:"name" json:"name"`
	Secret         *string  `graphql:"secret" json:"secret"`
	SecretHeaders  []string `graphql:"secretHeaders" json:"secretHeaders"`
	Labels         []string `graphql:"labels" json:"labels"`
	RetryOnFailure *bool    `graphql:"retryOnFailure" json:"retryOnFailure"`
}

// WebhooksIntegrationInput represents the input required to create or update a webhook.
type NamedWebhooksIntegrationInput struct {
	Enabled        graphql.Boolean  `json:"enabled"`
	Endpoint       graphql.String   `json:"endpoint"`
	Space          graphql.ID       `json:"space"`
	Name           graphql.String   `json:"name"`
	Secret         *graphql.String  `json:"secret"`
	Labels         []graphql.String `json:"labels"`
	RetryOnFailure *graphql.Boolean `json:"retryOnFailure"`
}

// NamedWebhookHeaderInput represents the input required to set a secret header.
type NamedWebhookHeaderInput struct {
	Entries []KeyValuePair `json:"entries"`
}

// KeyValuePair is used in map representations for graphql requests.
type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
