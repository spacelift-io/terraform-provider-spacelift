package structs

import "github.com/shurcooL/graphql"

// TemplateDeploymentTemplate represents a blueprint deployment group.
type TemplateDeploymentTemplate struct {
	ID string `graphql:"id"`
}

// TemplateDeploymentTemplateVersion represents the template version used for a deployment.
type TemplateDeploymentTemplateVersion struct {
	ID      string  `graphql:"id"`
	Version *string `graphql:"version"`
}

type TemplateDeploymentSpace struct {
	ID string `graphql:"id"`
}

type TemplateDeploymentStack struct {
	ID string `graphql:"id"`
}

type TemplateDeploymentInput struct {
	ID       string `graphql:"id"`
	Value    string `graphql:"value"`
	Secret   bool   `graphql:"secret"`
	Checksum string `graphql:"checksum"`
}

// TemplateDeployment represents a deployment of a Spacelift template.
type TemplateDeployment struct {
	ID              string                            `graphql:"id"`
	Name            string                            `graphql:"name"`
	Description     *string                           `graphql:"description"`
	CreatedAt       int32                             `graphql:"createdAt"`
	State           string                            `graphql:"state"`
	Inputs          []TemplateDeploymentInput         `graphql:"inputs"`
	Space           TemplateDeploymentSpace           `graphql:"space"`
	Template        TemplateDeploymentTemplate        `graphql:"blueprint"`
	Stacks          []TemplateDeploymentStack         `graphql:"stacks"`
	TemplateVersion TemplateDeploymentTemplateVersion `graphql:"blueprintVersion"`
}

// BlueprintDeploymentCreateInputPair represents a single input for deployment creation.
// Note: Uses Blueprint naming to match the GraphQL API.
type BlueprintDeploymentCreateInputPair struct {
	ID    graphql.String `json:"id"`
	Value graphql.String `json:"value"`
}

// BlueprintDeploymentCreateInput represents the input for creating a deployment.
// Note: Uses Blueprint naming to match the GraphQL API.
type BlueprintDeploymentCreateInput struct {
	Space       graphql.ID                           `json:"space"`
	Name        graphql.String                       `json:"name"`
	Description *graphql.String                      `json:"description,omitempty"`
	Inputs      []BlueprintDeploymentCreateInputPair `json:"inputs,omitempty"`
}

// BlueprintDeploymentUpdateInput represents the input for updating a deployment.
// Note: Uses Blueprint naming to match the GraphQL API.
type BlueprintDeploymentUpdateInput struct {
	VersionID   *graphql.ID                          `json:"versionId,omitempty"`
	Name        *graphql.String                      `json:"name,omitempty"`
	Description *graphql.String                      `json:"description,omitempty"`
	Inputs      []BlueprintDeploymentCreateInputPair `json:"inputs,omitempty"`
}
