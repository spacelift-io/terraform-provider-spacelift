package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

const (
	ghEnterpriseId            = "id"
	ghEnterpriseName          = "name"
	ghEnterpriseDescription   = "description"
	ghEnterpriseIsDefault     = "is_default"
	ghEnterpriseLabels        = "labels"
	ghEnterpriseSpaceID       = "space_id"
	ghEnterpriseAppID         = "app_id"
	ghEnterpriseAPIHost       = "api_host"
	ghEnterpriseWebhookSecret = "webhook_secret"
	ghEnterpriseWebhookURL    = "webhook_url"
)

func dataGithubEnterpriseIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_github_enterprise_integration` returns details about Github Enterprise integration",

		ReadContext: dataGithubEnterpriseIntegrationRead,

		Schema: map[string]*schema.Schema{
			ghEnterpriseId: {
				Type:        schema.TypeString,
				Description: "Github integration id. If not provided, the default integration will be returned",
				Optional:    true,
			},
			ghEnterpriseName: {
				Type:        schema.TypeString,
				Description: "Github integration name",
				Computed:    true,
			},
			ghEnterpriseDescription: {
				Type:        schema.TypeString,
				Description: "Github integration description",
				Computed:    true,
			},
			ghEnterpriseIsDefault: {
				Type:        schema.TypeBool,
				Description: "Github integration is default",
				Computed:    true,
			},
			ghEnterpriseLabels: {
				Type:        schema.TypeList,
				Description: "Github integration labels",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			ghEnterpriseSpaceID: {
				Type:        schema.TypeString,
				Description: "Github integration space id",
				Computed:    true,
			},
			ghEnterpriseAPIHost: {
				Type:        schema.TypeString,
				Description: "Github integration api host",
				Computed:    true,
			},
			ghEnterpriseWebhookSecret: {
				Type:        schema.TypeString,
				Description: "Github integration webhook secret",
				Computed:    true,
			},
			ghEnterpriseWebhookURL: {
				Type:        schema.TypeString,
				Description: "Github integration webhook url",
				Computed:    true,
			},
			ghEnterpriseAppID: {
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
			WebhookURL    string `graphql:"webhookUrl"`
			ID            string `graphql:"id"`
			Name          string `graphql:"name"`
			Description   string `graphql:"description"`
			IsDefault     bool   `graphql:"isDefault"`
			Space         struct {
				ID string `graphql:"id"`
			} `graphql:"space"`
			Labels []string `graphql:"labels"`
		} `graphql:"githubEnterpriseIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": ""}

	if id, ok := d.GetOk(ghEnterpriseId); ok && id != "" {
		variables["id"] = toID(id)
	}

	if err := meta.(*internal.Client).Query(ctx, "GithubEnterpriseIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for github enterprise integration: %v", err)
	}

	githubEnterpriseIntegration := query.GithubEnterpriseIntegration
	if githubEnterpriseIntegration == nil {
		return diag.Errorf("github enterprise integration not found")
	}

	d.SetId(githubEnterpriseIntegration.ID)
	d.Set(ghEnterpriseAPIHost, githubEnterpriseIntegration.APIHost)
	d.Set(ghEnterpriseWebhookSecret, githubEnterpriseIntegration.WebhookSecret)
	d.Set(ghEnterpriseWebhookURL, githubEnterpriseIntegration.WebhookURL)
	d.Set(ghEnterpriseAppID, githubEnterpriseIntegration.AppID)
	d.Set(ghEnterpriseId, githubEnterpriseIntegration.ID)
	d.Set(ghEnterpriseName, githubEnterpriseIntegration.Name)
	d.Set(ghEnterpriseDescription, githubEnterpriseIntegration.Description)
	d.Set(ghEnterpriseIsDefault, githubEnterpriseIntegration.IsDefault)
	d.Set(ghEnterpriseSpaceID, githubEnterpriseIntegration.Space.ID)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range githubEnterpriseIntegration.Labels {
		labels.Add(label)
	}

	d.Set(ghEnterpriseLabels, labels)

	return nil
}
