package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataAccount() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_account` represents the currently used Spacelift **account**",
		ReadContext: dataAccountRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "name of the account",
				Computed:    true,
			},
			"tier": {
				Type:        schema.TypeString,
				Description: "account billing tier",
				Computed:    true,
			},
		},
	}
}

func dataAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Name string `graphql:"name"`
		Tier string `graphql:"tier"`
	}

	if err := meta.(*internal.Client).Query(ctx, "AccountDetails", &query, nil); err != nil {
		d.SetId("")
		return diag.Errorf("could not query for account details: %v", err)
	}

	d.Set("name", query.Name)
	d.Set("tier", query.Tier)
	d.SetId("spacelift-account")

	return nil
}
