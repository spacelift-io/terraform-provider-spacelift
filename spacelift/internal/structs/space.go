package structs

import "github.com/shurcooL/graphql"

// Space represents the Space data relevant to the provider.
type Space struct {
	ID              string   `graphql:"id"`
	Description     string   `graphql:"description"`
	Name            string   `graphql:"name"`
	InheritEntities bool     `graphql:"inheritEntities"`
	ParentSpace     *string  `graphql:"parentSpace"`
	Labels          []string `graphql:"labels"`
}

// SpaceInput represents input relevant to creating or updating the Space.
type SpaceInput struct {
	Description     graphql.String    `json:"description"`
	Name            graphql.String    `json:"name"`
	InheritEntities graphql.Boolean   `json:"inheritEntities"`
	ParentSpace     graphql.ID        `json:"parentSpace"`
	Labels          *[]graphql.String `json:"labels"`
}
