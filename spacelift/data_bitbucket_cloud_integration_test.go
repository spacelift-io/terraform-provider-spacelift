package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBitbucketCloudIntegrationData(t *testing.T) {
	t.Run("without the id specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.BitbucketCloud.Default
		testSteps(t, []resource.TestStep{
			{
				Config: `
					data "spacelift_bitbucket_cloud_integration" "test" {}
				`,
				Check: Resource(
					"data.spacelift_bitbucket_cloud_integration.test",
					Attribute("id", Equals(cfg.ID)),
					Attribute("name", Equals(cfg.Name)),
					Attribute("is_default", Equals("true")),
					Attribute("space_id", Equals("root")),
					Attribute("username", Equals(cfg.Username)),
					Attribute("webhook_url", Equals(cfg.WebhookURL)),
					Attribute("vcs_checks", Equals(cfg.VCSChecks)),
				),
			},
		})
	})

	t.Run("with the id specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.BitbucketCloud.SpaceLevel
		testSteps(t, []resource.TestStep{
			{
				Config: `
					data "spacelift_bitbucket_cloud_integration" "test" {
						id = "` + cfg.ID + `"
					}
				`,
				Check: Resource(
					"data.spacelift_bitbucket_cloud_integration.test",
					Attribute("id", Equals(cfg.ID)),
					Attribute("name", Equals(cfg.Name)),
					Attribute("is_default", Equals("false")),
					Attribute("space_id", Equals(cfg.Space)),
					Attribute("username", Equals(cfg.Username)),
					Attribute("webhook_url", Equals(cfg.WebhookURL)),
					Attribute("vcs_checks", Equals(cfg.VCSChecks)),
				),
			},
		})
	})
}
