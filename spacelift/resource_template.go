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

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_template` represents a Spacelift template (versioned blueprint), " +
			"which is a collection of blueprint versions that can be used to create stacks.",

		CreateContext: resourceTemplateCreate,
		ReadContext:   resourceTemplateRead,
		UpdateContext: resourceTemplateUpdate,
		DeleteContext: resourceTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the template",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space": {
				Type:             schema.TypeString,
				Description:      "ID of the space the template is in",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the template",
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels of the template",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
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

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintVersionedGroup structs.BlueprintVersionedGroup `graphql:"blueprintVersionedGroupCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": templateCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "TemplateCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create template: %v", err)
	}

	d.SetId(mutation.BlueprintVersionedGroup.ID)

	return resourceTemplateRead(ctx, d, meta)
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BlueprintVersionedGroup *structs.BlueprintVersionedGroup `graphql:"blueprintVersionedGroup(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := meta.(*internal.Client).Query(ctx, "TemplateRead", &query, variables); err != nil {
		return diag.Errorf("could not query for template: %v", err)
	}

	if query.BlueprintVersionedGroup == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", query.BlueprintVersionedGroup.Name)
	d.Set("space", query.BlueprintVersionedGroup.Space.ID)
	d.Set("ulid", query.BlueprintVersionedGroup.ULID)
	d.Set("created_at", query.BlueprintVersionedGroup.CreatedAt)
	d.Set("updated_at", query.BlueprintVersionedGroup.UpdatedAt)

	if query.BlueprintVersionedGroup.Description == nil {
		d.Set("description", "")
	} else {
		d.Set("description", *query.BlueprintVersionedGroup.Description)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range query.BlueprintVersionedGroup.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintVersionedGroup structs.BlueprintVersionedGroup `graphql:"blueprintVersionedGroupUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    graphql.ID(d.Id()),
		"input": templateUpdateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "TemplateUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update template: %v", err)
	}

	return resourceTemplateRead(ctx, d, meta)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintVersionedGroup *structs.BlueprintVersionedGroup `graphql:"blueprintVersionedGroupDelete(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "TemplateDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete template: %v", err)
	}

	d.SetId("")

	return nil
}

func templateCreateInput(d *schema.ResourceData) structs.BlueprintVersionedGroupCreateInput {
	var input structs.BlueprintVersionedGroupCreateInput

	input.Space = graphql.ID(d.Get("space").(string))
	input.Name = graphql.String(d.Get("name").(string))

	if description, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(description)
	}

	input.Labels = []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			input.Labels = append(input.Labels, graphql.String(label.(string)))
		}
	}

	return input
}

func templateUpdateInput(d *schema.ResourceData) structs.BlueprintVersionedGroupUpdateInput {
	var input structs.BlueprintVersionedGroupUpdateInput

	input.Space = graphql.ID(d.Get("space").(string))
	input.Name = graphql.String(d.Get("name").(string))

	if description, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(description)
	}

	input.Labels = []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			input.Labels = append(input.Labels, graphql.String(label.(string)))
		}
	}

	return input
}
