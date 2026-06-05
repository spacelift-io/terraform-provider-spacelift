package spacelift

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAzureDevOpsIntegrationData(t *testing.T) {
	t.Run("without the id specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.AzureDevOps.Default

		testSteps(t, []resource.TestStep{{
			Config: `
			data "spacelift_azure_devops_integration" "test" {}
			`,
			Check: Resource(
				"data.spacelift_azure_devops_integration.test",
				Attribute("id", Equals(cfg.ID)),
				Attribute("name", Equals(cfg.Name)),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", Equals("root")),
				Attribute("organization_url", Equals(cfg.OrganizationURL)),
				Attribute("user_facing_host", Equals(cfg.UserFacingHost)),
				Attribute("webhook_password", Equals(cfg.WebhookSecret)),
				Attribute("webhook_url", Equals(cfg.WebhookURL)),
				Attribute("vcs_checks", Equals(cfg.VCSChecks)),
				Attribute("use_git_checkout", Equals(strconv.FormatBool(cfg.UseGitCheckout))),
			),
		}})
	})

	t.Run("with the id specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.AzureDevOps.SpaceLevel
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_azure_devops_integration" "test" {
				name                  = "test-integration-%s"
				space_id              = "root"
				organization_url      = "%s"
				user_facing_host      = "%s"
				personal_access_token = "%s"
				vcs_checks            = "AGGREGATED"
				use_git_checkout      = true
			}

			data "spacelift_azure_devops_integration" "test" {
				id = spacelift_azure_devops_integration.test.id
			}
			`, randomID, cfg.OrganizationURL, cfg.UserFacingHost, cfg.PersonalAccessToken),
			Check: Resource(
				"data.spacelift_azure_devops_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", Equals(fmt.Sprintf("test-integration-%s", randomID))),
				Attribute("is_default", Equals("false")),
				Attribute("space_id", Equals("root")),
				Attribute("organization_url", Equals(cfg.OrganizationURL)),
				Attribute("user_facing_host", Equals(cfg.UserFacingHost)),
				Attribute("webhook_password", IsNotEmpty()),
				Attribute("webhook_url", IsNotEmpty()),
				Attribute("vcs_checks", Equals("AGGREGATED")),
				Attribute("use_git_checkout", Equals("true")),
			),
		}})
	})
}
