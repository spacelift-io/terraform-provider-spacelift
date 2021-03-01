package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPolicyAttachmentResource(t *testing.T) {
	const resourceName = "spacelift_policy_attachment.test"

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with a stack", func(t *testing.T) {
		config := func(message string) string {
			return fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					body = "package spacelift"
					type = "PLAN"
				}

				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_policy_attachment" "test" {
					policy_id    = spacelift_policy.test.id
					stack_id     = spacelift_stack.test.id
					custom_input = jsonencode({ message = "%s" })
				}
			`, randomID, randomID, message)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("boom"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("policy_id", Contains(randomID)),
					Attribute("stack_id", Contains(randomID)),
					Attribute("custom_input", Contains("boom")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("my-first-policy-%s/test-stack-%s", randomID, randomID),
				ImportStateVerify: true,
			},
			{
				Config: config("bang"),
				Check:  Resource(resourceName, Attribute("custom_input", Contains("bang"))),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					body = "package spacelift"
					type = "PLAN"
				}
	
				resource "spacelift_module" "test" {
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}
	
				resource "spacelift_policy_attachment" "test" {
					policy_id = spacelift_policy.test.id
					module_id = spacelift_module.test.id
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("policy_id", Contains(randomID)),
					Attribute("module_id", Equals("terraform-bacon-tasty")),
					AttributeNotPresent("custom_input"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("my-first-policy-%s/terraform-bacon-tasty", randomID),
				ImportStateVerify: true,
			},
		})
	})
}
