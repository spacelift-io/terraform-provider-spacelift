package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestEnvironmentVariableResource(t *testing.T) {
	const resourceName = "spacelift_environment_variable.test"

	t.Run("with a context", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(writeOnly bool, description string) string {
			return fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name = "My first context %s"
				}

				resource "spacelift_environment_variable" "test" {
					context_id  = spacelift_context.test.id
					name        = "BACON"
					value       = "is tasty"
					write_only  = %t
					description = %s
				}
			`, randomID, writeOnly, description)
		}

		const resourceName = "spacelift_environment_variable.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config(true, `"Bacon is tasty"`),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("checksum", Equals("4d5d01ea427b10dd483e8fce5b5149fb5a9814e9ee614176b756ca4a65c8f154")),
					Attribute("context_id", Contains(randomID)),
					Attribute("name", Equals("BACON")),
					Attribute("value", Equals("4d5d01ea427b10dd483e8fce5b5149fb5a9814e9ee614176b756ca4a65c8f154")),
					Attribute("write_only", Equals("true")),
					Attribute("description", Equals("Bacon is tasty")),
					AttributeNotPresent("module_id"),
					AttributeNotPresent("stack_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config(false, "null"),
				Check: Resource(
					resourceName,
					Attribute("value", Equals("is tasty")),
					Attribute("write_only", Equals("false")),
					Attribute("description", IsEmpty()),
				),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
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
					description = "Bacon is tasty"
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("module_id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
					Attribute("value", Equals("4d5d01ea427b10dd483e8fce5b5149fb5a9814e9ee614176b756ca4a65c8f154")),
					Attribute("write_only", Equals("true")),
					Attribute("description", Equals("Bacon is tasty")),
					AttributeNotPresent("context_id"),
					AttributeNotPresent("stack_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}
	
				resource "spacelift_environment_variable" "test" {
					stack_id    = spacelift_stack.test.id
					value       = "is tasty"
					name        = "BACON"
					description = "Bacon is tasty"
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("stack_id", StartsWith("test-stack-")),
					Attribute("stack_id", Contains(randomID)),
					Attribute("value", Equals("4d5d01ea427b10dd483e8fce5b5149fb5a9814e9ee614176b756ca4a65c8f154")),
					Attribute("description", Equals("Bacon is tasty")),
					AttributeNotPresent("context_id"),
					AttributeNotPresent("module_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})
}

func TestEnvironmentVariableResourceNonsensitiveValue(t *testing.T) {
	const resourceName = "spacelift_environment_variable.test"

	t.Run("with a context", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name = "My first context %s"
				}

				resource "spacelift_environment_variable" "test" {
					context_id         = spacelift_context.test.id
					name               = "BACON"
					value_nonsensitive = "is tasty"
					write_only         = false
					description        = %s
				}
			`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(`"Bacon is tasty"`),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("checksum", Equals("4d5d01ea427b10dd483e8fce5b5149fb5a9814e9ee614176b756ca4a65c8f154")),
					Attribute("context_id", Contains(randomID)),
					Attribute("name", Equals("BACON")),
					Attribute("value_nonsensitive", Equals("is tasty")),
					Attribute("write_only", Equals("false")),
					Attribute("description", Equals("Bacon is tasty")),
					AttributeNotPresent("value"),
					AttributeNotPresent("module_id"),
					AttributeNotPresent("stack_id"),
				),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
	          			name           = "test-module-%s"
					branch         = "master"
					repository     = "terraform-bacon-tasty"
				}

				resource "spacelift_environment_variable" "test" {
					module_id          = spacelift_module.test.id
					name               = "BACON"
					value_nonsensitive = "is tasty"
					write_only         = false
					description        = "Bacon is tasty"
				}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("module_id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
					Attribute("value_nonsensitive", Equals("is tasty")),
					Attribute("write_only", Equals("false")),
					Attribute("description", Equals("Bacon is tasty")),
					AttributeNotPresent("value"),
					AttributeNotPresent("context_id"),
					AttributeNotPresent("stack_id"),
				),
			},
		})
	})

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_environment_variable" "test" {
					stack_id    	   = spacelift_stack.test.id
					value_nonsensitive = "is tasty"
					write_only         = false
					name        	   = "BACON"
					description   	   = "Bacon is tasty"
				}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("stack_id", StartsWith("test-stack-")),
					Attribute("stack_id", Contains(randomID)),
					Attribute("value_nonsensitive", Equals("is tasty")),
					Attribute("description", Equals("Bacon is tasty")),
					AttributeNotPresent("value"),
					AttributeNotPresent("context_id"),
					AttributeNotPresent("module_id"),
				),
			},
		})
	})

	t.Run("write only is not allowed", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_environment_variable" "test" {
					stack_id    	   = spacelift_stack.test.id
					value_nonsensitive = "is tasty"
					write_only         = true
					name        	   = "BACON"
					description   	   = "Bacon is tasty"
				}
				`, randomID),
				ExpectError: regexp.MustCompile("a non-sensitive environment variable cannot be write-only"),
			},
		})
	})
}
