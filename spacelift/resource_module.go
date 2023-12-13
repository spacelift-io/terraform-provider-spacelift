package spacelift

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceModule() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_module` is a special type of a stack used to test and " +
			"version Terraform modules.",

		CreateContext: resourceModuleCreate,
		ReadContext:   resourceModuleRead,
		UpdateContext: resourceModuleUpdate,
		DeleteContext: resourceModuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"administrative": {
				Type:        schema.TypeBool,
				Description: "Indicates whether this module can manage others. Defaults to `false`.",
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
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the Azure Devops integration. If not specified, the default integration will be used.",
							Optional:    true,
						},
						"project": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The name of the Azure DevOps project",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
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
			"branch": {
				Type:        schema.TypeString,
				Description: "GitHub branch to apply changes to",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form module description for users",
				Optional:    true,
			},
			"enable_local_preview": {
				Type:        schema.TypeBool,
				Description: "Indicates whether local preview versions can be triggered on this Module. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"github_enterprise": {
				Type:          schema.TypeList,
				Description:   "GitHub Enterprise (self-hosted) VCS settings",
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
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the GitHub Enterprise integration. If not specified, the default integration will be used.",
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
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the Gitlab integration. If not specified, the default integration will be used.",
							Optional:    true,
						},
						"namespace": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The GitLab namespace containing the repository",
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"labels": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The module name will by default be inferred from the repository name if it follows the terraform-provider-name naming convention. However, if the repository doesn't follow this convention, or you want to give it a custom name, you can provide it here.",
				Computed:    true,
				ForceNew:    true,
				Optional:    true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[a-z][a-z0-9-_]*[a-z0-9]$"),
					"invalid name format: must start and end with lowercase letter and may only contain lowercase letters, digits, dashes and underscores",
				),
			},
			"project_root": {
				Type:        schema.TypeString,
				Description: "Project root is the optional directory relative to the repository root containing the module source code.",
				Optional:    true,
			},
			"protect_from_deletion": {
				Type:        schema.TypeBool,
				Description: "Protect this module from accidental deletion. If set, attempts to delete this module will fail. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"repository": {
				Type:             schema.TypeString,
				Description:      "Name of the repository, without the owner part",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"shared_accounts": {
				Type:        schema.TypeSet,
				Description: "List of the accounts (subdomains) which should have access to the Module",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the module is in",
				Optional:    true,
				Computed:    true,
			},
			"terraform_provider": {
				Type:        schema.TypeString,
				Description: "The module provider will by default be inferred from the repository name if it follows the terraform-provider-name naming convention. However, if the repository doesn't follow this convention, or you gave the module a custom name, you can provide the provider name here.",
				Computed:    true,
				ForceNew:    true,
				Optional:    true,
			},
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to use. NOTE: worker_pool_id is required when using a self-hosted instance of Spacelift.",
				Optional:    true,
			},
			"workflow_tool": {
				Type:        schema.TypeString,
				Description: "Defines the tool that will be used to execute the workflow. This can be one of `OPEN_TOFU`, `TERRAFORM_FOSS` or `CUSTOM`. Defaults to `TERRAFORM_FOSS`.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceModuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateModule *structs.Module `graphql:"moduleCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": moduleCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ModuleCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create module: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateModule.ID)

	return resourceModuleRead(ctx, d, meta)
}

func resourceModuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, "ModuleRead", &query, variables); err != nil {
		return diag.Errorf("could not query for module: %v", err)
	}

	module := query.Module
	if module == nil {
		d.SetId("")
		return nil
	}

	d.Set("aws_assume_role_policy_statement", module.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("administrative", module.Administrative)
	d.Set("branch", module.Branch)
	d.Set("name", module.Name)
	d.Set("enable_local_preview", module.LocalPreviewEnabled)
	d.Set("protect_from_deletion", module.ProtectFromDeletion)
	d.Set("repository", module.Repository)
	d.Set("terraform_provider", module.TerraformProvider)
	d.Set("space_id", module.Space)

	if description := module.Description; description != nil {
		d.Set("description", *description)
	}

	if projectRoot := module.ProjectRoot; projectRoot != nil {
		d.Set("project_root", *projectRoot)
	}

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

func resourceModuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateModule structs.Module `graphql:"moduleUpdateV2(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": moduleUpdateV2Input(d),
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "ModuleUpdate", &mutation, variables); err != nil {
		ret = diag.FromErr(internal.FromSpaceliftError(err))
	}

	return append(ret, resourceModuleRead(ctx, d, meta)...)
}

func resourceModuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteModule *structs.Module `graphql:"moduleDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "ModuleDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete module: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func getSourceData(d *schema.ResourceData) (provider *graphql.String, namespace *graphql.String, vcsIntegrationID *graphql.ID) {
	provider = graphql.NewString("GITHUB")

	if azureDevOps, ok := d.Get("azure_devops").([]interface{}); ok && len(azureDevOps) > 0 {
		azureSettings := azureDevOps[0].(map[string]interface{})
		if id, ok := azureSettings["id"]; ok {
			vcsIntegrationID = graphql.NewID(id)
		}
		namespace = toOptionalString(azureSettings["project"])
		provider = graphql.NewString(graphql.String(structs.VCSProviderAzureDevOps))

		return
	}

	if bitbucketCloud, ok := d.Get("bitbucket_cloud").([]interface{}); ok && len(bitbucketCloud) > 0 {
		namespace = toOptionalString(bitbucketCloud[0].(map[string]interface{})["namespace"])
		provider = graphql.NewString(graphql.String(structs.VCSProviderBitbucketCloud))

		return
	}

	if bitbucketDatacenter, ok := d.Get("bitbucket_datacenter").([]interface{}); ok && len(bitbucketDatacenter) > 0 {
		namespace = toOptionalString(bitbucketDatacenter[0].(map[string]interface{})["namespace"])
		provider = graphql.NewString(graphql.String(structs.VCSProviderBitbucketDatacenter))

		return
	}

	if githubEnterprise, ok := d.Get("github_enterprise").([]interface{}); ok && len(githubEnterprise) > 0 {
		ghEnterpriseSettings := githubEnterprise[0].(map[string]interface{})
		if id, ok := ghEnterpriseSettings["id"]; ok {
			vcsIntegrationID = graphql.NewID(id)
		}
		namespace = toOptionalString(githubEnterprise[0].(map[string]interface{})["namespace"])
		provider = graphql.NewString(graphql.String(structs.VCSProviderGitHubEnterprise))

		return
	}

	if gitlab, ok := d.Get("gitlab").([]interface{}); ok && len(gitlab) > 0 {
		gitlabSettings := gitlab[0].(map[string]interface{})
		if id, ok := gitlabSettings["id"]; ok {
			vcsIntegrationID = graphql.NewID(id.(string))
		}
		namespace = toOptionalString(gitlabSettings["namespace"])
		provider = graphql.NewString(graphql.String(structs.VCSProviderGitlab))
	}

	return
}

func moduleCreateInput(d *schema.ResourceData) structs.ModuleCreateInput {
	ret := structs.ModuleCreateInput{
		UpdateInput: moduleUpdateInput(d),
		Repository:  toString(d.Get("repository")),
	}
	ret.Provider, ret.Namespace, ret.VCSIntegrationID = getSourceData(d)

	name, ok := d.GetOk("name")
	if ok {
		ret.Name = toOptionalString(name)
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.UpdateInput.WorkerPool = graphql.NewID(workerPoolID)
	}

	if space, ok := d.GetOk("space_id"); ok {
		ret.Space = toOptionalString(space)
	}

	terraformProvider, ok := d.GetOk("terraform_provider")
	if ok {
		ret.TerraformProvider = toOptionalString(terraformProvider)
	}

	if workflowTool, ok := d.GetOk("workflow_tool"); ok {
		ret.WorkflowTool = toOptionalString(workflowTool)
	}

	return ret
}

func moduleUpdateInput(d *schema.ResourceData) structs.ModuleUpdateInput {
	ret := structs.ModuleUpdateInput{
		Administrative:      graphql.Boolean(d.Get("administrative").(bool)),
		Branch:              toString(d.Get("branch")),
		LocalPreviewEnabled: graphql.Boolean(d.Get("enable_local_preview").(bool)),
		ProtectFromDeletion: graphql.Boolean(d.Get("protect_from_deletion").(bool)),
	}

	description, ok := d.GetOk("description")
	if ok {
		ret.Description = toOptionalString(description)
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		ret.Labels = &labels
	}

	if projectRoot := d.Get("project_root"); projectRoot != "" {
		ret.ProjectRoot = toOptionalString(projectRoot)
	}

	if sharedAccountsSet, ok := d.Get("shared_accounts").(*schema.Set); ok {
		var sharedAccounts []graphql.String
		for _, account := range sharedAccountsSet.List() {
			sharedAccounts = append(sharedAccounts, graphql.String(account.(string)))
		}
		ret.SharedAccounts = &sharedAccounts
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.WorkerPool = graphql.NewID(workerPoolID)
	}

	if space, ok := d.GetOk("space_id"); ok {
		ret.Space = toOptionalString(space)
	}

	return ret
}

func moduleUpdateV2Input(d *schema.ResourceData) structs.ModuleUpdateV2Input {
	ret := structs.ModuleUpdateV2Input{
		Administrative:      graphql.Boolean(d.Get("administrative").(bool)),
		Branch:              toString(d.Get("branch")),
		LocalPreviewEnabled: graphql.Boolean(d.Get("enable_local_preview").(bool)),
		ProtectFromDeletion: graphql.Boolean(d.Get("protect_from_deletion").(bool)),
		Repository:          toString(d.Get("repository")),
	}
	ret.Provider, ret.Namespace, ret.VCSIntegrationID = getSourceData(d)

	description, ok := d.GetOk("description")
	if ok {
		ret.Description = toOptionalString(description)
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		ret.Labels = &labels
	}

	if projectRoot := d.Get("project_root"); projectRoot != "" {
		ret.ProjectRoot = toOptionalString(projectRoot)
	}

	if sharedAccountsSet, ok := d.Get("shared_accounts").(*schema.Set); ok {
		var sharedAccounts []graphql.String
		for _, account := range sharedAccountsSet.List() {
			sharedAccounts = append(sharedAccounts, graphql.String(account.(string)))
		}
		ret.SharedAccounts = &sharedAccounts
	}

	if space, ok := d.GetOk("space_id"); ok {
		ret.Space = toOptionalString(space)
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.WorkerPool = graphql.NewID(workerPoolID)
	}

	if workflowTool, ok := d.GetOk("workflow_tool"); ok {
		ret.WorkflowTool = toOptionalString(workflowTool)
	}

	return ret
}
