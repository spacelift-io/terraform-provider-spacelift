package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
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
				Description: "Indicates whether this module can manage others",
				Optional:    true,
				Default:     false,
			},
			"aws_assume_role_policy_statement": {
				Type:        schema.TypeString,
				Description: "AWS IAM assume role policy statement setting up trust relationship",
				Computed:    true,
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
			"gitlab": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The module name will by default be inferred from the repository name if it follows the terraform-provider-name naming convention. However, if the repository doesn't follow this convention, or you want to give it a custom name, you can provide it here.",
				Computed:    true,
				ForceNew:    true,
				Optional:    true,
			},
			"project_root": {
				Type:        schema.TypeString,
				Description: "Project root is the optional directory relative to the repository root containing the module source code.",
				Optional:    true,
			},
			"repository": {
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Required:    true,
			},
			"shared_accounts": {
				Type:        schema.TypeSet,
				Description: "List of the accounts (subdomains) which should have access to the Module",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
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
				Description: "ID of the worker pool to use",
				Optional:    true,
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
	d.Set("repository", module.Repository)
	d.Set("terraform_provider", module.TerraformProvider)

	if description := module.Description; description != nil {
		d.Set("description", *description)
	}

	if projectRoot := module.ProjectRoot; projectRoot != nil {
		d.Set("project_root", *projectRoot)
	}

	if module.Provider == "GITLAB" {
		m := map[string]interface{}{
			"namespace": module.Namespace,
		}

		if err := d.Set("gitlab", []interface{}{m}); err != nil {
			return diag.Errorf("error setting gitlab (resource): %v", err)
		}
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

	return nil
}

func resourceModuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateModule structs.Module `graphql:"moduleUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": moduleUpdateInput(d),
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

func moduleCreateInput(d *schema.ResourceData) structs.ModuleCreateInput {
	ret := structs.ModuleCreateInput{
		UpdateInput: moduleUpdateInput(d),
		Repository:  toString(d.Get("repository")),
	}

	foundGitlab := false
	if gitlab, ok := d.Get("gitlab").([]interface{}); ok {
		if len(gitlab) > 0 {
			foundGitlab = true
			ret.Namespace = toOptionalString(gitlab[0].(map[string]interface{})["namespace"])
			ret.Provider = graphql.NewString(vcsProviderGitlab)
		}
	}
	if !foundGitlab {
		ret.Provider = graphql.NewString("GITHUB")
	}

	name, ok := d.GetOk("name")
	if ok {
		ret.Name = toOptionalString(name)
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.UpdateInput.WorkerPool = graphql.NewID(workerPoolID)
	}

	terraformProvider, ok := d.GetOk("terraform_provider")
	if ok {
		ret.TerraformProvider = toOptionalString(terraformProvider)
	}

	return ret
}

func moduleUpdateInput(d *schema.ResourceData) structs.ModuleUpdateInput {
	ret := structs.ModuleUpdateInput{
		Administrative: graphql.Boolean(d.Get("administrative").(bool)),
		Branch:         toString(d.Get("branch")),
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

	projectRoot, ok := d.GetOk("project_root")
	if ok {
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

	return ret
}
