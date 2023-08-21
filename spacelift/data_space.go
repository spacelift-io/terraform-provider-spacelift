package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataSpace() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_space` represents a Spacelift **space** - " +
			"a collection of resources such as stacks, modules, policies, etc. Allows for more granular access control. Can have a parent space.",

		ReadContext: dataSpaceRead,

		Schema: map[string]*schema.Schema{
			"space_id": {
				Type:             schema.TypeString,
				Description:      "immutable ID (slug) of the space",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"parent_space_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of parent space",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "free-form space description for users",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "name of the space",
				Computed:    true,
			},
			"inherit_entities": {
				Type:        schema.TypeBool,
				Description: "indication whether access to this space inherits read access to entities from the parent space",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "list of labels describing a space",
				Computed:    true,
			},
		},
	}
}

func dataSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Space *structs.Space `graphql:"space(id: $id)"`
	}

	spaceID := d.Get("space_id")

	variables := map[string]interface{}{"id": toID(spaceID)}
	if err := meta.(*internal.Client).Query(ctx, "SpaceRead", &query, variables); err != nil {
		return diag.Errorf("could not query for space: %v", err)
	}

	space := query.Space
	if space == nil {
		return diag.Errorf("could not find space %s", spaceID)
	}

	d.SetId(space.ID)
	d.Set("name", space.Name)
	d.Set("description", space.Description)
	d.Set("inherit_entities", space.InheritEntities)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range space.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if space.ParentSpace != nil {
		d.Set("parent_space_id", *space.ParentSpace)
	}

	return nil
}
