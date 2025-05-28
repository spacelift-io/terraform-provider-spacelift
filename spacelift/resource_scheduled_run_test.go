package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledRunResource(t *testing.T) {
	const resourceName = "spacelift_scheduled_run.test"

	t.Run("for scheduled run", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		runConfig := func(name string, every []string, timezone string) string {
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
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
		
					name       = "%s"
					every      = [%s]
					timezone   = "%s"
				}
			`, randomID, name, strings.Join(everyStrs, ", "), timezone)
		}

		runConfigWithoutTimezone := func(name string, every []string) string {
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
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
		
					name  = "%s"
					every = [%s]
				}
			`, randomID, name, strings.Join(everyStrs, ", "))
		}

		runConfigWithAt := func(name string, at string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
		
					name = "%s"
					at   = "%s"
				}
			`, randomID, name, at)
		}

		runConfigWithRuntimeConfig := func(name string, every []string, runtimeConfig string) string {
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
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
		
					name           = "%s"
					every          = [%s]
					runtime_config = "%s"
				}
			`, randomID, name, strings.Join(everyStrs, ", "), runtimeConfig)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: runConfig("test-run-apply", []string{"*/3 * * * *", "*/4 * * * *"}, "CET"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("name", Equals("test-run-apply")),
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
				Config: runConfigWithoutTimezone("test-run-apply", []string{"*/5 * * * *"}),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("name", Equals("test-run-apply")),
					Attribute("timezone", Equals("UTC")),
					Attribute("every.#", Equals("1")),
					Attribute("every.0", Equals("*/5 * * * *")),
				),
			},
			{
				Config: runConfigWithAt("test-run-apply", "1234567"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("name", Equals("test-run-apply")),
					Attribute("at", Equals("1234567")),
				),
			},
			{
				Config: runConfigWithRuntimeConfig("test-run-apply", []string{"0 7 * * 1-5"}, "terraform_version: \"1.0\""),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("name", Equals("test-run-apply")),
					Attribute("every.#", Equals("1")),
					Attribute("every.0", Equals("0 7 * * 1-5")),
					Attribute("runtime_config", Equals("terraform_version: \"1.0\"")),
				),
			},
		})
	})
}
