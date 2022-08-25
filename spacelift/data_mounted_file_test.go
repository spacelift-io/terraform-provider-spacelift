package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestMountedFileData(t *testing.T) {
	t.Run("with a context", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name = "My first context %s"
				}

				resource "spacelift_mounted_file" "test" {
					context_id    = spacelift_context.test.id
					content       = base64encode("bacon is tasty")
					relative_path = "bacon.txt"
				}

				data "spacelift_mounted_file" "test" {
					context_id    = spacelift_mounted_file.test.context_id
					relative_path = spacelift_mounted_file.test.relative_path
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_mounted_file.test",
				Attribute("id", IsNotEmpty()),
				Attribute("content", IsEmpty()),
				Attribute("context_id", Contains(randomID)),
				Attribute("relative_path", Equals("bacon.txt")),
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
	
				resource "spacelift_mounted_file" "test" {
					module_id     = spacelift_module.test.id
					content       = base64encode("bacon is tasty")
					relative_path = "bacon.txt"
					write_only    = false
				}

				data "spacelift_mounted_file" "test" {
					module_id    = spacelift_mounted_file.test.module_id
					relative_path = spacelift_mounted_file.test.relative_path
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_mounted_file.test",
				Attribute("module_id", Equals(fmt.Sprintf("test-module-%s", randomID))),
				Attribute("content", Equals("YmFjb24gaXMgdGFzdHk=")),
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
	
				resource "spacelift_mounted_file" "test" {
					stack_id      = spacelift_stack.test.id
					content       = base64encode("bacon is tasty")
					relative_path = "bacon.txt"
				}

				data "spacelift_mounted_file" "test" {
					stack_id      = spacelift_mounted_file.test.stack_id
					relative_path = spacelift_mounted_file.test.relative_path
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_mounted_file.test",
				Attribute("stack_id", StartsWith("test-stack-")),
				Attribute("stack_id", Contains(randomID)),
				Attribute("content", IsEmpty()),
				Attribute("write_only", Equals("true")),
				AttributeNotPresent("context_id"),
				AttributeNotPresent("module_id"),
			),
		}})
	})
}
