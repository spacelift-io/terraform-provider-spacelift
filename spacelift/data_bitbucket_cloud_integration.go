package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

var bitbucketCloudFields = struct {
	Username   string
	WebhookURL string
}{
	Username:   "username",
	WebhookURL: "webhook_url",
}

func dataBitbucketCloudIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_bitbucket_cloud_integration` returns details about Bitbucket Cloud integration",

		ReadContext: dataBitbucketCloudIntegrationRead,

		Schema: map[string]*schema.Schema{
			bitbucketCloudFields.Username: {
				Type:        schema.TypeString,
				Description: "Bitbucket Cloud username",
				Computed:    true,
			},
			bitbucketCloudFields.WebhookURL: {
				Type:        schema.TypeString,
				Description: "Bitbucket Cloud integration webhook URL",
				Computed:    true,
			},
		},
	}
}

func dataBitbucketCloudIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BitbucketCloudIntegration *struct {
			Username   string `graphql:"username"`
			WebhookURL string `graphql:"webhookUrl"`
		} `graphql:"bitbucketCloudIntegration"`
	}

	if err := meta.(*internal.Client).Query(ctx, "BitbucketCloudIntegrationRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for bitbucket cloud integration: %v", err)
	}

	bitbucketCloudIntegration := query.BitbucketCloudIntegration
	if bitbucketCloudIntegration == nil {
		return diag.Errorf("bitbucket cloud integration not found")
	}

	d.SetId("spacelift_bitbucket_cloud_integration_id") // TF expects id to be set otherwise it will fail
	d.Set(bitbucketCloudFields.Username, bitbucketCloudIntegration.Username)
	d.Set(bitbucketCloudFields.WebhookURL, bitbucketCloudIntegration.WebhookURL)

	return nil
}
