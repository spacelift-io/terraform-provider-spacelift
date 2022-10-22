package search

import "github.com/shurcooL/graphql"

type SearchInput struct {
	First      *graphql.Int            `json:"first"`
	After      *graphql.String         `json:"after"`
	Predicates *[]SearchQueryPredicate `json:"predicates"`
}

type SearchQueryPredicate struct {
	Field      graphql.String             `json:"field"`
	Constraint SearchQueryFieldConstraint `json:"constraint"`
}

type SearchQueryFieldConstraint struct {
	BooleanEquals *[]graphql.Boolean `json:"booleanEquals"`
	EnumEquals    *[]graphql.String  `json:"enumEquals"`
	StringMatches *[]graphql.String  `json:"stringMatches"`
}
