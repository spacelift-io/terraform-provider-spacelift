package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

var githubErpIntegrationFields = struct {
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
			githubErpIntegrationFields.APIHost: {
				Type:        schema.TypeString,
				Description: "Github integration api host",
				Computed:    true,
			},
			githubErpIntegrationFields.WebhookSecret: {
				Type:        schema.TypeString,
				Description: "Github integration webhook secret",
				Computed:    true,
			},
			githubErpIntegrationFields.AppID: {
				Type:        schema.TypeString,
				Description: "Github integration app id",
				Computed:    true,
			},
		},
	}
}

func dataGithubEnterpriseIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		GithubEnterpriseIntegration *structs.GithubEnterpriseIntegration `graphql:"githubEnterpriseIntegration"`
	}

	if err := meta.(*internal.Client).Query(ctx, "GithubEnterpriseIntegrationRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for github enterprise integration: %v", err)
	}

	githubEnterpriseIntegration := query.GithubEnterpriseIntegration
	if githubEnterpriseIntegration == nil {
		return diag.Errorf("github enterprise integration not found")
	}

	d.SetId("spacelift_github_enterprise_integration_id") // TF expects id to be set otherwise it will fail
	d.Set(githubErpIntegrationFields.APIHost, githubEnterpriseIntegration.APIHost)
	d.Set(githubErpIntegrationFields.WebhookSecret, githubEnterpriseIntegration.WebhookSecret)
	d.Set(githubErpIntegrationFields.AppID, githubEnterpriseIntegration.AppID)

	return nil
}
