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

func resourcePlugin() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_plugin` represents a Spacelift **plugin** - " +
			"an instance of a plugin template that can be used to extend " +
			"Spacelift functionality.",

		CreateContext: resourcePluginCreate,
		ReadContext:   resourcePluginRead,
		DeleteContext: resourcePluginDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the plugin",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
				ForceNew:         true,
			},
			"plugin_template_id": {
				Type:             schema.TypeString,
				Description:      "ID of the plugin template to use",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"parameters": {
				Type:        schema.TypeMap,
				Description: "Map of parameter ids to values.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels to apply to the plugin",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
				ForceNew: true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID of the space the plugin is in",
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"stack_label": {
				Type:             schema.TypeString,
				Description:      "Label to use when attaching the plugin to stacks",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"id": {
				Type:        schema.TypeString,
				Description: "Immutable ID of the plugin",
				Computed:    true,
			},
		},
	}
}

func resourcePluginCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		InstallPlugin structs.Plugin `graphql:"pluginInstall(input: $input)"`
	}

	input := structs.PluginInstallInput{
		Name:             toString(d.Get("name")),
		PluginTemplateID: toID(d.Get("plugin_template_id")),
		LabelIdentifier:  toString(d.Get("stack_label")),
	}

	// Always fetch the plugin template to validate parameters and get definitions
	var templateQuery struct {
		PluginTemplate *structs.PluginTemplate `graphql:"pluginTemplate(id: $id)"`
	}

	templateVars := map[string]interface{}{"id": toID(d.Get("plugin_template_id"))}
	if err := meta.(*internal.Client).Query(ctx, "PluginTemplateReadForParams", &templateQuery, templateVars); err != nil {
		return diag.Errorf("could not query for plugin template: %v", err)
	}

	if templateQuery.PluginTemplate == nil {
		return diag.Errorf("plugin template not found")
	}

	// Get user-provided parameters (may be empty)
	var params map[string]interface{}
	if parametersMap, ok := d.GetOk("parameters"); ok {
		params = parametersMap.(map[string]interface{})
	} else {
		params = make(map[string]interface{})
	}

	// Build a map of valid template parameter IDs for validation
	validParamIDs := make(map[string]bool)
	for _, templateParam := range templateQuery.PluginTemplate.Parameters {
		validParamIDs[templateParam.ID] = true
	}

	// Validate that all user-provided parameters exist in the template
	for userParamID := range params {
		if !validParamIDs[userParamID] {
			return diag.Errorf("unknown parameter '%s': not defined in plugin template", userParamID)
		}
	}

	// Build parameter list based on template parameter order
	paramList := []structs.PluginInstallParameterInput{}

	for _, templateParam := range templateQuery.PluginTemplate.Parameters {
		// Look for this parameter in the user's input by ID (not name!)
		if val, exists := params[templateParam.ID]; exists {
			// Convert the value to string and create parameter input
			paramList = append(paramList, structs.PluginInstallParameterInput{
				ID:    graphql.String(templateParam.ID),
				Value: graphql.String(convertToString(val)),
			})
		} else if templateParam.Required {
			return diag.Errorf("required parameter '%s' is missing", templateParam.ID)
		} else if templateParam.Default != nil {
			// Use default value if provided
			paramList = append(paramList, structs.PluginInstallParameterInput{
				ID:    graphql.String(templateParam.ID),
				Value: graphql.String(*templateParam.Default),
			})
		}
	}

	if len(paramList) > 0 {
		input.Parameters = &paramList
	}

	// Set space - default to "root" if not provided (space is required by API)
	if spaceID, ok := d.GetOk("space_id"); ok {
		input.Space = toID(spaceID)
	} else {
		input.Space = graphql.ID("root")
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		input.Labels = &labels
	}

	variables := map[string]interface{}{"input": input}

	if err := meta.(*internal.Client).Mutate(ctx, "PluginInstall", &mutation, variables); err != nil {
		return diag.Errorf("could not install plugin: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.InstallPlugin.ID)

	return resourcePluginRead(ctx, d, meta)
}

func resourcePluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Plugin *structs.Plugin `graphql:"plugin(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "PluginRead", &query, variables); err != nil {
		return diag.Errorf("could not query for plugin: %v", err)
	}

	plugin := query.Plugin
	if plugin == nil {
		d.SetId("")
		return nil
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

func resourcePluginDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeletePlugin *structs.Plugin `graphql:"pluginDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "PluginDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete plugin: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
