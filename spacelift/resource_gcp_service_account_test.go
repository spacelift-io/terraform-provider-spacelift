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

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

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
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
                    name       = "test-module-%s"
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}
				resource "spacelift_gcp_service_account" "test" {
					module_id    = spacelift_module.test.id
					token_scopes = ["https://www.googleapis.com/auth/compute"]
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("module_id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("module/test-module-%s", randomID),
				ImportStateVerify: true,
			},
		})
	})
}
