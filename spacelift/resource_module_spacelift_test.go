package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModuleResourceSpacelift(t *testing.T) {
	const resourceName = "spacelift_module.test"

	t.Run("attaches a module to a Spacelift repo", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("module-repo-%s", randID)

		config := repoWithFileConfig(repoName) + fmt.Sprintf(`
			resource "spacelift_module" "test" {
				name               = "spacelift-repo-module-%s"
				repository         = spacelift_repo.test.id
				branch             = "main"
				space_id           = "root"
				terraform_provider = "default"
				spacelift {
					id = spacelift_repo.test.id
				}

				depends_on = [spacelift_repo_file.test]
			}
		`, randID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("spacelift.0.id", Equals(repoName)),
					Attribute("repository", Equals(repoName)),
					Attribute("branch", Equals("main")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})

	t.Run("rejects a branch other than main", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := repoWithFileConfig(fmt.Sprintf("module-repo-branch-%s", randID)) + fmt.Sprintf(`
			resource "spacelift_module" "test" {
				name               = "spacelift-repo-module-bad-%s"
				repository         = spacelift_repo.test.id
				branch             = "develop"
				space_id           = "root"
				terraform_provider = "default"
				spacelift {
					id = spacelift_repo.test.id
				}
			}
		`, randID)

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`branch must be "main" when using a Spacelift repo`),
			},
		})
	})
}
