package structs

import "github.com/shurcooL/graphql"

type Action string

type Role struct {
	ID                string   `graphql:"id"`
	Slug              string   `graphql:"slug"`
	IsSystem          bool     `graphql:"isSystem"`
	Name              string   `graphql:"name"`
	Description       string   `graphql:"description"`
	Actions           []Action `graphql:"actions"`
	RoleBindingsCount int      `graphql:"roleBindingsCount"`
}

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
