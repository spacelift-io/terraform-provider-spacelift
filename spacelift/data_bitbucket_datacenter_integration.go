package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataBitbucketDatacenterIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_bitbucket_datacenter_integration` returns details about Bitbucket Datacenter integration",

		ReadContext: dataBitbucketDatacenterIntegrationRead,

		Schema: map[string]*schema.Schema{
			structs.BitbucketDatacenterFields.APIHost: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration api host",
				Computed:    true,
			},
			structs.BitbucketDatacenterFields.Username: {
				Type:        schema.TypeString,
				Description: "Username which will be used to authenticate requests for cloning repositories",
				Computed:    true,
			},
			structs.BitbucketDatacenterFields.WebhookSecret: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration webhook secret",
				Computed:    true,
			},
			structs.BitbucketDatacenterFields.WebhookURL: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration webhook URL",
				Computed:    true,
			},
			structs.BitbucketDatacenterFields.UserFacingHost: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration user facing host",
				Computed:    true,
			},
		},
	}
}

func dataBitbucketDatacenterIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BitbucketDataCenterIntegration *structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegration"`
	}

	if err := meta.(*internal.Client).Query(ctx, "BitbucketDatacenterIntegrationRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for bitbucket datacenter integration: %v", err)
	}

	bitbucketDatacenterIntegration := query.BitbucketDataCenterIntegration
	if bitbucketDatacenterIntegration == nil {
		return diag.Errorf("bitbucket datacenter integration not found")
	}

	d.SetId("spacelift_bitbucket_datacenter_integration_id") // TF expects id to be set otherwise it will fail
	d.Set(structs.BitbucketDatacenterFields.APIHost, bitbucketDatacenterIntegration.APIHost)
	d.Set(structs.BitbucketDatacenterFields.WebhookSecret, bitbucketDatacenterIntegration.WebhookSecret)
	d.Set(structs.BitbucketDatacenterFields.WebhookURL, bitbucketDatacenterIntegration.WebhookURL)
	d.Set(structs.BitbucketDatacenterFields.UserFacingHost, bitbucketDatacenterIntegration.UserFacingHost)
	d.Set(structs.BitbucketDatacenterFields.Username, bitbucketDatacenterIntegration.Username)

	return nil
}
