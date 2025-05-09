package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVCSIntegrationBitbucketCloud(t *testing.T) {
	const resourceName = "spacelift_stack.test"

	t.Run("with_default_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("bitbucket-cloud-with-default-integration-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "root"
				administrative     = false
				bitbucket_cloud {
					namespace = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.BitbucketCloud.Repository.Name,
			testConfig.SourceCode.BitbucketCloud.Repository.Branch,
			testConfig.SourceCode.BitbucketCloud.Repository.Namespace)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("bitbucket_cloud.0.id", Equals(testConfig.SourceCode.BitbucketCloud.Default.ID)),
				),
			},
		})
	})

	t.Run("with_space_level_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("bitbucket-cloud-with-space-level-integration-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "%s"
				administrative     = false
				bitbucket_cloud {
					namespace = "%s"
					id = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.BitbucketCloud.Repository.Name,
			testConfig.SourceCode.BitbucketCloud.Repository.Branch,
			testConfig.SourceCode.BitbucketCloud.SpaceLevel.Space,
			testConfig.SourceCode.BitbucketCloud.Repository.Namespace,
			testConfig.SourceCode.BitbucketCloud.SpaceLevel.ID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("bitbucket_cloud.0.id", Equals(testConfig.SourceCode.BitbucketCloud.SpaceLevel.ID)),
				),
			},
		})
	})
}
