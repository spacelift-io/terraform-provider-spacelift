package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackDependencyResource(t *testing.T) {
	const resourceName = "spacelift_stack_dependency.test"

	t.Run("creates and updates stack dependency", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func() string {
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
		`, randomID, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", StartsWith("my-first-stack")),
					Attribute("depends_on_stack_id", StartsWith("my-second-stack")),
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
