package structs

import "github.com/shurcooL/graphql"

type BlueprintVersionedGroup struct {
	ID          string   `graphql:"id"`
	Space       Space    `graphql:"space"`
	Name        string   `graphql:"name"`
	Description *string  `graphql:"description"`
	Labels      []string `graphql:"labels"`
}

type BlueprintVersionedGroupCreateInput struct {
	Space       graphql.ID       `json:"space"`
	Name        graphql.String   `json:"name"`
	State       graphql.String   `json:"state"`
	Description *graphql.String  `json:"description"`
	Labels      []graphql.String `json:"labels"`
	Template    *graphql.String  `json:"template"`
}

type BlueprintWithGroup struct {
	ID          string   `graphql:"id"`
	Description *string  `graphql:"description"`
	RawTemplate *string  `graphql:"rawTemplate"`
	State       string   `graphql:"state"`
	Labels      []string `graphql:"labels"`
	Version     string   `graphql:"version"`

	GroupDetails BlueprintVersionedGroup `graphql:"groupDetails"`
}

type BlueprintWithGroupCreateInput struct {
	Group graphql.ID `json:"group"`
	BlueprintWithGroupUpdateInput
}

type BlueprintWithGroupUpdateInput struct {
	Name        graphql.String   `json:"name"`
	State       graphql.String   `json:"state"`
	Description *graphql.String  `json:"description"`
	Labels      []graphql.String `json:"labels"`
	Template    *graphql.String  `json:"template"`
	Version     graphql.String   `json:"version"`
}
