package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAzureIntegrationAttachmentResource(t *testing.T) {
	const resourceName = "spacelift_azure_integration_attachment.test"

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_stack" "test" {
						branch     = "master"
						repository = "demo"
						name       = "Test stack %s"
					}

					resource "spacelift_azure_integration" "test" {
						name      = "test-integration-%s"
						tenant_id = "tenant-id"
					}

					resource "spacelift_azure_integration_attachment" "test" {
						stack_id       = spacelift_stack.test.id
						integration_id = spacelift_azure_integration.test.id
					}
				`, randomID, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					AttributeNotPresent("module_id"),
					Attribute("stack_id", IsNotEmpty()),
					Attribute("read", Equals("true")),
					AttributeNotPresent("subscription_id"),
					Attribute("write", Equals("true")),
					Attribute("attachment_id", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
					resource "spacelift_stack" "test" {
						branch     = "master"
						repository = "demo"
						name       = "Test stack %s"
					}

					resource "spacelift_azure_integration" "test" {
						name      = "test-integration-%s"
						tenant_id = "tenant-id"
					}

					resource "spacelift_azure_integration_attachment" "test" {
						stack_id = spacelift_stack.test.id
						integration_id  = spacelift_azure_integration.test.id
						read            = false
						subscription_id = "subscription-id"
					}
				`, randomID, randomID),
				Check: Resource(
					resourceName,
					Attribute("read", Equals("false")),
					Attribute("subscription_id", Equals("subscription-id")),
				),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_module" "test" {
                	    name       = "test-module-%s"
						branch     = "master"
						repository = "terraform-bacon-tasty"
					}

					resource "spacelift_azure_integration" "test" {
						name      = "test-integration-%s"
						tenant_id = "tenant-id"
					}

					resource "spacelift_azure_integration_attachment" "test" {
						module_id = spacelift_module.test.id
						integration_id = spacelift_azure_integration.test.id
					}
				`, randomID, randomID),
				Check: Resource(
					resourceName,
					Attribute("module_id", IsNotEmpty()),
					AttributeNotPresent("stack_id"),
				),
			},
		})
	})
}
