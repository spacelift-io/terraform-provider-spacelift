package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

// repoWithFileConfig is a repo holding one commit; a stack cannot attach to one with no revisions.
func repoWithFileConfig(name string) string {
	return repoConfig(name) + `
		resource "spacelift_repo_file" "test" {
			repo_id = spacelift_repo.test.id
			path    = "main.tf"
			content = "# managed by terraform"
		}
	`
}

func TestVCSIntegrationSpacelift(t *testing.T) {
	t.Parallel()

	const resourceName = "spacelift_stack.test"

	t.Run("attaches a Spacelift repo to a stack", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("stack-repo-%s", randID)

		config := repoWithFileConfig(repoName) + fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name       = "spacelift-repo-stack-%s"
				repository = spacelift_repo.test.id
				branch     = "main"
				space_id   = "root"
				spacelift_repo {}

				depends_on = [spacelift_repo_file.test]
			}
		`, randID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("spacelift_repo.#", Equals("1")),
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

		config := repoWithFileConfig(fmt.Sprintf("stack-repo-branch-%s", randID)) + fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name       = "spacelift-repo-bad-branch-%s"
				repository = spacelift_repo.test.id
				branch     = "develop"
				space_id   = "root"
				spacelift_repo {}
			}
		`, randID)

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`branch must be "main" when using a Spacelift repo`),
			},
		})
	})

	t.Run("conflicts with another VCS provider block", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name       = "spacelift-repo-conflict-%s"
				repository = "some-repo"
				branch     = "main"
				space_id   = "root"
				spacelift_repo {}
				gitlab {
					namespace = "some-namespace"
				}
			}
		`, randID)

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`conflicts with`),
			},
		})
	})
}
