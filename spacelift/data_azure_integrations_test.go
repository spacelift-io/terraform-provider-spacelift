package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAzureIntegrationsData(t *testing.T) {
	t.Run("when looking up integrations", func(t *testing.T) {
		subId1 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
		subId2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
		first := &structs.AzureIntegration{
			DefaultSubscriptionID: &subId1,
			Labels:                []string{"one", "two"},
			Name:                  acctest.RandStringFromCharSet(5, acctest.CharSetAlpha),
			TenantID:              acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
			Space:                 "root",
		}
		second := &structs.AzureIntegration{
			DefaultSubscriptionID: &subId2,
			Labels:                []string{"three", "four"},
			Name:                  acctest.RandStringFromCharSet(5, acctest.CharSetAlpha),
			TenantID:              acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
			Space:                 "legacy",
		}

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			%s
			%s

			data "spacelift_azure_integrations" "test" {
				depends_on = [spacelift_azure_integration.%s,spacelift_azure_integration.%s]
			}
			`,
				azureIntegrationToResource(first),
				azureIntegrationToResource(second),
				first.Name,
				second.Name),
			Check: resource.ComposeTestCheckFunc(
				Resource("data.spacelift_azure_integrations.test", Attribute("id", Equals("spacelift_azure_integrations"))),
				resource.ComposeTestCheckFunc(azureIntegrationChecks(first)...),
				resource.ComposeTestCheckFunc(azureIntegrationChecks(second)...),
			),
		}})
	})
}

func azureIntegrationToResource(i *structs.AzureIntegration) string {
	return fmt.Sprintf(`
 			resource "spacelift_azure_integration" "%s" {
				name                    = "%s"
				tenant_id               = "%s"
				default_subscription_id = "%s"
				labels                  =  %s
				space_id 				= "%s"
			}
`,
		i.Name,
		i.Name,
		i.TenantID,
		*i.DefaultSubscriptionID,
		fmt.Sprintf(`["%s"]`, strings.Join(i.Labels, `", "`)),
		i.Space,
	)
}

func azureIntegrationChecks(i *structs.AzureIntegration) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		Resource("data.spacelift_azure_integrations.test",
			Nested("integrations",
				CheckInList(
					Attribute("name", Equals(i.Name)),
					Attribute("tenant_id", Equals(i.TenantID)),
					Attribute("default_subscription_id", Equals(*i.DefaultSubscriptionID)),
					SetEquals("labels", i.Labels...),
					Attribute("space_id", Equals(i.Space)),
					Attribute("integration_id", IsNotEmpty()),
					Attribute("object_id", IsNotEmpty()),
				),
			),
		),
	}
}
