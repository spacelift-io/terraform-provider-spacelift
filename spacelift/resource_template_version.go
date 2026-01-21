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

func resourceTemplateVersion() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_template_version` represents a version of a Spacelift template. " +
			"Each template can have multiple versions, each with its own state (DRAFT or PUBLISHED) and template body.",

		CreateContext: resourceTemplateVersionCreate,
		ReadContext:   resourceTemplateVersionRead,
		UpdateContext: resourceTemplateVersionUpdate,
		DeleteContext: resourceTemplateVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:             schema.TypeString,
				Description:      "ID of the template this version belongs to",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"version_number": {
				Type:             schema.TypeString,
				Description:      "Version number (e.g., \"1.0.0\")",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"state": {
				Type:             schema.TypeString,
				Description:      "State of the template version. Value can be `DRAFT` or `PUBLISHED`.",
				Required:         true,
				ValidateDiagFunc: validateStateEnum,
			},
			"template": {
				Type:        schema.TypeString,
				Description: "Body of the template. Required if `state` is set to `PUBLISHED`.",
				Optional:    true,
			},
			"instructions": {
				Type:        schema.TypeString,
				Description: "Instructions for the template version",
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels of the template version",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"ulid": {
				Type:        schema.TypeString,
				Description: "Unique ULID of the template version",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp when the template version was created",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp when the template version was last updated",
				Computed:    true,
			},
			"published_at": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp when the template version was published",
				Computed:    true,
			},
		},
	}
}

func resourceTemplateVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		Blueprint structs.Blueprint `graphql:"blueprintVersionCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": templateVersionCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "TemplateVersionCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create template version: %v", err)
	}

	d.SetId(mutation.Blueprint.ID)

	return resourceTemplateVersionRead(ctx, d, meta)
}

func resourceTemplateVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Blueprint *structs.Blueprint `graphql:"blueprint(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := meta.(*internal.Client).Query(ctx, "TemplateVersionRead", &query, variables); err != nil {
		return diag.Errorf("could not query for template version: %v", err)
	}

	if query.Blueprint == nil {
		d.SetId("")
		return nil
	}

	d.Set("state", query.Blueprint.State)
	d.Set("ulid", query.Blueprint.ULID)
	d.Set("created_at", query.Blueprint.CreatedAt)
	d.Set("updated_at", query.Blueprint.UpdatedAt)

	if query.Blueprint.Version != nil {
		d.Set("version_number", *query.Blueprint.Version)
	}

	if query.Blueprint.RawTemplate == nil {
		d.Set("template", "")
	} else {
		d.Set("template", *query.Blueprint.RawTemplate)
	}

	if query.Blueprint.Instructions == nil {
		d.Set("instructions", "")
	} else {
		d.Set("instructions", *query.Blueprint.Instructions)
	}

	if query.Blueprint.PublishedAt != nil {
		d.Set("published_at", *query.Blueprint.PublishedAt)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range query.Blueprint.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}

func resourceTemplateVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		Blueprint structs.Blueprint `graphql:"blueprintVersionUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    graphql.ID(d.Id()),
		"input": templateVersionUpdateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "TemplateVersionUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update template version: %v", err)
	}

	return resourceTemplateVersionRead(ctx, d, meta)
}

func resourceTemplateVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		Blueprint *structs.Blueprint `graphql:"blueprintVersionDelete(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "TemplateVersionDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete template version: %v", err)
	}

	d.SetId("")

	return nil
}

func templateVersionCreateInput(d *schema.ResourceData) structs.BlueprintVersionCreateInput {
	var input structs.BlueprintVersionCreateInput

	input.BlueprintID = graphql.ID(d.Get("template_id").(string))
	input.State = graphql.String(d.Get("state").(string))
	input.VersionNumber = graphql.String(d.Get("version_number").(string))

	if instructions, ok := d.GetOk("instructions"); ok {
		input.Instructions = toOptionalString(instructions)
	}

	if template, ok := d.GetOk("template"); ok {
		input.Template = toOptionalString(template)
	}

	input.Labels = []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			input.Labels = append(input.Labels, graphql.String(label.(string)))
		}
	}

	return input
}

func templateVersionUpdateInput(d *schema.ResourceData) structs.BlueprintVersionUpdateInput {
	var input structs.BlueprintVersionUpdateInput

	input.State = graphql.String(d.Get("state").(string))
	input.VersionNumber = graphql.String(d.Get("version_number").(string))

	if instructions, ok := d.GetOk("instructions"); ok {
		input.Instructions = toOptionalString(instructions)
	}

	if template, ok := d.GetOk("template"); ok {
		input.Template = toOptionalString(template)
	}

	input.Labels = []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			input.Labels = append(input.Labels, graphql.String(label.(string)))
		}
	}

	return input
}
