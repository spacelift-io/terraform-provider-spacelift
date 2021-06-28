package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackDestructorResource(t *testing.T) {
	const resourceName = "spacelift_stack_destructor.test"

	t.Parallel()

	t.Run("test destructor", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(deactivated bool) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_stack_destructor" "test" {
					stack_id    = spacelift_stack.test.id
					deactivated = %t
				}
			`, randomID, deactivated)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(false),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("deactivated", Equals("false")),
				),
			},
			{
				Config: config(true),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("deactivated", Equals("true")),
				),
			},
		})
	})
}
