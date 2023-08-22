package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

var bitbucketDatacenterFields = struct {
	UserFacingHost string
	APIHost        string
	WebhookSecret  string
	WebhookURL     string
}{
	UserFacingHost: "user_facing_host",
	APIHost:        "api_host",
	WebhookSecret:  "webhook_secret",
	WebhookURL:     "webhook_url",
}

func dataBitbucketDatacenterIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_bitbucket_datacenter_integration` returns details about Bitbucket Datacenter integration",

		ReadContext: dataBitbucketDatacenterIntegrationRead,

		Schema: map[string]*schema.Schema{
			bitbucketDatacenterFields.APIHost: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration api host",
				Computed:    true,
			},
			bitbucketDatacenterFields.WebhookSecret: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration webhook secret",
				Computed:    true,
			},
			bitbucketDatacenterFields.WebhookURL: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration webhook URL",
				Computed:    true,
			},
			bitbucketDatacenterFields.UserFacingHost: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration user facing host",
				Computed:    true,
			},
		},
	}
}

func dataBitbucketDatacenterIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BitbucketDataCenterIntegration *struct {
			APIHost        string `graphql:"apiHost"`
			WebhookSecret  string `graphql:"webhookSecret"`
			UserFacingHost string `graphql:"userFacingHost"`
			WebhookURL     string `graphql:"webhookURL"`
		} `graphql:"bitbucketDatacenterIntegration"`
	}

	if err := meta.(*internal.Client).Query(ctx, "BitbucketDatacenterIntegrationRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for bitbucket datacenter integration: %v", err)
	}

	bitbucketDatacenterIntegration := query.BitbucketDataCenterIntegration
	if bitbucketDatacenterIntegration == nil {
		return diag.Errorf("bitbucket datacenter integration not found")
	}

	d.SetId("spacelift_bitbucket_datacenter_integration_id") // TF expects id to be set otherwise it will fail
	d.Set(bitbucketDatacenterFields.APIHost, bitbucketDatacenterIntegration.APIHost)
	d.Set(bitbucketDatacenterFields.WebhookSecret, bitbucketDatacenterIntegration.WebhookSecret)
	d.Set(bitbucketDatacenterFields.WebhookURL, bitbucketDatacenterIntegration.WebhookURL)
	d.Set(bitbucketDatacenterFields.UserFacingHost, bitbucketDatacenterIntegration.UserFacingHost)

	return nil
}
