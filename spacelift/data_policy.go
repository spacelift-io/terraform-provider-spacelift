package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataPolicyRead,

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

func dataPolicyRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Policy *structs.Policy `graphql:"policy(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("policy_id"))}
	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for policy")
	}

	policy := query.Policy
	if policy == nil {
		return errors.New("policy not found")
	}

	d.SetId(policy.ID)
	d.Set("name", policy.Name)
	d.Set("body", policy.Body)
	d.Set("type", policy.Type)

	return nil
}
