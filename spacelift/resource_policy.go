package spacelift

import (
	"github.com/fluxio/multierror"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

var policyTypes = []string{
	"ACCESS",
	"GIT_PUSH",
	"INITIALIZATION",
	"LOGIN",
	"PLAN",
	"STACK_ACCESS", // deprecated
	"TASK",
	"TASK_RUN",       // deprecated
	"TERRAFORM_PLAN", // deprecated
	"TRIGGER",
}

// This is a map of new policy type names to the ones they are replacing.
var typeNameReplacements = map[string]string{
	"ACCESS": "STACK_ACCESS",
	"PLAN":   "TERRAFORM_PLAN",
	"TASK":   "TASK_RUN",
}

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
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the policy - should be unique in one account",
				Required:    true,
				ForceNew:    true,
			},
			"body": {
				Type:        schema.TypeString,
				Description: "Body of the policy",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Body of the policy",
				Required:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					// If the backend responds with a new name, but we still have the old
					// name defined or stored in the state, let's not do the replacement.
					if previous, ok := typeNameReplacements[new]; ok && previous == old {
						return true
					}
					next, ok := typeNameReplacements[old]
					return ok && next == new
				},
				ValidateFunc: validation.StringInSlice(
					policyTypes,
					false, // case-sensitive match
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

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
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
	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
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

	var acc multierror.Accumulator

	acc.Push(errors.Wrap(meta.(*internal.Client).Mutate(&mutation, variables), "could not update policy"))
	acc.Push(errors.Wrap(resourcePolicyRead(d, meta), "could not read the current state"))

	return acc.Error()
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeletePolicy *structs.Policy `graphql:"policyDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete policy")
	}

	d.SetId("")

	return nil
}
