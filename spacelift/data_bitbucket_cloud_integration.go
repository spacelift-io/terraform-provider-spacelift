package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

const (
	bitbucketCloudID          = "id"
	bitbucketCloudName        = "name"
	bitbucketCloudDescription = "description"
	bitbucketCloudIsDefault   = "is_default"
	bitbucketCloudLabels      = "labels"
	bitbucketCloudSpaceID     = "space_id"
	bitbucketCloudUsername    = "username"
	bitbucketCloudWebhookURL  = "webhook_url"
	bitbucketCloudVCSChecks   = "vcs_checks"
)

func dataBitbucketCloudIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_bitbucket_cloud_integration` returns details about Bitbucket Cloud integration",

		ReadContext: dataBitbucketCloudIntegrationRead,

		Schema: map[string]*schema.Schema{
			bitbucketCloudID: {
				Type:        schema.TypeString,
				Description: "Bitbucket Cloud integration id. If not provided, the default integration will be returned",
				Optional:    true,
			},
			bitbucketCloudName: {
				Type:        schema.TypeString,
				Description: "Bitbucket Cloud integration name",
				Computed:    true,
			},
			bitbucketCloudDescription: {
				Type:        schema.TypeString,
				Description: "Bitbucket Cloud integration description",
				Computed:    true,
			},
			bitbucketCloudIsDefault: {
				Type:        schema.TypeBool,
				Description: "Bitbucket Cloud integration is default",
				Computed:    true,
			},
			bitbucketCloudLabels: {
				Type:        schema.TypeList,
				Description: "Bitbucket Cloud integration labels",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			bitbucketCloudSpaceID: {
				Type:        schema.TypeString,
				Description: "Bitbucket Cloud integration space id",
				Computed:    true,
			},
			bitbucketCloudUsername: {
				Type:        schema.TypeString,
				Description: "Bitbucket Cloud username",
				Computed:    true,
			},
			bitbucketCloudWebhookURL: {
				Type:        schema.TypeString,
				Description: "Bitbucket Cloud integration webhook URL",
				Computed:    true,
			},
			bitbucketCloudVCSChecks: {
				Type:        schema.TypeString,
				Description: "VCS checks configured for GitLab repositories. Possible values: INDIVIDUAL, AGGREGATED, ALL. Defaults to INDIVIDUAL.",
				Computed:    true,
			},
		},
	}
}

func dataBitbucketCloudIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BitbucketCloudIntegration *struct {
			ID          string `graphql:"id"`
			Name        string `graphql:"name"`
			Description string `graphql:"description"`
			IsDefault   bool   `graphql:"isDefault"`
			Space       struct {
				ID string `graphql:"id"`
			} `graphql:"space"`
			Labels     []string `graphql:"labels"`
			Username   string   `graphql:"username"`
			WebhookURL string   `graphql:"webhookUrl"`
			VCSChecks  string   `graphql:"vcsChecks"`
		} `graphql:"bitbucketCloudIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": ""}

	if id, ok := d.GetOk(bitbucketCloudID); ok && id != "" {
		variables["id"] = toID(id)
	}

	if err := meta.(*internal.Client).Query(ctx, "BitbucketCloudIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for bitbucket cloud integration: %v", err)
	}

	bitbucketCloudIntegration := query.BitbucketCloudIntegration
	if bitbucketCloudIntegration == nil {
		return diag.Errorf("bitbucket cloud integration not found")
	}

	d.SetId(bitbucketCloudIntegration.ID)
	d.Set(bitbucketCloudID, bitbucketCloudIntegration.ID)
	d.Set(bitbucketCloudName, bitbucketCloudIntegration.Name)
	d.Set(bitbucketCloudDescription, bitbucketCloudIntegration.Description)
	d.Set(bitbucketCloudIsDefault, bitbucketCloudIntegration.IsDefault)
	d.Set(bitbucketCloudSpaceID, bitbucketCloudIntegration.Space.ID)
	d.Set(bitbucketCloudUsername, bitbucketCloudIntegration.Username)
	d.Set(bitbucketCloudWebhookURL, bitbucketCloudIntegration.WebhookURL)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range bitbucketCloudIntegration.Labels {
		labels.Add(label)
	}

	d.Set(bitbucketCloudLabels, labels)
	d.Set(bitbucketCloudVCSChecks, bitbucketCloudIntegration.VCSChecks)

	return nil
}
