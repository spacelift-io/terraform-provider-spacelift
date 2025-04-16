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

func resourceBlueprintVersionedGroup() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_blueprint_versioned_group` represents a Spacelift BlueprintVersionedGroup, which allows you " +
			"to have groups in to which versioned blueprints can be assigned. " +
			"This resource is required for versioned blueprints. " +
			"For Terraform users it's preferable to use `spacelift_stack` instead. " +
			"This resource is mostly useful for those who do not use Terraform " +
			"to create stacks.",

		CreateContext: resourceBlueprintVersionedGroupCreate,
		ReadContext:   resourceBlueprintVersionedGroupRead,
		UpdateContext: resourceBlueprintVersionedGroupUpdate,
		DeleteContext: resourceBlueprintVersionedGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the BlueprintVersionedGroup",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space": {
				Type:             schema.TypeString,
				Description:      "ID of the space the BlueprintVersionedGroup is in",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the BlueprintVersionedGroup",
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels of the BlueprintVersionedGroup",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
		},
	}
}

func resourceBlueprintVersionedGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintVersionedGroup structs.BlueprintVersionedGroup `graphql:"blueprintVersionedGroupCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": BlueprintVersionedGroupCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "blueprintVersionedGroupCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create BlueprintVersionedGroup: %v", err)
	}

	d.SetId(mutation.BlueprintVersionedGroup.ID)

	return resourceBlueprintVersionedGroupRead(ctx, d, meta)
}

func resourceBlueprintVersionedGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BlueprintVersionedGroup *structs.BlueprintVersionedGroup `graphql:"blueprintVersionedGroup(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := meta.(*internal.Client).Query(ctx, "blueprintVersionedGroupRead", &query, variables); err != nil {
		return diag.Errorf("could not query for BlueprintVersionedGroup: %v", err)
	}

	if query.BlueprintVersionedGroup == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", query.BlueprintVersionedGroup.Name)
	d.Set("space", query.BlueprintVersionedGroup.Space.ID)
	d.Set("description", query.BlueprintVersionedGroup.Description)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range query.BlueprintVersionedGroup.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}

func resourceBlueprintVersionedGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintVersionedGroup structs.BlueprintVersionedGroup `graphql:"blueprintVersionedGroupUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    graphql.ID(d.Id()),
		"input": BlueprintVersionedGroupCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "blueprintVersionedGroupUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update BlueprintVersionedGroup: %v", err)
	}

	return resourceBlueprintVersionedGroupRead(ctx, d, meta)
}

func resourceBlueprintVersionedGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintVersionedGroup struct {
			ID string `graphql:"id"`
		} `graphql:"blueprintVersionedGroupDelete(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "blueprintVersionedGroupDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete BlueprintVersionedGroup: %v", err)
	}

	d.SetId("")

	return nil
}

func BlueprintVersionedGroupCreateInput(d *schema.ResourceData) structs.BlueprintVersionedGroupCreateInput {
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
