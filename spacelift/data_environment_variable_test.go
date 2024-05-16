package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestEnvironmentVariableData(t *testing.T) {
	t.Run("with a context", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name = "My first context %s"
				}

				resource "spacelift_environment_variable" "test" {
					context_id = spacelift_context.test.id
					name       = "BACON"
					value      = "is tasty"
					write_only = true
				}

				data "spacelift_environment_variable" "test" {
					context_id = spacelift_environment_variable.test.context_id
					name       = "BACON"
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_environment_variable.test",
				Attribute("id", IsNotEmpty()),
				Attribute("checksum", Equals("4d5d01ea427b10dd483e8fce5b5149fb5a9814e9ee614176b756ca4a65c8f154")),
				Attribute("name", Equals("BACON")),
				Attribute("value", IsEmpty()),
				Attribute("write_only", Equals("true")),
				AttributeNotPresent("module_id"),
				AttributeNotPresent("stack_id"),
			),
		}})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
                    name           = "test-module-%s"
					branch         = "master"
					repository     = "terraform-bacon-tasty"
				}
	
				resource "spacelift_environment_variable" "test" {
					module_id  = spacelift_module.test.id
					name       = "BACON"
					value      = "is tasty"
					write_only = false
				}

				data "spacelift_environment_variable" "test" {
					module_id = spacelift_environment_variable.test.module_id
					name       = "BACON"
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_environment_variable.test",
				Attribute("module_id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
				Attribute("value", Equals("is tasty")),
				Attribute("write_only", Equals("false")),
				AttributeNotPresent("context_id"),
				AttributeNotPresent("stack_id"),
			),
		}})
	})

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
	
				resource "spacelift_environment_variable" "test" {
					stack_id = spacelift_stack.test.id
					value    = "is tasty"
					name     = "BACON"
				}

				data "spacelift_environment_variable" "test" {
					stack_id = spacelift_environment_variable.test.stack_id
					name     = "BACON"
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_environment_variable.test",
				Attribute("stack_id", StartsWith("test-stack-")),
				Attribute("stack_id", Contains(randomID)),
				AttributeNotPresent("context_id"),
				AttributeNotPresent("module_id"),
			),
		}})
	})
}
