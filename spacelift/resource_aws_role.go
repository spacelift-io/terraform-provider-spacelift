package spacelift

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

// Deprecated! Used for backwards compatibility.
func resourceStackAWSRole() *schema.Resource {
	schema := resourceAWSRole()
	schema.DeprecationMessage = "use spacelift_aws_role resource instead"

	return schema
}

func resourceAWSRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAWSRoleCreate,
		ReadContext:   resourceAWSRoleRead,
		UpdateContext: resourceAWSRoleUpdate,
		DeleteContext: resourceAWSRoleDelete,

		Importer: &schema.ResourceImporter{StateContext: importIntegration},

		Schema: map[string]*schema.Schema{
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module which assumes the AWS IAM role",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
				ForceNew:     true,
			},
			"role_arn": {
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Required:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which assumes the AWS IAM role",
				Optional:    true,
				ForceNew:    true,
			},
			"generate_credentials_in_worker": {
				Type:        schema.TypeBool,
				Description: "Generate AWS credentials in the private worker",
				Optional:    true,
				Default:     false,
			},
			"external_id": {
				Type:        schema.TypeString,
				Description: "Custom external ID (works only for private workers).",
				Optional:    true,
			},
		},
	}
}

func resourceAWSRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var ID string

	if stackID, ok := d.GetOk("stack_id"); ok {
		ID = stackID.(string)
	} else {
		ID = d.Get("module_id").(string)
	}

	var err error

	for i := 0; i < 5; i++ {
		err = resourceAWSRoleSet(ctx, meta.(*internal.Client), ID, d)
		if err == nil || !strings.Contains(err.Error(), "AccessDenied") || i == 4 {
			break
		}

		// Yay for eventual consistency.
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		return diag.Errorf("could not create AWS role delegation: %v", err)
	}

	d.SetId(ID)

	return nil
}

func resourceAWSRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if _, ok := d.GetOk("module_id"); ok {
		return resourceModuleAWSRoleRead(ctx, d, meta)
	}

	return resourceStackAWSRoleRead(ctx, d, meta)
}

func resourceModuleAWSRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, "ModuleAWSRoleRead", &query, variables); err != nil {
		return diag.Errorf("could not query for module: %v", err)
	}

	if query.Module == nil {
		d.SetId("")
		return nil
	}

	resourceAWSRoleSetIntegration(d, &query.Module.Integrations)

	return nil
}

func resourceStackAWSRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, "StackAWSRoleRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	if query.Stack == nil {
		d.SetId("")
		return nil
	}

	resourceAWSRoleSetIntegration(d, query.Stack.Integrations)

	return nil
}

func resourceAWSRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var ID string

	if stackID, ok := d.GetOk("stack_id"); ok {
		ID = stackID.(string)
	} else {
		ID = d.Get("module_id").(string)
	}

	var ret diag.Diagnostics
	if err := resourceAWSRoleSet(ctx, meta.(*internal.Client), ID, d); err != nil {
		ret = append(ret, diag.FromErr(err)...)
	}
	ret = append(ret, resourceAWSRoleRead(ctx, d, meta)...)

	return ret
}

func resourceAWSRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		AttachAWSRole struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationAwsDelete(id: $id)"`
	}

	variables := map[string]interface{}{}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["id"] = stackID.(string)
	} else {
		variables["id"] = d.Get("module_id").(string)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AWSRoleDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete AWS role delegation: %v", err)
	}

	if mutation.AttachAWSRole.Activated {
		return diag.Errorf("did not disable AWS integration, still reporting as activated")
	}

	d.SetId("")
	return nil
}

func resourceAWSRoleSet(ctx context.Context, client *internal.Client, id string, d *schema.ResourceData) error {
	var mutation struct {
		AttachAWSRole struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationAwsCreate(id: $id, roleArn: $roleArn, generateCredentialsInWorker: $generateCredentialsInWorker, externalID: $externalID)"`
	}

	variables := map[string]interface{}{
		"id":                          toID(id),
		"roleArn":                     graphql.String(d.Get("role_arn").(string)),
		"generateCredentialsInWorker": graphql.Boolean(d.Get("generate_credentials_in_worker").(bool)),
	}

	if externalID, ok := d.GetOk("external_id"); ok {
		variables["externalID"] = toOptionalString(externalID)
	} else {
		variables["externalID"] = (*graphql.String)(nil)
	}

	if err := client.Mutate(ctx, "AWSRoleSet", &mutation, variables); err != nil {
		return errors.Wrap(err, "could not set AWS role delegation")
	}

	if !mutation.AttachAWSRole.Activated {
		return errors.New("AWS integration not activated")
	}

	return nil
}

func resourceAWSRoleSetIntegration(d *schema.ResourceData, integrations *structs.Integrations) {
	if roleARN := integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", roleARN)
	} else {
		d.Set("role_arn", nil)
	}

	d.Set("generate_credentials_in_worker", integrations.AWS.GenerateCredentialsInWorker)
	d.Set("external_id", integrations.AWS.ExternalID)
}
