package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataPolicies() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_policies` can find all policies that have certain labels.",

		ReadContext: dataPoliciesRead,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Description: "required policy type",
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "required labels to match",
				Optional:    true,
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the policy",
							Computed:    true,
						},
						"labels": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the policy",
							Computed:    true,
						},
						"space_id": {
							Type:        schema.TypeString,
							Description: "ID (slug) of the space the policy is in",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "Type of the policy",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "description of the policy",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataPoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(fmt.Sprintf("policies/%s/%s", d.Get("type").(string), d.Get("labels").(*schema.Set).List()))
	var query struct {
		Policies []struct {
			ID          string   `graphql:"id"`
			Labels      []string `graphql:"labels"`
			Name        string   `graphql:"name"`
			Type        string   `graphql:"type"`
			Space       string   `graphql:"space"`
			Description string   `graphql:"description"`
		} `graphql:"policies()"`
	}

	if err := meta.(*internal.Client).Query(ctx, "PoliciesRead", &query, nil); err != nil {
		return diag.Errorf("could not query for policy: %v", err)
	}

	typeRaw, typeSpecified := d.GetOk("type")
	requestedType := typeRaw.(string)
	labelsRaw, labelsSpecified := d.GetOk("labels")
	requestedLabels := labelsRaw.(*schema.Set).List()

	var policies []interface{}
	for _, policy := range query.Policies {
		if typeSpecified && policy.Type != requestedType {
			continue
		}
		if labelsSpecified {
			found := false
			for _, required := range requestedLabels {
				found = false
				for _, existing := range policy.Labels {
					if required == existing {
						found = true
					}
				}
				if !found {
					break // we didn't find a required label
				}
			}
			if !found {
				continue
			}
		}
		policies = append(policies, map[string]interface{}{
			"id":          policy.ID,
			"labels":      policy.Labels,
			"name":        policy.Name,
			"type":        policy.Type,
			"space_id":    policy.Space,
			"description": policy.Description,
		})
	}

	d.Set("policies", policies)

	return nil
}
