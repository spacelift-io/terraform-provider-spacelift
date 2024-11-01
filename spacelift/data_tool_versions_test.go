package spacelift

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestToolVersionsData(t *testing.T) {
	testSteps(t, []resource.TestStep{
		{
			Config: `
				data "spacelift_tool_versions" "kubectl" {
					tool = "KUBECTL"
				}
				`,
			Check: Resource(
				"data.spacelift_tool_versions.kubectl",
				Attribute("id", Contains("kubectl-versions")),
				Attribute("tool", Equals("KUBECTL")),
				SetLengthGreaterThanZero("versions"),
			),
		},
		{
			Config: `
				data "spacelift_tool_versions" "open_tofu" {
					tool = "OPEN_TOFU"
				}
				`,
			Check: Resource(
				"data.spacelift_tool_versions.open_tofu",
				Attribute("id", Contains("open-tofu-versions")),
				Attribute("tool", Equals("OPEN_TOFU")),
				SetLengthGreaterThanZero("versions"),
			),
		},
		{
			Config: `
				data "spacelift_tool_versions" "terraform_foss" {
					tool = "TERRAFORM_FOSS"
				}
				`,
			Check: Resource(
				"data.spacelift_tool_versions.terraform_foss",
				Attribute("id", Contains("terraform-foss-versions")),
				Attribute("tool", Equals("TERRAFORM_FOSS")),
				SetLengthGreaterThanZero("versions"),
			),
		},
		{
			Config: `
				data "spacelift_tool_versions" "terragrunt" {
					tool = "TERRAGRUNT"
				}
				`,
			Check: Resource(
				"data.spacelift_tool_versions.terragrunt",
				Attribute("id", Contains("terragrunt-versions")),
				Attribute("tool", Equals("TERRAGRUNT")),
				SetLengthGreaterThanZero("versions"),
			),
		},
	})

	t.Run("only allows specific tools", func(t *testing.T) {
		re, err := regexp.Compile(`tool must be one of \[.*]`)
		if err != nil {
			t.Fatalf("could not compile regexp: %v", err)
		}
		testSteps(t, []resource.TestStep{
			{
				Config: `
				data "spacelift_tool_versions" "test" {
					tool = "this-tool-should-error"
				}
				`,
				ExpectError: re,
			},
		})
	})
}
