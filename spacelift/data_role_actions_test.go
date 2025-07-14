package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestRoleActionsData(t *testing.T) {
	t.Run("retrieves role actions without an error", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
        data "spacelift_role_actions" "test" {}
      `,
			Check: Resource(
				"data.spacelift_role_actions.test",
				Attribute("id", Equals("role_actions")),
				SetContains("actions", "SPACE_READ", "SPACE_WRITE", "SPACE_ADMIN"),
			),
		}})
	})
}
