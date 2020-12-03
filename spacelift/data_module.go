package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataModule() *schema.Resource {
	return &schema.Resource{
		Read: dataModuleRead,

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
			"description": {
				Type:        schema.TypeString,
				Description: "free-form module description for human users (supports Markdown)",
				Computed:    true,
			},
			"gitlab": {
				Type:        schema.TypeList,
				Description: "GitLab configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
			"repository": {
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Computed:    true,
			},
			"module_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the module",
				Required:    true,
			},
			"shared_accounts": {
				Type:        schema.TypeSet,
				Description: "List of the accounts (subdomains) which should have access to the Module",
				Elem:        &schema.Schema{Type: schema.TypeString},
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

func dataModuleRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	moduleID := d.Get("module_id")
	variables := map[string]interface{}{"id": toID(moduleID)}
	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	module := query.Module
	if module == nil {
		return errors.New("module not found")
	}

	d.SetId(moduleID.(string))
	d.Set("administrative", module.Administrative)
	d.Set("aws_assume_role_policy_statement", module.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("branch", module.Branch)

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

	d.Set("repository", module.Repository)

	if module.Description != nil {
		d.Set("description", *module.Description)
	} else {
		d.Set("description", nil)
	}

	if workerPool := module.WorkerPool; workerPool != nil {
		d.Set("worker_pool_id", workerPool.ID)
	} else {
		d.Set("worker_pool_id", nil)
	}

	return nil
}
