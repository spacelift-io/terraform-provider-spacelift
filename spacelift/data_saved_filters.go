package spacelift

import (
	"context"
	"fmt"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataSavedFilters() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_saved_filters` can find all saved filters that have certain type or name",

		ReadContext: dataFiltersRead,

		Schema: map[string]*schema.Schema{
			"filter_type": {
				Type:        schema.TypeString,
				Description: "filter type to look for",
				Optional:    true,
			},
			"filter_name": {
				Type:        schema.TypeString,
				Description: "filter name to look for",
				Optional:    true,
			},
			"filters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Globally unique ID of the saved filter",
							Computed:    true,
						},
						"is_public": {
							Type:        schema.TypeBool,
							Description: "Toggle whether the filter is public or not",
							Computed:    true,
						},
						"created_by": {
							Type:        schema.TypeString,
							Description: "Login of the user who created the saved filter",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the filter",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "Type describes the type of the filter. It is used to determine which view the filter is for",
							Computed:    true,
						},
						"data": {
							Type:        schema.TypeString,
							Description: "Data is the JSON representation of the filter data",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataFiltersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Filters []structs.SavedFilter `graphql:"savedFilters()"`
	}

	typeRaw, typeSpecified := d.GetOk("filter_type")
	requestedType := typeRaw.(string)
	nameRaw, nameSpecified := d.GetOk("filter_name")
	requestedName := nameRaw.(string)

	if err := meta.(*internal.Client).Query(ctx, "savedFilters", &query, nil); err != nil {
		return diag.Errorf("could not query for filters: %v", err)
	}

	var filters []interface{}
	for _, filter := range query.Filters {
		if typeSpecified && filter.Type != requestedType {
			continue
		}
		if nameSpecified && filter.Name != requestedName {
			continue
		}
		filters = append(filters, map[string]interface{}{
			"id":         filter.ID,
			"is_public":  filter.IsPublic,
			"name":       filter.Name,
			"type":       filter.Type,
			"created_by": filter.CreatedBy,
			"data":       filter.Data,
		})
	}

	d.SetId(fmt.Sprintf("filters/%s/%s", requestedType, requestedName))
	d.Set("filters", filters)

	return nil
}
