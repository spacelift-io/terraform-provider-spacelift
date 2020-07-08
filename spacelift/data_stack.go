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
			"administrative": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Indicates whether this stack can administer others",
				Computed:    true,
			},
			"autodeploy": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Indicates whether changes to this stack can be automatically deployed",
				Computed:    true,
			},
			"aws_assume_role_policy_statement": &schema.Schema{
				Type:        schema.TypeString,
				Description: "AWS IAM assume role policy statement setting up trust relationship",
				Computed:    true,
			},
			"branch": &schema.Schema{
				Type:        schema.TypeString,
				Description: "GitHub branch to apply changes to",
				Computed:    true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Free-form stack description for users",
				Computed:    true,
			},
			"gitlab": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Gitlab-related attributes",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"labels": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"manage_state": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Determines if Spacelift should manage state for this stack",
				Computed:    true,
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the stack - should be unique in one account",
				Computed:    true,
			},
			"repository": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Computed:    true,
			},
			"stack_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID (slug) of the stack",
				Required:    true,
			},
			"terraform_version": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Terraform version to use",
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

	if stack.TerraformVersion != nil {
		d.Set("terraform_version", *stack.TerraformVersion)
	} else {
		d.Set("terraform_version", nil)
	}

	return nil
}
