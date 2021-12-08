package structs

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// AzureIntegrationAttachment is a single Azure integration stack or module
// attachment.
type AzureIntegrationAttachment struct {
	ID             string  `graphql:"id"`
	StackID        string  `graphql:"stackId"`
	IsModule       bool    `graphql:"isModule"`
	Read           bool    `graphql:"read"`
	SubscriptionID *string `graphql:"subscriptionId"`
	Write          bool    `graphql:"write"`
}

// PopulateResourceData populates Terraform resource data with the contents of
// the AzureIntegration attachment.
func (i *AzureIntegrationAttachment) PopulateResourceData(d *schema.ResourceData) {
	d.Set("attachment_id", i.ID)
	d.Set("read", i.Read)
	d.Set("write", i.Write)

	if i.IsModule {
		d.Set("module_id", i.StackID)
	} else {
		d.Set("stack_id", i.StackID)
	}

	if subID := i.SubscriptionID; subID != nil {
		d.Set("subscription_id", *subID)
	}
}
