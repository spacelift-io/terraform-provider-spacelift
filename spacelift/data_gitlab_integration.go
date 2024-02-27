package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

const (
	gitlabId            = "id"
	gitlabName          = "name"
	gitlabDescription   = "description"
	gitlabIsDefault     = "is_default"
	gitlabLabels        = "labels"
	gitlabSpaceID       = "space_id"
	gitlabAppID         = "app_id"
	gitlabAPIHost       = "api_host"
	gitlabWebhookSecret = "webhook_secret"
	gitlabWebhookURL    = "webhook_url"
)

func dataGitlabIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_gitlab_integration` returns details about Gitlab integration",

		ReadContext: dataGitlabIntegrationRead,

		Schema: map[string]*schema.Schema{
			gitlabId: {
				Type:        schema.TypeString,
				Description: "Gitlab integration id. If not provided, the default integration will be returned",
				Optional:    true,
			},
			gitlabName: {
				Type:        schema.TypeString,
				Description: "Gitlab integration name",
				Computed:    true,
			},
			gitlabDescription: {
				Type:        schema.TypeString,
				Description: "Gitlab integration description",
				Computed:    true,
			},
			gitlabIsDefault: {
				Type:        schema.TypeBool,
				Description: "Gitlab integration is default",
				Computed:    true,
			},
			gitlabLabels: {
				Type:        schema.TypeList,
				Description: "Gitlab integration labels",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			gitlabSpaceID: {
				Type:        schema.TypeString,
				Description: "Gitlab integration space id",
				Computed:    true,
			},
			gitlabAPIHost: {
				Type:        schema.TypeString,
				Description: "Gitlab integration api host",
				Computed:    true,
			},
			gitlabWebhookSecret: {
				Type:        schema.TypeString,
				Description: "Gitlab integration webhook secret",
				Computed:    true,
			},
			gitlabWebhookURL: {
				Type:        schema.TypeString,
				Description: "Gitlab integration webhook url",
				Computed:    true,
			},
		},
	}
}

func dataGitlabIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var query struct {
		GitlabIntegration *struct {
			ID          string `graphql:"id"`
			Name        string `graphql:"name"`
			Description string `graphql:"description"`
			IsDefault   bool   `graphql:"isDefault"`
			Space       struct {
				ID string `graphql:"id"`
			} `graphql:"space"`
			Labels        []string `graphql:"labels"`
			APIHost       string   `graphql:"apiHost"`
			WebhookSecret string   `graphql:"webhookSecret"`
			WebhookURL    string   `graphql:"webhookUrl"`
		} `graphql:"gitlabIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": ""}

	if id, ok := d.GetOk(gitlabId); ok && id != "" {
		variables["id"] = toID(id)
	}

	if err := meta.(*internal.Client).Query(ctx, "GitlabIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for gitlab integration: %v", err)
	}

	gitlabIntegration := query.GitlabIntegration
	if gitlabIntegration == nil {
		return diag.Errorf("gitlab integration not found")
	}

	d.SetId(gitlabIntegration.ID)
	d.Set(gitlabAPIHost, gitlabIntegration.APIHost)
	d.Set(gitlabWebhookSecret, gitlabIntegration.WebhookSecret)
	d.Set(gitlabWebhookURL, gitlabIntegration.WebhookURL)
	d.Set(gitlabId, gitlabIntegration.ID)
	d.Set(gitlabName, gitlabIntegration.Name)
	d.Set(gitlabDescription, gitlabIntegration.Description)
	d.Set(gitlabIsDefault, gitlabIntegration.IsDefault)
	d.Set(gitlabSpaceID, gitlabIntegration.Space.ID)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range gitlabIntegration.Labels {
		labels.Add(label)
	}

	d.Set(gitlabLabels, labels)

	return nil
}
