package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWebhookResource(t *testing.T) {

	t.Run("with a stack", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

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
					"spacelift_webhook.test",
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.org")),
					Attribute("secret", Equals("very-very-secret")),
				),
			},
			{
				Config: config("https://cabbage.org"),
				Check:  Resource("spacelift_webhook.test", Attribute("endpoint", Equals("https://cabbage.org"))),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
			resource "spacelift_module" "test" {
				branch     = "master"
				repository = "terraform-bacon-tasty"
			}

			resource "spacelift_webhook" "test" {
				module_id = spacelift_module.test.id
				endpoint  = "https://bacon.org"
				secret    = "very-very-secret"
			}
		`,
			Check: Resource(
				"spacelift_webhook.test",
				Attribute("id", IsNotEmpty()),
				Attribute("endpoint", Equals("https://bacon.org")),
				Attribute("secret", Equals("very-very-secret")),
			),
		}})
	})
}
