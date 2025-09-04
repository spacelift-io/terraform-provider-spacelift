package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

const (
	azureDevopsID              = "id"
	azureDevopsName            = "name"
	azureDevopsDescription     = "description"
	azureDevopsIsDefault       = "is_default"
	azureDevopsLabels          = "labels"
	azureDevopsSpaceID         = "space_id"
	azureDevopsOrganizationURL = "organization_url"
	azureDevopsWebhookPassword = "webhook_password"
	azureDevopsWebhookURL      = "webhook_url"
	azureDevopsVCSChecks       = "vcs_checks"
	azureDevopsUseGitCheckout  = "use_git_checkout"
)

func dataAzureDevopsIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_azure_devops_integration` returns details about Azure DevOps integration",

		ReadContext: dataAzureDevopsIntegrationRead,

		Schema: map[string]*schema.Schema{
			azureDevopsID: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration id. If not provided, the default integration will be returned",
				Optional:    true,
			},
			azureDevopsName: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration name",
				Computed:    true,
			},
			azureDevopsDescription: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration description",
				Computed:    true,
			},
			azureDevopsIsDefault: {
				Type:        schema.TypeBool,
				Description: "Azure DevOps integration is default",
				Computed:    true,
			},
			azureDevopsLabels: {
				Type:        schema.TypeList,
				Description: "Azure DevOps integration labels",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			azureDevopsSpaceID: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration space id",
				Computed:    true,
			},
			azureDevopsOrganizationURL: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration organization url",
				Computed:    true,
			},
			azureDevopsWebhookPassword: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration webhook password",
				Computed:    true,
			},
			azureDevopsWebhookURL: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration webhook url",
				Computed:    true,
			},
			azureDevopsVCSChecks: {
				Type:        schema.TypeString,
				Description: "VCS checks configured for Azure DevOps repositories. Possible values: INDIVIDUAL, AGGREGATED, ALL. Defaults to INDIVIDUAL.",
				Computed:    true,
			},
			azureDevopsUseGitCheckout: {
				Type:        schema.TypeBool,
				Description: "Indicates whether the integration should use git checkout. If false source code will be downloaded using the VCS API.",
				Computed:    true,
			},
		},
	}
}

func dataAzureDevopsIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AzureDevOpsIntegration *struct {
			ID          string `graphql:"id"`
			Name        string `graphql:"name"`
			Description string `graphql:"description"`
			IsDefault   bool   `graphql:"isDefault"`
			Space       struct {
				ID string `graphql:"id"`
			} `graphql:"space"`
			Labels          []string `graphql:"labels"`
			OrganizationURL string   `graphql:"organizationURL"`
			WebhookPassword string   `graphql:"webhookPassword"`
			WebhookURL      string   `graphql:"webhookUrl"`
			VCSChecks       string   `graphql:"vcsChecks"`
			UseGitCheckout  bool     `graphql:"useGitCheckout"`
		} `graphql:"azureDevOpsRepoIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": ""}

	if id, ok := d.GetOk(azureDevopsID); ok && id != "" {
		variables["id"] = toID(id)
	}

	if err := meta.(*internal.Client).Query(ctx, "AzureDevOpsIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for azure devops integration: %v", err)
	}

	azureDevopsIntegration := query.AzureDevOpsIntegration
	if azureDevopsIntegration == nil {
		return diag.Errorf("azure devops integration not found")
	}

	d.SetId(azureDevopsIntegration.ID)
	d.Set(azureDevopsID, azureDevopsIntegration.ID)
	d.Set(azureDevopsName, azureDevopsIntegration.Name)
	d.Set(azureDevopsDescription, azureDevopsIntegration.Description)
	d.Set(azureDevopsIsDefault, azureDevopsIntegration.IsDefault)
	d.Set(azureDevopsSpaceID, azureDevopsIntegration.Space.ID)
	d.Set(azureDevopsOrganizationURL, azureDevopsIntegration.OrganizationURL)
	d.Set(azureDevopsWebhookPassword, azureDevopsIntegration.WebhookPassword)
	d.Set(azureDevopsWebhookURL, azureDevopsIntegration.WebhookURL)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range azureDevopsIntegration.Labels {
		labels.Add(label)
	}

	d.Set(azureDevopsLabels, labels)
	d.Set(azureDevopsVCSChecks, azureDevopsIntegration.VCSChecks)
	d.Set(azureDevopsUseGitCheckout, azureDevopsIntegration.UseGitCheckout)

	return nil
}
