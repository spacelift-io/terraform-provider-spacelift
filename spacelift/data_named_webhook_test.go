package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestNamedWebhookData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
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

				data "spacelift_named_webhook" "test" {
					depends_on = [
						spacelift_named_webhook_secret_header.test-secret
					]
					webhook_id = spacelift_named_webhook.test.id
				}
			`, randomID),
		Check: Resource(
			"data.spacelift_named_webhook.test",
			Attribute("webhook_id", IsNotEmpty()),
			Attribute("endpoint", Equals("https://bacon.org")),
			Attribute("enabled", Equals("true")),
			SetContains("secret_header_keys", "thisisakey"),
		),
	}})
}
