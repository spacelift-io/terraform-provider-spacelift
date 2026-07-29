package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestRepoData(t *testing.T) {
	const datasourceName = "data.spacelift_repo.test"

	t.Run("reads an existing repo", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("repo-data-%s", randID)

		config := repoConfig(repoName) + `
			data "spacelift_repo" "test" {
				repo_id = spacelift_repo.test.id
			}
		`

		testSteps(t, []resource.TestStep{{
			Config: config,
			Check: Resource(
				datasourceName,
				Attribute("id", Equals(repoName)),
				Attribute("repo_id", Equals(repoName)),
				Attribute("name", Equals(repoName)),
				Attribute("space_id", Equals("root")),
				Attribute("vcs_checks", Equals("INDIVIDUAL")),
				Attribute("created_at", IsNotEmpty()),
			),
		}})
	})

	t.Run("errors out for a repo that does not exist", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_repo" "test" {
					repo_id = "this-repo-does-not-exist"
				}
			`,
			ExpectError: regexp.MustCompile(`could not find repo this-repo-does-not-exist`),
		}})
	})
}
