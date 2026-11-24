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
	// Omitted when unset, so queries that don't use them keep working on
	// backends that predate them.
	StringEquals  *[]graphql.String  `json:"stringEquals,omitempty"`
	HierarchyNode *HierarchyNodeList `json:"hierarchyNode,omitempty"`
}

type HierarchyNodeList struct {
	Nodes []HierarchyNode `json:"nodes"`
}

// HierarchyNode matches a space by its path of slugs, e.g. root/team/app.
type HierarchyNode struct {
	Equals             graphql.String  `json:"equals"`
	IncludeDescendants graphql.Boolean `json:"includeDescendants"`
}
