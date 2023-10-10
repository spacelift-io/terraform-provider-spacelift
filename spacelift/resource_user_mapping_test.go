package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

var userWithOneAccess = `
resource "spacelift_user_mapping" "test" {
  email = "%s"
  username = "%s"
  policy {
	space_id = "root"
    role     = "ADMIN"	
  }
}
`

func TestUserResource(t *testing.T) {
	const resourceName = "spacelift_user_mapping.test"

	t.Run("creates and updates a user mapping without an error", func(t *testing.T) {
		randomEmail := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		randomUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		newEmail := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		newUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(userWithOneAccess, randomEmail, randomUsername),
				Check: Resource(
					resourceName,
					Attribute("email", Equals(randomEmail)),
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
				Config: fmt.Sprintf(userWithOneAccess, newEmail, newUsername),
				Check: Resource(
					resourceName,
					Attribute("email", Equals(newEmail)),
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
