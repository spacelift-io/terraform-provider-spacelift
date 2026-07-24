package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

// repoConfig is a repo to hang files, stacks and modules off in tests.
func repoConfig(name string) string {
	return fmt.Sprintf(`
		resource "spacelift_repo" "test" {
			name     = "%s"
			space_id = "root"
		}
	`, name)
}

func TestRepoFileResource(t *testing.T) {
	const resourceName = "spacelift_repo_file.test"

	t.Run("creates, updates and imports a file", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("repo-file-test-%s", randID)

		config := func(content string, mode string) string {
			return repoConfig(repoName) + fmt.Sprintf(`
				resource "spacelift_repo_file" "test" {
					repo_id   = spacelift_repo.test.id
					path      = "modules/vpc/main.tf"
					content   = "%s"
					file_mode = "%s"
				}
			`, content, mode)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("# first", repoFileDefaultMode),
				Check: Resource(
					resourceName,
					Attribute("id", Equals(repoName+"/modules/vpc/main.tf")),
					Attribute("path", Equals("modules/vpc/main.tf")),
					Attribute("content", Equals("# first")),
					Attribute("file_mode", Equals("0644")),
					Attribute("encrypt", Equals("false")),
					Attribute("revision_sha", IsNotEmpty()),
					Attribute("size_bytes", Equals("7")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Commit metadata belongs to the revision, not the file, so it
				// cannot be read back.
				ImportStateVerifyIgnore: []string{"commit_message", "author_name", "author_email"},
			},
			{
				Config: config("# second", "0755"),
				Check: Resource(
					resourceName,
					Attribute("content", Equals("# second")),
					Attribute("file_mode", Equals("0755")),
					Attribute("size_bytes", Equals("8")),
				),
			},
		})
	})

	t.Run("records the commit metadata it is given", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("repo-file-commit-%s", randID)

		config := repoConfig(repoName) + `
			resource "spacelift_repo_file" "test" {
				repo_id        = spacelift_repo.test.id
				path           = "main.tf"
				content        = "# hello"
				commit_message = "Add the entrypoint"
				author_name    = "Francis Bacon"
				author_email   = "francis@example.com"
			}
		`

		testSteps(t, []resource.TestStep{{
			Config: config,
			Check: Resource(
				resourceName,
				Attribute("commit_message", Equals("Add the entrypoint")),
				Attribute("author_name", Equals("Francis Bacon")),
				Attribute("revision_sha", IsNotEmpty()),
			),
		}})
	})

	t.Run("does not read back the contents of an encrypted file", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("repo-file-encrypted-%s", randID)

		config := repoConfig(repoName) + `
			resource "spacelift_repo_file" "test" {
				repo_id = spacelift_repo.test.id
				path    = "secret.tfvars"
				content = "token = \"bacon\""
				encrypt = true
			}
		`

		testSteps(t, []resource.TestStep{{
			Config: config,
			Check: Resource(
				resourceName,
				Attribute("encrypt", Equals("true")),
				// State keeps what Terraform wrote, since Spacelift withholds it.
				Attribute("content", Equals(`token = "bacon"`)),
			),
		}})
	})
}
