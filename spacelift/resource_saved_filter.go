package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceSavedFilter() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_saved_filter` represents a Spacelift **filter** - a collection of " +
			"customer-defined criteria that are applied by Spacelift at one of the " +
			"decision points within the application.",

		CreateContext: resourceSavedFilterCreate,
		ReadContext:   resourceSavedFilterRead,
		UpdateContext: resourceSavedFilterUpdate,
		DeleteContext: resourceSavedFilterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Globally unique ID of the saved filter",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the saved filter",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"data": {
				Type:             schema.TypeString,
				Description:      "Data is the JSON representation of the filter data",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"is_public": {
				Type:        schema.TypeBool,
				Description: "Toggle whether the filter is public or not",
				Required:    true,
			},
			"created_by": {
				Type:        schema.TypeString,
				Description: "Login of the user who created the saved filter",
				Computed:    true,
			},
			"type": {
				Type: schema.TypeString,
				Description: "Type describes the type of the filter. It is used to determine which view the filter is for. " +
					"Possible values are `stacks`, `blueprints`, `contexts`, `webhooks`.",
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					structs.SavedFilterTypes,
					false, // case-sensitive match
				),
			},
		},
	}
}

func resourceSavedFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateFilter structs.SavedFilter `graphql:"savedFilterCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": savedFilterInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "savedFilterCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create saved filter %v: %v", toString(d.Get("name")), internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateFilter.ID)

	return resourceSavedFilterRead(ctx, d, meta)
}

func savedFilterInput(d *schema.ResourceData) structs.SavedFilterInput {
	return structs.SavedFilterInput{
		Name:     toString(d.Get("name")),
		IsPublic: toBool(d.Get("is_public")),
		Data:     toString(d.Get("data")),
		Type:     toString(d.Get("type")),
	}
}

func resourceSavedFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Filter *structs.SavedFilter `graphql:"savedFilter(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Id()),
	}
	if err := meta.(*internal.Client).Query(ctx, "savedFilter", &query, variables); err != nil {
		return diag.Errorf("could not query for saved filter: %v", err)
	}

	filter := query.Filter
	if filter == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", filter.Name)
	d.Set("data", filter.Data)
	d.Set("type", filter.Type)
	d.Set("is_public", filter.IsPublic)
	d.Set("created_by", filter.CreatedBy)
	return nil
}

func resourceSavedFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateFilter structs.SavedFilter `graphql:"savedFilterUpdate(id: $id, name: $name, data: $data, isPublic: $isPublic)"`
	}

	variables := map[string]interface{}{
		"id":       toID(d.Id()),
		"name":     toString(d.Get("name")),
		"isPublic": toBool(d.Get("is_public")),
		"data":     toString(d.Get("data")),
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "savedFilterUpdate", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update saved filter: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourceSavedFilterRead(ctx, d, meta)...)
}

func resourceSavedFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteFilter *structs.SavedFilter `graphql:"savedFilterDelete(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "SavedFilterDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete saved filter: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
