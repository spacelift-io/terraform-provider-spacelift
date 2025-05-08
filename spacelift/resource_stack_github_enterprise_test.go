package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVCSIntegrationGithubEnterprise(t *testing.T) {
	const resourceName = "spacelift_stack.test"

	t.Run("with_default_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("github-enterprise-with-default-integration-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "root"
				administrative     = false
				github_enterprise {
					namespace = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.GithubEnterprise.Repository.Name,
			testConfig.SourceCode.GithubEnterprise.Repository.Branch,
			testConfig.SourceCode.GithubEnterprise.Repository.Namespace)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("github_enterprise.0.id", Equals(testConfig.SourceCode.GithubEnterprise.Default.ID)),
				),
			},
		})
	})

	t.Run("with_space_level_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("github-enterprise-with-space-level-integration-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "%s"
				administrative     = false
				github_enterprise {
					namespace = "%s"
					id = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.GithubEnterprise.Repository.Name,
			testConfig.SourceCode.GithubEnterprise.Repository.Branch,
			testConfig.SourceCode.GithubEnterprise.SpaceLevel.Space,
			testConfig.SourceCode.GithubEnterprise.Repository.Namespace,
			testConfig.SourceCode.GithubEnterprise.SpaceLevel.ID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("github_enterprise.0.id", Equals(testConfig.SourceCode.GithubEnterprise.SpaceLevel.ID)),
				),
			},
		})
	})
}
