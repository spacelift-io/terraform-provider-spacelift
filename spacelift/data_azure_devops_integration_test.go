package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAzureDevOpsIntegrationData(t *testing.T) {
	testSteps(t, []resource.TestStep{{
		Config: `
			data "spacelift_azure_devops_integration" "test" {}
		`,
		Check: Resource(
			"data.spacelift_azure_devops_integration.test",
			Attribute("organization_url", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_AZURE_DEVOPS_ORGANIZATION_URL"))),
			Attribute("webhook_password", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_AZURE_DEVOPS_WEBHOOK_PASSWORD"))),
		),
	}})
}
