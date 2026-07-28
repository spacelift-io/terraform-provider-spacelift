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

		config := repoWithFileConfig(fmt.Sprintf("stack-repo-branch-%s", randID)) + fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name       = "spacelift-repo-bad-branch-%s"
				repository = spacelift_repo.test.id
				branch     = "develop"
				space_id   = "root"
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

	t.Run("rejects a repository that is not the repo ID", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name       = "spacelift-repo-mismatch-%s"
				repository = "not-the-repo"
				branch     = "main"
				space_id   = "root"
				spacelift {
					id = "some-repo"
				}
			}
		`, randID)

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`repository must be the Spacelift repo ID \(slug\) "some-repo"`),
			},
		})
	})

	t.Run("accepts a literal repository while the repo id is still unknown", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("stack-repo-unknown-%s", randID)

		// The literal repository matches the slug, while the reference to it stays unknown at plan time.
		config := repoWithFileConfig(repoName) + fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name       = "spacelift-repo-unknown-%s"
				repository = "%s"
				branch     = "main"
				space_id   = "root"
				spacelift {
					id = spacelift_repo.test.id
				}

				depends_on = [spacelift_repo_file.test]
			}
		`, randID, repoName)

		testSteps(t, []resource.TestStep{{
			Config: config,
			Check: Resource(
				resourceName,
				Attribute("repository", Equals(repoName)),
			),
		}})
	})

	t.Run("conflicts with another VCS provider block", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name       = "spacelift-repo-conflict-%s"
				repository = "some-repo"
				branch     = "main"
				space_id   = "root"
				spacelift {
					id = "some-repo"
				}
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
