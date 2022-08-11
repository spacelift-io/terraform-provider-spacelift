package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledTaskResource(t *testing.T) {
	const resourceName = "spacelift_scheduled_task.test"

	t.Run("for scheduled task", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		taskConfig := func(command string, every []string, timezone string) string {
			everyStrs := make([]string, len(every))
			for i := range every {
				everyStrs[i] = `"` + every[i] + `"`
			}

			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_task" "test" {
					stack_id = spacelift_stack.test.id
		
					command    = "%s"
					every      = [%s]
					timezone   = "%s"
				}
			`, randomID, command, strings.Join(everyStrs, ", "), timezone)
		}

		taskConfigWithoutTimezone := func(command string, every []string) string {
			everyStrs := make([]string, len(every))
			for i := range every {
				everyStrs[i] = `"` + every[i] + `"`
			}

			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_task" "test" {
					stack_id = spacelift_stack.test.id
		
					command    = "%s"
					every      = [%s]
				}
			`, randomID, command, strings.Join(everyStrs, ", "))
		}

		taskConfigWithAt := func(command string, at string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_task" "test" {
					stack_id = spacelift_stack.test.id
		
					command = "%s"
					at      = "%s"
				}
			`, randomID, command, at)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: taskConfig("terraform apply -auto-approve", []string{"*/3 * * * *", "*/4 * * * *"}, "CET"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("command", Equals("terraform apply -auto-approve")),
					Attribute("timezone", Equals("CET")),
					Attribute("every.#", Equals("2")),
					Attribute("every.0", Equals("*/3 * * * *")),
					Attribute("every.1", Equals("*/4 * * * *")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: taskConfigWithoutTimezone("terraform apply -auto-approve", []string{"*/5 * * * *"}),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("command", Equals("terraform apply -auto-approve")),
					Attribute("timezone", Equals("UTC")),
					Attribute("every.#", Equals("1")),
					Attribute("every.0", Equals("*/5 * * * *")),
				),
			},
			{
				Config: taskConfigWithAt("terraform apply -auto-approve", "1234567"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("command", Equals("terraform apply -auto-approve")),
					Attribute("at", Equals("1234567")),
				),
			},
		})
	})
}
