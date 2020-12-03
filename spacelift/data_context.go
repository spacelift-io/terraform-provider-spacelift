package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataContext() *schema.Resource {
	return &schema.Resource{
		Read: dataContextRead,

		Schema: map[string]*schema.Schema{
			"context_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of the context",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "free-form context description for users",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "name of the context",
				Computed:    true,
			},
		},
	}
}

func dataContextRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Context *structs.Context `graphql:"context(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("context_id"))}
	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for context")
	}

	context := query.Context
	if context == nil {
		return errors.New("context not found")
	}

	d.SetId(context.ID)
	d.Set("name", context.Name)

	if context.Description != nil {
		d.Set("description", *context.Description)
	} else {
		d.Set("description", nil)
	}

	return nil
}
