package structs

import "github.com/shurcooL/graphql"

// ModuleCreateInput represents the input required to create a Module.
type ModuleCreateInput struct {
	UpdateInput       ModuleUpdateInput `json:"updateInput"`
	Name              *graphql.String   `json:"name"`
	Namespace         *graphql.String   `json:"namespace"`
	Provider          *graphql.String   `json:"provider"`
	Repository        graphql.String    `json:"repository"`
	RepositoryURL     *graphql.String   `json:"repositoryURL"`
	TerraformProvider *graphql.String   `json:"terraformProvider"`
	Space             *graphql.String   `json:"space"`
	VCSIntegrationID  *graphql.ID       `json:"vcsIntegrationId"`
	WorkflowTool      *graphql.String   `json:"workflowTool"`
}

// ModuleUpdateInput represents the input required to update a Module.
type ModuleUpdateInput struct {
	Administrative      graphql.Boolean   `json:"administrative"`
	Branch              graphql.String    `json:"branch"`
	Description         *graphql.String   `json:"description"`
	Labels              *[]graphql.String `json:"labels"`
	LocalPreviewEnabled graphql.Boolean   `json:"localPreviewEnabled"`
	ProjectRoot         *graphql.String   `json:"projectRoot"`
	ProtectFromDeletion graphql.Boolean   `json:"protectFromDeletion"`
	SharedAccounts      *[]graphql.String `json:"sharedAccounts"`
	WorkerPool          *graphql.ID       `json:"workerPool"`
	Space               *graphql.String   `json:"space"`
}

// ModuleUpdateV2Input represents the input required to update a Module,
// including the source.
type ModuleUpdateV2Input struct {
	Administrative      graphql.Boolean   `json:"administrative"`
	Branch              graphql.String    `json:"branch"`
	Description         *graphql.String   `json:"description"`
	Labels              *[]graphql.String `json:"labels"`
	LocalPreviewEnabled graphql.Boolean   `json:"localPreviewEnabled"`
	Namespace           *graphql.String   `json:"namespace"`
	ProjectRoot         *graphql.String   `json:"projectRoot"`
	ProtectFromDeletion graphql.Boolean   `json:"protectFromDeletion"`
	Provider            *graphql.String   `json:"provider"`
	Repository          graphql.String    `json:"repository"`
	RepositoryURL       *graphql.String   `json:"repositoryURL"`
	SharedAccounts      *[]graphql.String `json:"sharedAccounts"`
	Space               *graphql.String   `json:"space"`
	VCSIntegrationID    *graphql.ID       `json:"vcsIntegrationId"`
	WorkerPool          *graphql.ID       `json:"workerPool"`
	WorkflowTool        *graphql.String   `json:"workflowTool"`
}
