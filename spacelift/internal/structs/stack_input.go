package structs

import "github.com/shurcooL/graphql"

// StackInput represents the input required to create or update a Stack.
type StackInput struct {
	Administrative   graphql.Boolean    `json:"administrative"`
	Autodeploy       graphql.Boolean    `json:"autodeploy"`
	Autoretry        graphql.Boolean    `json:"autoretry"`
	BeforeInit       *[]graphql.String  `json:"beforeInit"`
	Branch           graphql.String     `json:"branch"`
	Description      *graphql.String    `json:"description"`
	Labels           *[]graphql.String  `json:"labels"`
	Name             graphql.String     `json:"name"`
	Namespace        *graphql.String    `json:"namespace"`
	ProjectRoot      *graphql.String    `json:"projectRoot"`
	Provider         *graphql.String    `json:"provider"`
	VendorConfig     *VendorConfigInput `json:"vendorConfig"`
	Repository       graphql.String     `json:"repository"`
	RunnerImage      *graphql.String    `json:"runnerImage"`
	TerraformVersion *graphql.String    `json:"terraformVersion"`
	WorkerPool       *graphql.ID        `json:"workerPool"`
}

// VendorConfigInput represents vendor-specific configuration.
type VendorConfigInput struct {
	CloudFormationInput *CloudFormationInput `json:"cloudFormation"`
	Pulumi              *PulumiInput         `json:"pulumi"`
	Terraform           *struct {
	} `json:"terraform"`
}

// CloudFormationInput represents CloudFormation-specific configuration.
type CloudFormationInput struct {
	EntryTemplateFile graphql.String `json:"entryTemplateFile"`
	Region            graphql.String `json:"region"`
	StackName         graphql.String `json:"stackName"`
	TemplateBucket    graphql.String `json:"templateBucket"`
}

// PulumiInput represents Pulumi-specific configuration.
type PulumiInput struct {
	LoginURL  graphql.String `json:"loginURL"`
	StackName graphql.String `json:"stackName"`
}
