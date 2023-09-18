package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackActivatorResource(t *testing.T) {
	const resourceName = "spacelift_stack_activator.test"

	t.Run("test activator", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(enabled bool, stackID int) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test1" {
					branch     = "master"
					repository = "demo"
					name       = "Stack %s number 1"
				}

				resource "spacelift_stack" "test2" {
					branch     = "master"
					repository = "demo"
					name       = "Stack %s number 2"
				}

				resource "spacelift_stack_activator" "test" {
					stack_id = spacelift_stack.test%d.id
					enabled  = %t
				}
			`, randomID, randomID, stackID, enabled)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(false, 1),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains("number-1")),
					Attribute("enabled", Equals("false")),
				),
			},
			{
				Config: config(true, 2),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains("number-2")),
					Attribute("enabled", Equals("true")),
				),
			},
		})
	})
}
