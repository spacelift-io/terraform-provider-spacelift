package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledTaskData(t *testing.T) {
	t.Run("task scheduling config", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		taskConfigWithEvery := func(command string, every []string, timezone string) string {
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

				data "spacelift_scheduled_task" "test" {
					scheduled_task_id   = spacelift_scheduled_task.test.id
				}
			`, randomID, command, strings.Join(everyStrs, ", "), timezone)
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
		
				data "spacelift_scheduled_task" "test" {
					scheduled_task_id = spacelift_scheduled_task.test.id
				}
			`, randomID, command, at)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: taskConfigWithEvery("terraform apply -auto-approve", []string{"*/3 * * * *", "*/4 * * * *"}, "CET"),
				Check: Resource(
					"data.spacelift_scheduled_task.test",
					Attribute("scheduled_task_id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("command", Equals("terraform apply -auto-approve")),
					Attribute("timezone", Equals("CET")),
					Attribute("every.#", Equals("2")),
					Attribute("every.0", Equals("*/3 * * * *")),
					Attribute("every.1", Equals("*/4 * * * *")),
				),
			},
			{
				Config: taskConfigWithAt("terraform apply -auto-approve", "1234567"),
				Check: Resource(
					"data.spacelift_scheduled_task.test",
					Attribute("scheduled_task_id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("command", Equals("terraform apply -auto-approve")),
					Attribute("at", Equals("1234567")),
				),
			},
		})
	})
}
