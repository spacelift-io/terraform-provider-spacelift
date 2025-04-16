package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBlueprintVersionedGroupResource(t *testing.T) {
	const resourceName = "spacelift_blueprint_versioned_group.test"

	t.Run("Creates and updates a blueprint group", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_blueprint_versioned_group" "test" {
					name        = "test-blueprint-%s"
					space       = "root"
					description = "%s"
					labels      = ["one", "two"]
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
				),
			},
		})
	})
}
