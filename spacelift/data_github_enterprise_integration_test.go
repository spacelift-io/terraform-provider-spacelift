package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGithubEnterpriseIntegrationData(t *testing.T) {
	t.Parallel()

	t.Run("without the ID specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.GithubEnterprise.Default
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_github_enterprise_integration" "test" {}
			`,
			Check: Resource(
				"data.spacelift_github_enterprise_integration.test",
				Attribute("id", Equals(cfg.ID)),
				Attribute("name", Equals(cfg.Name)),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", Equals("root")),
				Attribute("api_host", Equals(cfg.APIHost)),
				Attribute("webhook_secret", Equals(cfg.WebhookSecret)),
				Attribute("webhook_url", Equals(cfg.WebhookURL)),
				Attribute("app_id", Equals(cfg.AppID)),
				Attribute(ghEnterpriseVCSChecks, Equals(cfg.VCSChecks)),
			),
		}})
	})

	t.Run("with the ID specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.GithubEnterprise.SpaceLevel
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_github_enterprise_integration" "test" {
					id = "` + cfg.ID + `"
				}
			`,
			Check: Resource(
				"data.spacelift_github_enterprise_integration.test",
				Attribute("id", Equals(cfg.ID)),
				Attribute("name", Equals(cfg.Name)),
				Attribute("is_default", Equals("false")),
				Attribute("space_id", Equals(cfg.Space)),
				Attribute("api_host", Equals(cfg.APIHost)),
				Attribute("webhook_secret", Equals(cfg.WebhookSecret)),
				Attribute("webhook_url", Equals(cfg.WebhookURL)),
				Attribute("app_id", Equals(cfg.AppID)),
				Attribute(ghEnterpriseVCSChecks, Equals(cfg.VCSChecks)),
			),
		}})
	})
}
