package spacelift

import (
	"strconv"
	"testing"

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
				Attribute("webhook_password", Equals(cfg.WebhookSecret)),
				Attribute("webhook_url", Equals(cfg.WebhookURL)),
				Attribute(azureDevopsVCSChecks, Equals(cfg.VCSChecks)),
				Attribute(azureDevopsUseGitCheckout, Equals(strconv.FormatBool(cfg.UseGitCheckout))),
			),
		}})
	})

	t.Run("with the id specified", func(t *testing.T) {
		cfg := testConfig.SourceCode.AzureDevOps.SpaceLevel
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_azure_devops_integration" "test" {
					id = "` + cfg.ID + `"
				}
			`,
			Check: Resource(
				"data.spacelift_azure_devops_integration.test",
				Attribute("id", Equals(cfg.ID)),
				Attribute("name", Equals(cfg.Name)),
				Attribute("is_default", Equals("false")),
				Attribute("space_id", Equals(cfg.Space)),
				Attribute("organization_url", Equals(cfg.OrganizationURL)),
				Attribute("webhook_password", Equals(cfg.WebhookSecret)),
				Attribute("webhook_url", Equals(cfg.WebhookURL)),
				Attribute(azureDevopsVCSChecks, Equals(cfg.VCSChecks)),
				Attribute(azureDevopsUseGitCheckout, Equals(strconv.FormatBool(cfg.UseGitCheckout))),
			),
		}})
	})

}
