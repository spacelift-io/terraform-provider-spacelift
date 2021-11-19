package structs

import "github.com/shurcooL/graphql"

// StackInput represents the input required to create or update a Stack.
type StackInput struct {
	Administrative      graphql.Boolean    `json:"administrative"`
	AfterApply          *[]graphql.String  `json:"afterApply"`
	AfterDestroy        *[]graphql.String  `json:"afterDestroy"`
	AfterInit           *[]graphql.String  `json:"afterInit"`
	AfterPerform        *[]graphql.String  `json:"afterPerform"`
	AfterPlan           *[]graphql.String  `json:"afterPlan"`
	Autodeploy          graphql.Boolean    `json:"autodeploy"`
	Autoretry           graphql.Boolean    `json:"autoretry"`
	BeforeApply         *[]graphql.String  `json:"beforeApply"`
	BeforeDestroy       *[]graphql.String  `json:"beforeDestroy"`
	BeforeInit          *[]graphql.String  `json:"beforeInit"`
	BeforePerform       *[]graphql.String  `json:"beforePerform"`
	BeforePlan          *[]graphql.String  `json:"beforePlan"`
	Branch              graphql.String     `json:"branch"`
	Description         *graphql.String    `json:"description"`
	GitHubActionDeploy  graphql.Boolean    `json:"githubActionDeploy"`
	Labels              *[]graphql.String  `json:"labels"`
	LocalPreviewEnabled graphql.Boolean    `json:"localPreviewEnabled"`
	Name                graphql.String     `json:"name"`
	Namespace           *graphql.String    `json:"namespace"`
	ProjectRoot         *graphql.String    `json:"projectRoot"`
	Provider            *graphql.String    `json:"provider"`
	Repository          graphql.String     `json:"repository"`
	RunnerImage         *graphql.String    `json:"runnerImage"`
	VendorConfig        *VendorConfigInput `json:"vendorConfig"`
	WorkerPool          *graphql.ID        `json:"workerPool"`
}

// VendorConfigInput represents vendor-specific configuration.
type VendorConfigInput struct {
	CloudFormationInput *CloudFormationInput `json:"cloudFormation"`
	Pulumi              *PulumiInput         `json:"pulumi"`
	Terraform           *TerraformInput      `json:"terraform"`
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

// TerraformInput represents Terraform-specific configuration.
type TerraformInput struct {
	Version   *graphql.String `json:"version"`
	Workspace *graphql.String `json:"workspace"`
}
