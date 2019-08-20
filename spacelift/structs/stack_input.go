package structs

import "github.com/shurcooL/graphql"

// StackInput represents the input required to create or update a Stack.
type StackInput struct {
	Administrative   graphql.Boolean `json:"administrative"`
	Branch           graphql.String  `json:"branch"`
	Description      *graphql.String `json:"description"`
	Name             graphql.String  `json:"name"`
	ReadersSlug      *graphql.String `json:"readersSlug"`
	Repo             graphql.String  `json:"repo"`
	TerraformVersion *graphql.String `json:"terraformVersion"`
	WritersSlug      *graphql.String `json:"writersSlug"`
}
