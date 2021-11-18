package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceVCSAgentPool() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_vcs_agent_pool` represents a Spacelift **VCS agent pool** - " +
			"a logical group of proxies allowing Spacelift to access private " +
			"VCS installations",

		CreateContext: resourceVCSAgentPoolCreate,
		ReadContext:   resourceVCSAgentPoolRead,
		UpdateContext: resourceVCSAgentPoolUpdate,
		DeleteContext: resourceVCSAgentPoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form VCS agent pool description for users",
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the VCS agent pool, must be unique within an account",
				Required:    true,
			},
			"config": {
				Type:        schema.TypeString,
				Description: "VCS agent pool configuration, encoded using base64",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceVCSAgentPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateVCSAgentPool structs.VCSAgentPool `graphql:"vcsAgentPoolCreate(name: $name, description: $description)"`
	}

	variables := map[string]interface{}{
		"name":        toString(d.Get("name")),
		"description": (*graphql.String)(nil),
	}

	if description, ok := d.GetOk("description"); ok {
		variables["description"] = toOptionalString(description)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "VCSAgentPoolCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create the VCS agent pool: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateVCSAgentPool.ID)
	d.Set("config", mutation.CreateVCSAgentPool.Config)

	return nil
}

func resourceVCSAgentPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		VCSAgentPool *structs.VCSAgentPool `graphql:"vcsAgentPool(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "VCSAgentPoolRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the VCS agent pool: %v", err)
	}

	vcsAgentPool := query.VCSAgentPool
	if vcsAgentPool == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", vcsAgentPool.Name)

	description := vcsAgentPool.Description
	if description != nil {
		d.Set("description", *description)
	} else {
		d.Set("description", nil)
	}

	return nil
}

func resourceVCSAgentPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateVCSAgentPool structs.VCSAgentPool `graphql:"vcsAgentPoolUpdate(id: $id, name: $name, description: $description)"`
	}

	variables := map[string]interface{}{
		"id":          toID(d.Id()),
		"name":        toString(d.Get("name")),
		"description": (*graphql.String)(nil),
	}

	if description, ok := d.GetOk("description"); ok {
		variables["description"] = toOptionalString(description)
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "VCSAgentPoolUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update the VCS agent pool: %v", internal.FromSpaceliftError(err))...)
	}

	ret = append(ret, resourceVCSAgentPoolRead(ctx, d, meta)...)

	return ret
}

func resourceVCSAgentPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteVCSAgentPool *structs.VCSAgentPool `graphql:"vcsAgentPoolDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "VCSAgentPoolDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete context: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
