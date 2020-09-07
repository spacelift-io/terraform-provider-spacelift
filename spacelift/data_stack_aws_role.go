package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func dataAWSRole() *schema.Resource {
	return &schema.Resource{
		Read: dataAWSRoleRead,

		Schema: map[string]*schema.Schema{
			"role_arn": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Computed:    true,
			},
			"stack_id": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "ID of the stack which assumes the AWS IAM role",
				ConflictsWith: []string{"module_id"},
				Optional:      true,
			},
			"module_id": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "ID of the stack which assumes the AWS IAM role",
				ConflictsWith: []string{"stack_id"},
				Optional:      true,
			},
		},
	}
}

func dataAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("stack_id"); ok {
		return dataStackAWSRoleRead(d, meta)
	}

	if _, ok := d.GetOk("module_id"); ok {
		return dataModuleAWSRoleRead(d, meta)
	}

	return errors.New("either module_id or stack_id must be provided")
}

func dataModuleAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	moduleID := d.Get("module_id")
	variables := map[string]interface{}{"id": toID(moduleID)}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	module := query.Module
	if module == nil {
		return errors.New("module not found")
	}

	d.SetId(moduleID.(string))

	if roleARN := module.Integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", *roleARN)
	} else {
		d.Set("role_arn", "")
	}

	return nil
}

func dataStackAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
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

	if roleARN := stack.Integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", *roleARN)
	} else {
		d.Set("role_arn", "")
	}

	return nil
}
