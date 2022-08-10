package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestMountedFileResource(t *testing.T) {
	const resourceName = "spacelift_mounted_file.test"

	t.Run("with a context", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(writeOnly bool) string {
			return fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name = "My first context %s"
				}

				resource "spacelift_mounted_file" "test" {
					context_id    = spacelift_context.test.id
					content       = base64encode("bacon is tasty")
					relative_path = "bacon.txt"
					write_only    = %t
				}
			`, randomID, writeOnly)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(true),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("checksum", Equals("fb13e7977b7548a324b598e155b5b5ba3dcca2dad5789abe1411a88fa544be9b")),
					Attribute("context_id", Contains(randomID)),
					Attribute("relative_path", Equals("bacon.txt")),
					Attribute("write_only", Equals("true")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config(false),
				Check:  Resource(resourceName, Attribute("write_only", Equals("false"))),
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
	
				resource "spacelift_mounted_file" "test" {
					module_id     = spacelift_module.test.id
					content       = base64encode("bacon is tasty")
					relative_path = "bacon.txt"
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("module_id", Equals(fmt.Sprintf("test-module-%s", randomID))),
					Attribute("write_only", Equals("true")),
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
	
				resource "spacelift_mounted_file" "test" {
					stack_id      = spacelift_stack.test.id
					content       = base64encode("bacon is tasty")
					relative_path = "bacon.txt"
					file_mode     = "755"
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("stack_id", StartsWith("test-stack-")),
					Attribute("stack_id", Contains(randomID)),
					Attribute("file_mode", Equals("755")),
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
