package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackDependencyReferenceResource(t *testing.T) {
	const resourceName = "spacelift_stack_dependency_reference.test"

	t.Run("creates and updates stack dependency reference", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(outputName, inputName string) string {
			return fmt.Sprintf(`
			resource "spacelift_stack" "test1" {
				branch     = "master"
				repository = "demo"
				name       = "my-first-stack-%s"
			}

			resource "spacelift_stack" "test2" {
				branch     = "master"
				repository = "demo"
				name       = "my-second-stack-%s"
			}

			resource "spacelift_stack_dependency" "test" {
				stack_id = spacelift_stack.test1.id
				depends_on_stack_id = spacelift_stack.test2.id
			}

			resource "spacelift_stack_dependency_reference" "test" {
				stack_dependency_id = spacelift_stack_dependency.test.id
				output_name = "%s"
				input_name = "%s"
			}
		`, randomID, randomID, outputName, inputName)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("output_abc", "input_123"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("output_name", Equals("output_abc")),
					Attribute("input_name", Equals("input_123")),
				),
			},
			{
				Config: config("output_abc", "input_456"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("output_name", Equals("output_abc")),
					Attribute("input_name", Equals("input_456")),
				),
			},
			{
				Config: config("output_xyz", "input_456"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("output_name", Equals("output_xyz")),
					Attribute("input_name", Equals("input_456")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})
}
