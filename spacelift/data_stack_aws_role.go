package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func dataStackAWSRole() *schema.Resource {
	return &schema.Resource{
		Read: dataStackAWSRoleRead,

		Schema: map[string]*schema.Schema{
			"role_arn": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Computed:    true,
			},
			"stack_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the stack which assumes the AWS IAM role",
				Required:    true,
			},
		},
	}
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

	if roleARN := stack.AWSAssumedRoleARN; roleARN != nil {
		d.Set("role_arn", *roleARN)
	} else {
		d.Set("role_arn", "")
	}

	return nil
}
