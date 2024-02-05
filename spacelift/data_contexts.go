package spacelift

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search/predicates"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataContexts() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_contexts` represents all the contexts in the Spacelift " +
			"account visible to the API user.",

		ReadContext: dataContextsRead,

		Schema: map[string]*schema.Schema{
			"labels": predicates.StringField("Require contexts to have one of the labels", 0),
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

	var labelFilters [][]string

	for _, labelFilter := range d.Get("labels").([]interface{}) {
		possibleValues := labelFilter.(map[string]interface{})["any_of"]
		if possibleValues == nil {
			continue
		}

		var labels []string
		for _, label := range possibleValues.([]interface{}) {
			labels = append(labels, label.(string))
		}

		labelFilters = append(labelFilters, labels)
	}

	wps := flattenDataContextsList(contexts, labelFilters)
	if err := d.Set("contexts", wps); err != nil {
		d.SetId("")
		return diag.Errorf("could not set contexts: %v", err)
	}

	return nil
}

func flattenDataContextsList(contexts []*structs.Context, labelFilters [][]string) []map[string]interface{} {
	wps := []map[string]interface{}{}

	for _, ctx := range contexts {
		if !matchesLabels(ctx.Labels, labelFilters) {
			continue
		}

		wps = append(wps, map[string]interface{}{
			"context_id":  ctx.ID,
			"description": ctx.Description,
			"labels":      ctx.Labels,
			"name":        ctx.Name,
			"space_id":    ctx.Space,
		})
	}

	return wps
}

func matchesLabels(labelSet []string, labelFilters [][]string) bool {
	if len(labelFilters) == 0 {
		return true
	}

	for _, labelFilter := range labelFilters {
		if len(labelFilter) == 0 {
			continue
		}

		var matchesFilter bool
		for _, label := range labelFilter {
			if slices.Contains(labelSet, label) {
				matchesFilter = true
			}
		}

		if !matchesFilter {
			return false
		}
	}

	return true
}
