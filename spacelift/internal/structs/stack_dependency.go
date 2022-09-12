package structs

import "github.com/shurcooL/graphql"

type StackDependency struct {
	ID               string `graphql:"id"`
	StackID          string `graphql:"stackId"`
	DependsOnStackID string `graphql:"dependsOnStackId"`
}

type StackDependencyInput struct {
	StackID          graphql.ID `json:"stackId"`
	DependsOnStackID graphql.ID `json:"dependsOnStackId"`
}
