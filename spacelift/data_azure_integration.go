package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataAzureIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_azure_integration` represents an integration with an Azure " +
			"AD tenant. This integration is account-level and needs to be explicitly " +
			"attached to individual stacks in order to take effect. Note that you will " +
			"need to provide admin consent manually for the integration to work",

		ReadContext: dataAzureIntegrationRead,

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:        schema.TypeString,
				Description: "immutable ID of the integration",
				Required:    true,
			},
			"admin_consent_provided": {
				Type: schema.TypeBool,
				Description: "" +
					"Indicates whether admin consent has been performed for the " +
					"AAD Application.",
				Computed: true,
			},
			"admin_consent_url": {
				Type: schema.TypeString,
				Description: "" +
					"The URL to use to provide admin consent to the application in " +
					"the customer's tenant",
				Computed: true,
			},
			"application_id": {
				Type: schema.TypeString,
				Description: "" +
					"The applicationId of the Azure AD application used by the " +
					"integration.",
				Computed: true,
			},
			"default_subscription_id": {
				Type: schema.TypeString,
				Description: "" +
					"The default subscription ID to use, if one isn't specified " +
					"at the stack/module level",
				Computed: true,
			},
			"display_name": {
				Type: schema.TypeString,
				Description: "" +
					"The display name for the application in Azure. This is " +
					"automatically generated when the integration is created, and " +
					"cannot be changed without deleting and recreating the " +
					"integration.",
				Computed: true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels to set on the integration",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The friendly name of the integration",
				Computed:    true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the integration is in",
				Computed:    true,
			},
			"tenant_id": {
				Type:        schema.TypeString,
				Description: "The Azure AD tenant ID",
				Computed:    true,
			},
		},
	}
}

func dataAzureIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AzureIntegration *structs.AzureIntegration `graphql:"azureIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Get("integration_id").(string))}
	if err := meta.(*internal.Client).Query(ctx, "AzureIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the Azure integration: %v", err)
	}

	integration := query.AzureIntegration
	if integration == nil {
		return diag.Errorf("Azure integration not found: %s", d.Id())
	}

	d.SetId(integration.ID)
	integration.PopulateResourceData(d)

	return nil
}
