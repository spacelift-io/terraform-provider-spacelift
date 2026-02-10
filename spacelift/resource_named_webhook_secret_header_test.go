package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestNamedWebhookSecretHeaderResource(t *testing.T) {
	const resourceName = "spacelift_named_webhook_secret_header.test-secret"

	t.Run("attach a webhook to root space with all fields filled", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func() string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint = "https://bacon.org"
					space_id = "root"
					name     = "%s"
					labels   = ["1", "2"]
					secret   = "super-secret"
					enabled  = true
				}

				resource "spacelift_named_webhook_secret_header" "test-secret" {
					webhook_id = spacelift_named_webhook.test.id
					key        = "thisisakey"
					value      = "thisisavalue"
				}
			`, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(),
				Check: Resource(
					resourceName,
					Attribute("id", Equals(fmt.Sprintf("%s/%s", randomID, "thisisakey"))),
					Attribute("key", Equals("thisisakey")),
					Attribute("value", Equals("thisisavalue")),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
			},
		})
	})

	t.Run("attach a webhook to root space with all write-only fields", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func() string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint          = "https://bacon.org"
					space_id          = "root"
					name              = "%s"
					labels            = ["1", "2"]
					secret_wo         = "super-secret"
					secret_wo_version = 1
					enabled           = true
				}

				resource "spacelift_named_webhook_secret_header" "test-secret" {
					webhook_id            = spacelift_named_webhook.test.id
					key                   = "thisisakey"
					value_wo              = "thisisavalue"
					value_wo_version      = 1
				}
			`, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(),
				Check: Resource(
					resourceName,
					Attribute("id", Equals(fmt.Sprintf("%s/%s", randomID, "thisisakey"))),
					Attribute("key", Equals("thisisakey")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"secret", "secret_wo_version", "value", "value_wo_version"},
			},
		})
	})
}
