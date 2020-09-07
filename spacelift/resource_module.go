package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
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
				Description: "Indicates whether this stack can manage others",
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
				Description: "Free-form stack description for users",
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
			"namespace": {
				Type:        schema.TypeString,
				Description: "Name of the stack - should be unique in one account",
				Optional:    true,
			},
			"repository": {
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Required:    true,
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

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
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

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	stack := query.Module
	if stack == nil {
		d.SetId("")
		return nil
	}

	d.Set("administrative", stack.Administrative)
	d.Set("branch", stack.Branch)
	d.Set("repository", stack.Repository)

	if description := stack.Description; description != nil {
		d.Set("description", *description)
	}

	if stack.Provider == "GITLAB" {
		m := map[string]interface{}{
			"namespace": stack.Namespace,
		}

		if err := d.Set("gitlab", []interface{}{m}); err != nil {
			return errors.Wrap(err, "error setting gitlab (resource)")
		}
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range stack.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if workerPool := stack.WorkerPool; workerPool != nil {
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

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not update module")
	}

	return resourceModuleRead(d, meta)
}

func resourceModuleDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeleteModule *structs.Module `graphql:"stackDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete stack")
	}

	d.SetId("")

	return nil
}

func moduleCreateInput(d *schema.ResourceData) structs.ModuleCreateInput {
	ret := structs.ModuleCreateInput{
		UpdateInput: structs.ModuleUpdateInput{
			Administrative: graphql.Boolean(d.Get("administrative").(bool)),
			Branch:         toString(d.Get("branch")),
		},
		Repository: toString(d.Get("repository")),
	}

	description, ok := d.GetOk("description")
	if ok {
		ret.UpdateInput.Description = toOptionalString(description)
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

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		ret.UpdateInput.Labels = &labels
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

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.WorkerPool = graphql.NewID(workerPoolID)
	}

	return ret
}
