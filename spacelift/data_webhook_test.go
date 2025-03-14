package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWebhookData(t *testing.T) {
	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_webhook" "test" {
					stack_id = spacelift_stack.test.id
					endpoint = "https://bacon.org"
					secret   = "very-very-secret"
				}

				data "spacelift_webhook" "test" {
					stack_id   = spacelift_webhook.test.stack_id
					webhook_id = spacelift_webhook.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_webhook.test",
				Attribute("id", IsNotEmpty()),
				Attribute("endpoint", Equals("https://bacon.org")),
				Attribute("enabled", Equals("true")),
				Attribute("stack_id", StartsWith("test-stack-")),
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

				resource "spacelift_webhook" "test" {
					module_id = spacelift_module.test.id
					endpoint  = "https://bacon.org"
					secret    = "very-very-secret"
				}

				data "spacelift_webhook" "test" {
					module_id  = spacelift_webhook.test.module_id
					webhook_id = spacelift_webhook.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_webhook.test",
				Attribute("id", IsNotEmpty()),
				Attribute("endpoint", Equals("https://bacon.org")),
				Attribute("enabled", Equals("true")),
				Attribute("module_id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
				AttributeNotPresent("stack_id"),
			),
		}})
	})
}
