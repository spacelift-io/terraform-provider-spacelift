package structs

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// AWSIntegration represents an integration with AWS.
type AWSIntegration struct {
	ID                          string   `graphql:"id"`
	DurationSeconds             int      `graphql:"durationSeconds"`
	GenerateCredentialsInWorker bool     `graphql:"generateCredentialsInWorker"`
	ExternalID                  string   `graphql:"externalId"`
	Labels                      []string `graphql:"labels"`
	Name                        string   `graphql:"name"`
	RoleARN                     string   `graphql:"roleArn"`
	Space                       string   `graphql:"space"`
	Region                      *string  `graphql:"region"`
	AutoattachEnabled           bool     `graphql:"autoattachEnabled"`
	TagAssumeRole               bool     `graphql:"tagAssumeRole"`
}

// PopulateResourceData populates Terraform resource data with the contents of
// the AWSIntegration.
func (i *AWSIntegration) PopulateResourceData(d *schema.ResourceData) {
	for key, value := range i.ToMap() {
		d.Set(key, value)
	}
}

func (i *AWSIntegration) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"duration_seconds":               i.DurationSeconds,
		"generate_credentials_in_worker": i.GenerateCredentialsInWorker,
		"external_id":                    i.ExternalID,
		"name":                           i.Name,
		"role_arn":                       i.RoleARN,
		"space_id":                       i.Space,
		"labels":                         i.getLabelsSet(),
		"region":                         i.Region,
		"autoattach_enabled":             i.AutoattachEnabled,
		"tag_assume_role":                i.TagAssumeRole,
	}
}

func (i *AWSIntegration) getLabelsSet() *schema.Set {
	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range i.Labels {
		labels.Add(label)
	}
	return labels
}
