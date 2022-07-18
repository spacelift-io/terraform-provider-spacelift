package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataSpace() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_space` represents a Spacelift **space** -",

		ReadContext: dataSpaceRead,

		Schema: map[string]*schema.Schema{
			"space_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of the space",
				Required:    true,
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
				Description: "indication whether this space inherits entities from the parent space",
				Computed:    true,
			},
		},
	}
}

func dataSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Space *structs.Space `graphql:"space(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("space_id"))}
	if err := meta.(*internal.Client).Query(ctx, "SpaceRead", &query, variables); err != nil {
		return diag.Errorf("could not query for space: %v", err)
	}

	space := query.Space
	if space == nil {
		return diag.Errorf("space not found")
	}

	d.SetId(space.ID)
	d.Set("name", space.Name)
	d.Set("description", space.Description)
	d.Set("inherit_entities", space.InheritEntities)
	if space.ParentSpace != nil {
		d.Set("parent_space_id", *space.ParentSpace)
	}

	return nil
}
