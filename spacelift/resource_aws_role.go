package spacelift

import (
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceAWSRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceAWSRoleCreate,
		Read:   resourceAWSRoleRead,
		Update: resourceAWSRoleUpdate,
		Delete: resourceAWSRoleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"module_id": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "ID of the module which assumes the AWS IAM role",
				ConflictsWith: []string{"stack_id"},
				Optional:      true,
				ForceNew:      true,
			},
			"role_arn": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Required:    true,
			},
			"stack_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the stack which assumes the AWS IAM role",
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceAWSRoleCreate(d *schema.ResourceData, meta interface{}) error {
	var ID string

	roleARN := d.Get("role_arn").(string)

	if stackID, ok := d.GetOk("stack_id"); ok {
		ID = stackID.(string)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		ID = moduleID.(string)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	var err error

	for i := 0; i < 5; i++ {
		err = resourceAWSRoleSet(meta.(*Client), ID, roleARN)
		if err == nil || !strings.Contains(err.Error(), "AccessDenied") || i == 4 {
			break
		}

		// Yay for eventual consistency.
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		return errors.Wrap(err, "could not create AWS role delegation")
	}

	d.SetId(ID)

	return nil
}

func resourceAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("module_id"); ok {
		return resourceModuleRead(d, meta)
	}

	return resourceStackRead(d, meta)
}

func resourceModuleAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	if query.Module == nil {
		d.SetId("")
		return nil
	}

	if roleARN := query.Module.Integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", roleARN)
	} else {
		d.Set("role_arn", nil)
	}

	return nil
}

func resourceStackAWSRoleRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	if query.Stack == nil {
		d.SetId("")
		return nil
	}

	if roleARN := query.Stack.Integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.Set("role_arn", roleARN)
	} else {
		d.Set("role_arn", nil)
	}

	return nil
}

func resourceAWSRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	var ID string
	roleARN := d.Get("role_arn").(string)

	if stackID, ok := d.GetOk("stack_id"); ok {
		ID = stackID.(string)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		ID = moduleID.(string)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	if err := resourceAWSRoleSet(meta.(*Client), ID, roleARN); err != nil {
		return errors.Wrap(err, "could not update AWS role delegation")
	}

	return nil
}

func resourceAWSRoleDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		AttachAWSRole struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationAwsDelete(id: $id)"`
	}

	variables := map[string]interface{}{}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["id"] = stackID.(string)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		variables["id"] = moduleID.(string)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete AWS role delegation")
	}

	if mutation.AttachAWSRole.Activated {
		return errors.New("did not disable AWS integration, still reporting as activated")
	}

	d.SetId("")
	return nil
}

func resourceAWSRoleSet(client *Client, ID, roleARN string) error {
	var mutation struct {
		AttachAWSRole struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationAwsCreate(id: $id, roleArn: $roleArn)"`
	}

	variables := map[string]interface{}{
		"id":      toID(ID),
		"roleArn": graphql.String(roleARN),
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not set AWS role delegation")
	}

	if !mutation.AttachAWSRole.Activated {
		return errors.New("AWS integration not activated")
	}

	return nil
}
