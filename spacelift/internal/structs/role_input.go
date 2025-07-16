package structs

import "github.com/shurcooL/graphql"

type RoleInput struct {
	Name        graphql.String   `json:"name"`
	Description *graphql.String  `json:"description"`
	Actions     []graphql.String `json:"actions"`
}

type RoleUpdateInput struct {
	Name        *graphql.String   `json:"name"`
	Description *graphql.String   `json:"description"`
	Actions     *[]graphql.String `json:"Actions"`
}
