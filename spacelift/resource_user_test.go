package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

var userWithOneAccess = `
resource "spacelift_user" "test" {
  invitation_email = "%s"
  username = "%s"
  policy {
	space_id = "root"
    role     = "ADMIN"
  }
}
`

var userWithTwoAccesses = `
resource "spacelift_user" "test" {
  invitation_email = "%s"
  username = "%s"
  policy {
    space_id = "root"
    role     = "ADMIN"
  }
  policy {
    space_id = "legacy"
    role     = "READ"
  }
}
`

func TestUserResource(t *testing.T) {
	const resourceName = "spacelift_user.test"

	t.Run("creates a user without an error", func(t *testing.T) {
		randomUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		exampleEmail := fmt.Sprintf("%s@example.com", randomUsername)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(userWithOneAccess, exampleEmail, randomUsername),
				Check: Resource(
					resourceName,
					Attribute("invitation_email", Equals(exampleEmail)),
					Attribute("username", Equals(randomUsername)),
					SetContains("policy", "root"),
					SetContains("policy", "ADMIN"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})

	// Note: the api doesn't allow for the username or email to be updated
	t.Run("can remove one access", func(t *testing.T) {
		randomUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		exampleEmail := fmt.Sprintf("%s@example.com", randomUsername)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(userWithTwoAccesses, exampleEmail, randomUsername),
				Check: Resource(
					resourceName,
					Attribute("invitation_email", Equals(exampleEmail)),
					Attribute("username", Equals(randomUsername)),
					SetContains("policy", "root", "ADMIN"),
					SetContains("policy", "legacy", "READ")),
			},
			{
				Config: fmt.Sprintf(userWithOneAccess, exampleEmail, randomUsername),
				Check: Resource(
					resourceName,
					Attribute("invitation_email", Equals(exampleEmail)),
					Attribute("username", Equals(randomUsername)),
					SetContains("policy", "root", "ADMIN"),
					SetDoesNotContain("policy", "legacy", "READ")),
			},
		})

	})

}
