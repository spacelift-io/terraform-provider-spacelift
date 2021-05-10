package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataStack() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataStackRead,

		Schema: map[string]*schema.Schema{
			"administrative": {
				Type:        schema.TypeBool,
				Description: "indicates whether this stack can administer others",
				Computed:    true,
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
			"before_apply": {
				Type:        schema.TypeList,
				Description: "List of before-apply scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"before_init": {
				Type:        schema.TypeList,
				Description: "List of before-init scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
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
			"gitlab": {
				Type:        schema.TypeList,
				Description: "GitLab-related attributes",
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
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the stack",
				Required:    true,
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
	d.Set("autodeploy", stack.Autodeploy)
	d.Set("autoretry", stack.Autoretry)
	d.Set("aws_assume_role_policy_statement", stack.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("before_apply", stack.BeforeApply)
	d.Set("before_init", stack.BeforeInit)
	d.Set("branch", stack.Branch)
	d.Set("description", stack.Description)
	d.Set("enable_local_preview", stack.LocalPreviewEnabled)
	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("project_root", stack.ProjectRoot)
	d.Set("repository", stack.Repository)
	d.Set("runner_image", stack.RunnerImage)
	d.Set("terraform_version", stack.TerraformVersion)

	if stack.Provider == "GITLAB" {
		m := map[string]interface{}{
			"namespace": stack.Namespace,
		}

		if err := d.Set("gitlab", []interface{}{m}); err != nil {
			return diag.Errorf("error setting gitlab (resource): %v", err)
		}
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range stack.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	switch stack.VendorConfig.Typename {
	case structs.StackConfigVendorCloudFormation:
		m := map[string]interface{}{
			"entry_template_name": stack.VendorConfig.CloudFormation.EntryTemplateName,
			"region":              stack.VendorConfig.CloudFormation.Region,
			"stack_name":          stack.VendorConfig.CloudFormation.StackName,
			"template_bucket":     stack.VendorConfig.CloudFormation.TemplateBucket,
		}

		d.Set("cloudformation", []interface{}{m})
	case structs.StackConfigVendorPulumi:
		m := map[string]interface{}{
			"login_url":  stack.VendorConfig.Pulumi.LoginURL,
			"stack_name": stack.VendorConfig.Pulumi.StackName,
		}

		d.Set("pulumi", []interface{}{m})
	default:
		d.Set("terraform_version", stack.VendorConfig.Terraform.Version)
		d.Set("terraform_workspace", stack.VendorConfig.Terraform.Workspace)
	}

	if workerPool := stack.WorkerPool; workerPool != nil {
		d.Set("worker_pool_id", workerPool.ID)
	} else {
		d.Set("worker_pool_id", nil)
	}

	return nil
}
