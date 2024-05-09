package spacelift

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceStack() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_stack` combines source code and configuration to create a " +
			"runtime environment where resources are managed. In this way it's " +
			"similar to a stack in AWS CloudFormation, or a project on generic " +
			"CI/CD platforms.",

		CreateContext: resourceStackCreate,
		ReadContext:   resourceStackRead,
		UpdateContext: resourceStackUpdate,
		DeleteContext: resourceStackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceStackImport,
		},

		Schema: map[string]*schema.Schema{
			"administrative": {
				Type:        schema.TypeBool,
				Description: "Indicates whether this stack can manage others. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"ansible": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"cloudformation", "kubernetes", "pulumi", "terraform_version", "terraform_workflow_tool", "terraform_workspace", "terragrunt"},
				Description:   "Ansible-specific configuration. Presence means this Stack is an Ansible Stack.",
				Optional:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"playbook": {
							Type:             schema.TypeString,
							Description:      "The playbook Ansible should run.",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"after_apply": {
				Type:        schema.TypeList,
				Description: "List of after-apply scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_destroy": {
				Type:        schema.TypeList,
				Description: "List of after-destroy scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_init": {
				Type:        schema.TypeList,
				Description: "List of after-init scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_perform": {
				Type:        schema.TypeList,
				Description: "List of after-perform scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_plan": {
				Type:        schema.TypeList,
				Description: "List of after-plan scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_run": {
				Type:        schema.TypeList,
				Description: "List of after-run scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"autodeploy": {
				Type:        schema.TypeBool,
				Description: "Indicates whether changes to this stack can be automatically deployed. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"autoretry": {
				Type:        schema.TypeBool,
				Description: "Indicates whether obsolete proposed changes should automatically be retried. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"aws_assume_role_policy_statement": {
				Type:        schema.TypeString,
				Description: "AWS IAM assume role policy statement setting up trust relationship",
				Computed:    true,
			},
			"azure_devops": {
				Type:          schema.TypeList,
				Description:   "Azure DevOps VCS settings",
				Optional:      true,
				ConflictsWith: conflictingVCSProviders("azure_devops"),
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the Azure Devops integration. If not specified, the default integration will be used.",
							DiffSuppressFunc: func(_, _, new string, res *schema.ResourceData) bool {
								isDefault := res.Get("azure_devops.0.is_default").(bool)

								return isDefault && new == ""
							},
						},
						"project": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The name of the Azure DevOps project",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether this is the default Azure DevOps integration",
						},
					},
				},
			},
			"before_apply": {
				Type:        schema.TypeList,
				Description: "List of before-apply scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_destroy": {
				Type:        schema.TypeList,
				Description: "List of before-destroy scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_init": {
				Type:        schema.TypeList,
				Description: "List of before-init scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_perform": {
				Type:        schema.TypeList,
				Description: "List of before-perform scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_plan": {
				Type:        schema.TypeList,
				Description: "List of before-plan scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"branch": {
				Type:             schema.TypeString,
				Description:      "Git branch to apply changes to",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"bitbucket_cloud": {
				Type:          schema.TypeList,
				Description:   "Bitbucket Cloud VCS settings",
				Optional:      true,
				ConflictsWith: conflictingVCSProviders("bitbucket_cloud"),
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the Bitbucket Cloud integration. If not specified, the default integration will be used.",
							DiffSuppressFunc: func(_, _, new string, res *schema.ResourceData) bool {
								isDefault := res.Get("bitbucket_cloud.0.is_default").(bool)

								return isDefault && new == ""
							},
						},
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The Bitbucket project containing the repository",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether this is the default Bitbucket Cloud integration",
						},
					},
				},
			},
			"bitbucket_datacenter": {
				Type:          schema.TypeList,
				Description:   "Bitbucket Datacenter VCS settings",
				Optional:      true,
				ConflictsWith: conflictingVCSProviders("bitbucket_datacenter"),
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the Bitbucket Datacenter integration. If not specified, the default integration will be used.",
							DiffSuppressFunc: func(_, _, new string, res *schema.ResourceData) bool {
								isDefault := res.Get("bitbucket_datacenter.0.is_default").(bool)

								return isDefault && new == ""
							},
						},
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The Bitbucket project containing the repository",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether this is the default Bitbucket Datacenter integration",
						},
					},
				},
			},
			"cloudformation": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"ansible", "kubernetes", "pulumi", "terraform_version", "terraform_workflow_tool", "terraform_workspace", "terragrunt"},
				Description:   "CloudFormation-specific configuration. Presence means this Stack is a CloudFormation Stack.",
				Optional:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry_template_file": {
							Type:             schema.TypeString,
							Description:      "Template file `cloudformation package` will be called on",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"region": {
							Type:             schema.TypeString,
							Description:      "AWS region to use",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"stack_name": {
							Type:             schema.TypeString,
							Description:      "CloudFormation stack name",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"template_bucket": {
							Type:             schema.TypeString,
							Description:      "S3 bucket to save CloudFormation templates to",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form stack description for users",
				Optional:    true,
			},
			"enable_local_preview": {
				Type:        schema.TypeBool,
				Description: "Indicates whether local preview runs can be triggered on this Stack. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"enable_well_known_secret_masking": {
				Type:        schema.TypeBool,
				Description: "Indicates whether well-known secret masking is enabled.",
				Optional:    true,
				Default:     false,
			},
			"github_action_deploy": {
				Type:        schema.TypeBool,
				Description: "Indicates whether GitHub users can deploy from the Checks API. Defaults to `true`. This is called allow run promotion in the UI.",
				Optional:    true,
				Default:     true,
			},
			"github_enterprise": {
				Type:          schema.TypeList,
				Description:   "VCS settings for [GitHub custom application](https://docs.spacelift.io/integrations/source-control/github#setting-up-the-custom-application)",
				Optional:      true,
				ConflictsWith: conflictingVCSProviders("github_enterprise"),
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The GitHub organization / user the repository belongs to",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the GitHub Enterprise integration. If not specified, the default integration will be used.",
							DiffSuppressFunc: func(_, _, new string, res *schema.ResourceData) bool {
								isDefault := res.Get("github_enterprise.0.is_default").(bool)

								return isDefault && new == ""
							},
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether this is the default GitHub Enterprise integration",
						},
					},
				},
			},
			"gitlab": {
				Type:          schema.TypeList,
				Description:   "GitLab VCS settings",
				Optional:      true,
				ConflictsWith: conflictingVCSProviders("gitlab"),
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the Gitlab integration. If not specified, the default integration will be used.",
							DiffSuppressFunc: func(_, _, new string, res *schema.ResourceData) bool {
								isDefault := res.Get("gitlab.0.is_default").(bool)

								return isDefault && new == ""
							},
						},
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The GitLab namespace containing the repository",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether this is the default GitLab integration",
						},
					},
				},
			},
			"import_state": {
				Type:             schema.TypeString,
				Description:      "State file to upload when creating a new stack",
				ConflictsWith:    []string{"import_state_file"},
				Optional:         true,
				DiffSuppressFunc: ignoreOnceCreated,
				Sensitive:        true,
			},
			"import_state_file": {
				Type:             schema.TypeString,
				Description:      "Path to the state file to upload when creating a new stack",
				ConflictsWith:    []string{"import_state"},
				Optional:         true,
				DiffSuppressFunc: ignoreOnceCreated,
			},
			"kubernetes": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"ansible", "cloudformation", "pulumi", "terraform_version", "terraform_workflow_tool", "terraform_workspace", "terragrunt"},
				Description:   "Kubernetes-specific configuration. Presence means this Stack is a Kubernetes Stack.",
				Optional:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:             schema.TypeString,
							Description:      "Namespace of the Kubernetes cluster to run commands on. Leave empty for multi-namespace Stacks.",
							Optional:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"kubectl_version": {
							Type:             schema.TypeString,
							Description:      "Kubectl version.",
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"manage_state": {
				Type:        schema.TypeBool,
				Description: "Determines if Spacelift should manage state for this stack. Defaults to `true`.",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the stack - should be unique in one account",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"project_root": {
				Type:        schema.TypeString,
				Description: "Project root is the optional directory relative to the workspace root containing the entrypoint to the Stack.",
				Optional:    true,
			},
			"additional_project_globs": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Project globs is an optional list of paths to track changes of in addition to the project root.",
			},
			"protect_from_deletion": {
				Type:        schema.TypeBool,
				Description: "Protect this stack from accidental deletion. If set, attempts to delete this stack will fail. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"pulumi": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"ansible", "cloudformation", "kubernetes", "terraform_version", "terraform_workflow_tool", "terraform_workspace", "terragrunt"},
				Description:   "Pulumi-specific configuration. Presence means this Stack is a Pulumi Stack.",
				Optional:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"login_url": {
							Type:             schema.TypeString,
							Description:      "State backend to log into on Run initialize.",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"stack_name": {
							Type:             schema.TypeString,
							Description:      "Pulumi stack name to use with the state backend.",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "Allows setting the custom ID (slug) for the stack",
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"raw_git": {
				Type:          schema.TypeList,
				Description:   "One-way VCS integration using a raw Git repository link",
				Optional:      true,
				ConflictsWith: conflictingVCSProviders("raw_git"),
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "User-friendly namespace for the repository, this is for cosmetic purposes only",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"url": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "HTTPS URL of the Git repository",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"repository": {
				Type:             schema.TypeString,
				Description:      "Name of the repository, without the owner part",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"runner_image": {
				Type:        schema.TypeString,
				Description: "Name of the Docker image used to process Runs",
				Optional:    true,
			},
			"showcase": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the stack is in. Defaults to `legacy` if it exists, otherwise `root`.",
				Optional:    true,
				Computed:    true,
			},
			"terraform_external_state_access": {
				Type:        schema.TypeBool,
				Description: "Indicates whether you can access the Stack state file from other stacks or outside of Spacelift. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"terraform_smart_sanitization": {
				Type:        schema.TypeBool,
				Description: "Indicates whether runs on this will use terraform's sensitive value system to sanitize the outputs of Terraform state and plans in spacelift instead of sanitizing all fields. Note: Requires the terraform version to be v1.0.1 or above. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"terraform_version": {
				Type:             schema.TypeString,
				Description:      "Terraform version to use",
				Optional:         true,
				DiffSuppressFunc: onceTheVersionIsSetDoNotUnset,
			},
			"terraform_workflow_tool": {
				Type:        schema.TypeString,
				Description: "Defines the tool that will be used to execute the workflow. This can be one of `OPEN_TOFU`, `TERRAFORM_FOSS` or `CUSTOM`. Defaults to `TERRAFORM_FOSS`.",
				Optional:    true,
				Computed:    true,
			},
			"terraform_workspace": {
				Type:        schema.TypeString,
				Description: "Terraform workspace to select",
				Optional:    true,
			},
			"terragrunt": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"ansible", "cloudformation", "kubernetes", "pulumi", "terraform_version", "terraform_workflow_tool", "terraform_workspace"},
				Description:   "Terragrunt-specific configuration. Presence means this Stack is an Terragrunt Stack.",
				Optional:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					/*CustomizeDiff: customdiff.All(
						customdiff.ComputedIf("terraform_version", func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) bool {
							f, err := os.OpenFile("/Users/ptru/fabryka/spacelift/terraform-provider-spacelift-extras/test-terragrunt/terragrunt.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
							if err != nil {
								panic(err)
							}
							defer f.Close()
							fmt.Fprintf(f, "checked! %v\n", diff.HasChange("tool"))
							return diff.HasChange("tool")
						}),
					),*/
					Schema: map[string]*schema.Schema{
						"terraform_version": {
							Type:             schema.TypeString,
							Description:      "The Terraform version. Must not be provided when tool is set to MANUALLY_PROVISIONED. Defaults to the latest available OpenTofu/Terraform version.",
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"terragrunt_version": {
							Type:             schema.TypeString,
							Description:      "The Terragrunt version. Defaults to the latest Terragrunt version.",
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"use_run_all": {
							Type:        schema.TypeBool,
							Description: "Whether to use `terragrunt run-all` instead of `terragrunt`.",
							Optional:    true,
							Default:     false,
						},
						"use_smart_sanitization": {
							Type:        schema.TypeBool,
							Description: "Indicates whether runs on this will use Terraform's sensitive value system to sanitize the outputs of Terraform state and plans in spacelift instead of sanitizing all fields.",
							Optional:    true,
							Default:     false,
						},
						"tool": {
							Type:        schema.TypeString,
							Description: "The IaC tool used by Terragrunt. Valid values are OPEN_TOFU, TERRAFORM_FOSS or MANUALLY_PROVISIONED. Defaults to TERRAFORM_FOSS if not specified.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to use. NOTE: worker_pool_id is required when using a self-hosted instance of Spacelift.",
				Optional:    true,
			},
		},
	}
}

func resourceStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateStack structs.Stack `graphql:"stackCreate(input: $input, manageState: $manageState, stackObjectID: $stackObjectID, slug: $slug)"`
	}

	manageState := d.Get("manage_state").(bool)

	variables := map[string]interface{}{
		"input":         stackInput(d),
		"manageState":   graphql.Boolean(manageState),
		"stackObjectID": (*graphql.String)(nil),
		"slug":          (*graphql.String)(nil),
	}

	if slug, ok := d.GetOk("slug"); ok {
		variables["slug"] = toOptionalString(slug)
	}

	var stateContent string

	content, ok := d.GetOk("import_state")
	if ok && !manageState {
		return diag.Errorf(`"import_state" requires "manage_state" to be true`)
	} else if ok {
		stateContent = content.(string)
	}

	path, ok := d.GetOk("import_state_file")
	if ok && !manageState {
		return diag.Errorf(`"import_state_file" requires "manage_state" to be true`)
	} else if ok {
		data, err := os.ReadFile(path.(string))
		if err != nil {
			return diag.Errorf("failed to read imported state file: %s", err)
		}
		stateContent = string(data)
	}

	if stateContent != "" {
		objectID, err := uploadStateFile(ctx, stateContent, meta)
		if err != nil {
			return diag.FromErr(err)
		}
		variables["stackObjectID"] = toOptionalString(objectID)
	}

	if v, ok := d.GetOk("terraform_external_state_access"); ok {
		if v.(bool) && !manageState {
			return diag.Errorf(`"terraform_external_state_access" requires "manage_state" to be true`)
		}
	}

	if err := meta.(*internal.Client).Mutate(ctx, "StackCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create stack: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateStack.ID)

	return resourceStackRead(ctx, d, meta)
}

func getStackByID(ctx context.Context, client *internal.Client, stackID string) (*structs.Stack, error) {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(stackID)}

	if err := client.Query(ctx, "StackRead", &query, variables); err != nil {
		return nil, errors.Wrap(err, "could not query for stack")
	}

	return query.Stack, nil
}

func resourceStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	stack, err := getStackByID(ctx, meta.(*internal.Client), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if stack == nil {
		d.SetId("")
		return nil
	}

	if err := structs.PopulateStack(d, stack); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateStack structs.Stack `graphql:"stackUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": stackInput(d),
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "StackUpdate", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update stack: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourceStackRead(ctx, d, meta)...)
}

func resourceStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteStack *structs.Stack `graphql:"stackDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete stack: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func stackInput(d *schema.ResourceData) structs.StackInput {
	ret := structs.StackInput{
		Administrative:               graphql.Boolean(d.Get("administrative").(bool)),
		Autodeploy:                   graphql.Boolean(d.Get("autodeploy").(bool)),
		Autoretry:                    graphql.Boolean(d.Get("autoretry").(bool)),
		Branch:                       toString(d.Get("branch")),
		GitHubActionDeploy:           graphql.Boolean(d.Get("github_action_deploy").(bool)),
		LocalPreviewEnabled:          graphql.Boolean(d.Get("enable_local_preview").(bool)),
		EnableWellKnownSecretMasking: graphql.Boolean(d.Get("enable_well_known_secret_masking").(bool)),
		Name:                         toString(d.Get("name")),
		ProtectFromDeletion:          graphql.Boolean(d.Get("protect_from_deletion").(bool)),
		Repository:                   toString(d.Get("repository")),
	}

	afterApplies := getStrings(d, "after_apply")
	ret.AfterApply = &afterApplies

	afterDestroys := getStrings(d, "after_destroy")
	ret.AfterDestroy = &afterDestroys

	afterInits := getStrings(d, "after_init")
	ret.AfterInit = &afterInits

	afterPerforms := getStrings(d, "after_perform")
	ret.AfterPerform = &afterPerforms

	afterPlans := getStrings(d, "after_plan")
	ret.AfterPlan = &afterPlans

	afterRuns := getStrings(d, "after_run")
	ret.AfterRun = &afterRuns

	beforeApplies := getStrings(d, "before_apply")
	ret.BeforeApply = &beforeApplies

	beforeDestroys := getStrings(d, "before_destroy")
	ret.BeforeDestroy = &beforeDestroys

	beforeInits := getStrings(d, "before_init")
	ret.BeforeInit = &beforeInits

	beforePerforms := getStrings(d, "before_perform")
	ret.BeforePerform = &beforePerforms

	beforePlans := getStrings(d, "before_plan")
	ret.BeforePlan = &beforePlans

	description, ok := d.GetOk("description")
	if ok {
		ret.Description = toOptionalString(description)
	}

	ret.Provider = graphql.NewString("GITHUB")

	if azureDevOps, ok := d.Get("azure_devops").([]interface{}); ok && len(azureDevOps) > 0 {
		azureSettings := azureDevOps[0].(map[string]interface{})
		if id, ok := azureSettings["id"]; ok && id != nil && id.(string) != "" {
			ret.VCSIntegrationID = graphql.NewID(id.(string))
		}
		ret.Namespace = toOptionalString(azureDevOps[0].(map[string]interface{})["project"])
		ret.Provider = graphql.NewString(graphql.String(structs.VCSProviderAzureDevOps))
	}

	if bitbucketCloud, ok := d.Get("bitbucket_cloud").([]interface{}); ok && len(bitbucketCloud) > 0 {
		bitbucketCloudSettings := bitbucketCloud[0].(map[string]interface{})
		if id, ok := bitbucketCloudSettings["id"]; ok && id != nil && id.(string) != "" {
			ret.VCSIntegrationID = graphql.NewID(id.(string))
		}
		ret.Namespace = toOptionalString(bitbucketCloud[0].(map[string]interface{})["namespace"])
		ret.Provider = graphql.NewString(graphql.String(structs.VCSProviderBitbucketCloud))
	}

	if bitbucketDatacenter, ok := d.Get("bitbucket_datacenter").([]interface{}); ok && len(bitbucketDatacenter) > 0 {
		bitbucketDatacenterSettings := bitbucketDatacenter[0].(map[string]interface{})
		if id, ok := bitbucketDatacenterSettings["id"]; ok && id != nil && id.(string) != "" {
			ret.VCSIntegrationID = graphql.NewID(id.(string))
		}
		ret.Namespace = toOptionalString(bitbucketDatacenter[0].(map[string]interface{})["namespace"])
		ret.Provider = graphql.NewString(graphql.String(structs.VCSProviderBitbucketDatacenter))
	}

	if githubEnterprise, ok := d.Get("github_enterprise").([]interface{}); ok && len(githubEnterprise) > 0 {
		ghEnterpriseSettings := githubEnterprise[0].(map[string]interface{})
		if id, ok := ghEnterpriseSettings["id"]; ok && id != nil && id.(string) != "" {
			ret.VCSIntegrationID = graphql.NewID(id)
		}
		ret.Namespace = toOptionalString(ghEnterpriseSettings["namespace"])
		ret.Provider = graphql.NewString(graphql.String(structs.VCSProviderGitHubEnterprise))
	}

	if gitlab, ok := d.Get("gitlab").([]interface{}); ok && len(gitlab) > 0 {
		gitlabSettings := gitlab[0].(map[string]interface{})
		if id, ok := gitlabSettings["id"]; ok && id != nil && id.(string) != "" {
			ret.VCSIntegrationID = graphql.NewID(id.(string))
		}
		ret.Namespace = toOptionalString(gitlabSettings["namespace"])
		ret.Provider = graphql.NewString(graphql.String(structs.VCSProviderGitlab))
	}

	if rawGit, ok := d.Get("raw_git").([]interface{}); ok && len(rawGit) > 0 {
		ret.Provider = graphql.NewString(graphql.String(structs.VCSProviderRawGit))
		ret.Namespace = toOptionalString(rawGit[0].(map[string]interface{})["namespace"])
		ret.RepositoryURL = toOptionalString(rawGit[0].(map[string]interface{})["url"])
	}

	if showcase, ok := d.Get("showcase").([]interface{}); ok && len(showcase) > 0 {
		ret.Namespace = toOptionalString(showcase[0].(map[string]interface{})["namespace"])
		ret.Provider = graphql.NewString(graphql.String(structs.VCSProviderShowcases))
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		ret.Labels = &labels
	}

	if space, ok := d.GetOk("space_id"); ok {
		ret.Space = toOptionalString(space)
	}

	if projectRoot, ok := d.GetOk("project_root"); ok {
		ret.ProjectRoot = toOptionalString(projectRoot)
	}

	if globsSet, ok := d.Get("additional_project_globs").(*schema.Set); ok {
		var gbs []graphql.String
		for _, gb := range globsSet.List() {
			gbs = append(gbs, graphql.String(gb.(string)))
		}
		ret.AddditionalProjectGlobs = &gbs
	}

	if runnerImage, ok := d.GetOk("runner_image"); ok {
		ret.RunnerImage = toOptionalString(runnerImage)
	}

	ret.VendorConfig = getVendorConfig(d)

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.WorkerPool = graphql.NewID(workerPoolID)
	}

	return ret
}

func getVendorConfig(d *schema.ResourceData) *structs.VendorConfigInput {
	if cloudFormation, ok := d.Get("cloudformation").([]interface{}); ok && len(cloudFormation) > 0 {
		return &structs.VendorConfigInput{
			CloudFormationInput: &structs.CloudFormationInput{
				EntryTemplateFile: toString(cloudFormation[0].(map[string]interface{})["entry_template_file"]),
				Region:            toString(cloudFormation[0].(map[string]interface{})["region"]),
				StackName:         toString(cloudFormation[0].(map[string]interface{})["stack_name"]),
				TemplateBucket:    toString(cloudFormation[0].(map[string]interface{})["template_bucket"]),
			},
		}
	}

	if kubernetes, ok := d.Get("kubernetes").([]interface{}); ok && len(kubernetes) > 0 {
		vendorConfig := &structs.VendorConfigInput{
			Kubernetes: &structs.KubernetesInput{},
		}

		if kubernetesSettings, ok := kubernetes[0].(map[string]interface{}); ok {
			vendorConfig.Kubernetes.Namespace = toString(kubernetesSettings["namespace"])
			if s := toOptionalString(kubernetesSettings["kubectl_version"]); *s != "" {
				vendorConfig.Kubernetes.KubectlVersion = s
			}
		}
		return vendorConfig
	}

	if pulumi, ok := d.Get("pulumi").([]interface{}); ok && len(pulumi) > 0 {
		return &structs.VendorConfigInput{
			Pulumi: &structs.PulumiInput{
				LoginURL:  toString(pulumi[0].(map[string]interface{})["login_url"]),
				StackName: toString(pulumi[0].(map[string]interface{})["stack_name"]),
			},
		}
	}

	if ansible, ok := d.Get("ansible").([]interface{}); ok && len(ansible) > 0 {
		return &structs.VendorConfigInput{
			AnsibleInput: &structs.AnsibleInput{
				Playbook: toString(ansible[0].(map[string]interface{})["playbook"]),
			},
		}
	}

	if terragrunt, ok := d.Get("terragrunt").([]interface{}); ok && len(terragrunt) > 0 {
		terragruntConfig := structs.TerragruntInput{
			UseRunAll:            toBool(terragrunt[0].(map[string]interface{})["use_run_all"]),
			UseSmartSanitization: toBool(terragrunt[0].(map[string]interface{})["use_smart_sanitization"]),
		}

		if version, ok := terragrunt[0].(map[string]interface{})["terraform_version"]; ok && version.(string) != "" {
			terragruntConfig.TerraformVersion = toOptionalString(version)
		}

		if version, ok := terragrunt[0].(map[string]interface{})["terragrunt_version"]; ok && version.(string) != "" {
			terragruntConfig.TerragruntVersion = toOptionalString(version)
		}

		if tool, ok := terragrunt[0].(map[string]interface{})["tool"]; ok && tool.(string) != "" {
			terragruntConfig.Tool = toOptionalString(tool)
		}

		if shouldWeReComputeTerraformVersion(d) {
			terragruntConfig.TerraformVersion = nil
		}

		return &structs.VendorConfigInput{
			TerragruntInput: &terragruntConfig,
		}
	}

	terraformConfig := &structs.TerraformInput{}

	if terraformVersion, ok := d.GetOk("terraform_version"); ok {
		terraformConfig.Version = toOptionalString(terraformVersion)
	}

	if terraformWorkflowTool, ok := d.GetOk("terraform_workflow_tool"); ok {
		terraformConfig.WorkflowTool = toOptionalString(terraformWorkflowTool)
	}

	if terraformWorkspace, ok := d.GetOk("terraform_workspace"); ok {
		terraformConfig.Workspace = toOptionalString(terraformWorkspace)
	}

	if terraformSmartSanitization, ok := d.GetOk("terraform_smart_sanitization"); ok {
		terraformConfig.UseSmartSanitization = toOptionalBool(terraformSmartSanitization)
	} else {
		terraformConfig.UseSmartSanitization = toOptionalBool(false)
	}

	if v, ok := d.GetOk("terraform_external_state_access"); ok {
		terraformConfig.ExternalStateAccessEnabled = toOptionalBool(v)
	} else {
		terraformConfig.ExternalStateAccessEnabled = toOptionalBool(false)
	}

	return &structs.VendorConfigInput{Terraform: terraformConfig}
}

func shouldWeReComputeTerraformVersion(d *schema.ResourceData) bool {
	// When tool is changed, we need to recompute terraform version
	oldTool, newTool := d.GetChange("terragrunt.0.tool")
	if oldTool.(string) != newTool.(string) {
		// but only if version isn't provided manually in the config
		inConf := d.GetRawConfig().AsValueMap()["terragrunt"].AsValueSlice()[0].AsValueMap()
		if value, ok := inConf["terraform_version"]; ok {
			if value.IsNull() || value.AsString() == "" {
				return true
			}
		}
	}

	return false
}

func getStrings(d *schema.ResourceData, fieldName string) []graphql.String {
	values := []graphql.String{}
	if commands, ok := d.GetOk(fieldName); ok {
		for _, cmd := range commands.([]interface{}) {
			values = append(values, graphql.String(cmd.(string)))
		}
	}
	return values
}

func uploadStateFile(ctx context.Context, content string, meta interface{}) (string, error) {
	var mutation struct {
		StateUploadURL struct {
			ObjectID string `graphql:"objectId"`
			URL      string `graphql:"url"`
		} `graphql:"stateUploadUrl"`
	}

	if err := meta.(*internal.Client).Mutate(ctx, "StateUploadUrl", &mutation, nil); err != nil {
		return "", errors.Wrap(err, "could not generate state upload URL")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, mutation.StateUploadURL.URL, strings.NewReader(content))
	if err != nil {
		return "", errors.Wrap(err, "could not create state upload request")
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", errors.Wrap(err, "could not upload the state to remote URL")
	}

	if (response.StatusCode / 100) != 2 {
		return "", errors.Errorf("unexpected HTTP status code when uploading the state: %d", response.StatusCode)
	}

	return mutation.StateUploadURL.ObjectID, nil
}

func onceTheVersionIsSetDoNotUnset(_, _, new string, _ *schema.ResourceData) bool {
	return new == ""
}

func ignoreOnceCreated(_, _, _ string, d *schema.ResourceData) bool {
	return d.Id() != ""
}

func resourceStackImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	stackID := d.Id()
	if stackID == "" {
		return nil, errors.New("stack ID is required to import a stack")
	}

	stack, err := getStackByID(ctx, meta.(*internal.Client), stackID)
	if err != nil {
		return nil, fmt.Errorf("could not query for stack with ID %q: %v", stackID, err)
	}

	if stack == nil {
		return nil, fmt.Errorf("stack with ID %q does not exist (or you may not have access to it)", stackID)
	}

	if err := structs.PopulateStack(d, stack); err != nil {
		return nil, errors.Wrap(err, "could not import stack into state")
	}

	return []*schema.ResourceData{d}, nil
}

func conflictingVCSProviders(me string) (out []string) {
	available := []string{
		"azure_devops",
		"bitbucket_cloud",
		"bitbucket_datacenter",
		"github_enterprise",
		"gitlab",
		"raw_git",
	}

	for _, v := range available {
		if v != me {
			out = append(out, v)
		}
	}

	return
}
