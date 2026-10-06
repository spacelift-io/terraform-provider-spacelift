package structs

import "github.com/shurcooL/graphql"

type StackDependencyDetail struct {
	ID string `graphql:"id"`
}

type StackDependency struct {
	ID             string                `graphql:"id"`
	Stack          StackDependencyDetail `graphql:"stack"`
	DependsOnStack StackDependencyDetail `graphql:"dependsOnStack"`
	TriggerAlways  bool                  `graphql:"triggerAlways"`
}

type StackDependencyInput struct {
	StackID          graphql.ID      `json:"stackId"`
	DependsOnStackID graphql.ID      `json:"dependsOnStackId"`
	TriggerAlways    graphql.Boolean `json:"triggerAlways"`
}

type StackDependencyUpdateInput struct {
	ID            graphql.ID      `json:"id"`
	TriggerAlways graphql.Boolean `json:"triggerAlways"`
}
