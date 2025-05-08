package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVCSIntegrationBitbucketDatacenter(t *testing.T) {
	const resourceName = "spacelift_stack.test"

	t.Run("with_default_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("bitbucket-datacenter-with-default-integration-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "root"
				administrative     = false
				bitbucket_datacenter {
					namespace = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.BitbucketDatacenter.Repository.Name,
			testConfig.SourceCode.BitbucketDatacenter.Repository.Branch,
			testConfig.SourceCode.BitbucketDatacenter.Repository.Namespace)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("bitbucket_datacenter.0.id", Equals(testConfig.SourceCode.BitbucketDatacenter.Default.ID)),
				),
			},
		})
	})

	t.Run("with_space_level_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("bitbucket-datacenter-with-space-level-integration-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "%s"
				administrative     = false
				bitbucket_datacenter {
					namespace = "%s"
					id = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.BitbucketDatacenter.Repository.Name,
			testConfig.SourceCode.BitbucketDatacenter.Repository.Branch,
			testConfig.SourceCode.BitbucketDatacenter.SpaceLevel.Space,
			testConfig.SourceCode.BitbucketDatacenter.Repository.Namespace,
			testConfig.SourceCode.BitbucketDatacenter.SpaceLevel.ID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("bitbucket_datacenter.0.id", Equals(testConfig.SourceCode.BitbucketDatacenter.SpaceLevel.ID)),
				),
			},
		})
	})
}
