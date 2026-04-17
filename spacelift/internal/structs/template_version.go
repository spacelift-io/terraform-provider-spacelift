package structs

import "github.com/shurcooL/graphql"

// TemplateVersion represents a version of a template in the GraphQL API.
type TemplateVersion struct {
	ID           string   `graphql:"id"`
	ULID         string   `graphql:"ulid"`
	Space        Space    `graphql:"space"`
	Name         string   `graphql:"name"`
	Instructions *string  `graphql:"instructions"`
	RawTemplate  *string  `graphql:"rawTemplate"`
	State        string   `graphql:"state"`
	Labels       []string `graphql:"labels"`
	Version      *string  `graphql:"version"`
	CreatedAt    int      `graphql:"createdAt"`
	UpdatedAt    int      `graphql:"updatedAt"`
	PublishedAt  *int     `graphql:"publishedAt"`
}

type TemplateVersionCreateInput struct {
	TemplateID    graphql.ID       `json:"templateID"`
	State         graphql.String   `json:"state"`
	Instructions  *graphql.String  `json:"instructions"`
	Labels        []graphql.String `json:"labels"`
	Template      *graphql.String  `json:"template"`
	VersionNumber graphql.String   `json:"versionNumber"`
}

type TemplateVersionUpdateInput struct {
	State         graphql.String   `json:"state"`
	Instructions  *graphql.String  `json:"instructions"`
	Labels        []graphql.String `json:"labels"`
	Template      *graphql.String  `json:"template"`
	VersionNumber graphql.String   `json:"versionNumber"`
}
