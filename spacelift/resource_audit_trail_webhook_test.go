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
	enabled = false
	endpoint = "%s"
	include_runs = true
	secret = "secret"
	retry_on_failure = false
}
`

const auditTrailWebhookCustomHeaders = `
resource "spacelift_audit_trail_webhook" "test" {
	enabled = false
	endpoint = "%s"
	include_runs = true
	secret = "secret"
	retry_on_failure = false
	custom_headers = {
		"X-Some-Header" = "some-value"
		"X-Some-Header-2" = "some-value-2"
	}
}
`

const auditTrailWebhookWriteOnly = `
resource "spacelift_audit_trail_webhook" "test" {
	enabled = false
	endpoint = "%s"
	include_runs = true
	secret_wo = "secret"
	secret_wo_version = 1
	retry_on_failure = false
}
`
// There can only be one audit trail webhook per account, so these tests must run sequentially
func Test_resourceAuditTrailWebhook(t *testing.T) {
	const resourceName = "spacelift_audit_trail_webhook.test"

	t.Run("creates an audit trail webhook with write_only secret", func(t *testing.T) {
		testStepsSequential(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(auditTrailWebhookWriteOnly, "https://example.com"),
				Check: Resource(
					resourceName,
					Attribute("enabled", Equals("false")),
					Attribute("endpoint", Equals("https://example.com")),
					Attribute("include_runs", Equals("true")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "secret_wo_version"},
			},
		})
	})

	t.Run("creates an audit trail webhook without an error", func(t *testing.T) {
		testStepsSequential(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(auditTrailWebhookSimple, "https://example.com"),
				Check: Resource(
					resourceName,
					Attribute("enabled", Equals("false")),
					Attribute("endpoint", Equals("https://example.com")),
					Attribute("include_runs", Equals("true")),
					Attribute("secret", Equals("")),
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
					Attribute("secret", Equals("")),
					Attribute("custom_headers.%", Equals("2")),
					Attribute("custom_headers.X-Some-Header", Equals("some-value")),
					Attribute("custom_headers.X-Some-Header-2", Equals("some-value-2")),
				),
			},
		})
	})

	t.Run("endpoint has to be valid", func(t *testing.T) {
		testStepsSequential(t, []resource.TestStep{
			{
				Config:      fmt.Sprintf(auditTrailWebhookSimple, "https:/invalidendpoint.com/"),
				ExpectError: regexp.MustCompile(`endpoint must be a valid URL`),
			},
		})
	})

}
