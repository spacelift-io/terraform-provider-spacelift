package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestScheduledRunResource_WhenEveryDefinedAndUpdate_OK(t *testing.T) {
	resourceType := "spacelift_scheduled_run"
	resourceName := "test"
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "%s" "%s" {
					stack_id = spacelift_stack.test.id
					name = "test-run-apply"
		
					every      = [ "*/3 * * * *", "*/4 * * * *" ]
					timezone   = "CET"
				}
			`, randomID, resourceType, resourceName),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					fmt.Sprintf("%s.%s", resourceType, resourceName),
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
		
				resource "%s" "%s" {
					stack_id = spacelift_stack.test.id
					name = "test-run-apply"
		
					every      = [ "*/3 * * * *" ]
					timezone   = "CET"
				}
			`, randomID, resourceType, resourceName),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					fmt.Sprintf("%s.%s", resourceType, resourceName),
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
	resourceType := "spacelift_scheduled_run"
	resourceName := "test"
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "%s" "%s" {
					stack_id = spacelift_stack.test.id
		
					at      = 1234
				}
			`, randomID, resourceType, resourceName),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					fmt.Sprintf("%s.%s", resourceType, resourceName),
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
	resourceType := "spacelift_scheduled_run"
	resourceName := "test"
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "%s" "%s" {
					stack_id = spacelift_stack.test.id
		
					every      = [ "*/3 * * * *", "*/4 * * * *" ]
				}
			`, randomID, resourceType, resourceName),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					fmt.Sprintf("%s.%s", resourceType, resourceName),
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
	resourceType := "spacelift_scheduled_run"
	resourceName := "test"
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
		
				resource "%s" "%s" {
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
			`, randomID, resourceType, resourceName),
			Check: resource.ComposeTestCheckFunc(
				Resource(
					fmt.Sprintf("%s.%s", resourceType, resourceName),
					Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
					Attribute("stack_id", Contains(randomID)),
					Attribute("schedule_id", IsNotEmpty()),
					Attribute("timezone", Equals("UTC")),
					Attribute("every.#", Equals("2")),
					Attribute("every.0", Equals("*/3 * * * *")),
					Attribute("every.1", Equals("*/4 * * * *")),
					Attribute("runtime_config.#", Equals("1")),
					Nested("runtime_config.0", Attribute("project_root", Equals("root"))),
					Nested("runtime_config.0", Attribute("runner_image", Equals("image"))),
					Nested("runtime_config.0", Attribute("after_apply.#", Equals("2"))),
					Nested("runtime_config.0", Attribute("after_apply.0", Equals("cmd1"))),
					Nested("runtime_config.0", Attribute("after_apply.1", Equals("cmd2"))),
					Nested("runtime_config.0", Attribute("environment.#", Equals("2"))),
					Nested("runtime_config.0", Attribute("environment.0.key", Equals("ENV_1"))),
					Nested("runtime_config.0", Attribute("environment.0.value", Equals("ENV_1_VAL"))),
					Nested("runtime_config.0", Attribute("environment.1.key", Equals("ENV_2"))),
					Nested("runtime_config.0", Attribute("environment.1.value", Equals("ENV_2_VAL"))),
				),
			),
		},
	})
}
