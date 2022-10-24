package structs

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// AzureIntegration represents an Azure identity provided by the Spacelift
// integration.
type AzureIntegration struct {
	ID                    string   `graphql:"id"`
	AdminConsentProvided  bool     `graphql:"adminConsentProvided"`
	AdminConsentURL       string   `graphql:"adminConsentURL"`
	ApplicationID         string   `graphql:"applicationId"`
	DefaultSubscriptionID *string  `graphql:"defaultSubscriptionId"`
	DisplayName           string   `graphql:"displayName"`
	Labels                []string `graphql:"labels"`
	Name                  string   `graphql:"name"`
	TenantID              string   `graphql:"tenantId"`
	Space                 string   `graphql:"space"`
}

// PopulateResourceData populates Terraform resource data with the contents of
// the AzureIntegration.
func (i *AzureIntegration) PopulateResourceData(d *schema.ResourceData) {
	for key, value := range i.ToMap() {
		d.Set(key, value)
	}
}

func (i *AzureIntegration) ToMap() map[string]interface{} {
	fields := map[string]interface{}{
		"admin_consent_provided": i.AdminConsentProvided,
		"admin_consent_url":      i.AdminConsentURL,
		"application_id":         i.ApplicationID,
		"display_name":           i.DisplayName,
		"name":                   i.Name,
		"tenant_id":              i.TenantID,
		"space_id":               i.Space,
	}
	if subID := i.DefaultSubscriptionID; subID != nil {
		fields["default_subscription_id"] = *subID
	}

	fields["labels"] = i.getLabelsSet()

	return fields
}

func (i *AzureIntegration) getLabelsSet() *schema.Set {
	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range i.Labels {
		labels.Add(label)
	}
	return labels
}
