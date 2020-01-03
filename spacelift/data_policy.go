package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func dataPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataPolicyRead,

		Schema: map[string]*schema.Schema{
			"policy_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Immutable ID (slug) of the policy",
				Required:    true,
			},
			"body": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Body of the policy",
				Computed:    true,
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the policy",
				Computed:    true,
			},
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Type of the policy",
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
	if err := meta.(*Client).Query(&query, variables); err != nil {
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
