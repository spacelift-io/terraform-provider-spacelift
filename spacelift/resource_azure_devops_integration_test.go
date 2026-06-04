package spacelift

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/vcs"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

type terraformVersionOutput struct {
	Version string `json:"terraform_version"`
}

func skipAzureDevOpsIntegrationResourceTestsIfUnconfigured(t *testing.T) {
	t.Helper()

	if testIsMachineSession() {
		t.Skip("skipping Azure DevOps integration resource tests: create/update is not available in machine sessions")
	}

	cfg := testConfig.SourceCode.AzureDevOps.SpaceLevel
	missing := make([]string, 0)

	if cfg.PersonalAccessToken == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_PERSONALACCESSTOKEN")
	}

	if cfg.Space == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_SPACE")
	}

	if cfg.OrganizationURL == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_ORGANIZATIONURL")
	}

	if cfg.UserFacingHost == "" {
		missing = append(missing, "SPACELIFT_PROVIDER_TEST_SOURCECODE_AZUREDEVOPS_SPACELEVEL_USERFACINGHOST")
	}

	if len(missing) > 0 {
		t.Skipf("skipping Azure DevOps integration resource tests: missing required fixtures: %s", strings.Join(missing, ", "))
	}
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
	skipAzureDevOpsIntegrationResourceTestsIfUnconfigured(t)

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
			{
				Config: configAzureDevOps("new descr", "null", "null", `"`+vcs.CheckTypeAggregated+`"`, useGitCheckout),
				Check: Resource(
					resourceName,
					AttributeNotPresent(azureDevopsLabels),
					AttributeNotPresent(azureDevopsAccessibleProjects),
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
					personal_access_token_wo_version = 1
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

func TestHandleAzureDevopsIntegrationUpdateResult(t *testing.T) {
	t.Run("keeps state intact on update error", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, resourceAzureDevopsIntegration().Schema, map[string]interface{}{})
		data.SetId("existing-id")

		diags := handleAzureDevopsIntegrationUpdateResult(data, &structs.AzureDevOpsRepoIntegration{}, errors.New("boom"))

		if len(diags) == 0 {
			t.Fatal("expected diagnostics, got none")
		}

		if data.Id() != "existing-id" {
			t.Fatalf("expected ID to remain %q, got %q", "existing-id", data.Id())
		}
	})

	t.Run("applies mutation results on success", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, resourceAzureDevopsIntegration().Schema, map[string]interface{}{})
		data.SetId("existing-id")

		integration := &structs.AzureDevOpsRepoIntegration{
			ID:                 "updated-id",
			Name:               "updated-name",
			Description:        "updated-description",
			OrganizationURL:    "https://dev.azure.com/example",
			UserFacingHost:     "https://dev.azure.com/example",
			WebhookPassword:    "secret",
			WebhookURL:         "https://webhooks.example",
			VCSChecks:          vcs.CheckTypeAggregated,
			UseGitCheckout:     true,
			Labels:             []string{"label-1"},
			AccessibleProjects: []string{"Project One"},
		}
		integration.Space.ID = "root"

		diags := handleAzureDevopsIntegrationUpdateResult(data, integration, nil)

		if len(diags) != 0 {
			t.Fatalf("expected no diagnostics, got %v", diags)
		}

		if data.Id() != "updated-id" {
			t.Fatalf("expected ID %q, got %q", "updated-id", data.Id())
		}

		if got := data.Get(azureDevopsName).(string); got != "updated-name" {
			t.Fatalf("expected name %q, got %q", "updated-name", got)
		}
	})
}

func TestValidateAzureDevopsDefaultSpace(t *testing.T) {
	t.Run("allows non-default integration in non-root space", func(t *testing.T) {
		if err := validateAzureDevopsDefaultSpace(false, "team-space"); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("allows default integration in root space", func(t *testing.T) {
		if err := validateAzureDevopsDefaultSpace(true, "root"); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("rejects default integration in non-root space", func(t *testing.T) {
		err := validateAzureDevopsDefaultSpace(true, "non-root-space")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		expected := `the default integration must be in the space "root" not in "non-root-space"`
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})
}
