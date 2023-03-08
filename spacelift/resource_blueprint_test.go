package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBlueprintResource(t *testing.T) {
	const resourceName = "spacelift_blueprint.test"

	t.Run("Creates and updates a blueprint in DRAFT state", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_blueprint" "test" {
					name        = "test-blueprint-%s"
					space       = "root"
					description = "%s"
					labels      = ["one", "two"]
					state       = "DRAFT"
					template    = "not validated for drafts"
				}`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("test description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals("test-blueprint-"+randomID)),
					Attribute("space", Equals("root")),
					Attribute("description", Equals("test description")),
					Attribute("labels.#", Equals("2")),
					Attribute("state", Equals("DRAFT")),
					Attribute("template", Equals("not validated for drafts")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("updated description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals("test-blueprint-"+randomID)),
					Attribute("space", Equals("root")),
					Attribute("description", Equals("updated description")),
					Attribute("labels.#", Equals("2")),
					Attribute("state", Equals("DRAFT")),
					Attribute("template", Equals("not validated for drafts")),
				),
			},
		})
	})

	t.Run("Creates and updates a blueprint in PUBLISHED state", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		validTemplate1 := `stack:\n  name: stackerino\n  space: root\n  vcs:\n    branch: main\n    repository: spacelift-io/terraform-provider-spacelift\n    provider: GITHUB\n  vendor:\n    terraform:\n      manage_state: true\n      version: 0.12.0`
		validTemplate2 := `stack:\n  name: stackerino\n  space: root\n  vcs:\n    branch: main\n    repository: spacelift-io/terraform-provider-spacelift\n    provider: GITHUB\n  vendor:\n    terraform:\n      manage_state: true\n      version: 0.13.0`

		config := func(template, description string) string {
			return fmt.Sprintf(`
				resource "spacelift_blueprint" "test" {
					name        = "test-blueprint-%s"
					space       = "root"
					description = "%s"
					labels      = ["one", "two"]
					state       = "PUBLISHED"
					template    = "%s"
				}`, randomID, description, template)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(validTemplate1, "test description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals("test-blueprint-"+randomID)),
					Attribute("space", Equals("root")),
					Attribute("description", Equals("test description")),
					Attribute("labels.#", Equals("2")),
					Attribute("state", Equals("PUBLISHED")),
					Attribute("template", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config(validTemplate2, "updated description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals("test-blueprint-"+randomID)),
					Attribute("space", Equals("root")),
					Attribute("description", Equals("updated description")),
					Attribute("labels.#", Equals("2")),
					Attribute("state", Equals("PUBLISHED")),
					Attribute("template", IsNotEmpty()),
				),
			},
		})
	})
}
