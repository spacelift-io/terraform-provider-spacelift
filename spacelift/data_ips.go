package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataIPs() *schema.Resource {
	return &schema.Resource{
		ReadContext: ipsRead,
		Schema: map[string]*schema.Schema{
			"ips": {
				Type:        schema.TypeSet,
				Description: "the list of spacelift.io outgoing IP addresses",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
		},
	}
}

func ipsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("spacelift-ips")

	var query struct {
		IPs []string `graphql:"outgoingIPAddresses"`
	}

	if err := meta.(*internal.Client).Query(ctx, &query, nil); err != nil {
		d.SetId("")
		return diag.Errorf("could not query for outgoing IP addresses: %v", err)
	}

	ips := schema.NewSet(schema.HashString, []interface{}{})
	for _, ip := range query.IPs {
		ips.Add(ip)
	}

	d.Set("ips", ips)

	return nil
}
