package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataIPs() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_ips` returns the list of Spacelift's outgoing IP addresses, " +
			"which you can use to whitelist connections coming from the " +
			"Spacelift's \"mothership\". **NOTE:** this does not include the IP addresses " +
			"of the workers in Spacelift's public worker pool. If you need to ensure " +
			"that requests made during runs originate from a known set of IP addresses, " +
			"please consider setting up a [private worker pool](https://docs.spacelift.io/concepts/worker-pools).",
		ReadContext: ipsRead,
		Schema: map[string]*schema.Schema{
			"ips": {
				Type:        schema.TypeSet,
				Description: "the list of spacelift.io outgoing IP addresses",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"cidrs": {
				Type:        schema.TypeList,
				Description: "list of Spacelift IP addresses in CIDR notation (/32) for easy use in security group rules",
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

	if err := meta.(*internal.Client).Query(ctx, "ReadIPs", &query, nil); err != nil {
		d.SetId("")
		return diag.Errorf("could not query for outgoing IP addresses: %v", err)
	}

	ips := schema.NewSet(schema.HashString, []interface{}{})
	for _, ip := range query.IPs {
		ips.Add(ip)
	}

	d.Set("ips", ips)

	// Create CIDR list by appending /32 to each IP
	cidrs := make([]string, len(query.IPs))
	for i, ip := range query.IPs {
		cidrs[i] = fmt.Sprintf("%s/32", ip)
	}
	d.Set("cidrs", cidrs)

	return nil
}
