package spacelift

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/vcs"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

type terraformVersionOutput struct {
	Version string `json:"terraform_version"`
}

func terraformVersionAtLeast(major, minor int) (bool, string, error) {
	cmd := exec.Command("terraform", "version", "-json")
	output, err := cmd.Output()
	if err != nil {
		return false, "", err
	}

	var versionOutput terraformVersionOutput
	if err := json.Unmarshal(output, &versionOutput); err != nil {
		return false, "", fmt.Errorf("failed to parse terraform version json: %w", err)
	}

	parts := strings.SplitN(versionOutput.Version, ".", 3)
	if len(parts) < 2 {
		return false, versionOutput.Version, fmt.Errorf("unexpected terraform version format: %q", versionOutput.Version)
	}

	majorVersion, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, versionOutput.Version, fmt.Errorf("invalid terraform major version %q: %w", parts[0], err)
	}

	minorVersion, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, versionOutput.Version, fmt.Errorf("invalid terraform minor version %q: %w", parts[1], err)
	}

	if majorVersion > major {
		return true, versionOutput.Version, nil
	}

	if majorVersion == major && minorVersion >= minor {
		return true, versionOutput.Version, nil
	}

	return false, versionOutput.Version, nil
}

func TestAzureDevOpsIntegrationResource(t *testing.T) {
	const resourceName = "spacelift_azure_devops_integration.test"

	t.Run("creates and updates an Azure DevOps integration without an error", func(t *testing.T) {
		random := func() string { return acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum) }

		var (
			name         = "my-test-azure-devops-integration-" + random()
			organization = testConfig.SourceCode.AzureDevOps.SpaceLevel.OrganizationURL
			token        = testConfig.SourceCode.AzureDevOps.SpaceLevel.PersonalAccessToken
			descr        = "description " + random()
		)

		configAzureDevOps := func(descr, labels, accessibleProjects, vcsChecks string, useGitCheckout bool) string {
			return `
				resource "spacelift_azure_devops_integration" "test" {
					name                  = "` + name + `"
					space_id              = "` + testConfig.SourceCode.AzureDevOps.SpaceLevel.Space + `"
					organization_url      = "` + organization + `"
					user_facing_host      = "` + testConfig.SourceCode.AzureDevOps.SpaceLevel.UserFacingHost + `"
					personal_access_token = "` + token + `"
					description           = "` + descr + `"
					labels                = ` + labels + `
					accessible_projects   = ` + accessibleProjects + `
					vcs_checks            = ` + vcsChecks + `
					use_git_checkout      = ` + strconv.FormatBool(useGitCheckout) + `
				}
			`
		}

		testSteps(t, []resource.TestStep{
			{
				Config: configAzureDevOps(descr, "null", "null", "null", !useGitCheckout),
				Check: Resource(
					resourceName,
					Attribute(azureDevopsName, Equals(name)),
					Attribute(azureDevopsOrganizationURL, Equals(organization)),
					Attribute(azureDevopsUserFacingHost, Equals(testConfig.SourceCode.AzureDevOps.SpaceLevel.UserFacingHost)),
					Attribute(azureDevopsWebhookURL, IsNotEmpty()),
					Attribute(azureDevopsWebhookPassword, IsNotEmpty()),
					Attribute(azureDevopsIsDefault, Equals("false")),
					Attribute(azureDevopsDescription, Equals(descr)),
					Attribute(azureDevopsVCSChecks, Equals(vcs.CheckTypeDefault)),
					Attribute(azureDevopsUseGitCheckout, Equals("false")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{azureDevopsPersonalAccessToken},
			},
			{
				Config: configAzureDevOps("new descr", `["new label1"]`, `["Project One"]`, `"`+vcs.CheckTypeAggregated+`"`, useGitCheckout),
				Check: Resource(
					resourceName,
					Attribute(azureDevopsDescription, Equals("new descr")),
					Attribute(azureDevopsLabels+".#", Equals("1")),
					Attribute(azureDevopsAccessibleProjects+".#", Equals("1")),
					Attribute(azureDevopsVCSChecks, Equals(vcs.CheckTypeAggregated)),
					Attribute(azureDevopsUseGitCheckout, Equals("true")),
				),
			},
		})
	})

	t.Run("creates and updates an Azure DevOps integration with write-only fields and without an error", func(t *testing.T) {
		supported, version, err := terraformVersionAtLeast(1, 11)
		if err != nil {
			t.Skipf("skipping write-only test: unable to detect terraform version: %v", err)
		}

		if !supported {
			t.Skipf("skipping write-only test: Terraform 1.11+ required, detected %s", version)
		}

		random := func() string { return acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum) }

		var (
			name         = "my-test-azure-devops-integration-" + random()
			organization = testConfig.SourceCode.AzureDevOps.SpaceLevel.OrganizationURL
			token        = testConfig.SourceCode.AzureDevOps.SpaceLevel.PersonalAccessToken
			descr        = "description " + random()
		)

		configAzureDevOps := func(descr, labels, vcsChecks string, useGitCheckout bool) string {
			return `
				resource "spacelift_azure_devops_integration" "test" {
					name                            = "` + name + `"
					space_id                        = "` + testConfig.SourceCode.AzureDevOps.SpaceLevel.Space + `"
					organization_url                = "` + organization + `"
					user_facing_host                = "` + testConfig.SourceCode.AzureDevOps.SpaceLevel.UserFacingHost + `"
					personal_access_token_wo        = "` + token + `"
					personal_access_token_wo_version = "1"
					description                     = "` + descr + `"
					labels                          = ` + labels + `
					vcs_checks                      = ` + vcsChecks + `
					use_git_checkout                = ` + strconv.FormatBool(useGitCheckout) + `
				}
			`
		}

		testSteps(t, []resource.TestStep{
			{
				Config: configAzureDevOps(descr, "null", "null", !useGitCheckout),
				Check: Resource(
					resourceName,
					Attribute(azureDevopsName, Equals(name)),
					Attribute(azureDevopsOrganizationURL, Equals(organization)),
					Attribute(azureDevopsWebhookURL, IsNotEmpty()),
					Attribute(azureDevopsWebhookPassword, IsNotEmpty()),
					Attribute(azureDevopsDescription, Equals(descr)),
					Attribute(azureDevopsVCSChecks, Equals(vcs.CheckTypeDefault)),
					Attribute(azureDevopsUseGitCheckout, Equals("false")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{azureDevopsPersonalAccessToken, azureDevopsPersonalAccessTokenWoVer},
			},
		})
	})
}
