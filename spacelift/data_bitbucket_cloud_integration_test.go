package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBitbucketCloudIntegrationData(t *testing.T) {
	t.Run("without the id specified", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: `
				data "spacelift_bitbucket_cloud_integration" "test" {}
			`,
				Check: Resource(
					"data.spacelift_bitbucket_cloud_integration.test",
					Attribute("id", IsNotEmpty()),
					Attribute("name", IsNotEmpty()),
					Attribute("is_default", Equals("true")),
					Attribute("space_id", IsNotEmpty()),
					Attribute("username", IsNotEmpty()),
					Attribute("webhook_url", IsNotEmpty()),
				),
			},
		})
	})

	t.Run("with the id specified", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: `
				data "spacelift_bitbucket_cloud_integration" "test" {
					id = "bitbucket-cloud-default-integration"
				}
			`,
				Check: Resource(
					"data.spacelift_bitbucket_cloud_integration.test",
					Attribute("id", IsNotEmpty()),
					Attribute("name", IsNotEmpty()),
					Attribute("is_default", Equals("true")),
					Attribute("space_id", IsNotEmpty()),
					Attribute("username", IsNotEmpty()),
					Attribute("webhook_url", IsNotEmpty()),
				),
			},
		})
	})
}
