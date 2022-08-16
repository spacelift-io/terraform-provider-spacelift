package structs

import "github.com/shurcooL/graphql"

type StackDependency struct {
	ID               string `graphql:"id"`
	StackID          string `graphql:"stackId"`
	DependsOnStackID string `graphql:"dependsOnStackId"`
	Triggers         bool   `graphql:"triggers"`
}

type StackDependencyInput struct {
	StackID          graphql.ID      `json:"stackId"`
	DependsOnStackID graphql.ID      `json:"dependsOnStackId"`
	Triggers         graphql.Boolean `json:"triggers"`
}
