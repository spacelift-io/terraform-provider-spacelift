package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataStack() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_stack` combines source code and configuration to create a " +
			"runtime environment where resources are managed. In this way it's " +
			"similar to a stack in AWS CloudFormation, or a project on generic " +
			"CI/CD platforms.",

		ReadContext: dataStackRead,

		Schema: map[string]*schema.Schema{
			"administrative": {
				Type:        schema.TypeBool,
				Description: "indicates whether this stack can administer others",
				Computed:    true,
			},
			"ansible": {
				Type:        schema.TypeList,
				Description: "Ansible-specific configuration. Presence means this Stack is an Ansible Stack.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"playbook": {
							Type:        schema.TypeString,
							Description: "The playbook the Ansible stack should run.",
							Computed:    true,
						},
					},
				},
			},
			"after_apply": {
				Type:        schema.TypeList,
				Description: "List of after-apply scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_destroy": {
				Type:        schema.TypeList,
				Description: "List of after-destroy scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_init": {
				Type:        schema.TypeList,
				Description: "List of after-init scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_perform": {
				Type:        schema.TypeList,
				Description: "List of after-perform scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_plan": {
				Type:        schema.TypeList,
				Description: "List of after-plan scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_run": {
				Type:        schema.TypeList,
				Description: "List of after-run scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"autodeploy": {
				Type:        schema.TypeBool,
				Description: "indicates whether changes to this stack can be automatically deployed",
				Computed:    true,
			},
			"autoretry": {
				Type:        schema.TypeBool,
				Description: "indicates whether obsolete proposed changes should automatically be retried",
				Computed:    true,
			},
			"aws_assume_role_policy_statement": {
				Type:        schema.TypeString,
				Description: "AWS IAM assume role policy statement setting up trust relationship",
				Computed:    true,
			},
			"azure_devops": {
				Type:        schema.TypeList,
				Description: "Azure DevOps VCS settings",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the Azure DevOps project",
						},
					},
				},
			},
			"before_apply": {
				Type:        schema.TypeList,
				Description: "List of before-apply scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"before_destroy": {
				Type:        schema.TypeList,
				Description: "List of before-destroy scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"before_init": {
				Type:        schema.TypeList,
				Description: "List of before-init scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"before_perform": {
				Type:        schema.TypeList,
				Description: "List of before-perform scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"before_plan": {
				Type:        schema.TypeList,
				Description: "List of before-plan scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"bitbucket_cloud": {
				Type:        schema.TypeList,
				Description: "Bitbucket Cloud VCS settings",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:        schema.TypeString,
							Description: "Bitbucket Cloud namespace of the stack's repository",
							Required:    true,
						},
					},
				},
			},
			"bitbucket_datacenter": {
				Type:        schema.TypeList,
				Description: "Bitbucket Datacenter VCS settings",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:        schema.TypeString,
							Description: "Bitbucket Datacenter namespace of the stack's repository",
							Required:    true,
						},
					},
				},
			},
			"branch": {
				Type:        schema.TypeString,
				Description: "Repository branch to treat as the default 'main' branch",
				Computed:    true,
			},
			"cloudformation": {
				Type:        schema.TypeList,
				Description: "CloudFormation-specific configuration. Presence means this Stack is a CloudFormation Stack.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry_template_file": {
							Type:        schema.TypeString,
							Description: "Template file `cloudformation package` will be called on",
							Computed:    true,
						},
						"region": {
							Type:        schema.TypeString,
							Description: "AWS region to use",
							Computed:    true,
						},
						"stack_name": {
							Type:        schema.TypeString,
							Description: "CloudFormation stack name",
							Computed:    true,
						},
						"template_bucket": {
							Type:        schema.TypeString,
							Description: "S3 bucket to save CloudFormation templates to",
							Computed:    true,
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "free-form stack description for users",
				Computed:    true,
			},
			"github_enterprise": {
				Type:        schema.TypeList,
				Description: "GitHub Enterprise (self-hosted) VCS settings",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:        schema.TypeString,
							Description: "GitHub Enterprise namespace of the stack's repository",
							Required:    true,
						},
					},
				},
			},
			"gitlab": {
				Type:        schema.TypeList,
				Description: "GitLab VCS settings",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:        schema.TypeString,
							Description: "GitLab namespace of the stack's repository",
							Computed:    true,
						},
					},
				},
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"enable_local_preview": {
				Type:        schema.TypeBool,
				Description: "Indicates whether local preview runs can be triggered on this Stack.",
				Computed:    true,
			},
			"kubernetes": {
				Type:        schema.TypeList,
				Description: "Kubernetes-specific configuration. Presence means this Stack is a Kubernetes Stack.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:        schema.TypeString,
							Description: "Namespace of the Kubernetes cluster to run commands on. Leave empty for multi-namespace Stacks.",
							Computed:    true,
						},
						"kubectl_version": {
							Type:        schema.TypeString,
							Description: "Kubectl version.",
							Computed:    true,
						},
					},
				},
			},
			"manage_state": {
				Type:        schema.TypeBool,
				Description: "Determines if Spacelift should manage state for this stack",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the stack - should be unique in one account",
				Computed:    true,
			},
			"project_root": {
				Type:        schema.TypeString,
				Description: "Project root is the optional directory relative to the workspace root containing the entrypoint to the Stack.",
				Computed:    true,
			},
			"protect_from_deletion": {
				Type:        schema.TypeBool,
				Description: "Protect this stack from accidental deletion. If set, attempts to delete this stack will fail.",
				Computed:    true,
			},
			"pulumi": {
				Type:        schema.TypeList,
				Description: "Pulumi-specific configuration. Presence means this Stack is a Pulumi Stack.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"login_url": {
							Type:        schema.TypeString,
							Description: "State backend to log into on Run initialize.",
							Computed:    true,
						},
						"stack_name": {
							Type:        schema.TypeString,
							Description: "Pulumi stack name to use with the state backend.",
							Computed:    true,
						},
					},
				},
			},
			"repository": {
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Computed:    true,
			},
			"runner_image": {
				Type:        schema.TypeString,
				Description: "Name of the Docker image used to process Runs",
				Computed:    true,
			},
			"showcase": {
				Type:        schema.TypeList,
				Description: "Showcase-related attributes",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:        schema.TypeString,
							Description: "GitHub namespace of the stack's repository",
							Computed:    true,
						},
					},
				},
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the stack is in",
				Computed:    true,
			},
			"stack_id": {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the stack",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"terraform_smart_sanitization": {
				Type:        schema.TypeBool,
				Description: "Indicates whether runs on this will use terraform's sensitive value system to sanitize the outputs of Terraform state and plans in spacelift instead of sanitizing all fields.",
				Computed:    true,
			},
			"terraform_version": {
				Type:        schema.TypeString,
				Description: "Terraform version to use",
				Computed:    true,
			},
			"terraform_workspace": {
				Type:        schema.TypeString,
				Description: "Terraform workspace to select",
				Computed:    true,
			},
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to use",
				Computed:    true,
			},
		},
	}
}

func dataStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id")
	variables := map[string]interface{}{"id": toID(stackID)}
	if err := meta.(*internal.Client).Query(ctx, "StackRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	stack := query.Stack
	if stack == nil {
		return diag.Errorf("stack not found")
	}

	d.SetId(stackID.(string))
	d.Set("administrative", stack.Administrative)
	d.Set("after_apply", stack.AfterApply)
	d.Set("after_destroy", stack.AfterDestroy)
	d.Set("after_init", stack.AfterInit)
	d.Set("after_perform", stack.AfterPerform)
	d.Set("after_plan", stack.AfterPlan)
	d.Set("after_run", stack.AfterRun)
	d.Set("autodeploy", stack.Autodeploy)
	d.Set("autoretry", stack.Autoretry)
	d.Set("aws_assume_role_policy_statement", stack.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("before_apply", stack.BeforeApply)
	d.Set("before_destroy", stack.BeforeDestroy)
	d.Set("before_init", stack.BeforeInit)
	d.Set("before_perform", stack.BeforePerform)
	d.Set("before_plan", stack.BeforePlan)
	d.Set("branch", stack.Branch)
	d.Set("description", stack.Description)
	d.Set("enable_local_preview", stack.LocalPreviewEnabled)
	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("project_root", stack.ProjectRoot)
	d.Set("protect_from_deletion", stack.ProtectFromDeletion)
	d.Set("repository", stack.Repository)
	d.Set("runner_image", stack.RunnerImage)
	d.Set("terraform_version", stack.TerraformVersion)
	d.Set("space_id", stack.Space)

	if err := stack.ExportVCSSettings(d); err != nil {
		return diag.FromErr(err)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range stack.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if iacKey, iacSettings := stack.IaCSettings(); iacKey != "" {
		if err := d.Set(iacKey, []interface{}{iacSettings}); err != nil {
			return diag.Errorf("could not set IaC settings: %v", err)
		}
	} else { // this is a Terraform stack
		d.Set("terraform_version", stack.VendorConfig.Terraform.Version)
		d.Set("terraform_workspace", stack.VendorConfig.Terraform.Workspace)
		d.Set("terraform_smart_sanitization", stack.VendorConfig.Terraform.UseSmartSanitization)
	}

	if workerPool := stack.WorkerPool; workerPool != nil {
		d.Set("worker_pool_id", workerPool.ID)
	} else {
		d.Set("worker_pool_id", nil)
	}

	return nil
}
