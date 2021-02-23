package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

// Deprecated! Used for backwards compatibility.
func dataStackAWSRole() *schema.Resource {
	schema := dataAWSRole()
	schema.DeprecationMessage = "use spacelift_aws_role data source instead"

	return schema
}

func dataAWSRole() *schema.Resource {
	return &schema.Resource{
		Read: dataAWSRoleRead,

		Schema: map[string]*schema.Schema{
			"role_arn": {
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Computed:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the module which assumes the AWS IAM role",
				ConflictsWith: []string{"stack_id"},
				Optional:      true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which assumes the AWS IAM role",
				Optional:    true,
			},
			"generate_credentials_in_worker": {
				Type:        schema.TypeBool,
				Description: "Generate AWS credentials in the private worker",
				Optional:    true,
			},
		},
	}
}

func dataAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("module_id"); ok {
		return dataModuleAWSRoleRead(d, meta)
	}

	if _, ok := d.GetOk("stack_id"); ok {
		return dataStackAWSRoleRead(d, meta)
	}

	return errors.New("either module_id or stack_id must be provided")
}

func dataModuleAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
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

	if roleARN := module.Integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", *roleARN)
	} else {
		return errors.New("this module is missing the AWS integration")
	}

	d.SetId(moduleID.(string))
	d.Set("generate_credentials_in_worker", query.Module.Integrations.AWS.GenerateCredentialsInWorker)

	return nil
}

func dataStackAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id")
	variables := map[string]interface{}{"id": toID(stackID)}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	stack := query.Stack
	if stack == nil {
		return errors.New("stack not found")
	}

	if roleARN := stack.Integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", *roleARN)
	} else {
		return errors.New("this stack is missing the AWS integration")
	}

	d.SetId(stackID.(string))
	d.Set("generate_credentials_in_worker", query.Stack.Integrations.AWS.GenerateCredentialsInWorker)

	return nil
}
