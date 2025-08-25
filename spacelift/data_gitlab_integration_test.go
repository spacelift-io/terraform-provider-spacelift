package spacelift

import (
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGitlabIntegrationData(t *testing.T) {
	t.Run("without the id specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.Gitlab.Default
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_gitlab_integration" "test" {}
			`,
			Check: Resource(
				"data.spacelift_gitlab_integration.test",
				Attribute("id", Equals(cfg.ID)),
				Attribute("name", Equals(cfg.Name)),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", Equals("root")),
				Attribute("api_host", Equals(cfg.APIHost)),
				Attribute("webhook_secret", Equals(cfg.WebhookSecret)),
				Attribute("webhook_url", Equals(cfg.WebhookURL)),
				Attribute("vcs_checks", Equals(cfg.VCSChecks)),
				Attribute("use_git_checkout", Equals(strconv.FormatBool(cfg.UseGitCheckout))),
			),
		}})
	})

	t.Run("with the id specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.Gitlab.SpaceLevel
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_gitlab_integration" "test" {
					id = "` + cfg.ID + `"
				}
			`,
			Check: Resource(
				"data.spacelift_gitlab_integration.test",
				Attribute("id", Equals(cfg.ID)),
				Attribute("name", Equals(cfg.Name)),
				Attribute("is_default", Equals("false")),
				Attribute("space_id", Equals(cfg.Space)),
				Attribute("api_host", Equals(cfg.APIHost)),
				Attribute("webhook_secret", Equals(cfg.WebhookSecret)),
				Attribute("webhook_url", Equals(cfg.WebhookURL)),
				Attribute("vcs_checks", Equals(cfg.VCSChecks)),
				Attribute("use_git_checkout", Equals(strconv.FormatBool(cfg.UseGitCheckout))),
			),
		}})
	})
}
