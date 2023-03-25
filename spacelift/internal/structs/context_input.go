package structs

import "github.com/shurcooL/graphql"

// ContextInput represents the input required to create or update a Context.
type ContextInput struct {
	Name        graphql.String    `json:"name"`
	Description *graphql.String   `json:"description"`
	Hooks       *HooksInput       `json:"hooks"`
	Labels      *[]graphql.String `json:"labels"`
	Space       *graphql.ID       `json:"space"`
}

// HooksInput represents the input required to create or update Hooks.
type HooksInput struct {
	AfterApply    []graphql.String `json:"afterApply"`
	AfterDestroy  []graphql.String `json:"afterDestroy"`
	AfterInit     []graphql.String `json:"afterInit"`
	AfterPerform  []graphql.String `json:"afterPerform"`
	AfterPlan     []graphql.String `json:"afterPlan"`
	AfterRun      []graphql.String `json:"afterRun"`
	BeforeApply   []graphql.String `json:"beforeApply"`
	BeforeDestroy []graphql.String `json:"beforeDestroy"`
	BeforeInit    []graphql.String `json:"beforeInit"`
	BeforePerform []graphql.String `json:"beforePerform"`
	BeforePlan    []graphql.String `json:"beforePlan"`
}
