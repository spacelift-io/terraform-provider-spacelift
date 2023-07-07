package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestDriftDetectionResource(t *testing.T) {
	const resourceName = "spacelift_drift_detection.test"

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(reconcile, ignoreState bool, schedule []string) string {
			scheduleStrs := make([]string, len(schedule))
			for i := range schedule {
				scheduleStrs[i] = `"` + schedule[i] + `"`
			}

			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_drift_detection" "test" {
					stack_id     = spacelift_stack.test.id
					reconcile    = %t
					ignore_state = %t
					schedule     = [%s]
				}
			`, randomID, reconcile, ignoreState, strings.Join(scheduleStrs, ", "))
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(true, true, []string{"*/3 * * * *", "*/4 * * * *"}),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
					Attribute("reconcile", Equals("true")),
					Attribute("ignore_state", Equals("true")),
					Attribute("timezone", Equals("UTC")),
					Attribute("schedule.#", Equals("2")),
					Attribute("schedule.0", Equals("*/3 * * * *")),
					Attribute("schedule.1", Equals("*/4 * * * *")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("stack/test-stack-%s", randomID),
				ImportStateVerify: true,
			},
			{
				Config: config(false, false, []string{"*/5 * * * *"}),
				Check: Resource(
					resourceName,
					Attribute("reconcile", Equals("false")),
					Attribute("ignore_state", Equals("false")),
					Attribute("schedule.#", Equals("1")),
					Attribute("schedule.0", Equals("*/5 * * * *")),
				),
			},
		})
	})
}
