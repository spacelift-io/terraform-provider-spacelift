package structs

import "github.com/shurcooL/graphql"

// StackInput represents the input required to create or update a Stack.
type StackInput struct {
	Administrative               graphql.Boolean    `json:"administrative"`
	AfterApply                   *[]graphql.String  `json:"afterApply"`
	AfterDestroy                 *[]graphql.String  `json:"afterDestroy"`
	AfterInit                    *[]graphql.String  `json:"afterInit"`
	AfterPerform                 *[]graphql.String  `json:"afterPerform"`
	AfterPlan                    *[]graphql.String  `json:"afterPlan"`
	AfterRun                     *[]graphql.String  `json:"afterRun"`
	Autodeploy                   graphql.Boolean    `json:"autodeploy"`
	Autoretry                    graphql.Boolean    `json:"autoretry"`
	BeforeApply                  *[]graphql.String  `json:"beforeApply"`
	BeforeDestroy                *[]graphql.String  `json:"beforeDestroy"`
	BeforeInit                   *[]graphql.String  `json:"beforeInit"`
	BeforePerform                *[]graphql.String  `json:"beforePerform"`
	BeforePlan                   *[]graphql.String  `json:"beforePlan"`
	Branch                       graphql.String     `json:"branch"`
	Description                  *graphql.String    `json:"description"`
	GitHubActionDeploy           graphql.Boolean    `json:"githubActionDeploy"`
	Labels                       *[]graphql.String  `json:"labels"`
	LocalPreviewEnabled          graphql.Boolean    `json:"localPreviewEnabled"`
	EnableWellKnownSecretMasking graphql.Boolean    `json:"enableWellKnownSecretMasking"`
	Name                         graphql.String     `json:"name"`
	Namespace                    *graphql.String    `json:"namespace"`
	ProjectRoot                  *graphql.String    `json:"projectRoot"`
	AddditionalProjectGlobs      *[]graphql.String  `json:"additionalProjectGlobs"`
	ProtectFromDeletion          graphql.Boolean    `json:"protectFromDeletion"`
	Provider                     *graphql.String    `json:"provider"`
	Repository                   graphql.String     `json:"repository"`
	RepositoryURL                *graphql.String    `json:"repositoryURL"`
	RunnerImage                  *graphql.String    `json:"runnerImage"`
	Space                        *graphql.String    `json:"space"`
	VCSIntegrationID             *graphql.ID        `json:"vcsIntegrationId"`
	VendorConfig                 *VendorConfigInput `json:"vendorConfig"`
	WorkerPool                   *graphql.ID        `json:"workerPool"`
}

// VendorConfigInput represents vendor-specific configuration.
type VendorConfigInput struct {
	AnsibleInput        *AnsibleInput        `json:"ansible"`
	CloudFormationInput *CloudFormationInput `json:"cloudFormation"`
	Kubernetes          *KubernetesInput     `json:"kubernetes"`
	Pulumi              *PulumiInput         `json:"pulumi"`
	Terraform           *TerraformInput      `json:"terraform"`
	TerragruntInput     *TerragruntInput     `json:"terragrunt"`
}

// AnsibleInput represents Ansible-specific configuration.
type AnsibleInput struct {
	Playbook graphql.String `json:"playbook"`
}

// CloudFormationInput represents CloudFormation-specific configuration.
type CloudFormationInput struct {
	EntryTemplateFile graphql.String `json:"entryTemplateFile"`
	Region            graphql.String `json:"region"`
	StackName         graphql.String `json:"stackName"`
	TemplateBucket    graphql.String `json:"templateBucket"`
}

// KubernetesInput represents Kubernetes-specific configuration.
type KubernetesInput struct {
	Namespace      graphql.String  `json:"namespace"`
	KubectlVersion *graphql.String `json:"kubectlVersionkubectlVersion"`
	KubernetesWorkflowTool *graphql.String `json:"kubernetesWorkflowTool"`
}

// PulumiInput represents Pulumi-specific configuration.
type PulumiInput struct {
	LoginURL  graphql.String `json:"loginURL"`
	StackName graphql.String `json:"stackName"`
}

type TerragruntInput struct {
	TerraformVersion     *graphql.String `json:"terraformVersion"`
	TerragruntVersion    *graphql.String `json:"terragruntVersion"`
	UseRunAll            graphql.Boolean `json:"useRunAll"`
	UseSmartSanitization graphql.Boolean `json:"useSmartSanitization"`
	Tool                 *graphql.String `json:"tool"`
}

// TerraformInput represents Terraform-specific configuration.
type TerraformInput struct {
	UseSmartSanitization       *graphql.Boolean `json:"useSmartSanitization"`
	Version                    *graphql.String  `json:"version"`
	WorkflowTool               *graphql.String  `json:"workflowTool"`
	Workspace                  *graphql.String  `json:"workspace"`
	ExternalStateAccessEnabled *graphql.Boolean `json:"externalStateAccessEnabled"`
}
