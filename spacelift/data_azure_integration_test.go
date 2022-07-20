package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAzureIntegrationData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
		resource "spacelift_azure_integration" "test" {
			name                    = "test-integration-%s"
			tenant_id               = "tenant-id"
			default_subscription_id = "subscription-id"
			labels                  = ["one", "two"]
		}
		data "spacelift_azure_integration" "test" {
			integration_id = spacelift_azure_integration.test.id
		}
		`, randomID),
		Check: Resource(
			"data.spacelift_azure_integration.test",
			Attribute("admin_consent_provided", Equals("false")),
			Attribute("admin_consent_url", IsNotEmpty()),
			Attribute("application_id", IsNotEmpty()),
			Attribute("default_subscription_id", Equals("subscription-id")),
			Attribute("display_name", IsNotEmpty()),
			Attribute("id", IsNotEmpty()),
			Attribute("name", Equals(fmt.Sprintf("test-integration-%s", randomID))),
			Attribute("tenant_id", Equals("tenant-id")),
			SetEquals("labels", "one", "two"),
		),
	}})
}

func TestAzureIntegrationDataSpace(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
		resource "spacelift_azure_integration" "test" {
			name                    = "test-integration-%s"
			tenant_id               = "tenant-id"
			default_subscription_id = "subscription-id"
			labels                  = ["one", "two"]
			space_id                = "root"
		}

		data "spacelift_azure_integration" "test" {
			integration_id = spacelift_azure_integration.test.id
		}
		`, randomID),
			Check: Resource(
				"data.spacelift_azure_integration.test",
				Attribute("admin_consent_provided", Equals("false")),
				Attribute("admin_consent_url", IsNotEmpty()),
				Attribute("application_id", IsNotEmpty()),
				Attribute("default_subscription_id", Equals("subscription-id")),
				Attribute("display_name", IsNotEmpty()),
				Attribute("id", IsNotEmpty()),
				Attribute("name", Equals(fmt.Sprintf("test-integration-%s", randomID))),
				Attribute("tenant_id", Equals("tenant-id")),
				Attribute("space_id", Equals("root")),
				SetEquals("labels", "one", "two"),
			),
		},
	})
}
