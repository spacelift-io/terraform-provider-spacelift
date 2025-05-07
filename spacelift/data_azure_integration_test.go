package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAzureIntegrationData(t *testing.T) {
	t.Parallel()

	t.Run("when looking up integration by ID", func(t *testing.T) {
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
	})

	t.Run("when looking up integration by name", func(t *testing.T) {
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
				name = spacelift_azure_integration.test.name
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
	})

	t.Run("when integration ID does not exist", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
			data "spacelift_azure_integration" "test" {
				integration_id = "01GBASTWAEPJ1HDMXDMWTRC8DN"
			}`,
			ExpectError: regexp.MustCompile(`Azure integration not found: 01GBASTWAEPJ1HDMXDMWTRC8DN`),
		}})
	})

	t.Run("when integration name does not exist", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
			data "spacelift_azure_integration" "test" {
				name = "non-existent integration"
			}`,
			ExpectError: regexp.MustCompile(`Azure integration not found: non-existent integration`),
		}})
	})

	t.Run("when setting both integration_id and name it errors", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
	data "spacelift_azure_integration" "test" {
        integration_id = "01GBAME4P2BS72ZQRA9HJYWRCK"
		name           = "Test Integration"
    }
    `,
			ExpectError: regexp.MustCompile("only one of `integration_id,name` can be specified"),
		}})
	})
}

func TestAzureIntegrationDataSpace(t *testing.T) {
	t.Parallel()
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
