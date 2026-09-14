package structs

import "github.com/shurcooL/graphql"

// AIIntegration represents an integration with an AI/LLM provider.
type AIIntegration struct {
	ID          string   `graphql:"id"`
	Name        string   `graphql:"name"`
	Description *string  `graphql:"description"`
	Labels      []string `graphql:"labels"`
	Provider    string   `graphql:"provider"`
	BaseURL     *string  `graphql:"baseURL"`
	Models      []string `graphql:"models"`
	Enabled     bool     `graphql:"enabled"`

	// Null for API-key based providers.
	ProviderIntegrationID   *string `graphql:"providerIntegrationId"`
	ProviderIntegrationName *string `graphql:"providerIntegrationName"`

	// A union whose only member is currently AWSBedrockConfig, null otherwise.
	ProviderConfig *struct {
		Bedrock struct {
			Region   string   `graphql:"region"`
			Profiles []string `graphql:"profiles"`
		} `graphql:"... on AWSBedrockConfig"`
	} `graphql:"providerConfig"`

	// Null for Spacelift-provided integrations, which no account owns.
	Space *struct {
		ID string `graphql:"id"`
	} `graphql:"space"`

	// Read-only integrations shared with every account. They can be toggled,
	// but not updated or deleted.
	IsSpaceliftProvided bool `graphql:"isSpaceliftProvided"`
}

// AIProviderConfigInput wraps the per-provider config inputs. Exactly one
// field should be set.
type AIProviderConfigInput struct {
	Bedrock *AWSBedrockConfigInput `json:"bedrock"`
}

// Profiles are inference profile ARNs, which Spacelift validates against AWS
// through the attached integration on save.
type AWSBedrockConfigInput struct {
	Region   graphql.String   `json:"region"`
	Profiles []graphql.String `json:"profiles"`
}
