package spacelift

import (
	"github.com/fluxio/multierror"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceModule() *schema.Resource {
	return &schema.Resource{
		Create: resourceModuleCreate,
		Read:   resourceModuleRead,
		Update: resourceModuleUpdate,
		Delete: resourceModuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"repository": {
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Required:    true,
			},
			"shared_accounts": {
				Type:        schema.TypeSet,
				Description: "List of the accounts (subdomains) which have access to the Module ",
				Elem:        &schema.Schema{Type: schema.TypeString},
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

func resourceModuleCreate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		CreateModule *structs.Module `graphql:"moduleCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": moduleCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create module")
	}

	d.SetId(mutation.CreateModule.ID)

	return resourceModuleRead(d, meta)
}

func resourceModuleRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	module := query.Module
	if module == nil {
		d.SetId("")
		return nil
	}

	d.Set("aws_assume_role_policy_statement", module.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("administrative", module.Administrative)
	d.Set("branch", module.Branch)
	d.Set("repository", module.Repository)

	if description := module.Description; description != nil {
		d.Set("description", *description)
	}

	if module.Provider == "GITLAB" {
		m := map[string]interface{}{
			"namespace": module.Namespace,
		}

		if err := d.Set("gitlab", []interface{}{m}); err != nil {
			return errors.Wrap(err, "error setting gitlab (resource)")
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

func resourceModuleUpdate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		UpdateModule structs.Module `graphql:"moduleUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": moduleUpdateInput(d),
	}

	var acc multierror.Accumulator

	acc.Push(errors.Wrap(meta.(*internal.Client).Mutate(&mutation, variables), "could not update module"))
	acc.Push(errors.Wrap(resourceModuleRead(d, meta), "could not read the current state"))

	return acc.Error()
}

func resourceModuleDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeleteModule *structs.Module `graphql:"moduleDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete module")
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

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.UpdateInput.WorkerPool = graphql.NewID(workerPoolID)
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
