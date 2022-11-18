package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestNamedWebhookResource(t *testing.T) {
	const resourceName = "spacelift_named_webhook.test"

	t.Run("attach a webhook to root space", func(t *testing.T) {
		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint = "%s"
					space_id = "root"
					name     = "testing-named-hooks"
					labels   = ["1", "2"]
					secret   = "super-secret"
					enabled  = true
				}
			`, endpoint)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("secret", Equals("super-secret")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals("testing-named-hooks")),
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
}
