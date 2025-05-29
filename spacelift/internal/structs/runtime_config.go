package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/shurcooL/graphql"
)

// TerraformWorkflowTool represents the workflow tool used by Terraform.
type TerraformWorkflowTool string

const (
	TerraformWorkflowToolTerraformFoss TerraformWorkflowTool = "TERRAFORM_FOSS"
	TerraformWorkflowToolCustom        TerraformWorkflowTool = "CUSTOM"
	TerraformWorkflowToolOpenTofu      TerraformWorkflowTool = "OPEN_TOFU"
)

// TerragruntTool represents the tool used by Terragrunt.
type TerragruntTool string

const (
	TerragruntToolTerraformFoss       TerragruntTool = "TERRAFORM_FOSS"
	TerragruntToolOpenTofu            TerragruntTool = "OPEN_TOFU"
	TerragruntToolManuallyProvisioned TerragruntTool = "MANUALLY_PROVISIONED"
)

// EnvVar represents an environment variable.
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TerragruntRuntimeConfig represents Terragrunt-specific runtime configuration.
type TerragruntRuntimeConfig struct {
	TerraformVersion  *string        `json:"terraformVersion"`
	TerragruntVersion string         `json:"terragruntVersion"`
	UseRunAll         bool           `json:"useRunAll"`
	Tool              TerragruntTool `json:"tool"`
}

// TerraformRuntimeConfig represents Terraform-specific runtime configuration.
type TerraformRuntimeConfig struct {
	WorkflowTool TerraformWorkflowTool `json:"workflowTool"`
	Version      *string               `json:"version"`
}

// RuntimeConfig represents the runtime configuration for a run.
type RuntimeConfig struct {
	Environment           []EnvVar                 `json:"environment"`
	ProjectRoot           string                   `json:"projectRoot"`
	RunnerImage           string                   `json:"runnerImage"`
	TerraformVersion      string                   `json:"terraformVersion"`
	TerraformWorkflowTool TerraformWorkflowTool    `json:"terraformWorkflowTool"`
	AfterApply            []string                 `json:"afterApply"`
	BeforeApply           []string                 `json:"beforeApply"`
	AfterInit             []string                 `json:"afterInit"`
	BeforeInit            []string                 `json:"beforeInit"`
	AfterPlan             []string                 `json:"afterPlan"`
	BeforePlan            []string                 `json:"beforePlan"`
	AfterPerform          []string                 `json:"afterPerform"`
	BeforePerform         []string                 `json:"beforePerform"`
	AfterDestroy          []string                 `json:"afterDestroy"`
	AfterRun              []string                 `json:"afterRun"`
	BeforeDestroy         []string                 `json:"beforeDestroy"`
	Terragrunt            *TerragruntRuntimeConfig `json:"terragrunt"`
	Terraform             *TerraformRuntimeConfig  `json:"terraform"`
	Yaml                  *string                  `json:"yaml"`
}

// EnvVarInput represents input for an environment variable.
type EnvVarInput struct {
	Key   graphql.String `json:"key"`
	Value graphql.String `json:"value"`
}

// RuntimeConfigInput represents input for creating or updating runtime configuration.
type RuntimeConfigInput struct {
	Yaml          *graphql.String   `json:"yaml,omitempty"`
	Environment   *[]EnvVarInput    `json:"environment,omitempty"`
	ProjectRoot   *graphql.String   `json:"projectRoot,omitempty"`
	RunnerImage   *graphql.String   `json:"runnerImage,omitempty"`
	AfterApply    *[]graphql.String `json:"afterApply,omitempty"`
	AfterDestroy  *[]graphql.String `json:"afterDestroy,omitempty"`
	AfterInit     *[]graphql.String `json:"afterInit,omitempty"`
	AfterPerform  *[]graphql.String `json:"afterPerform,omitempty"`
	AfterPlan     *[]graphql.String `json:"afterPlan,omitempty"`
	AfterRun      *[]graphql.String `json:"afterRun,omitempty"`
	BeforeApply   *[]graphql.String `json:"beforeApply,omitempty"`
	BeforeDestroy *[]graphql.String `json:"beforeDestroy,omitempty"`
	BeforeInit    *[]graphql.String `json:"beforeInit,omitempty"`
	BeforePerform *[]graphql.String `json:"beforePerform,omitempty"`
	BeforePlan    *[]graphql.String `json:"beforePlan,omitempty"`
}

func ExportRuntimeConfigToMap(r *RuntimeConfig) (map[string]interface{}, diag.Diagnostics) {
	result := make(map[string]interface{})

	if r.ProjectRoot != "" {
		result["project_root"] = r.ProjectRoot
	}

	if r.RunnerImage != "" {
		result["runner_image"] = r.RunnerImage
	}

	if r.TerraformVersion != "" {
		result["terraform_version"] = r.TerraformVersion
	}

	if r.TerraformWorkflowTool != "" {
		result["terraform_workflow_tool"] = r.TerraformWorkflowTool
	}

	if len(r.Environment) > 0 {
		l := make([]map[string]interface{}, 0, len(r.Environment))
		for _, e := range r.Environment {
			l = append(l, map[string]interface{}{
				"key":   e.Key,
				"value": e.Value,
			})
		}
		result["environment"] = l
	}

	if len(r.AfterApply) > 0 {
		result["after_apply"] = r.AfterApply
	}

	if len(r.AfterDestroy) > 0 {
		result["after_destroy"] = r.AfterDestroy
	}

	if len(r.AfterInit) > 0 {
		result["after_init"] = r.AfterInit
	}

	if len(r.AfterPerform) > 0 {
		result["after_perform"] = r.AfterPerform
	}

	if len(r.AfterPlan) > 0 {
		result["after_plan"] = r.AfterPlan
	}

	if len(r.AfterRun) > 0 {
		result["after_run"] = r.AfterRun
	}

	if len(r.BeforeApply) > 0 {
		result["before_apply"] = r.BeforeApply
	}

	if len(r.BeforeDestroy) > 0 {
		result["before_destroy"] = r.BeforeDestroy
	}

	if len(r.BeforeInit) > 0 {
		result["before_init"] = r.BeforeInit
	}

	if len(r.BeforePerform) > 0 {
		result["before_perform"] = r.BeforePerform
	}

	if len(r.BeforePlan) > 0 {
		result["before_plan"] = r.BeforePlan
	}

	if r.Terragrunt != nil {
		result["terragrunt"] = r.Terragrunt
	}

	if r.Terraform != nil {
		result["terraform"] = r.Terraform
	}

	return result, nil
}
