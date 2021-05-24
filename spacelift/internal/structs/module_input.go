package structs

import "github.com/shurcooL/graphql"

// ModuleCreateInput represents the input required to create a Module.
type ModuleCreateInput struct {
	UpdateInput ModuleUpdateInput `json:"updateInput"`
	Namespace   *graphql.String   `json:"namespace"`
	Provider    *graphql.String   `json:"provider"`
	Repository  graphql.String    `json:"repository"`
}

// ModuleUpdateInput represents the input required to update a Module.
type ModuleUpdateInput struct {
	Administrative    graphql.Boolean   `json:"administrative"`
	Branch            graphql.String    `json:"branch"`
	Description       *graphql.String   `json:"description"`
	Labels            *[]graphql.String `json:"labels"`
	Name              *graphql.String   `json:"name"`
	ProjectRoot       *graphql.String   `json:"projectRoot"`
	SharedAccounts    *[]graphql.String `json:"sharedAccounts"`
	TerraformProvider *graphql.String   `json:"terraformProvider"`
	WorkerPool        *graphql.ID       `json:"workerPool"`
}
