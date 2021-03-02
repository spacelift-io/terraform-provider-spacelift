package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGCPServiceAccountResource(t *testing.T) {
	const resourceName = "spacelift_gcp_service_account.test"

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with a stack", func(t *testing.T) {
		config := func(scope string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_gcp_service_account" "test" {
					stack_id     = spacelift_stack.test.id
					token_scopes = ["%s"]
				}
			`, randomID, scope)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://www.googleapis.com/auth/compute"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
					Attribute("service_account_email", IsNotEmpty()),
					SetEquals("token_scopes", "https://www.googleapis.com/auth/compute"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("stack/test-stack-%s", randomID),
				ImportStateVerify: true,
			},
			{
				Config: config("https://www.googleapis.com/auth/cloud-platform"),
				Check: Resource(
					resourceName,
					SetEquals("token_scopes", "https://www.googleapis.com/auth/cloud-platform"),
				),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: `
				resource "spacelift_module" "test" {
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}
				resource "spacelift_gcp_service_account" "test" {
					module_id    = spacelift_module.test.id
					token_scopes = ["https://www.googleapis.com/auth/compute"]
				}
			`,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("module_id", Equals("terraform-bacon-tasty")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "module/terraform-bacon-tasty",
				ImportStateVerify: true,
			},
		})
	})
}
