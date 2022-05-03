package structs

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// AWSIntegrationAttachment is a single AWS integration stack or module
// attachment.
type AWSIntegrationAttachment struct {
	ID       string `graphql:"id"`
	StackID  string `graphql:"stackId"`
	IsModule bool   `graphql:"isModule"`
	Read     bool   `graphql:"read"`
	Write    bool   `graphql:"write"`
}

// PopulateResourceData populates Terraform resource data with the contents of
// the AWSIntegration attachment.
func (i *AWSIntegrationAttachment) PopulateResourceData(d *schema.ResourceData) {
	d.Set("attachment_id", i.ID)
	d.Set("read", i.Read)
	d.Set("write", i.Write)

	if i.IsModule {
		d.Set("module_id", i.StackID)
	} else {
		d.Set("stack_id", i.StackID)
	}
}
