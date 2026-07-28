package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestRepoResource(t *testing.T) {
	const resourceName = "spacelift_repo.test"

	t.Run("creates, updates and imports a repo", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("repo-test-%s", randID)

		config := func(description string, labels string) string {
			return fmt.Sprintf(`
				resource "spacelift_repo" "test" {
					name        = "%s"
					space_id    = "root"
					description = "%s"
					labels      = [%s]
				}
			`, name, description, labels)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description", `"one", "two"`),
				Check: Resource(
					resourceName,
					Attribute("id", Equals(name)),
					Attribute("name", Equals(name)),
					Attribute("space_id", Equals("root")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("vcs_checks", Equals("INDIVIDUAL")),
					Attribute("created_at", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description", `"three"`),
				Check: Resource(
					resourceName,
					// The slug is derived once at creation, so it survives edits.
					Attribute("id", Equals(name)),
					Attribute("description", Equals("new description")),
					SetEquals("labels", "three"),
				),
			},
		})
	})

	t.Run("keeps its ID when renamed", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		originalName := fmt.Sprintf("repo-rename-%s", randID)
		newName := fmt.Sprintf("repo-renamed-%s", randID)

		config := func(name string) string {
			return fmt.Sprintf(`
				resource "spacelift_repo" "test" {
					name     = "%s"
					space_id = "root"
				}
			`, name)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(originalName),
				Check:  Resource(resourceName, Attribute("id", Equals(originalName))),
			},
			{
				Config: config(newName),
				Check: Resource(
					resourceName,
					Attribute("name", Equals(newName)),
					Attribute("id", Equals(originalName)),
				),
			},
		})
	})
}
