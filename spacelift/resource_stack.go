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
				ConflictsWith: []string{"cloudformation", "kubernetes", "pulumi", "terraform_version", "terraform_workspace"},
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
				ConflictsWith: []string{"bitbucket_cloud", "bitbucket_datacenter", "github_enterprise", "gitlab"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The name of the Azure DevOps project",
							ValidateDiagFunc: validations.DisallowEmptyString,
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
				Description:      "GitHub branch to apply changes to",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"bitbucket_cloud": {
				Type:          schema.TypeList,
				Description:   "Bitbucket Cloud VCS settings",
				Optional:      true,
				ConflictsWith: []string{"azure_devops", "bitbucket_datacenter", "github_enterprise", "gitlab"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The Bitbucket project containing the repository",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"bitbucket_datacenter": {
				Type:          schema.TypeList,
				Description:   "Bitbucket Datacenter VCS settings",
				Optional:      true,
				ConflictsWith: []string{"azure_devops", "bitbucket_cloud", "github_enterprise", "gitlab"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The Bitbucket project containing the repository",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"cloudformation": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"ansible", "kubernetes", "pulumi", "terraform_version", "terraform_workspace"},
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
				ConflictsWith: []string{"azure_devops", "bitbucket_cloud", "bitbucket_datacenter", "gitlab"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The GitHub organization / user the repository belongs to",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"gitlab": {
				Type:          schema.TypeList,
				Description:   "GitLab VCS settings",
				Optional:      true,
				ConflictsWith: []string{"azure_devops", "bitbucket_cloud", "bitbucket_datacenter", "github_enterprise"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The GitLab namespace containing the repository",
							ValidateDiagFunc: validations.DisallowEmptyString,
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
				ConflictsWith: []string{"ansible", "cloudformation", "pulumi", "terraform_version", "terraform_workspace"},
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
			"protect_from_deletion": {
				Type:        schema.TypeBool,
				Description: "Protect this stack from accidental deletion. If set, attempts to delete this stack will fail. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"pulumi": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"ansible", "cloudformation", "kubernetes", "terraform_version", "terraform_workspace"},
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
				Description: "ID (slug) of the space the stack is in",
				Optional:    true,
				Computed:    true,
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
			"terraform_workspace": {
				Type:        schema.TypeString,
				Description: "Terraform workspace to select",
				Optional:    true,
			},
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to use",
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

	if err := meta.(*internal.Client).Mutate(ctx, "StackCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create stack: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateStack.ID)

	return resourceStackRead(ctx, d, meta)
}

func resourceStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, "StackRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	stack := query.Stack
	if stack == nil {
		d.SetId("")
		return nil
	}

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
	d.Set("github_action_deploy", stack.GitHubActionDeploy)
	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("project_root", stack.ProjectRoot)
	d.Set("protect_from_deletion", stack.ProtectFromDeletion)
	d.Set("repository", stack.Repository)
	d.Set("runner_image", stack.RunnerImage)
	d.Set("space_id", stack.Space)
	d.Set("slug", stack.ID)

	if err := stack.ExportVCSSettings(d); err != nil {
		return diag.FromErr(err)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range stack.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	switch stack.VendorConfig.Typename {
	case structs.StackConfigVendorAnsible:
		m := map[string]interface{}{
			"playbook": stack.VendorConfig.Ansible.Playbook,
		}

		d.Set("ansible", []interface{}{m})
	case structs.StackConfigVendorCloudFormation:
		m := map[string]interface{}{
			"entry_template_file": stack.VendorConfig.CloudFormation.EntryTemplateName,
			"region":              stack.VendorConfig.CloudFormation.Region,
			"stack_name":          stack.VendorConfig.CloudFormation.StackName,
			"template_bucket":     stack.VendorConfig.CloudFormation.TemplateBucket,
		}

		d.Set("cloudformation", []interface{}{m})
	case structs.StackConfigVendorKubernetes:
		m := map[string]interface{}{
			"namespace": stack.VendorConfig.Kubernetes.Namespace,
		}

		d.Set("kubernetes", []interface{}{m})
	case structs.StackConfigVendorPulumi:
		m := map[string]interface{}{
			"login_url":  stack.VendorConfig.Pulumi.LoginURL,
			"stack_name": stack.VendorConfig.Pulumi.StackName,
		}

		d.Set("pulumi", []interface{}{m})
	default:
		d.Set("terraform_smart_sanitization", stack.VendorConfig.Terraform.UseSmartSanitization)
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
		Administrative:      graphql.Boolean(d.Get("administrative").(bool)),
		Autodeploy:          graphql.Boolean(d.Get("autodeploy").(bool)),
		Autoretry:           graphql.Boolean(d.Get("autoretry").(bool)),
		Branch:              toString(d.Get("branch")),
		GitHubActionDeploy:  graphql.Boolean(d.Get("github_action_deploy").(bool)),
		LocalPreviewEnabled: graphql.Boolean(d.Get("enable_local_preview").(bool)),
		Name:                toString(d.Get("name")),
		ProtectFromDeletion: graphql.Boolean(d.Get("protect_from_deletion").(bool)),
		Repository:          toString(d.Get("repository")),
	}

	afterApplies := []graphql.String{}
	if commands, ok := d.GetOk("after_apply"); ok {
		for _, cmd := range commands.([]interface{}) {
			afterApplies = append(afterApplies, graphql.String(cmd.(string)))
		}
	}
	ret.AfterApply = &afterApplies

	afterDestroys := []graphql.String{}
	if commands, ok := d.GetOk("after_destroy"); ok {
		for _, cmd := range commands.([]interface{}) {
			afterDestroys = append(afterDestroys, graphql.String(cmd.(string)))
		}
	}
	ret.AfterDestroy = &afterDestroys

	afterInits := []graphql.String{}
	if commands, ok := d.GetOk("after_init"); ok {
		for _, cmd := range commands.([]interface{}) {
			afterInits = append(afterInits, graphql.String(cmd.(string)))
		}
	}
	ret.AfterInit = &afterInits

	afterPerforms := []graphql.String{}
	if commands, ok := d.GetOk("after_perform"); ok {
		for _, cmd := range commands.([]interface{}) {
			afterPerforms = append(afterPerforms, graphql.String(cmd.(string)))
		}
	}
	ret.AfterPerform = &afterPerforms

	afterPlans := []graphql.String{}
	if commands, ok := d.GetOk("after_plan"); ok {
		for _, cmd := range commands.([]interface{}) {
			afterPlans = append(afterPlans, graphql.String(cmd.(string)))
		}
	}
	ret.AfterPlan = &afterPlans

	afterRuns := []graphql.String{}
	if commands, ok := d.GetOk("after_run"); ok {
		for _, cmd := range commands.([]interface{}) {
			afterRuns = append(afterRuns, graphql.String(cmd.(string)))
		}
	}
	ret.AfterRun = &afterRuns

	beforeApplies := []graphql.String{}
	if commands, ok := d.GetOk("before_apply"); ok {
		for _, cmd := range commands.([]interface{}) {
			beforeApplies = append(beforeApplies, graphql.String(cmd.(string)))
		}
	}
	ret.BeforeApply = &beforeApplies

	beforeDestroys := []graphql.String{}
	if commands, ok := d.GetOk("before_destroy"); ok {
		for _, cmd := range commands.([]interface{}) {
			beforeDestroys = append(beforeDestroys, graphql.String(cmd.(string)))
		}
	}
	ret.BeforeDestroy = &beforeDestroys

	beforeInits := []graphql.String{}
	if commands, ok := d.GetOk("before_init"); ok {
		for _, cmd := range commands.([]interface{}) {
			beforeInits = append(beforeInits, graphql.String(cmd.(string)))
		}
	}
	ret.BeforeInit = &beforeInits

	beforePerforms := []graphql.String{}
	if commands, ok := d.GetOk("before_perform"); ok {
		for _, cmd := range commands.([]interface{}) {
			beforePerforms = append(beforePerforms, graphql.String(cmd.(string)))
		}
	}
	ret.BeforePerform = &beforePerforms

	beforePlans := []graphql.String{}
	if commands, ok := d.GetOk("before_plan"); ok {
		for _, cmd := range commands.([]interface{}) {
			beforePlans = append(beforePlans, graphql.String(cmd.(string)))
		}
	}
	ret.BeforePlan = &beforePlans

	description, ok := d.GetOk("description")
	if ok {
		ret.Description = toOptionalString(description)
	}

	ret.Provider = graphql.NewString("GITHUB")

	if azureDevOps, ok := d.Get("azure_devops").([]interface{}); ok && len(azureDevOps) > 0 {
		ret.Namespace = toOptionalString(azureDevOps[0].(map[string]interface{})["project"])
		ret.Provider = graphql.NewString(structs.VCSProviderAzureDevOps)
	}

	if bitbucketCloud, ok := d.Get("bitbucket_cloud").([]interface{}); ok && len(bitbucketCloud) > 0 {
		ret.Namespace = toOptionalString(bitbucketCloud[0].(map[string]interface{})["namespace"])
		ret.Provider = graphql.NewString(structs.VCSProviderBitbucketCloud)
	}

	if bitbucketDatacenter, ok := d.Get("bitbucket_datacenter").([]interface{}); ok && len(bitbucketDatacenter) > 0 {
		ret.Namespace = toOptionalString(bitbucketDatacenter[0].(map[string]interface{})["namespace"])
		ret.Provider = graphql.NewString(structs.VCSProviderBitbucketDatacenter)
	}

	if githubEnterprise, ok := d.Get("github_enterprise").([]interface{}); ok && len(githubEnterprise) > 0 {
		ret.Namespace = toOptionalString(githubEnterprise[0].(map[string]interface{})["namespace"])
		ret.Provider = graphql.NewString(structs.VCSProviderGitHubEnterprise)
	}

	if gitlab, ok := d.Get("gitlab").([]interface{}); ok && len(gitlab) > 0 {
		ret.Namespace = toOptionalString(gitlab[0].(map[string]interface{})["namespace"])
		ret.Provider = graphql.NewString(structs.VCSProviderGitlab)
	}

	if showcase, ok := d.Get("showcase").([]interface{}); ok && len(showcase) > 0 {
		ret.Namespace = toOptionalString(showcase[0].(map[string]interface{})["namespace"])
		ret.Provider = graphql.NewString(structs.VCSProviderShowcases)
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

	if runnerImage, ok := d.GetOk("runner_image"); ok {
		ret.RunnerImage = toOptionalString(runnerImage)
	}

	if terraformVersion, ok := d.GetOk("terraform_version"); ok {
		ret.VendorConfig = &structs.VendorConfigInput{Terraform: &structs.TerraformInput{
			Version: toOptionalString(terraformVersion),
		}}
	}

	if cloudFormation, ok := d.Get("cloudformation").([]interface{}); ok && len(cloudFormation) > 0 {
		ret.VendorConfig = &structs.VendorConfigInput{
			CloudFormationInput: &structs.CloudFormationInput{
				EntryTemplateFile: toString(cloudFormation[0].(map[string]interface{})["entry_template_file"]),
				Region:            toString(cloudFormation[0].(map[string]interface{})["region"]),
				StackName:         toString(cloudFormation[0].(map[string]interface{})["stack_name"]),
				TemplateBucket:    toString(cloudFormation[0].(map[string]interface{})["template_bucket"]),
			},
		}
	} else if kubernetes, ok := d.Get("kubernetes").([]interface{}); ok && len(kubernetes) > 0 {
		ret.VendorConfig = &structs.VendorConfigInput{
			Kubernetes: &structs.KubernetesInput{},
		}

		if kubernetesSettings, ok := kubernetes[0].(map[string]interface{}); ok {
			ret.VendorConfig.Kubernetes.Namespace = toString(kubernetesSettings["namespace"])
		}
	} else if pulumi, ok := d.Get("pulumi").([]interface{}); ok && len(pulumi) > 0 {
		ret.VendorConfig = &structs.VendorConfigInput{
			Pulumi: &structs.PulumiInput{
				LoginURL:  toString(pulumi[0].(map[string]interface{})["login_url"]),
				StackName: toString(pulumi[0].(map[string]interface{})["stack_name"]),
			},
		}
	} else if ansible, ok := d.Get("ansible").([]interface{}); ok && len(ansible) > 0 {
		ret.VendorConfig = &structs.VendorConfigInput{
			AnsibleInput: &structs.AnsibleInput{
				Playbook: toString(ansible[0].(map[string]interface{})["playbook"]),
			},
		}
	} else {
		terraformConfig := &structs.TerraformInput{}

		if terraformVersion, ok := d.GetOk("terraform_version"); ok {
			terraformConfig.Version = toOptionalString(terraformVersion)
		}

		if terraformWorkspace, ok := d.GetOk("terraform_workspace"); ok {
			terraformConfig.Workspace = toOptionalString(terraformWorkspace)
		}

		if terraformSmartSanitization, ok := d.GetOk("terraform_smart_sanitization"); ok {
			terraformConfig.UseSmartSanitization = toOptionalBool(terraformSmartSanitization)
		} else {
			terraformConfig.UseSmartSanitization = toOptionalBool(false)
		}

		ret.VendorConfig = &structs.VendorConfigInput{Terraform: terraformConfig}
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.WorkerPool = graphql.NewID(workerPoolID)
	}

	return ret
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
	if d.Id() == "" {
		return nil, errors.New("stack ID is required to import a stack")
	}

	diag := resourceStackRead(ctx, d, meta)
	if diag.HasError() {
		return nil, fmt.Errorf("could not import stack: %v", diag)
	}

	return []*schema.ResourceData{d}, nil
}
