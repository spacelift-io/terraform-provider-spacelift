package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAzureIntegrationResource(t *testing.T) {
	const resourceName = "spacelift_azure_integration.test"

	t.Run("Creates and updates an integration", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_azure_integration" "test" {
					name      = "test-integration-%s"
					tenant_id = "tenant-id"
					labels    = ["one", "two"]
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("test-integration-%s", randomID))),
					Attribute("tenant_id", Equals("tenant-id")),
					AttributeNotPresent("default_subscription_id"),
					SetEquals("labels", "one", "two"),
					Attribute("admin_consent_provided", Equals("false")),
					Attribute("admin_consent_url", IsNotEmpty()),
					Attribute("application_id", IsNotEmpty()),
					Attribute("object_id", IsNotEmpty()),
					Attribute("display_name", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
				resource "spacelift_azure_integration" "test" {
					name                    = "test-integration-%s"
					tenant_id               = "tenant-id"
					default_subscription_id = "subscription-id"
					labels                  = ["three", "four"]
				}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("default_subscription_id", Equals("subscription-id")),
					SetEquals("labels", "three", "four"),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_azure_integration" "test" {
					name      = "test-integration-%s"
					tenant_id = "tenant-id"
					labels    = ["one", "two"]
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_azure_integration" "test" {
					name      = "test-integration-%s"
					tenant_id = "tenant-id"
					labels    = []
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels"),
				),
			},
		})
	})
}

func TestAzureIntegrationResourceSpace(t *testing.T) {
	const resourceName = "spacelift_azure_integration.test"

	t.Run("Creates and updates an integration", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_azure_integration" "test" {
					name      = "test-integration-%s"
					tenant_id = "tenant-id"
					labels    = ["one", "two"]
					space_id  = "root"
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("test-integration-%s", randomID))),
					Attribute("tenant_id", Equals("tenant-id")),
					AttributeNotPresent("default_subscription_id"),
					SetEquals("labels", "one", "two"),
					Attribute("admin_consent_provided", Equals("false")),
					Attribute("admin_consent_url", IsNotEmpty()),
					Attribute("application_id", IsNotEmpty()),
					Attribute("object_id", IsNotEmpty()),
					Attribute("display_name", IsNotEmpty()),
					Attribute("space_id", Equals("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
				resource "spacelift_azure_integration" "test" {
					name                    = "test-integration-%s"
					tenant_id               = "tenant-id"
					default_subscription_id = "subscription-id"
					labels                  = ["three", "four"]
				}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("default_subscription_id", Equals("subscription-id")),
					SetEquals("labels", "three", "four"),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_azure_integration" "test" {
					name      = "test-integration-%s"
					tenant_id = "tenant-id"
					labels    = ["one", "two"]
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_azure_integration" "test" {
					name      = "test-integration-%s"
					tenant_id = "tenant-id"
					labels    = []
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels"),
				),
			},
		})
	})
}
