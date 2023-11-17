package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataModule() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_module` is a special type of a stack used to test and " +
			"version Terraform modules.",

		ReadContext: dataModuleRead,

		Schema: map[string]*schema.Schema{
			"administrative": {
				Type:        schema.TypeBool,
				Description: "indicates whether this module can administer others",
				Computed:    true,
			},
			"aws_assume_role_policy_statement": {
				Type:        schema.TypeString,
				Description: "AWS IAM assume role policy statement setting up trust relationship",
				Computed:    true,
			},
			"branch": {
				Type:        schema.TypeString,
				Description: "GitHub branch to apply changes to",
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
			"description": {
				Type:        schema.TypeString,
				Description: "free-form module description for human users (supports Markdown)",
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
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the GitHub Enterprise application",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the GitHub Enterprise application",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Description of the GitHub Enterprise application",
							Computed:    true,
						},
						"is_default": {
							Type:        schema.TypeBool,
							Description: "Indicates whether this is the default GitHub Enterprise application",
							Computed:    true,
						},
						"labels": {
							Type:        schema.TypeSet,
							Description: "Labels of the GitHub Enterprise application",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
						"space_id": {
							Type:        schema.TypeString,
							Description: "ID (slug) of the space the GitHub Enterprise application is in",
							Computed:    true,
						},
					},
				},
			},
			"gitlab": {
				Type:        schema.TypeList,
				Description: "GitLab VCS settings",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the Gitlab integration",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the Gitlab integration",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Description of the Gitlab integration",
							Computed:    true,
						},
						"is_default": {
							Type:        schema.TypeBool,
							Description: "Indicates whether this is the default Gitlab integration",
							Computed:    true,
						},
						"labels": {
							Type:        schema.TypeSet,
							Description: "Labels of the Gitlab integration",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
						"space_id": {
							Type:        schema.TypeString,
							Description: "ID (slug) of the space the Gitlab integration is in",
							Computed:    true,
						},
						"namespace": {
							Type:        schema.TypeString,
							Description: "GitLab namespace of the repository",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"enable_local_preview": {
				Type:        schema.TypeBool,
				Description: "Indicates whether local preview versions can be triggered on this Module.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The module name will by default be inferred from the repository name if it follows the terraform-provider-name naming convention. However, if the repository doesn't follow this convention, or you want to give it a custom name, you can provide it here.",
				Computed:    true,
			},
			"project_root": {
				Type:        schema.TypeString,
				Description: "Project root is the optional directory relative to the repository root containing the module source code.",
				Computed:    true,
			},
			"protect_from_deletion": {
				Type:        schema.TypeBool,
				Description: "Protect this module from accidental deletion. If set, attempts to delete this module will fail.",
				Computed:    true,
			},
			"repository": {
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Computed:    true,
			},
			"module_id": {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the module",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"shared_accounts": {
				Type:        schema.TypeSet,
				Description: "List of the accounts (subdomains) which should have access to the Module",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the module is in",
				Computed:    true,
			},
			"terraform_provider": {
				Type:        schema.TypeString,
				Description: "The module provider will by default be inferred from the repository name if it follows the terraform-provider-name naming convention. However, if the repository doesn't follow this convention, or you gave the module a custom name, you can provide the provider name here.",
				Computed:    true,
			},
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to use",
				Computed:    true,
			},
			"workflow_tool": {
				Type:        schema.TypeString,
				Description: "Defines the tool that will be used to execute the workflow. This can be one of `OPEN_TOFU`, `TERRAFORM_FOSS` or `CUSTOM`.",
				Computed:    true,
			},
		},
	}
}

func dataModuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	moduleID := d.Get("module_id")
	variables := map[string]interface{}{"id": toID(moduleID)}
	if err := meta.(*internal.Client).Query(ctx, "ModuleRead", &query, variables); err != nil {
		return diag.Errorf("could not query for module: %v", err)
	}

	module := query.Module
	if module == nil {
		return diag.Errorf("module not found")
	}

	d.SetId(moduleID.(string))
	d.Set("administrative", module.Administrative)
	d.Set("aws_assume_role_policy_statement", module.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("branch", module.Branch)
	d.Set("name", module.Name)
	d.Set("enable_local_preview", module.LocalPreviewEnabled)
	d.Set("protect_from_deletion", module.ProtectFromDeletion)
	d.Set("terraform_provider", module.TerraformProvider)
	d.Set("space_id", module.Space)

	if err := module.ExportVCSSettings(d); err != nil {
		return diag.FromErr(err)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range module.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	sharedAccounts := schema.NewSet(schema.HashString, []interface{}{})
	for _, account := range module.SharedAccounts {
		sharedAccounts.Add(account)
	}
	d.Set("shared_accounts", sharedAccounts)

	d.Set("repository", module.Repository)

	if module.Description != nil {
		d.Set("description", *module.Description)
	} else {
		d.Set("description", nil)
	}

	if module.ProjectRoot != nil {
		d.Set("project_root", *module.ProjectRoot)
	} else {
		d.Set("project_root", nil)
	}

	if workerPool := module.WorkerPool; workerPool != nil {
		d.Set("worker_pool_id", workerPool.ID)
	} else {
		d.Set("worker_pool_id", nil)
	}

	if workflowTool := module.WorkflowTool; workflowTool != nil {
		d.Set("workflow_tool", *workflowTool)
	}

	return nil
}
