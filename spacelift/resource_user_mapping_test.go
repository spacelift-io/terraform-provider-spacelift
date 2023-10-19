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

func TestUserResource(t *testing.T) {
	const resourceName = "spacelift_user.test"

	t.Run("creates and updates a user without an error", func(t *testing.T) {
		randomUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		exampleEmail := fmt.Sprintf("%s@example.com", randomUsername)

		newUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		exampleEmailNew := fmt.Sprintf("%s@example.com", newUsername)

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
			{
				Config: fmt.Sprintf(userWithOneAccess, exampleEmailNew, newUsername),
				Check: Resource(
					resourceName,
					Attribute("invitation_email", Equals(exampleEmailNew)),
					Attribute("username", Equals(newUsername)),
					SetContains("policy", "root"),
					SetContains("policy", "ADMIN"),
				),
			},
		})
	})

	t.Run("can remove one access", func(t *testing.T) {

	})

}
