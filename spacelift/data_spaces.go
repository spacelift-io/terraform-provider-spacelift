package spacelift

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataSpaces() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_spaces` can find all spaces in the spacelift organization.",

		ReadContext: dataSpacesRead,

		Schema: map[string]*schema.Schema{
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "required labels to match",
				Optional:    true,
			},
			"spaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
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
				},
			},
		},
	}
}

func dataSpacesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Spaces []structs.Space `graphql:"spaces()"`
	}

	if err := meta.(*internal.Client).Query(ctx, "SpacesRead", &query, nil); err != nil {
		return diag.Errorf("could not query for space: %v", err)
	}

	var spaces []interface{}
	for _, space := range internal.FilterByRequiredLabels(d, query.Spaces, func(space structs.Space) []string { return space.Labels }) {
		spaces = append(spaces, map[string]interface{}{
			"space_id":         space.ID,
			"name":             space.Name,
			"description":      space.Description,
			"parent_space_id":  space.ParentSpace,
			"inherit_entities": space.InheritEntities,
			"labels":           space.Labels,
		})
	}

	d.SetId(fmt.Sprintf("spaces-%d", time.Now().UnixNano()))

	if err := d.Set("spaces", spaces); err != nil {
		return diag.Errorf("could not set stacks: %v", err)
	}

	return nil
}
