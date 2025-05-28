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

	// runConfigWithRuntimeConfig := func(name string, every []string, runtimeConfig string) string {
	// 	everyStrs := make([]string, len(every))
	// 	for i := range every {
	// 		everyStrs[i] = `"` + every[i] + `"`
	// 	}
	//
	// 	return fmt.Sprintf(`
	// 		resource "spacelift_stack" "test" {
	// 			branch     = "master"
	// 			repository = "demo"
	// 			name       = "Test stack %s"
	// 		}
	//
	// 		resource "spacelift_scheduled_run" "test" {
	// 			stack_id = spacelift_stack.test.id
	//
	// 			name           = "%s"
	// 			every          = [%s]
	// 			runtime_config = "%s"
	// 		}
	// 	`, randomID, name, strings.Join(everyStrs, ", "), runtimeConfig)
	// }

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
		// {
		// 	Config: runConfigWithRuntimeConfig("test-run-apply", []string{"0 7 * * 1-5"}, "terraform_version: \"1.0\""),
		// 	Check: resource.ComposeTestCheckFunc(
		// 		Resource(
		// 			resourceName,
		// 			Attribute("id", StartsWith(fmt.Sprintf("test-stack-%s", randomID))),
		// 			Attribute("stack_id", Contains(randomID)),
		// 			Attribute("name", Equals("test-run-apply")),
		// 			Attribute("every.#", Equals("1")),
		// 			Attribute("every.0", Equals("0 7 * * 1-5")),
		// 			// Attribute("runtime_config.environment", Equals("terraform_version: \"1.0\"")),
		// 		),
		// 	),
		// },
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
				),
			),
		},
	})
}
