package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataAzureIntegrations() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_azure_integrations` represents a list of all the Azure integrations in " +
			"the Spacelift account visible to the API user.",

		ReadContext: dataAzureIntegrationsRead,
		Schema: map[string]*schema.Schema{
			"integrations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"integration_id": {
							Type:        schema.TypeString,
							Description: "Immutable ID of the integration.",
							Computed:    true,
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
						"object_id": {
							Type:        schema.TypeString,
							Description: "The objectId of the Azure AD application used by the integration.",
							Computed:    true,
							Deprecated:  "This field will be removed in a future version. Use `service_principal_object_id` instead.",
						},
						"service_principal_object_id": {
							Type: schema.TypeString,
							Description: "This is the unique ID of the service principal object associated with this application. " +
								"This ID can be useful when performing management operations against this application using programmatic interfaces.",
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
							Description: "The friendly name of the integration.",
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
						"autoattach_enabled": {
							Type:        schema.TypeBool,
							Description: "Enables `autoattach:` labels functionality for this integration.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataAzureIntegrationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AzureIntegrations []*structs.AzureIntegration `graphql:"azureIntegrations()"`
	}
	if err := meta.(*internal.Client).Query(ctx, "azureIntegrations", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for azure integrations: %v", err)
	}

	d.SetId("spacelift_azure_integrations")

	integrations := query.AzureIntegrations
	if integrations == nil {
		d.Set("integrations", nil)
		return nil
	}

	mapped := flattenDataAzureIntegrationsList(integrations)
	if err := d.Set("integrations", mapped); err != nil {
		d.SetId("")
		return diag.Errorf("could not set azure integrations: %v", err)
	}

	return nil
}

func flattenDataAzureIntegrationsList(integrations []*structs.AzureIntegration) []map[string]interface{} {
	mapped := make([]map[string]interface{}, len(integrations))

	for index, integration := range integrations {
		integrationToMap := integration.ToMap()
		integrationToMap["integration_id"] = integration.ID
		mapped[index] = integrationToMap
	}

	return mapped
}
