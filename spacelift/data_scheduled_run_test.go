package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledRunData_WhenEveryDefined_OK(t *testing.T) {
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

					name       = "test-run-apply"
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

				data "spacelift_scheduled_run" "test" {
					scheduled_run_id = spacelift_scheduled_run.test.id
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					"data.spacelift_scheduled_run.test",
					Attribute("scheduled_run_id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("name", Equals("test-run-apply")),
					Attribute("timezone", Equals("UTC")),
					Attribute("next_schedule", IsNotEmpty()),
					Attribute("every.#", Equals("2")),
					Attribute("every.0", Equals("*/3 * * * *")),
					Attribute("every.1", Equals("*/4 * * * *")),
					Attribute("runtime_config.0.project_root", Equals("root")),
					Attribute("runtime_config.0.runner_image", Equals("image")),
					Attribute("runtime_config.0.after_apply.#", Equals("2")),
					Attribute("runtime_config.0.after_apply.0", Equals("cmd1")),
					Attribute("runtime_config.0.after_apply.1", Equals("cmd2")),
					Attribute("runtime_config.0.environment.#", Equals("2")),
					Attribute("runtime_config.0.terraform_version", IsNotEmpty()),
					Attribute("runtime_config.0.terraform_workflow_tool", Equals("TERRAFORM_FOSS")),
				),
			),
		},
	})
}

func TestScheduledRunData_WhenAtDefined_OK(t *testing.T) {
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

					name       = "test-run-apply"
					at         = 1234

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

				data "spacelift_scheduled_run" "test" {
					scheduled_run_id = spacelift_scheduled_run.test.id
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					"data.spacelift_scheduled_run.test",
					Attribute("scheduled_run_id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("name", Equals("test-run-apply")),
					Attribute("next_schedule", IsNotEmpty()),
					Attribute("at", Equals("1234")),
					Attribute("runtime_config.0.project_root", Equals("root")),
					Attribute("runtime_config.0.runner_image", Equals("image")),
					Attribute("runtime_config.0.after_apply.#", Equals("2")),
					Attribute("runtime_config.0.after_apply.0", Equals("cmd1")),
					Attribute("runtime_config.0.after_apply.1", Equals("cmd2")),
					Attribute("runtime_config.0.environment.#", Equals("2")),
					Attribute("runtime_config.0.terraform_version", IsNotEmpty()),
					Attribute("runtime_config.0.terraform_workflow_tool", Equals("TERRAFORM_FOSS")),
				),
			),
		},
	})
}
