package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledRunResource_WhenEveryDefinedAndUpdate_OK(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
					name = "test-run-apply"
		
					every      = [ "*/3 * * * *", "*/4 * * * *" ]
					timezone   = "CET"
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					"spacelift_scheduled_run.test",
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("name", Equals("test-run-apply")),
					Attribute("timezone", Equals("CET")),
					Attribute("every.#", Equals("2")),
					Attribute("every.0", Equals("*/3 * * * *")),
					Attribute("every.1", Equals("*/4 * * * *")),
				),
			),
		},
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
					name = "test-run-apply"
		
					every      = [ "*/3 * * * *" ]
					timezone   = "CET"
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					"spacelift_scheduled_run.test",
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("name", Equals("test-run-apply")),
					Attribute("timezone", Equals("CET")),
					Attribute("every.#", Equals("1")),
					Attribute("every.0", Equals("*/3 * * * *")),
				),
			),
		},
	})
}

func TestScheduledRunResource_WhenAtDefined_OK(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
		
					at      = 1234
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					"spacelift_scheduled_run.test",
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("timezone", Equals("UTC")),
					Attribute("at", Equals("1234")),
				),
			),
		},
	})
}

func TestScheduledRunResource_WhenTimezoneNotDefined_OK(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
		
					every      = [ "*/3 * * * *", "*/4 * * * *" ]
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					"spacelift_scheduled_run.test",
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("timezone", Equals("UTC")),
					Attribute("every.#", Equals("2")),
					Attribute("every.0", Equals("*/3 * * * *")),
					Attribute("every.1", Equals("*/4 * * * *")),
				),
			),
		},
	})
}

func TestScheduledRunResource_WhenRuntimeConfigDefined_OK(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "spacelift_scheduled_run" "test" {
					stack_id = spacelift_stack.test.id
		
					every      = [ "*/3 * * * *", "*/4 * * * *" ]

					runtime_config {
						project_root = "root"
						runner_image = "image"
						after_apply = [ "cmd1", "cmd2" ]

						environment { 
							key = "ENV_1"
							value = "ENV_1_VAL"
						}
						environment { 
							key = "ENV_2"
							value = "ENV_2_VAL"
						}
				    }
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					"spacelift_scheduled_run.test",
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("timezone", Equals("UTC")),
					Attribute("next_schedule", IsNotEmpty()),
					Attribute("every.#", Equals("2")),
					Attribute("every.0", Equals("*/3 * * * *")),
					Attribute("every.1", Equals("*/4 * * * *")),
					Attribute("runtime_config.#", Equals("1")),
					Attribute("runtime_config.0.project_root", Equals("root")),
					Attribute("runtime_config.0.runner_image", Equals("image")),
					Attribute("runtime_config.0.after_apply.#", Equals("2")),
					Attribute("runtime_config.0.after_apply.0", Equals("cmd1")),
					Attribute("runtime_config.0.after_apply.1", Equals("cmd2")),
					Attribute("runtime_config.0.environment.#", Equals("2")),
					Attribute("runtime_config.0.environment.0.key", Equals("ENV_1")),
					Attribute("runtime_config.0.environment.0.value", Equals("ENV_1_VAL")),
					Attribute("runtime_config.0.environment.1.key", Equals("ENV_2")),
					Attribute("runtime_config.0.environment.1.value", Equals("ENV_2_VAL")),
					Attribute("runtime_config.0.yaml", IsNotEmpty()),
					Attribute("runtime_config.0.terraform_version", IsNotEmpty()),
					Attribute("runtime_config.0.terraform_workflow_tool", Equals("TERRAFORM_FOSS")),
				),
			),
		},
	})
}
