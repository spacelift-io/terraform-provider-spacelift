package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackDestructorResource(t *testing.T) {
	t.Parallel()

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
			importByStackIDStep(resourceName, "spacelift_stack.test2", map[string]string{
				"discard_runs": "false",
			}),
		})
	})
}

// importByStackIDStep imports a resource whose import ID is the ID of the
// stack it points at, and checks that stack_id ends up in the imported state.
// ImportStateVerify can't be used: these resources get a new ID on every read.
func importByStackIDStep(resourceName, stackResourceName string, want map[string]string) resource.TestStep {
	var stackID string

	return resource.TestStep{
		ResourceName: resourceName,
		ImportState:  true,
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			stack, ok := s.RootModule().Resources[stackResourceName]
			if !ok {
				return "", fmt.Errorf("%s not found in state", stackResourceName)
			}
			stackID = stack.Primary.ID
			return stackID, nil
		},
		ImportStateCheck: func(states []*terraform.InstanceState) error {
			if len(states) != 1 {
				return fmt.Errorf("expected 1 imported resource, got %d", len(states))
			}
			attributes := states[0].Attributes
			if got := attributes["stack_id"]; got != stackID {
				return fmt.Errorf("stack_id: got %q, want %q", got, stackID)
			}
			for key, value := range want {
				if got := attributes[key]; got != value {
					return fmt.Errorf("%s: got %q, want %q", key, got, value)
				}
			}
			return nil
		},
	}
}

func TestDestroyStackDestructor(t *testing.T) {
	t.Parallel()

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
				wait {
					continue_on_state = ["unconfirmed"]
				}
			}

			resource "spacelift_stack_destructor" "test" {
				depends_on = [
					spacelift_run.test,
				]
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
