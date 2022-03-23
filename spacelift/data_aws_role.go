package spacelift

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

// Deprecated! Used for backwards compatibility.
func dataStackAWSRole() *schema.Resource {
	schema := dataAWSRole()
	schema.Description = "" +
		"~> **Note:** `spacelift_stack_aws_role` is deprecated. Please use `spacelift_aws_role` instead. The functionality is identical." +
		"\n\n" +
		strings.ReplaceAll(schema.Description, "spacelift_aws_role", "spacelift_stack_aws_role")

	return schema
}

func dataAWSRole() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_aws_role` represents [cross-account IAM role delegation](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) " +
			"between the Spacelift worker and an individual stack or module. " +
			"If this is set, Spacelift will use AWS STS to assume the supplied IAM role and " +
			"put its temporary credentials in the runtime environment." +
			"\n\n" +
			"If you use private workers, you can also assume IAM role on the worker side using " +
			"your own AWS credentials (e.g. from EC2 instance profile)." +
			"\n\n" +
			"Note: when assuming credentials for **shared worker**, Spacelift will use `$accountName@$stackID` " +
			"or `$accountName@$moduleID` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) " +
			"and Run ID as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole).",

		ReadContext: dataAWSRoleRead,

		Schema: map[string]*schema.Schema{
			"role_arn": {
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Computed:    true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module which assumes the AWS IAM role",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which assumes the AWS IAM role",
				Optional:    true,
			},
			"generate_credentials_in_worker": {
				Type:        schema.TypeBool,
				Description: "Generate AWS credentials in the private worker",
				Computed:    true,
			},
			"external_id": {
				Type:        schema.TypeString,
				Description: "Custom external ID (works only for private workers).",
				Computed:    true,
			},
		},
	}
}

func dataAWSRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if _, ok := d.GetOk("module_id"); ok {
		return dataModuleAWSRoleRead(ctx, d, meta)
	}

	return dataStackAWSRoleRead(ctx, d, meta)
}

func dataModuleAWSRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	moduleID := d.Get("module_id")
	variables := map[string]interface{}{"id": toID(moduleID)}

	if err := meta.(*internal.Client).Query(ctx, "ModuleAWSRoleRead", &query, variables); err != nil {
		return diag.Errorf("could not query for module: %v", err)
	}

	module := query.Module
	if module == nil {
		return diag.Errorf("module not found")
	}

	if roleARN := module.Integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", *roleARN)
	} else {
		return diag.Errorf("this module is missing the AWS integration")
	}

	d.SetId(moduleID.(string))
	d.Set("generate_credentials_in_worker", query.Module.Integrations.AWS.GenerateCredentialsInWorker)
	d.Set("external_id", module.Integrations.AWS.ExternalID)

	return nil
}

func dataStackAWSRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id")
	variables := map[string]interface{}{"id": toID(stackID)}

	if err := meta.(*internal.Client).Query(ctx, "StackAWSRoleRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	stack := query.Stack
	if stack == nil {
		return diag.Errorf("stack not found")
	}

	if roleARN := stack.Integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", *roleARN)
	} else {
		return diag.Errorf("this stack is missing the AWS integration")
	}

	d.SetId(stackID.(string))
	d.Set("generate_credentials_in_worker", query.Stack.Integrations.AWS.GenerateCredentialsInWorker)
	d.Set("external_id", stack.Integrations.AWS.ExternalID)

	return nil
}
