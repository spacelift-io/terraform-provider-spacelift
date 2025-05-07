package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGitlabWebhookEndpointData(t *testing.T) {
	t.Parallel()

	testSteps(t, []resource.TestStep{{
		Config: `
			data "spacelift_gitlab_webhook_endpoint" "test" {}
		`,
		Check: Resource(
			"data.spacelift_gitlab_webhook_endpoint.test",
			Attribute("webhook_endpoint", Equals(testConfig.SourceCode.Gitlab.Default.WebhookURL)),
		),
	}})
}
