package search

import "github.com/shurcooL/graphql"

type Input struct {
	First      *graphql.Int      `json:"first"`
	After      *graphql.String   `json:"after"`
	Predicates *[]QueryPredicate `json:"predicates"`
}

type QueryPredicate struct {
	Field      graphql.String       `json:"field"`
	Constraint QueryFieldConstraint `json:"constraint"`
}

type QueryFieldConstraint struct {
	BooleanEquals *[]graphql.Boolean `json:"booleanEquals"`
	EnumEquals    *[]graphql.String  `json:"enumEquals"`
	StringMatches *[]graphql.String  `json:"stringMatches"`
}
