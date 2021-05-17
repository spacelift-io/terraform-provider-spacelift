package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_policy` represents a Spacelift **policy** - a collection of " +
			"customer-defined rules that are applied by Spacelift at one of the " +
			"decision points within the application.",

		ReadContext: dataPolicyRead,

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of the policy",
				Required:    true,
			},
			"body": {
				Type:        schema.TypeString,
				Description: "body of the policy",
				Computed:    true,
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "name of the policy",
				Computed:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "type of the policy",
				Computed:    true,
			},
		},
	}
}

func dataPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Policy *structs.Policy `graphql:"policy(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("policy_id"))}
	if err := meta.(*internal.Client).Query(ctx, "PolicyRead", &query, variables); err != nil {
		return diag.Errorf("could not query for policy: %v", err)
	}

	policy := query.Policy
	if policy == nil {
		return diag.Errorf("policy not found")
	}

	d.SetId(policy.ID)
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
