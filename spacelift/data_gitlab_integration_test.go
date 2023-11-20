package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGitlabIntegrationData(t *testing.T) {
	t.Run("without the id specified", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_gitlab_integration" "test" {}
			`,
			Check: Resource(
				"data.spacelift_gitlab_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", IsNotEmpty()),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", IsNotEmpty()),
				Attribute("api_host", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITLAB_API_HOST"))),
				Attribute("webhook_secret", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITLAB_WEBHOOK_SECRET"))),
				Attribute("webhook_url", IsNotEmpty()),
			),
		}})
	})

	t.Run("with the id specified", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_gitlab_integration" "test" {
					id = "gitlab-default-integration"
				}
			`,
			Check: Resource(
				"data.spacelift_gitlab_integration.test",
				Attribute("id", Equals("gitlab-default-integration")),
				Attribute("name", IsNotEmpty()),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", IsNotEmpty()),
				Attribute("api_host", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITLAB_API_HOST"))),
				Attribute("webhook_secret", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITLAB_WEBHOOK_SECRET"))),
				Attribute("webhook_url", IsNotEmpty()),
			),
		}})
	})
}
