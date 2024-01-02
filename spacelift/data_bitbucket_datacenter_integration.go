package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

const (
	bitbucketDatacenterID             = "id"
	bitbucketDatacenterName           = "name"
	bitbucketDatacenterDescription    = "description"
	bitbucketDatacenterIsDefault      = "is_default"
	bitbucketDatacenterLabels         = "labels"
	bitbucketDatacenterSpaceID        = "space_id"
	bitbucketDatacenterUserFacingHost = "user_facing_host"
	bitbucketDatacenterAPIHost        = "api_host"
	bitbucketDatacenterUsername       = "username"
	bitbucketDatacenterWebhookURL     = "webhook_url"
	bitbucketDatacenterWebhookSecret  = "webhook_secret"
)

func dataBitbucketDatacenterIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_bitbucket_datacenter_integration` returns details about Bitbucket Datacenter integration",

		ReadContext: dataBitbucketDatacenterIntegrationRead,

		Schema: map[string]*schema.Schema{
			bitbucketDatacenterID: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration id. If not provided, the default integration will be returned",
				Optional:    true,
			},
			bitbucketDatacenterName: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration name",
				Computed:    true,
			},
			bitbucketDatacenterDescription: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration description",
				Computed:    true,
			},
			bitbucketDatacenterIsDefault: {
				Type:        schema.TypeBool,
				Description: "Bitbucket Datacenter integration is default",
				Computed:    true,
			},
			bitbucketDatacenterLabels: {
				Type:        schema.TypeList,
				Description: "Bitbucket Datacenter integration labels",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			bitbucketDatacenterSpaceID: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration space id",
				Computed:    true,
			},
			bitbucketDatacenterUsername: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter username",
				Computed:    true,
			},
			bitbucketDatacenterAPIHost: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration api host",
				Computed:    true,
			},
			structs.BitbucketDatacenterFields.Username: {
				Type:        schema.TypeString,
				Description: "Username which will be used to authenticate requests for cloning repositories",
				Computed:    true,
			},
			structs.BitbucketDatacenterFields.WebhookSecret: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration webhook secret",
				Computed:    true,
			},
			bitbucketDatacenterWebhookURL: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration webhook URL",
				Computed:    true,
			},
			bitbucketDatacenterUserFacingHost: {
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
			ID          string `graphql:"id"`
			Name        string `graphql:"name"`
			Description string `graphql:"description"`
			IsDefault   bool   `graphql:"isDefault"`
			Space       struct {
				ID string `graphql:"id"`
			} `graphql:"space"`
			Labels         []string `graphql:"labels"`
			APIHost        string   `graphql:"apiHost"`
			WebhookSecret  string   `graphql:"webhookSecret"`
			UserFacingHost string   `graphql:"userFacingHost"`
			WebhookURL     string   `graphql:"webhookURL"`
			Username       string   `graphql:"username"`
		} `graphql:"bitbucketDatacenterIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": ""}
	if id, ok := d.GetOk(bitbucketDatacenterID); ok && id != "" {
		variables["id"] = toID(id)
	}

	if err := meta.(*internal.Client).Query(ctx, "BitbucketDatacenterIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for bitbucket datacenter integration: %v", err)
	}

	bitbucketDatacenterIntegration := query.BitbucketDataCenterIntegration
	if bitbucketDatacenterIntegration == nil {
		return diag.Errorf("bitbucket datacenter integration not found")
	}

	d.SetId(bitbucketDatacenterIntegration.ID)
	d.Set(bitbucketDatacenterID, bitbucketDatacenterIntegration.ID)
	d.Set(bitbucketDatacenterName, bitbucketDatacenterIntegration.Name)
	d.Set(bitbucketDatacenterDescription, bitbucketDatacenterIntegration.Description)
	d.Set(bitbucketDatacenterIsDefault, bitbucketDatacenterIntegration.IsDefault)
	d.Set(bitbucketDatacenterSpaceID, bitbucketDatacenterIntegration.Space.ID)
	d.Set(bitbucketDatacenterAPIHost, bitbucketDatacenterIntegration.APIHost)
	d.Set(bitbucketDatacenterWebhookSecret, bitbucketDatacenterIntegration.WebhookSecret)
	d.Set(bitbucketDatacenterWebhookURL, bitbucketDatacenterIntegration.WebhookURL)
	d.Set(bitbucketDatacenterUserFacingHost, bitbucketDatacenterIntegration.UserFacingHost)
	d.Set(bitbucketDatacenterUsername, bitbucketDatacenterIntegration.Username)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range bitbucketDatacenterIntegration.Labels {
		labels.Add(label)
	}

	d.Set(bitbucketDatacenterLabels, labels)

	return nil
}
