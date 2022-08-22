package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

var gitlabWebhookEndpointFields = struct {
	WebhookEndpoint string
}{
	WebhookEndpoint: "webhook_endpoint",
}

func dataGitlabWebhookEndpoint() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_gitlab_webhook_endpoint` returns details about Gitlab webhook endpoint",

		ReadContext: dataGitlabWebhookEndpointRead,

		Schema: map[string]*schema.Schema{
			gitlabWebhookEndpointFields.WebhookEndpoint: {
				Type:        schema.TypeString,
				Description: "Gitlab webhook endpoint address", 
				Computed:    true,
			},
		},
	}
}

func dataGitlabWebhookEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var query struct {
		GitlabWebhooksEndpoint string `graphql:"gitlabWebhooksEndpoint"`
	}

	if err := meta.(*internal.Client).Query(ctx, "GitlabWebhookEndpointRead", &query, nil); err != nil {
		d.SetId("")
		return diag.Errorf("could not query for gitlab webhook endpoint: %v", err)
	}

	d.SetId("spacelift_gitlab_webhook_endpoint_id") // TF expects id to be set otherwise it will fail
	d.Set(gitlabWebhookEndpointFields.WebhookEndpoint, query.GitlabWebhooksEndpoint)

	return nil
}
