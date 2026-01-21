package structs

import "github.com/shurcooL/graphql"

type BlueprintVersionCreateInput struct {
	BlueprintID   graphql.ID       `json:"blueprintID"`
	State         graphql.String   `json:"state"`
	Instructions  *graphql.String  `json:"instructions"`
	Labels        []graphql.String `json:"labels"`
	Template      *graphql.String  `json:"template"`
	VersionNumber graphql.String   `json:"versionNumber"`
}

type BlueprintVersionUpdateInput struct {
	State         graphql.String   `json:"state"`
	Instructions  *graphql.String  `json:"instructions"`
	Labels        []graphql.String `json:"labels"`
	Template      *graphql.String  `json:"template"`
	VersionNumber graphql.String   `json:"versionNumber"`
}
