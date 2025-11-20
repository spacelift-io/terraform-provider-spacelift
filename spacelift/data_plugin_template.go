package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataPluginTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_plugin_template` represents a Spacelift **plugin template** - " +
			"a reusable template that defines a plugin's behavior and can be instantiated " +
			"multiple times as plugins.",

		ReadContext: dataPluginTemplateRead,

		Schema: map[string]*schema.Schema{
			"plugin_template_id": {
				Type:             schema.TypeString,
				Description:      "ID of the plugin template",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the plugin template",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form plugin template description",
				Computed:    true,
			},
			"manifest": {
				Type:        schema.TypeString,
				Description: "The plugin manifest",
				Computed:    true,
			},
			"is_global": {
				Type:        schema.TypeBool,
				Description: "Whether this is a global (Spacelift-provided) plugin template",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels applied to the plugin template",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"parameters": {
				Type:        schema.TypeList,
				Description: "Parameters extracted from the manifest",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Parameter ID",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Parameter name",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "Parameter type",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Parameter description",
							Computed:    true,
						},
						"sensitive": {
							Type:        schema.TypeBool,
							Description: "Whether the parameter is sensitive",
							Computed:    true,
						},
						"required": {
							Type:        schema.TypeBool,
							Description: "Whether the parameter is required",
							Computed:    true,
						},
						"default": {
							Type:        schema.TypeString,
							Description: "Default value for the parameter",
							Computed:    true,
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeInt,
				Description: "Timestamp when the plugin template was created",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeInt,
				Description: "Timestamp when the plugin template was last updated",
				Computed:    true,
			},
		},
	}
}

func dataPluginTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		PluginTemplate *structs.PluginTemplate `graphql:"pluginTemplate(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("plugin_template_id"))}
	if err := meta.(*internal.Client).Query(ctx, "PluginTemplateRead", &query, variables); err != nil {
		return diag.Errorf("could not query for plugin template: %v", err)
	}

	template := query.PluginTemplate
	if template == nil {
		return diag.Errorf("plugin template not found")
	}

	d.SetId(template.ID)
	d.Set("name", template.Name)
	d.Set("manifest", template.Manifest)
	d.Set("is_global", template.IsGlobal)
	d.Set("created_at", template.CreatedAt)
	d.Set("updated_at", template.UpdatedAt)

	if template.Description != nil {
		d.Set("description", *template.Description)
	} else {
		d.Set("description", nil)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range template.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	// Set parameters
	parameters := make([]map[string]interface{}, len(template.Parameters))
	for i, param := range template.Parameters {
		p := map[string]interface{}{
			"id":        param.ID,
			"name":      param.Name,
			"type":      param.Type,
			"sensitive": param.Sensitive,
			"required":  param.Required,
		}

		if param.Description != nil {
			p["description"] = *param.Description
		}

		if param.Default != nil {
			p["default"] = *param.Default
		}

		parameters[i] = p
	}
	d.Set("parameters", parameters)

	return nil
}
