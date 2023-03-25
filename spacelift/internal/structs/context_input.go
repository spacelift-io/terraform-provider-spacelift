package structs

import "github.com/shurcooL/graphql"

// ContextInput represents the input required to create or update a Context.
type ContextInput struct {
	Name        graphql.String    `graphql:"name"`
	Description *graphql.String   `graphql:"description"`
	Hooks       *HooksInput       `graphql:"hooks"`
	Labels      *[]graphql.String `json:"labels"`
	Space       *graphql.ID       `json:"space"`
}

// HooksInput represents the input required to create or update Hooks.
type HooksInput struct {
	AfterApply    []graphql.String `graphql:"afterApply"`
	AfterDestroy  []graphql.String `graphql:"afterDestroy"`
	AfterInit     []graphql.String `graphql:"afterInit"`
	AfterPerform  []graphql.String `graphql:"afterPerform"`
	AfterPlan     []graphql.String `graphql:"afterPlan"`
	AfterRun      []graphql.String `graphql:"afterRun"`
	BeforeApply   []graphql.String `graphql:"beforeApply"`
	BeforeDestroy []graphql.String `graphql:"beforeDestroy"`
	BeforeInit    []graphql.String `graphql:"beforeInit"`
	BeforePerform []graphql.String `graphql:"beforePerform"`
	BeforePlan    []graphql.String `graphql:"beforePlan"`
}
