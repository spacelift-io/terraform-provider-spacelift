package structs

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// AzureIntegration represents an Azure identity provided by the Spacelift
// integration.
type AzureIntegration struct {
	ID                    string   `graphql:"id"`
	AdminConsentProvided  bool     `graphql:"adminConsentProvided"`
	AdminConsentURL       string   `grapqhl:"adminConsentURL"`
	ApplicationID         string   `graphql:"applicationId"`
	DefaultSubscriptionID *string  `graphql:"defaultSubscriptionId"`
	DisplayName           string   `graphql:"displayName"`
	Labels                []string `graphql:"labels"`
	Name                  string   `graphql:"name"`
	TenantID              string   `graphql:"tenantId"`
}

func (i *AzureIntegration) PopulateResourceData(d *schema.ResourceData) {
	d.Set("admin_consent_provided", i.AdminConsentProvided)
	d.Set("admin_consent_url", i.AdminConsentURL)
	d.Set("application_id", i.ApplicationID)
	d.Set("display_name", i.DisplayName)
	d.Set("name", i.Name)
	d.Set("tenant_id", i.TenantID)

	if subID := i.DefaultSubscriptionID; subID != nil {
		d.Set("default_subscription_id", *subID)
	} else {
		d.Set("default_subscription_id", nil)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range i.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)
}
