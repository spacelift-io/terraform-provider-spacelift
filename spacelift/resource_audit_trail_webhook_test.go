package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

const auditTrailWebhookSimple = `
resource "spacelift_audit_trail_webhook" "test" {
	enabled = true
	endpoint = "%s"
	include_runs = true
	secret = "secret"
}
`

const auditTrailWebhookCustomHeaders = `
resource "spacelift_audit_trail_webhook" "test" {
	enabled = true
	endpoint = "%s"
	include_runs = true
	secret = "secret"
	custom_headers = {
		"X-Some-Header" = "some-value"
		"X-Some-Header-2" = "some-value-2"
	}
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
					Attribute("enabled", Equals("false")),
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
			{
				Config: fmt.Sprintf(auditTrailWebhookCustomHeaders, "https://example.com"),
				Check: Resource(
					resourceName,
					Attribute("enabled", Equals("false")),
					Attribute("endpoint", Equals("https://example.com")),
					Attribute("include_runs", Equals("true")),
					Attribute("secret", Equals("secret")),
					Attribute("custom_headers.%", Equals("2")),
					Attribute("custom_headers.X-Some-Header", Equals("some-value")),
					Attribute("custom_headers.X-Some-Header-2", Equals("some-value-2")),
				),
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
