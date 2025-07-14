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

	t.Run("test destructor", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(cancelPendingRuns, deactivated bool, stackID int) string {
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

				resource "spacelift_stack_destructor" "test" {
					stack_id     = spacelift_stack.test%d.id
					deactivated  = %t
					discard_runs = %t
				}
			`, randomID, randomID, stackID, deactivated, cancelPendingRuns)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(false, false, 1),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains("number-1")),
					Attribute("deactivated", Equals("false")),
					Attribute("discard_runs", Equals("false")),
				),
			},
			{
				Config: config(true, true, 2),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains("number-2")),
					Attribute("deactivated", Equals("true")),
					Attribute("discard_runs", Equals("true")),
				),
			},
		})
	})
}

func TestDestroyStackDestructor(t *testing.T) {
	const resourceName = "spacelift_stack_destructor.test"

	t.Run("test destroy with runs", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch     = "master"
				repository = "demo"
				name       = "Stack %s"
				autodeploy = false
			}

			resource "spacelift_run" "test" {
				stack_id = spacelift_stack.test.id
				keepers = {
					"test": "value"
				}
			}

			resource "spacelift_stack_destructor" "test" {
				stack_id     = spacelift_stack.test.id
				discard_runs = true
			}
		`, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
					Attribute("discard_runs", Equals("true")),
				),
			},
		})
	})
}
