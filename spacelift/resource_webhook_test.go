package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWebhookResource(t *testing.T) {
	const resourceName = "spacelift_webhook.test"
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with a stack", func(t *testing.T) {
		t.Parallel()

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_webhook" "test" {
					stack_id = spacelift_stack.test.id
					endpoint = "%s"
					secret   = "very-very-secret"
				}
			`, randomID, endpoint)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.org"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.org")),
					Attribute("secret", Equals("very-very-secret")),
				),
			},
			{
				ResourceName:        resourceName,
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("stack/test-stack-%s/", randomID),
				ImportStateVerify:   true,
			},
			{
				Config: config("https://cabbage.org"),
				Check:  Resource(resourceName, Attribute("endpoint", Equals("https://cabbage.org"))),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			resource "spacelift_module" "test" {
                name       = "test-module-%s"
				branch     = "master"
				repository = "terraform-bacon-tasty"
			}

			resource "spacelift_webhook" "test" {
				module_id = spacelift_module.test.id
				endpoint  = "https://bacon.org"
				secret    = "very-very-secret"
			}
		`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.org")),
					Attribute("secret", Equals("very-very-secret")),
				),
			},
			{
				ResourceName:        resourceName,
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("module/test-module-%s/", randomID),
				ImportStateVerify:   true,
			},
		})
	})
}
