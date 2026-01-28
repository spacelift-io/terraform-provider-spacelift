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

func dataTemplateVersion() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_template_version` represents a version of a Spacelift template. " +
			"Each template can have multiple versions, each with its own state (DRAFT or PUBLISHED) and template body.",

		ReadContext: dataTemplateVersionRead,

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:             schema.TypeString,
				Description:      "ID of the template this version belongs to",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"version_id": {
				Type:             schema.TypeString,
				Description:      "ID of the template version",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"version_number": {
				Type:        schema.TypeString,
				Description: "Version number (e.g., \"1.0.0\")",
				Computed:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "State of the template version. Value can be `DRAFT` or `PUBLISHED`.",
				Computed:    true,
			},
			"template": {
				Type:        schema.TypeString,
				Description: "Body of the template.",
				Computed:    true,
			},
			"instructions": {
				Type:        schema.TypeString,
				Description: "Instructions for the template version",
				Computed:    true,
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

func dataTemplateVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	templateID := d.Get("template_id").(string)
	versionID := d.Get("version_id").(string)

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

	if query.BlueprintVersionedGroup == nil {
		return diag.Errorf("could not find template %s", templateID)
	}

	if query.BlueprintVersionedGroup.BlueprintVersion == nil {
		return diag.Errorf("could not find template version %s in template %s", versionID, templateID)
	}

	blueprint := query.BlueprintVersionedGroup.BlueprintVersion

	d.SetId(blueprint.ID)
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
