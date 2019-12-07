package spacelift

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
	roleARN := aws.String(d.Get("role_arn").(string))

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

	if roleARN := query.Stack.AWSAssumedRoleARN; roleARN != nil {
		d.Set("role_arn", roleARN)
	} else {
		d.Set("role_arn", nil)
	}

	return nil
}

func resourceStackAWSRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	stackID := d.Get("stack_id").(string)
	roleARN := d.Get("role_arn").(string)

	if err := resourceStackAWSRoleSet(meta.(*Client), stackID, aws.String(roleARN)); err != nil {
		return errors.Wrap(err, "could not update AWS role delegation")
	}

	return nil
}

func resourceStackAWSRoleDelete(d *schema.ResourceData, meta interface{}) error {
	stackID := d.Get("stack_id").(string)

	if err := resourceStackAWSRoleSet(meta.(*Client), stackID, nil); err != nil {
		return errors.Wrap(err, "could not delete AWS role delegation")
	}

	d.SetId("")
	return nil
}

func resourceStackAWSRoleSet(client *Client, stackID string, roleARN *string) error {
	var mutation struct {
		AttachAWSRole *structs.Stack `graphql:"stackSetAwsRoleDelegation(id: $id, roleArn: $roleArn)"`
	}

	variables := map[string]interface{}{
		"id":      toID(stackID),
		"roleArn": (*graphql.String)(nil),
	}

	if roleARN != nil {
		variables["roleArn"] = toOptionalString(*roleARN)
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not set AWS role delegation on the stack")
	}

	return nil
}
