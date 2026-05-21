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
					Attribute("id", Contains("my-first-space")),
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
					SetEquals("labels"),
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
					Attribute("id", Contains("my-second-space")),
					Attribute("parent_space_id", Contains("my-first-space")),
				),
			},
		})
	})
	t.Run("creates a space with labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
				resource "spacelift_space" "test" {
					name = "My first space %s"
					parent_space_id = "root"
					inherit_entities = true
					labels = ["label1", "label2"]
				}
			`, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("id", Contains("my-first-space")),
					SetEquals("labels", "label1", "label2"),
				),
			},
		})
	})
	t.Run("adopts an existing space when adopt_existing is set", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		spaceName := fmt.Sprintf("Adopted space %s", randomID)

		createConfig := fmt.Sprintf(`
				resource "spacelift_space" "original" {
					name = "%s"
					parent_space_id = "root"
					description = "created by the first stack"
					inherit_entities = true
				}
			`, spaceName)

		// Second step swaps the resource block for an adopt-existing one declaring the
		// same name+parent. Terraform sees a different resource address (`adopter` vs.
		// `original`) so the new resource starts with no state and must take over the
		// row inserted by the first step via the create-time adopt path.
		adoptConfig := fmt.Sprintf(`
				resource "spacelift_space" "adopter" {
					name = "%s"
					parent_space_id = "root"
					description = "adopted by the second stack"
					inherit_entities = true
					adopt_existing = true
				}
			`, spaceName)

		testSteps(t, []resource.TestStep{
			{
				Config: createConfig,
				Check: Resource(
					"spacelift_space.original",
					Attribute("id", Contains("adopted-space")),
				),
			},
			{
				Config: adoptConfig,
				Check: Resource(
					"spacelift_space.adopter",
					Attribute("id", Contains("adopted-space")),
					Attribute("adopt_existing", Equals("true")),
				),
			},
		})
	})
}
