package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourcePluginTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_plugin_template` represents a Spacelift **plugin template** - " +
			"a reusable template that defines a plugin's behavior and can be instantiated " +
			"multiple times as plugins.",

		CreateContext: resourcePluginTemplateCreate,
		ReadContext:   resourcePluginTemplateRead,
		DeleteContext: resourcePluginTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the plugin template",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form plugin template description",
				Optional:    true,
				ForceNew:    true,
			},
			"manifest": {
				Type:             schema.TypeString,
				Description:      "The plugin manifest defining the plugin's behavior",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels to apply to the plugin template",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
				ForceNew: true,
			},
			"id": {
				Type:        schema.TypeString,
				Description: "Immutable ID of the plugin template",
				Computed:    true,
			},
			"is_global": {
				Type:        schema.TypeBool,
				Description: "Whether this is a global (Spacelift-provided) plugin template",
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

func resourcePluginTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreatePluginTemplate structs.PluginTemplate `graphql:"pluginTemplateCreate(input: $input)"`
	}

	input := structs.PluginTemplateCreateInput{
		Name:     toString(d.Get("name")),
		Manifest: toString(d.Get("manifest")),
	}

	if description, ok := d.GetOk("description"); ok {
		desc := graphql.String(description.(string))
		input.Description = &desc
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		input.Labels = &labels
	}

	variables := map[string]interface{}{"input": input}

	if err := meta.(*internal.Client).Mutate(ctx, "PluginTemplateCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create plugin template: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreatePluginTemplate.ID)

	return resourcePluginTemplateRead(ctx, d, meta)
}

func resourcePluginTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		PluginTemplate *structs.PluginTemplate `graphql:"pluginTemplate(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "PluginTemplateRead", &query, variables); err != nil {
		return diag.Errorf("could not query for plugin template: %v", err)
	}

	template := query.PluginTemplate
	if template == nil {
		d.SetId("")
		return nil
	}
	d.SetId(template.ID)

	d.Set("name", template.Name)
	d.Set("manifest", template.Manifest)
	d.Set("is_global", template.IsGlobal)
	d.Set("created_at", template.CreatedAt)
	d.Set("updated_at", template.UpdatedAt)

	if description := template.Description; description != nil {
		d.Set("description", *description)
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

func resourcePluginTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeletePluginTemplate *structs.PluginTemplate `graphql:"pluginTemplateDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "PluginTemplateDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete plugin template: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
