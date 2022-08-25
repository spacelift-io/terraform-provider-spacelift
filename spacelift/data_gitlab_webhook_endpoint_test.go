package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGitlabWebhookEndpointData(t *testing.T) {
	testSteps(t, []resource.TestStep{{
		Config: `
			data "spacelift_gitlab_webhook_endpoint" "test" {}
		`,
		Check: Resource(
			"data.spacelift_gitlab_webhook_endpoint.test",
			Attribute("webhook_endpoint", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITLAB_WEBHOOK_ENDPOINT"))),
		),
	}})
}
