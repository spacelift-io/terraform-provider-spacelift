package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledDeleteStack(t *testing.T) {
	t.Run("scheduled delete_stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		at := "123"

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					name       = "Test stack %s"
					repository = "demo"
				}
	
				resource "spacelift_scheduled_delete_stack" "test" {
					stack_id = spacelift_stack.test.id
	
					at = "%s"
				}

				data "spacelift_scheduled_delete_stack" "test" {
					scheduled_delete_stack_id = spacelift_scheduled_delete_stack.test.id
				}
			`, randomID, at),
			Check: Resource(
				"data.spacelift_scheduled_delete_stack.test",
				Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
				Attribute("stack_id", StartsWith("test-stack-")),
				Attribute("stack_id", Contains(randomID)),
				Attribute("at", Equals(at)),
				Attribute("delete_resources", Equals("true")),
			),
		}})
	})
}
