package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataContexts() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_contexts` represents all the contexts created in " +
			"the Spacelift account.",

		ReadContext: dataContextsRead,

		Schema: map[string]*schema.Schema{
			"contexts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"context_id": {
							Type:             schema.TypeString,
							Description:      "immutable ID (slug) of the context",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "free-form context description for users",
							Computed:    true,
						},
						"labels": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "name of the context",
							Computed:    true,
						},
						"space_id": {
							Type:        schema.TypeString,
							Description: "ID (slug) of the space the context is in",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataContextsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var query struct {
		Contexts []*structs.Context `graphql:"contexts()"`
	}
	variables := map[string]interface{}{}

	if err := meta.(*internal.Client).Query(ctx, "ContextsRead", &query, variables); err != nil {
		return diag.Errorf("could not query for contexts: %v", err)
	}

	d.SetId("spacelift-contexts")

	contexts := query.Contexts
	if contexts == nil {
		d.Set("contexts", nil)
		return nil
	}

	wps := flattenDataContextsList(contexts)
	if err := d.Set("contexts", wps); err != nil {
		d.SetId("")
		return diag.Errorf("could not set contexts: %v", err)
	}

	return nil
}

func flattenDataContextsList(contexts []*structs.Context) []map[string]interface{} {
	wps := make([]map[string]interface{}, len(contexts))

	for index, ctx := range contexts {
		wps[index] = map[string]interface{}{
			"context_id":  ctx.ID,
			"description": ctx.Description,
			"labels":      ctx.Labels,
			"name":        ctx.Name,
			"space_id":    ctx.Space,
		}
	}

	return wps
}
