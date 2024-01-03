package spacelift

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataSavedFilter() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_saved_filter` represents a Spacelift saved filter.",

		ReadContext: dataFilterRead,

		Schema: map[string]*schema.Schema{
			"filter_id": {
				Type:        schema.TypeString,
				Description: " immutable ID (slug) of the filter",
				Required:    true,
			},
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
	}
}

func dataFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Filter struct {
			ID        string `graphql:"id"`
			IsPublic  bool   `graphql:"isPublic"`
			CreatedBy string `graphql:"createdBy"`
			Name      string `graphql:"name"`
			Type      string `graphql:"type"`
			Data      string `graphql:"data"`
		} `graphql:"savedFilter(id: $id)"`
	}

	variables := map[string]interface{}{"id": d.Get("filter_id")}

	if err := meta.(*internal.Client).Query(ctx, "savedFilter", &query, variables); err != nil {
		return diag.Errorf("could not query for filter: %v", err)
	}

	d.SetId(query.Filter.ID)
	d.Set("is_public", query.Filter.IsPublic)
	d.Set("name", query.Filter.Name)
	d.Set("type", query.Filter.Type)
	d.Set("created_by", query.Filter.CreatedBy)
	d.Set("data", query.Filter.Data)

	return nil
}
