package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGithubEnterpriseIntegrationData(t *testing.T) {
	t.Run("without the ID specified", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_github_enterprise_integration" "test" {}
			`,
			Check: Resource(
				"data.spacelift_github_enterprise_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", IsNotEmpty()),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", IsNotEmpty()),
				Attribute("api_host", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITHUB_ENTERPRISE_API_HOST"))),
				Attribute("webhook_secret", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITHUB_ENTERPRISE_WEBHOOK_SECRET"))),
				Attribute("app_id", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITHUB_ENTERPRISE_APP_ID"))),
			),
		}})
	})

	t.Run("with the ID specified", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_github_enterprise_integration" "test" {
					id = "github-enterprise-default-integration"
				}
			`,
			Check: Resource(
				"data.spacelift_github_enterprise_integration.test",
				Attribute("id", Equals("github-enterprise-default-integration")),
				Attribute("name", IsNotEmpty()),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", IsNotEmpty()),
				Attribute("api_host", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITHUB_ENTERPRISE_API_HOST"))),
				Attribute("webhook_secret", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITHUB_ENTERPRISE_WEBHOOK_SECRET"))),
				Attribute("app_id", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITHUB_ENTERPRISE_APP_ID"))),
			),
		}})
	})
}
