package spacelift

import (
	"context"
	"fmt"

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

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			// Only validate on updates (not creates)
			if diff.Id() == "" {
				return nil
			}

			oldState, newState := diff.GetChange("state")
			if oldState == "PUBLISHED" && oldState != newState {
				return fmt.Errorf("cannot change the state of a published template version")
			}
			if oldState.(string) == "PUBLISHED" && diff.HasChange("template") {
				return fmt.Errorf("cannot modify 'template' field when the template version is already PUBLISHED")
			}

			return nil
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceTemplateVersionImport,
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
	templateID := d.Get("template_id").(string)
	versionID := d.Id()

	var query struct {
		BlueprintVersionedGroup *struct {
			BlueprintVersion *structs.Blueprint `graphql:"blueprintVersion(id: $versionId)"`
		} `graphql:"blueprintVersionedGroup(id: $templateId)"`
	}

	variables := map[string]interface{}{
		"templateId": graphql.ID(templateID),
		"versionId":  graphql.ID(versionID),
	}

	if err := meta.(*internal.Client).Query(ctx, "TemplateVersionRead", &query, variables); err != nil {
		return diag.Errorf("could not query for template version: %v", err)
	}

	if query.BlueprintVersionedGroup == nil || query.BlueprintVersionedGroup.BlueprintVersion == nil {
		d.SetId("")
		return nil
	}

	blueprint := query.BlueprintVersionedGroup.BlueprintVersion

	d.Set("state", blueprint.State)
	d.Set("ulid", blueprint.ULID)
	d.Set("created_at", blueprint.CreatedAt)
	d.Set("updated_at", blueprint.UpdatedAt)

	if blueprint.Version != nil {
		d.Set("version_number", *blueprint.Version)
	}

	if blueprint.RawTemplate == nil {
		d.Set("template", "")
	} else {
		d.Set("template", *blueprint.RawTemplate)
	}

	if blueprint.Instructions == nil {
		d.Set("instructions", "")
	} else {
		d.Set("instructions", *blueprint.Instructions)
	}

	if blueprint.PublishedAt != nil {
		d.Set("published_at", *blueprint.PublishedAt)
	}

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

func resourceTemplateVersionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()

	// ULID is always 26 characters, format is: template_id-ulid
	// We need to find the last dash that separates template_id from ulid
	if len(importID) < 28 { // minimum: "x-" + 26 char ULID
		return nil, fmt.Errorf("invalid import ID format: %s, expected format: template_id-version_ulid", importID)
	}

	// Split from the right: template_id is everything before the last 27 characters (dash + 26 char ULID)
	templateID := importID[:len(importID)-27]
	versionULID := importID[len(importID)-26:]

	// Validate that we have a dash separator
	if importID[len(importID)-27] != '-' {
		return nil, fmt.Errorf("invalid import ID format: %s, expected format: template_id-version_ulid", importID)
	}

	var query struct {
		BlueprintVersionedGroup *struct {
			BlueprintVersion *structs.Blueprint `graphql:"blueprintVersion(id: $versionId)"`
		} `graphql:"blueprintVersionedGroup(id: $templateId)"`
	}

	variables := map[string]interface{}{
		"templateId": graphql.ID(templateID),
		"versionId":  graphql.ID(versionULID),
	}

	if err := meta.(*internal.Client).Query(ctx, "TemplateVersionImport", &query, variables); err != nil {
		return nil, fmt.Errorf("could not query for template version: %v", err)
	}

	if query.BlueprintVersionedGroup == nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	if query.BlueprintVersionedGroup.BlueprintVersion == nil {
		return nil, fmt.Errorf("template version not found: %s in template %s", versionULID, templateID)
	}

	d.Set("template_id", templateID)
	d.SetId(query.BlueprintVersionedGroup.BlueprintVersion.ID)

	return []*schema.ResourceData{d}, nil
}

func templateVersionCreateInput(d *schema.ResourceData) structs.BlueprintVersionCreateInput {
	var input structs.BlueprintVersionCreateInput

	input.BlueprintID = graphql.ID(d.Get("template_id").(string))
	input.State = graphql.String(d.Get("state").(string))
	input.VersionNumber = graphql.String(d.Get("version_number").(string))
	input.Labels = []graphql.String{}

	if instructions, ok := d.GetOk("instructions"); ok {
		input.Instructions = toOptionalString(instructions)
	}

	if template, ok := d.GetOk("template"); ok {
		input.Template = toOptionalString(template)
	}

	return input
}

func templateVersionUpdateInput(d *schema.ResourceData) structs.BlueprintVersionUpdateInput {
	var input structs.BlueprintVersionUpdateInput

	input.State = graphql.String(d.Get("state").(string))
	input.VersionNumber = graphql.String(d.Get("version_number").(string))
	input.Labels = []graphql.String{}

	if instructions, ok := d.GetOk("instructions"); ok {
		input.Instructions = toOptionalString(instructions)
	}

	if template, ok := d.GetOk("template"); ok {
		input.Template = toOptionalString(template)
	}

	return input
}
