package structs

import "github.com/shurcooL/graphql"

// BitbucketDatacenterIntegration represents the bitbucket datacenter integration data relevant to the provider.
type BitbucketDatacenterIntegration struct {
	ID        string `graphql:"id"`
	Name      string `graphql:"name"`
	IsDefault bool   `graphql:"isDefault"`
	Space     struct {
		ID string `graphql:"id"`
	} `graphql:"space"`
	Labels         []string `graphql:"labels"`
	Description    *string  `graphql:"description"`
	APIHost        string   `graphql:"apiHost"`
	Username       string   `graphql:"username"`
	UserFacingHost string   `graphql:"userFacingHost"`
	WebhookSecret  string   `graphql:"webhookSecret"`
	WebhookURL     string   `graphql:"webhookURL"`
}

// CustomVCSInput represents the custom VCS input data.
type CustomVCSInput struct {
	Name        graphql.String    `json:"name"`
	SpaceID     graphql.ID        `json:"spaceID"`
	Labels      *[]graphql.String `json:"labels"`
	Description *graphql.String   `json:"description"`
	IsDefault   *graphql.Boolean  `json:"isDefault"`
}

// CustomVCSUpdateInput represents the custom VCS update input data.
type CustomVCSUpdateInput struct {
	ID          graphql.ID        `json:"id"`
	SpaceID     graphql.ID        `json:"space"`
	Labels      *[]graphql.String `json:"labels"`
	Description *graphql.String   `json:"description"`
}
