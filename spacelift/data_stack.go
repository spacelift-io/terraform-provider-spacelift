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
			"readers_team": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Slug of the GitHub team whose members get read-only access",
				Computed:    true,
			},
			"repository": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the GitHub repository, without the owner part",
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
			"writers_team": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Slug of the GitHub team whose members get read-write access",
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
	d.Set("aws_assume_role_policy_statement", stack.AWSAssumeRolePolicyStatement)
	d.Set("branch", stack.Branch)
	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("repository", stack.Repository)

	if stack.Description != nil {
		d.Set("description", *stack.Description)
	} else {
		d.Set("description", nil)
	}

	if stack.Readers != nil {
		d.Set("readers_team", stack.Readers.Slug)
	} else {
		d.Set("readers_team", nil)
	}

	if stack.TerraformVersion != nil {
		d.Set("terraform_version", *stack.TerraformVersion)
	} else {
		d.Set("terraform_version", nil)
	}

	if stack.Writers != nil {
		d.Set("writers_team", stack.Writers.Slug)
	} else {
		d.Set("writers_team", nil)
	}

	return nil
}
