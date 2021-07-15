package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

var gitlabIntegrationFields = struct {
	APIHost       string
	WebhookSecret string
}{
	APIHost:       "api_host",
	WebhookSecret: "webhook_secret",
}

func dataGitlabIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_gitlab_integration` returns details about Gitlab integration",

		ReadContext: dataGitlabIntegrationRead,

		Schema: map[string]*schema.Schema{
			gitlabIntegrationFields.APIHost: {
				Type:        schema.TypeString,
				Description: "Gitlab integration api host",
				Computed:    true,
			},
			gitlabIntegrationFields.WebhookSecret: {
				Type:        schema.TypeString,
				Description: "Gitlab integration webhook secret",
				Computed:    true,
			},
		},
	}
}

func dataGitlabIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var query struct {
		GitlabIntegration *structs.GitlabIntegration `graphql:"gitlabIntegration"`
	}

	if err := meta.(*internal.Client).Query(ctx, "GitlabIntegrationRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for gitlab integration: %v", err)
	}

	gitlabIntegration := query.GitlabIntegration
	if gitlabIntegration == nil {
		return diag.Errorf("gitlab integration not found")
	}

	d.SetId("spacelift_gitlab_integration_id") // TF expects id to be set otherwise it will fail
	d.Set(gitlabIntegrationFields.APIHost, gitlabIntegration.APIHost)
	d.Set(gitlabIntegrationFields.WebhookSecret, gitlabIntegration.WebhookSecret)

	return nil
}
