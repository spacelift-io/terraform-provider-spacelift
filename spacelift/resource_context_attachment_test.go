package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextAttachmentResource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with a stack", func(t *testing.T) {
		config := func(priority int) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_context" "test" {
					name = "Test context %s"
				}

				resource "spacelift_context_attachment" "test" {
					context_id = spacelift_context.test.id
					stack_id   = spacelift_stack.test.id
					priority   = %d
				}
			`, randomID, randomID, priority)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(1),
				Check: Resource(
					"spacelift_context_attachment.test",
					Attribute("id", IsNotEmpty()),
					Attribute("context_id", Contains(randomID)),
					AttributeNotPresent("module_id"),
					Attribute("stack_id", Contains(randomID)),
					Attribute("priority", Equals("1")),
				),
			},
			{
				Config: config(2),
				Check: Resource(
					"spacelift_context_attachment.test",
					Attribute("priority", Equals("2")),
				),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}
				resource "spacelift_context" "test" {
					name = "Test context %s"
				}

				resource "spacelift_context_attachment" "test" {
					context_id = spacelift_context.test.id
					module_id  = spacelift_module.test.id
					priority   = 1
				}
			`, randomID),
			Check: Resource(
				"spacelift_context_attachment.test",
				Attribute("id", IsNotEmpty()),
				Attribute("module_id", Equals("terraform-bacon-tasty")),
				AttributeNotPresent("stack_id"),
			),
		}})
	})
}
