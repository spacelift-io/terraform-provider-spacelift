package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledDeleteStackResource(t *testing.T) {
	t.Parallel()
	const resourceName = "spacelift_scheduled_delete_stack.test"

	t.Run("for scheduled delete_stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		deleteStackConfig := func(at string, shouldDeleteResources bool) string {
			return fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch     = "master"
				repository = "demo"
				name       = "Test stack %s"
			}
	
			resource "spacelift_scheduled_delete_stack" "test" {
				stack_id = spacelift_stack.test.id
	
				at               = "%s"
				delete_resources = %t
			}
		`, randomID, at, shouldDeleteResources)
		}

		deleteStackConfigWithOnlyRequiredAttributes := func(at string) string {
			return fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch     = "master"
				repository = "demo"
				name       = "Test stack %s"
			}
		
			resource "spacelift_scheduled_delete_stack" "test" {
				stack_id = spacelift_stack.test.id
		
				at = "%s"
			}
		`, randomID, at)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: deleteStackConfig("123", false),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", StartsWith("test-stack-")),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("at", Equals("123")),
					Attribute("delete_resources", Equals("false")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: deleteStackConfigWithOnlyRequiredAttributes("321"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", StartsWith("test-stack-")),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("at", Equals("321")),
					Attribute("delete_resources", Equals("true")),
				),
			},
		})
	})
}
