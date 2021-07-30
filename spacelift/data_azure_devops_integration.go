package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

var azureDevopsIntegrationFields = struct {
	OrganizationURL string
	WebhookPassword string
}{
	OrganizationURL: "organization_url",
	WebhookPassword: "webhook_password",
}

func dataAzureDevopsIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_azure_devops_integration` returns details about Azure DevOps integration",

		ReadContext: dataAzureDevopsIntegrationRead,

		Schema: map[string]*schema.Schema{
			azureDevopsIntegrationFields.OrganizationURL: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration organization url",
				Computed:    true,
			},
			azureDevopsIntegrationFields.WebhookPassword: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration webhook password",
				Computed:    true,
			},
		},
	}
}

func dataAzureDevopsIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AzureDevOpsIntegration *struct {
			OrganizationURL string `graphql:"organizationURL"`
			WebhookPassword string `graphql:"webhookPassword"`
		} `graphql:"azureDevOpsRepoIntegration"`
	}

	if err := meta.(*internal.Client).Query(ctx, "AzureDevOpsIntegrationRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for azure devops integration: %v", err)
	}

	azureDevopsIntegration := query.AzureDevOpsIntegration
	if azureDevopsIntegration == nil {
		return diag.Errorf("azure devops integration not found")
	}

	d.SetId("spacelift_azure_devops_integration_id") // TF expects id to be set otherwise it will fail
	d.Set(azureDevopsIntegrationFields.OrganizationURL, azureDevopsIntegration.OrganizationURL)
	d.Set(azureDevopsIntegrationFields.WebhookPassword, azureDevopsIntegration.WebhookPassword)

	return nil
}
