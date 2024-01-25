package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

var auditTrailWebhookSimple = `
resource "spacelift_audit_trail_webhook" "test" {
	enabled = true
	endpoint = "%s"
	include_runs = true
	secret = "secret"
}
`

func Test_resourceAuditTrailWebhook(t *testing.T) {
	const resourceName = "spacelift_audit_trail_webhook.test"

	t.Run("creates an audit trail webhook without an error", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(auditTrailWebhookSimple, "https://example.com"),
				Check: Resource(
					resourceName,
					Attribute("enabled", Equals("true")),
					Attribute("endpoint", Equals("https://example.com")),
					Attribute("include_runs", Equals("true")),
					Attribute("secret", Equals("secret")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})

	t.Run("endpoint has to exist", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config:      fmt.Sprintf(auditTrailWebhookSimple, "https://invalidendpoint.com/"),
				ExpectError: regexp.MustCompile(`could not send webhook to given endpoint`),
			},
		})
	})
}
