package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPolicyAttachmentResource(t *testing.T) {
	t.Parallel()
	const resourceName = "spacelift_policy_attachment.test"

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
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
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("policy_id", Contains(randomID)),
					Attribute("stack_id", Contains(randomID)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("my-first-policy-%s/test-stack-%s", randomID, randomID),
				ImportStateVerify: true,
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					body = "package spacelift"
					type = "PLAN"
				}
	
				resource "spacelift_module" "test" {
                    name       = "test-module-%s"
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}
	
				resource "spacelift_policy_attachment" "test" {
					policy_id = spacelift_policy.test.id
					module_id = spacelift_module.test.id
				}
			`, randomID, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("policy_id", Contains(randomID)),
					Attribute("module_id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("my-first-policy-%s/terraform-default-test-module-%s", randomID, randomID),
				ImportStateVerify: true,
			},
		})
	})
}
