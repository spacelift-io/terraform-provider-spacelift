package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func dataStack() *schema.Resource {
	return &schema.Resource{
		Read: dataStackRead,

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
			"aws_assume_role_policy_statement": {
				Type:        schema.TypeString,
				Description: "AWS IAM assume role policy statement setting up trust relationship",
				Computed:    true,
			},
			"branch": {
				Type:        schema.TypeString,
				Description: "Repository branch to treat as the default 'main' branch",
				Computed:    true,
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
			"repository": {
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
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
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to use",
				Computed:    true,
			},
		},
	}
}

func dataStackRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id")
	variables := map[string]interface{}{"id": toID(stackID)}
	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	stack := query.Stack
	if stack == nil {
		return errors.New("stack not found")
	}

	d.SetId(stackID.(string))
	d.Set("administrative", stack.Administrative)
	d.Set("autodeploy", stack.Autodeploy)
	d.Set("aws_assume_role_policy_statement", stack.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("branch", stack.Branch)

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

	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("repository", stack.Repository)

	if stack.Description != nil {
		d.Set("description", *stack.Description)
	} else {
		d.Set("description", nil)
	}

	if stack.ProjectRoot != nil {
		d.Set("project_root", *stack.ProjectRoot)
	} else {
		d.Set("project_root", nil)
	}

	if stack.TerraformVersion != nil {
		d.Set("terraform_version", *stack.TerraformVersion)
	} else {
		d.Set("terraform_version", nil)
	}

	if workerPool := stack.WorkerPool; workerPool != nil {
		d.Set("worker_pool_id", workerPool.ID)
	} else {
		d.Set("worker_pool_id", nil)
	}

	return nil
}
