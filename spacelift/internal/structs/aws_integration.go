package structs

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// AWSIntegration represents an integration with AWS.
type AWSIntegration struct {
	ID                          string   `graphql:"id"`
	DurationSeconds             int      `graphql:"durationSeconds"`
	GenerateCredentialsInWorker bool     `graphql:"generateCredentialsInWorker"`
	ExternalID                  string   `graphql:"externalId"`
	Labels                      []string `graphql:"labels"`
	Legacy                      bool     `graphql:"legacy"`
	Name                        string   `graphql:"name"`
	RoleARN                     string   `graphql:"roleArn"`
}

// PopulateResourceData populates Terraform resource data with the contents of
// the AWSIntegration.
func (i *AWSIntegration) PopulateResourceData(d *schema.ResourceData) {
	d.Set("duration_seconds", i.DurationSeconds)
	d.Set("generate_credentials_in_worker", i.GenerateCredentialsInWorker)
	d.Set("external_id", i.ExternalID)
	d.Set("legacy", i.Legacy)
	d.Set("name", i.Name)
	d.Set("role_arn", i.RoleARN)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range i.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

}
