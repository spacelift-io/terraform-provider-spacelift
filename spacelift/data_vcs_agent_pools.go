package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataVCSAgentPools() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_vcs_agent_pools` represents the VCS agent pools assigned " +
			"to the Spacelift account.",

		ReadContext: dataVCSAgentPoolsRead,

		Schema: map[string]*schema.Schema{
			"vcs_agent_pools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcs_agent_pool_id": {
							Type:        schema.TypeString,
							Description: "ID of the VCS agent pool to retrieve",
							Required:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Free-form VCS agent pool description for users",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the VCS agent pool",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataVCSAgentPoolsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		VCSAgentPools []*structs.VCSAgentPool `graphql:"vcsAgentPools()"`
	}
	variables := map[string]interface{}{}

	if err := meta.(*internal.Client).Query(ctx, "VCSAgentPoolsRead", &query, variables); err != nil {
		return diag.Errorf("could not query for VCS Agent pools: %v", err)
	}

	d.SetId("spacelift-vcs-agent-pools")

	vcsAgentPools := query.VCSAgentPools
	if vcsAgentPools == nil {
		d.Set("vcs_agent_pools", nil)
		return nil
	}

	aps := flattenDataVCSAgentPoolsList(vcsAgentPools)
	if err := d.Set("vcs_agent_pools", aps); err != nil {
		d.SetId("")
		return diag.Errorf("could not set VCS agent pools: %v", err)
	}

	return nil
}

func flattenDataVCSAgentPoolsList(vcsAgentPools []*structs.VCSAgentPool) []map[string]interface{} {
	out := make([]map[string]interface{}, len(vcsAgentPools))

	for index, ap := range vcsAgentPools {
		var description *string

		if ap.Description != nil {
			description = ap.Description
		} else {
			description = nil
		}

		out[index] = map[string]interface{}{
			"vcs_agent_pool_id": ap.ID,
			"name":              ap.Name,
			"description":       description,
		}
	}

	return out
}
