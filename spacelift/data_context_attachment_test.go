package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextAttachmentData(t *testing.T) {
	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
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
					priority   = 1
				}

				data "spacelift_context_attachment" "test" {
					context_id = spacelift_context_attachment.test.context_id
					stack_id   = spacelift_context_attachment.test.stack_id
				}
			`, randomID, randomID),
			Check: Resource(
				"data.spacelift_context_attachment.test",
				Attribute("id", IsNotEmpty()),
				Attribute("context_id", Contains(randomID)),
				Attribute("stack_id", Contains(randomID)),
				Attribute("priority", Equals("1")),
				AttributeNotPresent("module_id"),
			),
		}})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name       = "test-module-%s"
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

				data "spacelift_context_attachment" "test" {
					context_id = spacelift_context_attachment.test.context_id
					module_id  = spacelift_context_attachment.test.module_id
				}
			`, randomID, randomID),
			Check: Resource(
				"data.spacelift_context_attachment.test",
				Attribute("id", IsNotEmpty()),
				Attribute("module_id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
				Attribute("priority", Equals("1")),
				AttributeNotPresent("stack_id"),
			),
		}})
	})
}
