package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

var withOneAccess = `
resource "spacelift_user_group" "test" {
  name = "%s"
  access {
    space_id = "root"
    level    = "ADMIN"
  }
}
`

var withTwoAccesses = `
resource "spacelift_user_group" "test" {
  name = "%s"
  access {
    space_id = "root"
    level    = "ADMIN"
  }
  access {
    space_id = "legacy"
    level    = "READ"
  }
}
`

func TestUserGroupResource(t *testing.T) {
	const resourceName = "spacelift_user_group.test"

	t.Run("creates and updates a user group without an error", func(t *testing.T) {
		oldName := "old name"
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(withOneAccess, oldName),
				Check: Resource(
					resourceName,
					Attribute("name", Equals(oldName)),
					SetContains("access", "root"),
					SetContains("access", "ADMIN"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(withOneAccess, randomID),
				Check: Resource(
					resourceName,
					Attribute("name", Equals(randomID)),
				),
			},
		})
	})

	t.Run("can remove one access", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(withTwoAccesses, randomID),
				Check: Resource(
					resourceName,
					Attribute("name", Equals(randomID)),
					SetContains("access", "root"),
					SetContains("access", "ADMIN"),
					SetContains("access", "legacy"),
					SetContains("access", "READ"),
				),
			},
			{
				Config: fmt.Sprintf(withOneAccess, randomID),
				Check: Resource(
					resourceName,
					SetContains("access", "root"),
					SetContains("access", "ADMIN"),
					SetDoesNotContain("access", "legacy"),
					SetDoesNotContain("access", "READ"),
				),
			},
		})
	})

}
