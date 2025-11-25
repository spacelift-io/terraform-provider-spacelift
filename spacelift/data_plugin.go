package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataPlugin() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_plugin` represents a Spacelift **plugin** - " +
			"an instance of a plugin template that can be used to extend " +
			"Spacelift functionality.",

		ReadContext: dataPluginRead,

		Schema: map[string]*schema.Schema{
			"plugin_id": {
				Type:             schema.TypeString,
				Description:      "ID of the plugin",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the plugin",
				Computed:    true,
			},
			"plugin_template_id": {
				Type:        schema.TypeString,
				Description: "ID of the plugin template",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels applied to the plugin",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID of the space the plugin is in",
				Computed:    true,
			},
			"stack_label": {
				Type:        schema.TypeString,
				Description: "Label used when attaching the plugin to stacks",
				Computed:    true,
			},
		},
	}
}

func dataPluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Plugin *structs.Plugin `graphql:"plugin(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("plugin_id"))}
	if err := meta.(*internal.Client).Query(ctx, "PluginRead", &query, variables); err != nil {
		return diag.Errorf("could not query for plugin: %v", err)
	}

	plugin := query.Plugin
	if plugin == nil {
		return diag.Errorf("plugin not found")
	}

	d.SetId(plugin.ID)
	d.Set("name", plugin.Name)
	d.Set("stack_label", plugin.LabelIdentifier)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range plugin.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	d.Set("space_id", plugin.SpaceDetails.ID)

	if plugin.PluginTemplateDetails != nil {
		d.Set("plugin_template_id", plugin.PluginTemplateDetails.ID)
	}

	return nil
}
