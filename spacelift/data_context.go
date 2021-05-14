package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataContext() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_context` represents a Spacelift **context** - " +
			"a collection of configuration elements (either environment variables or " +
			"mounted files) that can be administratively attached to multiple " +
			"stacks (`spacelift_stack`) or modules (`spacelift_module`) using " +
			"a context attachment (`spacelift_context_attachment`)`",

		ReadContext: dataContextRead,

		Schema: map[string]*schema.Schema{
			"context_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of the context",
				Required:    true,
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
		},
	}
}

func dataContextRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Context *structs.Context `graphql:"context(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("context_id"))}
	if err := meta.(*internal.Client).Query(ctx, "ContextRead", &query, variables); err != nil {
		return diag.Errorf("could not query for context: %v", err)
	}

	context := query.Context
	if context == nil {
		return diag.Errorf("context not found")
	}

	d.SetId(context.ID)
	d.Set("name", context.Name)

	if context.Description != nil {
		d.Set("description", *context.Description)
	} else {
		d.Set("description", nil)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range context.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}
