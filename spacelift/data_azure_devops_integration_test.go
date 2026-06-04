package spacelift

import (
	"fmt"
	"strings"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func skipAzureDevOpsIntegrationDataTestsIfUnconfigured(t *testing.T) {
	t.Helper()

	defaultCfg := testConfig.SourceCode.AzureDevOps.Default
	spaceCfg := testConfig.SourceCode.AzureDevOps.SpaceLevel

	missing := make([]string, 0)

	if defaultCfg.ID == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_DEFAULT_ID")
	}

	if defaultCfg.Name == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_DEFAULT_NAME")
	}

	if defaultCfg.OrganizationURL == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_DEFAULT_ORGANIZATIONURL")
	}

	if defaultCfg.UserFacingHost == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_DEFAULT_USERFACINGHOST")
	}

	if defaultCfg.WebhookSecret == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_DEFAULT_WEBHOOKSECRET")
	}

	if defaultCfg.WebhookURL == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_DEFAULT_WEBHOOKURL")
	}

	if defaultCfg.VCSChecks == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_DEFAULT_VCSCHECKS")
	}

	if spaceCfg.ID == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_ID")
	}

	if spaceCfg.Name == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_NAME")
	}

	if spaceCfg.Space == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_SPACE")
	}

	if spaceCfg.OrganizationURL == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_ORGANIZATIONURL")
	}

	if spaceCfg.UserFacingHost == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_USERFACINGHOST")
	}

	if spaceCfg.WebhookSecret == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_WEBHOOKSECRET")
	}

	if spaceCfg.WebhookURL == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_WEBHOOKURL")
	}

	if spaceCfg.VCSChecks == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_VCSCHECKS")
	}

	if len(missing) > 0 {
		t.Skipf("skipping Azure DevOps integration data tests: missing required fixtures: %s", strings.Join(missing, ", "))
	}
}

func azureDevopsOrganizationURLCheck(expected string) ValueCheck {
	return func(actual string) error {
		if testIsMachineSession() {
			if actual == "***" {
				return nil
			}

			return fmt.Errorf("expected organization_url to be redacted in machine sessions, got %q", actual)
		}

		return Equals(expected)(actual)
	}
}

func TestAzureDevOpsIntegrationData(t *testing.T) {
	skipAzureDevOpsIntegrationDataTestsIfUnconfigured(t)

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
				Attribute("organization_url", azureDevopsOrganizationURLCheck(cfg.OrganizationURL)),
				Attribute("user_facing_host", Equals(cfg.UserFacingHost)),
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
				Attribute("organization_url", azureDevopsOrganizationURLCheck(cfg.OrganizationURL)),
				Attribute("user_facing_host", Equals(cfg.UserFacingHost)),
				Attribute("webhook_password", Equals(cfg.WebhookSecret)),
				Attribute("webhook_url", Equals(cfg.WebhookURL)),
				Attribute(azureDevopsVCSChecks, Equals(cfg.VCSChecks)),
				Attribute(azureDevopsUseGitCheckout, Equals(strconv.FormatBool(cfg.UseGitCheckout))),
			),
		}})
	})

}
