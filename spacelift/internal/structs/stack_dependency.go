package structs

import "github.com/shurcooL/graphql"

type StackDependencyDetail struct {
	ID string `graphql:"id"`
}

type StackDependency struct {
	ID             string                `graphql:"id"`
	Stack          StackDependencyDetail `graphql:"stack"`
	DependsOnStack StackDependencyDetail `graphql:"dependsOnStack"`
}

type StackDependencyInput struct {
	StackID          graphql.ID `json:"stackId"`
	DependsOnStackID graphql.ID `json:"dependsOnStackId"`
}
