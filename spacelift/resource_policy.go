package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
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

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreatePolicy structs.Policy `graphql:"policyCreate(name: $name, body: $body, type: $type, labels: $labels)"`
	}

	variables := map[string]interface{}{
		"name":   toString(d.Get("name")),
		"body":   toString(d.Get("body")),
		"type":   structs.PolicyType(d.Get("type").(string)),
		"labels": (*[]graphql.String)(nil),
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String

		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}

		variables["labels"] = &labels
	}

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not create policy %v: %v", toString(d.Get("name")), err)
	}

	d.SetId(mutation.CreatePolicy.ID)

	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Policy *structs.Policy `graphql:"policy(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, &query, variables); err != nil {
		return diag.Errorf("could not query for policy: %v", err)
	}

	policy := query.Policy
	if policy == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", policy.Name)
	d.Set("body", policy.Body)
	d.Set("type", policy.Type)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range policy.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdatePolicy structs.Policy `graphql:"policyUpdate(id: $id, name: $name, body: $body, labels: $labels)"`
	}

	variables := map[string]interface{}{
		"id":     toID(d.Id()),
		"name":   toString(d.Get("name")),
		"body":   toString(d.Get("body")),
		"labels": (*[]graphql.String)(nil),
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String

		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}

		variables["labels"] = &labels
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		ret = diag.Errorf("could not update policy: %v", err)
	}

	return append(ret, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeletePolicy *structs.Policy `graphql:"policyDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not delete policy: %v", err)
	}

	d.SetId("")

	return nil
}
