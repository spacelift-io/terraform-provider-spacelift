package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestNamedWebhookHeaderResource(t *testing.T) {
	const resourceName = "spacelift_named_webhook_header.test"

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint = "https://bacon.net"
					space_id = "root"
					name     = "testing-named-%s"
				}

				resource "spacelift_named_webhook_header" "test" {
					named_webhook_id = spacelift_named_webhook.test.id
					name             = "X-Test-Header"
					value            = "test-value"
				}
			`, randomID)

	testSteps(t, []resource.TestStep{
		{
			Config: config,
			Check: Resource(
				resourceName,
				Attribute("id", IsNotEmpty()),
				Attribute("name", Equals("X-Test-Header")),
			),
		},
		{
			ResourceName:      resourceName,
			ImportState:       true,
			ImportStateVerify: true,
		},
	})
}
