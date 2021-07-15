package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

var githubEnterpriseIntegrationFields = struct {
	AppID         string
	APIHost       string
	WebhookSecret string
}{
	AppID:         "app_id",
	APIHost:       "api_host",
	WebhookSecret: "webhook_secret",
}

func dataGithubEnterpriseIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_github_enterprise_integration` returns details about Github Enterprise integration",

		ReadContext: dataGithubEnterpriseIntegrationRead,

		Schema: map[string]*schema.Schema{
			githubEnterpriseIntegrationFields.APIHost: {
				Type:        schema.TypeString,
				Description: "Github integration api host",
				Computed:    true,
			},
			githubEnterpriseIntegrationFields.WebhookSecret: {
				Type:        schema.TypeString,
				Description: "Github integration webhook secret",
				Computed:    true,
			},
			githubEnterpriseIntegrationFields.AppID: {
				Type:        schema.TypeString,
				Description: "Github integration app id",
				Computed:    true,
			},
		},
	}
}

func dataGithubEnterpriseIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		GithubEnterpriseIntegration *struct {
			AppID         string `graphql:"appID"`
			APIHost       string `graphql:"apiHost"`
			WebhookSecret string `graphql:"webhookSecret"`
		} `graphql:"githubEnterpriseIntegration"`
	}

	if err := meta.(*internal.Client).Query(ctx, "GithubEnterpriseIntegrationRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for github enterprise integration: %v", err)
	}

	githubEnterpriseIntegration := query.GithubEnterpriseIntegration
	if githubEnterpriseIntegration == nil {
		return diag.Errorf("github enterprise integration not found")
	}

	d.SetId("spacelift_github_enterprise_integration_id") // TF expects id to be set otherwise it will fail
	d.Set(githubEnterpriseIntegrationFields.APIHost, githubEnterpriseIntegration.APIHost)
	d.Set(githubEnterpriseIntegrationFields.WebhookSecret, githubEnterpriseIntegration.WebhookSecret)
	d.Set(githubEnterpriseIntegrationFields.AppID, githubEnterpriseIntegration.AppID)

	return nil
}
