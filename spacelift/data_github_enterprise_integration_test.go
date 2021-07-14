package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGithubEnterpriseIntegrationData(t *testing.T) {
	t.Parallel()

	testSteps(t, []resource.TestStep{{
		Config: `
			data "spacelift_github_enterprise_integration" "test" {}
		`,
		Check: Resource(
			"data.spacelift_github_enterprise_integration.test",
			Attribute("app_id", Equals("6")),
			Attribute("api_host", Equals("https://github.liftspace.net")),
			Attribute("webhook_secret", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITHUB_ENTERPRISE_WEBHOOK_SECRET"))),
		),
	}})
}
