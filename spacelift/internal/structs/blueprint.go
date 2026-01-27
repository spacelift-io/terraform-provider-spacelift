package structs

import "github.com/shurcooL/graphql"

type Blueprint struct {
	ID           string   `graphql:"id"`
	ULID         string   `graphql:"ulid"`
	Space        Space    `graphql:"space"`
	Name         string   `graphql:"name"`
	Description  *string  `graphql:"description"`
	Instructions *string  `graphql:"instructions"`
	RawTemplate  *string  `graphql:"rawTemplate"`
	State        string   `graphql:"state"`
	Labels       []string `graphql:"labels"`
	Version      *string  `graphql:"version"`
	CreatedAt    int      `graphql:"createdAt"`
	UpdatedAt    int      `graphql:"updatedAt"`
	PublishedAt  *int     `graphql:"publishedAt"`
}

type BlueprintCreateInput struct {
	Space       graphql.ID       `json:"space"`
	Name        graphql.String   `json:"name"`
	State       graphql.String   `json:"state"`
	Description *graphql.String  `json:"description"`
	Labels      []graphql.String `json:"labels"`
	Template    *graphql.String  `json:"template"`
}
