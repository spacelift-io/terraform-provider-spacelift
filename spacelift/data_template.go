package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_template` represents a Spacelift template (versioned blueprint), " +
			"which is a collection of blueprint versions that can be used to create stacks.",

		ReadContext: dataTemplateRead,

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:             schema.TypeString,
				Description:      "ID of the template",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the template",
				Computed:    true,
			},
			"space": {
				Type:        schema.TypeString,
				Description: "ID of the space the template is in",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the template",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels of the template",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"ulid": {
				Type:        schema.TypeString,
				Description: "Unique ULID of the template",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp when the template was created",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp when the template was last updated",
				Computed:    true,
			},
		},
	}
}

func dataTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BlueprintVersionedGroup *structs.BlueprintVersionedGroup `graphql:"blueprintVersionedGroup(id: $id)"`
	}

	templateID := d.Get("template_id")

	variables := map[string]interface{}{"id": toID(templateID)}
	if err := meta.(*internal.Client).Query(ctx, "TemplateRead", &query, variables); err != nil {
		return diag.Errorf("could not query for template: %v", err)
	}

	template := query.BlueprintVersionedGroup
	if template == nil {
		return diag.Errorf("could not find template %s", templateID)
	}

	d.SetId(template.ID)
	d.Set("name", template.Name)
	d.Set("space", template.Space.ID)
	d.Set("ulid", template.ULID)
	d.Set("created_at", template.CreatedAt)
	d.Set("updated_at", template.UpdatedAt)

	if template.Description == nil {
		d.Set("description", "")
	} else {
		d.Set("description", *template.Description)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range template.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}
