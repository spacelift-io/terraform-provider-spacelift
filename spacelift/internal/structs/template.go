package structs

import "github.com/shurcooL/graphql"

type BlueprintVersionedGroup struct {
	ID          string   `graphql:"id"`
	ULID        string   `graphql:"ulid"`
	Space       Space    `graphql:"space"`
	Name        string   `graphql:"name"`
	Description *string  `graphql:"description"`
	Labels      []string `graphql:"labels"`
	CreatedAt   int      `graphql:"createdAt"`
	UpdatedAt   int      `graphql:"updatedAt"`
}

type BlueprintVersionedGroupCreateInput struct {
	Space       graphql.ID       `json:"space"`
	Name        graphql.String   `json:"name"`
	Description *graphql.String  `json:"description"`
	Labels      []graphql.String `json:"labels"`
}

type BlueprintVersionedGroupUpdateInput struct {
	Space       graphql.ID       `json:"space"`
	Name        graphql.String   `json:"name"`
	Description *graphql.String  `json:"description"`
	Labels      []graphql.String `json:"labels"`
}
