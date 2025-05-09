package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVCSIntegrationAzureDevOps(t *testing.T) {
	const resourceName = "spacelift_stack.test"

	t.Run("with_default_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("azure-devops-with-default-integration-implicit-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "root"
				administrative     = false
				azure_devops {
					project = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.AzureDevOps.Repository.Name,
			testConfig.SourceCode.AzureDevOps.Repository.Branch,
			testConfig.SourceCode.AzureDevOps.Repository.Namespace)

		var tfstateSerial int64
		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: func(tfstate *terraform.State) error {
					tfstateSerial = tfstate.Serial
					return Resource(
						resourceName,
						Attribute("azure_devops.0.id", Equals(testConfig.SourceCode.AzureDevOps.Default.ID)),
					)(tfstate)
				},
			},
			{
				Config: config,
				Check: func(tfstate *terraform.State) error {
					// We need to check the serials to make sure nothing changed
					if serial := tfstate.Serial; serial != tfstateSerial {
						return fmt.Errorf("serials do not match: %d != %d", serial, tfstateSerial)
					}
					return nil
				},
			},
		})
	})

	t.Run("with_space_level_integration", func(t *testing.T) {
		randID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("azure-devops-with-space-level-integration-%s", randID)

		config := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name               = "%s"
				repository         = "%s"
				branch             = "%s"
				space_id           = "%s"
				administrative     = false
				azure_devops {
					project = "%s"
					id = "%s"
				}
			}
		`, name,
			testConfig.SourceCode.AzureDevOps.Repository.Name,
			testConfig.SourceCode.AzureDevOps.Repository.Branch,
			testConfig.SourceCode.AzureDevOps.SpaceLevel.Space,
			testConfig.SourceCode.AzureDevOps.Repository.Namespace,
			testConfig.SourceCode.AzureDevOps.SpaceLevel.ID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("azure_devops.0.id", Equals(testConfig.SourceCode.AzureDevOps.SpaceLevel.ID)),
				),
			},
		})
	})
}
