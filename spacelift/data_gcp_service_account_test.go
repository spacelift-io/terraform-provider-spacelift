package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGCPServiceAccountData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with a stack", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_gcp_service_account" "test" {
					stack_id     = spacelift_stack.test.id
					token_scopes = ["https://www.googleapis.com/auth/compute"]
				}

				data "spacelift_gcp_service_account" "test" {
					stack_id = spacelift_gcp_service_account.test.stack_id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_gcp_service_account.test",
				Attribute("id", IsNotEmpty()),
				Attribute("stack_id", IsNotEmpty()),
				Attribute("service_account_email", IsNotEmpty()),
				SetEquals("token_scopes", "https://www.googleapis.com/auth/compute"),
				AttributeNotPresent("module_id"),
			),
		}})
	})

	t.Run("with a module", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				resource "spacelift_module" "test" {
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}
				resource "spacelift_gcp_service_account" "test" {
					module_id    = spacelift_module.test.id
					token_scopes = ["https://www.googleapis.com/auth/compute"]
				}

				data "spacelift_gcp_service_account" "test" {
					module_id = spacelift_gcp_service_account.test.module_id
				}
			`,
			Check: Resource(
				"data.spacelift_gcp_service_account.test",
				Attribute("id", IsNotEmpty()),
				Attribute("module_id", Equals("terraform-bacon-tasty")),
				Attribute("service_account_email", IsNotEmpty()),
				SetEquals("token_scopes", "https://www.googleapis.com/auth/compute"),
				AttributeNotPresent("stack_id"),
			),
		}})
	})
}
