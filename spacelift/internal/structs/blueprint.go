package structs

import "github.com/shurcooL/graphql"

type Blueprint struct {
	ID          string   `graphql:"id"`
	Space       Space    `graphql:"space"`
	Name        string   `graphql:"name"`
	Description *string  `graphql:"description"`
	RawTemplate *string  `graphql:"rawTemplate"`
	State       string   `graphql:"state"`
	Labels      []string `graphql:"labels"`
}

type BlueprintCreateInput struct {
	Space       graphql.ID       `json:"space"`
	Name        graphql.String   `json:"name"`
	State       graphql.String   `json:"state"`
	Description *graphql.String  `json:"description"`
	Labels      []graphql.String `json:"labels"`
	Template    *graphql.String  `json:"template"`
}
