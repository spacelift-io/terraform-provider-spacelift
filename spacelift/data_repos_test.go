package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestReposData(t *testing.T) {
	const datasourceName = "data.spacelift_repos.test"

	t.Run("finds a repo by label in its space", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("repos-data-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_repo" "test" {
				name     = "%s"
				space_id = "root"
				labels   = ["%s"]
			}

			data "spacelift_repos" "test" {
				space_id = "root"
				labels   = ["%s"]

				depends_on = [spacelift_repo.test]
			}
		`, repoName, randID, randID)

		testSteps(t, []resource.TestStep{{
			Config: config,
			Check: Resource(
				datasourceName,
				Attribute("id", IsNotEmpty()),
				Attribute("repos.#", Equals("1")),
				Nested("repos", CheckInList(
					Attribute("repo_id", Equals(repoName)),
					Attribute("name", Equals(repoName)),
				)),
			),
		}})
	})

	t.Run("returns nothing when no repo matches the labels", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := repoConfig(fmt.Sprintf("repos-data-empty-%s", randID)) + fmt.Sprintf(`
			data "spacelift_repos" "test" {
				space_id = "root"
				labels   = ["no-repo-carries-%s"]

				depends_on = [spacelift_repo.test]
			}
		`, randID)

		testSteps(t, []resource.TestStep{{
			Config: config,
			Check: Resource(
				datasourceName,
				Attribute("repos.#", Equals("0")),
			),
		}})
	})
}
