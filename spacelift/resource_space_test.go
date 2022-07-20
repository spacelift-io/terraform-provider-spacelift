package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSpaceResource(t *testing.T) {
	const resourceName = "spacelift_space.test"

	t.Run("creates and updates a space", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(message string) string {
			return fmt.Sprintf(`
				resource "spacelift_space" "test" {
					name = "My first space %s"
					parent_space_id = "root"
					description = "%s"
					inherit_entities = true
				}
			`, randomID, message)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("boom"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("my-first-space")),
					Attribute("description", Contains("boom")),
					Attribute("parent_space_id", Equals("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("bang"),
				Check: Resource(
					resourceName,
					Attribute("description", Contains("bang")),
				),
			},
		})
	})
	t.Run("creates a space and a child", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
				resource "spacelift_space" "test" {
					name = "My first space %s"
					parent_space_id = "root"
					inherit_entities = true
				}
				resource "spacelift_space" "test-child" {
					name = "My second space %s"
					parent_space_id = spacelift_space.test.id
					inherit_entities = true
				}
			`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					"spacelift_space.test-child",
					Attribute("id", StartsWith("my-second-space")),
					Attribute("description", Contains("boom")),
					Attribute("parent_space_id", StartsWith("my-first-space")),
				),
			},
		})
	})
}
