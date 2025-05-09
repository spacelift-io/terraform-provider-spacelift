package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVCSIntegrationGitlab(t *testing.T) {
	const resourceName = "spacelift_stack.test"

	t.Run("with_default_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("gitlab-with-default-integration-implicit-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "root"
				administrative     = false
				gitlab {
					namespace = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.Gitlab.Repository.Name,
			testConfig.SourceCode.Gitlab.Repository.Branch,
			testConfig.SourceCode.Gitlab.Repository.Namespace)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("gitlab.0.id", Equals(testConfig.SourceCode.Gitlab.Default.ID)),
				),
			},
		})
	})

	t.Run("with_space_level_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("gitlab-with-space-level-integration-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "%s"
				administrative     = false
				gitlab {
					namespace = "%s"
					id = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.Gitlab.Repository.Name,
			testConfig.SourceCode.Gitlab.Repository.Branch,
			testConfig.SourceCode.Gitlab.SpaceLevel.Space,
			testConfig.SourceCode.Gitlab.Repository.Namespace,
			testConfig.SourceCode.Gitlab.SpaceLevel.ID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("gitlab.0.id", Equals(testConfig.SourceCode.Gitlab.SpaceLevel.ID)),
				),
			},
		})
	})
}
