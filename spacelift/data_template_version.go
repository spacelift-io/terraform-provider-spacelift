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

func dataTemplateVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	templateID := d.Get("template_id").(string)
	versionID := d.Get("version_id").(string)

	var query struct {
		Template *struct {
			TemplateVersion *structs.TemplateVersion `graphql:"templateVersion(id: $versionId)"`
		} `graphql:"template(id: $templateId)"`
	}

	variables := map[string]any{
		"templateId": graphql.ID(templateID),
		"versionId":  graphql.ID(versionID),
	}

	if err := meta.(*internal.Client).Query(ctx, "TemplateVersionRead", &query, variables); err != nil {
		return diag.Errorf("could not query for template version: %v", err)
	}

	if query.Template == nil {
		return diag.Errorf("could not find template %s", templateID)
	}

	if query.Template.TemplateVersion == nil {
		return diag.Errorf("could not find template version %s in template %s", versionID, templateID)
	}

	tv := query.Template.TemplateVersion

	d.SetId(tv.ID)
	d.Set("state", tv.State)
	d.Set("ulid", tv.ULID)
	d.Set("created_at", tv.CreatedAt)
	d.Set("updated_at", tv.UpdatedAt)

	if tv.Version != nil {
		d.Set("version_number", *tv.Version)
	}

	if tv.RawTemplate == nil {
		d.Set("template", "")
	} else {
		d.Set("template", *tv.RawTemplate)
	}

	if tv.Instructions == nil {
		d.Set("instructions", "")
	} else {
		d.Set("instructions", *tv.Instructions)
	}

	if tv.PublishedAt != nil {
		d.Set("published_at", *tv.PublishedAt)
	}

	return nil
}
