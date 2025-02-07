package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestNamedWebhookResource(t *testing.T) {
	const resourceName = "spacelift_named_webhook.test"

	t.Run("attach a webhook to root space with all fields filled", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint = "%s"
					space_id = "root"
					name     = "testing-named-%s"
					labels   = ["1", "2"]
					secret   = "super-secret"
					enabled  = true
				}
			`, endpoint, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("secret", Equals("")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals(fmt.Sprintf("testing-named-%s", randomID))),
					Attribute("enabled", Equals("true")),
					SetEquals("labels", "1", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("https://cabbage.org"),
				Check:  Resource(resourceName, Attribute("endpoint", Equals("https://cabbage.org"))),
			},
		})
	})

	t.Run("attach a webhook to root space with default values", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint = "%s"
					space_id = "root"
					name     = "testing-named-%s"
					enabled  = true
				}
			`, endpoint, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("secret", Equals("")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals(fmt.Sprintf("testing-named-%s", randomID))),
					Attribute("enabled", Equals("true")),
					AttributeNotPresent("labels"),
				),
			},
		})
	})
}
