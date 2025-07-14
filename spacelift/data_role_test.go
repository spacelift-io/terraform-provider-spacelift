package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestRoleData(t *testing.T) {
	t.Run("reads a system role (filter by name)", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_role" "test" {
					name = "Space admin"
				}
			`,
			Check: Resource(
				"data.spacelift_role.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", Equals("Space admin")),
				Attribute("description", IsNotEmpty()),
				Attribute("is_system", Equals("true")),
				SetEquals("actions", "SPACE_READ", "SPACE_WRITE", "SPACE_ADMIN"),
			),
		}})
	})

	t.Run("reads a custom role (filter by ID)", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_role" "test" {
					name        = "Custom test role %s"
					description = "A custom role for testing"
					actions     = ["SPACE_READ", "SPACE_WRITE"]
				}

				data "spacelift_role" "test" {
					role_id = spacelift_role.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_role.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", StartsWith("Custom test role")),
				Attribute("description", Equals("A custom role for testing")),
				Attribute("is_system", Equals("false")),
				SetEquals("actions", "SPACE_READ", "SPACE_WRITE"),
			),
		}})
	})

	t.Run("no filter is provided", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_role" "test" {}
			`,
			ExpectError: regexp.MustCompile(`either 'role_id' or 'name' must be specified to read a role`),
		}})
	})
}
