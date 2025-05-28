package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledRunData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	datasourceName := "data.spacelift_scheduled_run.test"

	runConfigWithEvery := func(name string, every []string, timezone string) string {
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

			data "spacelift_scheduled_run" "test" {
				scheduled_run_id   = spacelift_scheduled_run.test.id
			}
		`, randomID, name, strings.Join(everyStrs, ", "), timezone)
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
	
			data "spacelift_scheduled_run" "test" {
				scheduled_run_id = spacelift_scheduled_run.test.id
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

			data "spacelift_scheduled_run" "test" {
				scheduled_run_id = spacelift_scheduled_run.test.id
			}
		`, randomID, name, strings.Join(everyStrs, ", "), runtimeConfig)
	}

	testSteps(t, []resource.TestStep{
		{
			Config: runConfigWithEvery("test-run-apply", []string{"*/3 * * * *", "*/4 * * * *"}, "CET"),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					datasourceName,
					Attribute("scheduled_run_id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("name", Equals("test-run-apply")),
					Attribute("timezone", Equals("CET")),
					Attribute("every.#", Equals("2")),
					Attribute("every.0", Equals("*/3 * * * *")),
					Attribute("every.1", Equals("*/4 * * * *")),
				),
			),
		},
		{
			Config: runConfigWithAt("test-run-apply", "1234567"),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					datasourceName,
					Attribute("scheduled_run_id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("name", Equals("test-run-apply")),
					Attribute("at", Equals("1234567")),
				),
			),
		},
		{
			Config: runConfigWithRuntimeConfig("test-run-apply", []string{"0 7 * * 1-5"}, "terraform_version: \"1.0\""),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					datasourceName,
					Attribute("scheduled_run_id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("name", Equals("test-run-apply")),
					Attribute("every.#", Equals("1")),
					Attribute("every.0", Equals("0 7 * * 1-5")),
					Attribute("runtime_config", Equals("terraform_version: \"1.0\"")),
				),
			),
		},
	})
}