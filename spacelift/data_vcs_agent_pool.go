package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataVCSAgentPool() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_vcs_agent_pool` represents a Spacelift **VCS agent pool** - " +
			"a logical group of proxies allowing Spacelift to access private " +
			"VCS installations",

		ReadContext: dataVcsAgentPoolRead,

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
	}
}

func dataVcsAgentPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		VCSAgentPool *structs.VCSAgentPool `graphql:"vcsAgentPool(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("vcs_agent_pool_id"))}
	if err := meta.(*internal.Client).Query(ctx, "VCSAgentPoolRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the VCS agent pool: %v", err)
	}

	vcsAgentPool := query.VCSAgentPool
	if vcsAgentPool == nil {
		return diag.Errorf("VCS agent pool not found")
	}

	d.SetId(vcsAgentPool.ID)
	d.Set("name", vcsAgentPool.Name)

	if vcsAgentPool.Description != nil {
		d.Set("description", *vcsAgentPool.Description)
	} else {
		d.Set("description", nil)
	}

	return nil
}
