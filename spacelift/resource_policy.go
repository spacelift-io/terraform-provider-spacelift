package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyCreate,
		Read:   resourcePolicyRead,
		Update: resourcePolicyUpdate,
		Delete: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the policy - should be unique in one account",
				Required:    true,
				ForceNew:    true,
			},
			"body": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Body of the policy",
				Required:    true,
			},
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Body of the policy",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						"LOGIN",
						"STACK_ACCESS",
						"INITIALIZATION",
						"TERRAFORM_PLAN",
						"TASK_RUN",
					},
					false,
				),
			},
		},
	}
}

func resourcePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		CreatePolicy structs.Policy `graphql:"policyCreate(name: $name, body: $body, type: $type)"`
	}

	variables := map[string]interface{}{
		"name": toString(d.Get("name")),
		"body": toString(d.Get("body")),
		"type": structs.PolicyType(d.Get("type").(string)),
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create policy")
	}

	d.SetId(mutation.CreatePolicy.ID)

	return resourcePolicyRead(d, meta)
}

func resourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Policy *structs.Policy `graphql:"policy(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for policy")
	}

	policy := query.Policy
	if policy == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", policy.Name)
	d.Set("body", policy.Body)
	d.Set("type", policy.Type)

	return nil
}

func resourcePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		UpdatePolicy structs.Policy `graphql:"policyUpdate(id: $id, name: $name, body: $body)"`
	}

	variables := map[string]interface{}{
		"id":   toID(d.Id()),
		"name": toString(d.Get("name")),
		"body": toString(d.Get("body")),
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not update policy")
	}

	return resourcePolicyRead(d, meta)
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeletePolicy *structs.Policy `graphql:"policyDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete policy")
	}

	d.SetId("")

	return nil
}
