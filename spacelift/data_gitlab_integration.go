package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataGitlabIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_gitlab_integration` returns details about Gitlab integration",

		ReadContext: dataGitlabIntegrationRead,

		Schema: map[string]*schema.Schema{
			gitLabID: {
				Type:        schema.TypeString,
				Description: "Gitlab integration id. If not provided, the default integration will be returned",
				Optional:    true,
			},
			gitLabName: {
				Type:        schema.TypeString,
				Description: "Gitlab integration name",
				Computed:    true,
			},
			gitLabDescription: {
				Type:        schema.TypeString,
				Description: "Gitlab integration description",
				Computed:    true,
			},
			gitLabIsDefault: {
				Type:        schema.TypeBool,
				Description: "Gitlab integration is default",
				Computed:    true,
			},
			gitLabLabels: {
				Type:        schema.TypeList,
				Description: "Gitlab integration labels",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			gitLabSpaceID: {
				Type:        schema.TypeString,
				Description: "Gitlab integration space id",
				Computed:    true,
			},
			gitLabAPIHost: {
				Type:        schema.TypeString,
				Description: "Gitlab integration api host",
				Computed:    true,
			},
			gitLabWebhookSecret: {
				Type:        schema.TypeString,
				Description: "Gitlab integration webhook secret",
				Computed:    true,
			},
			gitLabWebhookURL: {
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

	if id, ok := d.GetOk(gitLabID); ok && id != "" {
		variables["id"] = toID(id)
	}

	if err := meta.(*internal.Client).Query(ctx, "GitlabIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for gitlab integration: %v", err)
	}

	gitLabIntegration := query.GitlabIntegration
	if gitLabIntegration == nil {
		return diag.Errorf("gitlab integration not found")
	}

	d.SetId(gitLabIntegration.ID)
	d.Set(gitLabAPIHost, gitLabIntegration.APIHost)
	d.Set(gitLabWebhookSecret, gitLabIntegration.WebhookSecret)
	d.Set(gitLabWebhookURL, gitLabIntegration.WebhookURL)
	d.Set(gitLabID, gitLabIntegration.ID)
	d.Set(gitLabName, gitLabIntegration.Name)
	d.Set(gitLabDescription, gitLabIntegration.Description)
	d.Set(gitLabIsDefault, gitLabIntegration.IsDefault)
	d.Set(gitLabSpaceID, gitLabIntegration.Space.ID)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range gitLabIntegration.Labels {
		labels.Add(label)
	}

	d.Set(gitLabLabels, labels)

	return nil
}
