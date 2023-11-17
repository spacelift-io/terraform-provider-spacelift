package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAzureDevOpsIntegrationData(t *testing.T) {
	t.Run("without the id specified", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_azure_devops_integration" "test" {}
			`,
			Check: Resource(
				"data.spacelift_azure_devops_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", IsNotEmpty()),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", IsNotEmpty()),
				Attribute("organization_url", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_AZURE_DEVOPS_ORGANIZATION_URL"))),
				Attribute("webhook_password", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_AZURE_DEVOPS_WEBHOOK_PASSWORD"))),
				Attribute("webhook_url", IsNotEmpty()),
			),
		}})
	})

	t.Run("with the id specified", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_azure_devops_integration" "test" {
					id = "azure-devops-repo-default-integration"
				}
			`,
			Check: Resource(
				"data.spacelift_azure_devops_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", IsNotEmpty()),
				Attribute("is_default", Equals("true")),
				Attribute("space_id", IsNotEmpty()),
				Attribute("organization_url", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_AZURE_DEVOPS_ORGANIZATION_URL"))),
				Attribute("webhook_password", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_AZURE_DEVOPS_WEBHOOK_PASSWORD"))),
				Attribute("webhook_url", IsNotEmpty()),
			),
		}})
	})

}
