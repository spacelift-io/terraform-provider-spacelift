package spacelift

import (
	"strings"
	"time"

	"github.com/fluxio/multierror"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceStackAWSRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackAWSRoleCreate,
		Read:   resourceStackAWSRoleRead,
		Update: resourceStackAWSRoleUpdate,
		Delete: resourceStackAWSRoleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"role_arn": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Required:    true,
			},
			"stack_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the stack which assumes the AWS IAM role",
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceStackAWSRoleCreate(d *schema.ResourceData, meta interface{}) error {
	stackID := d.Get("stack_id").(string)
	roleARN := d.Get("role_arn").(string)

	var err error

	for i := 0; i < 5; i++ {
		err = resourceStackAWSRoleSet(meta.(*Client), stackID, roleARN)
		if err == nil || !strings.Contains(err.Error(), "AccessDenied") || i == 4 {
			break
		}

		// Yay for eventual consistency.
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		return errors.Wrap(err, "could not create AWS role delegation")
	}

	d.SetId(stackID)

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

func resourceStackAWSRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	stackID := d.Get("stack_id").(string)
	roleARN := d.Get("role_arn").(string)

	var acc multierror.Accumulator

	acc.Push(errors.Wrap(resourceStackAWSRoleSet(meta.(*Client), stackID, roleARN), "could not update AWS role delegation"))
	acc.Push(errors.Wrap(resourceStackAWSRoleRead(d, meta), "could not read the current state"))

	return acc.Error()
}

func resourceStackAWSRoleDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		AttachAWSRole struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationAwsDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("stack_id").(string))}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete AWS role delegation")
	}

	if mutation.AttachAWSRole.Activated {
		return errors.New("did not disable AWS integration, still reporting as activated")
	}

	d.SetId("")
	return nil
}

func resourceStackAWSRoleSet(client *Client, stackID, roleARN string) error {
	var mutation struct {
		AttachAWSRole struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationAwsCreate(id: $id, roleArn: $roleArn)"`
	}

	variables := map[string]interface{}{
		"id":      toID(stackID),
		"roleArn": graphql.String(roleARN),
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not set AWS role delegation on the stack")
	}

	if !mutation.AttachAWSRole.Activated {
		return errors.New("AWS integration not activated")
	}

	return nil
}
