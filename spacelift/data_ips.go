package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataIPs() *schema.Resource {
	return &schema.Resource{
		Read: ipsRead,
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

func ipsRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId("spacelift-ips")

	var query struct {
		IPs []string `graphql:"outgoingIPAddresses"`
	}

	if err := meta.(*internal.Client).Query(&query, nil); err != nil {
		d.SetId("")
		return errors.Wrap(err, "could not query for outgoing IP addresses")
	}

	ips := schema.NewSet(schema.HashString, []interface{}{})
	for _, ip := range query.IPs {
		ips.Add(ip)
	}

	d.Set("ips", ips)

	return nil
}
