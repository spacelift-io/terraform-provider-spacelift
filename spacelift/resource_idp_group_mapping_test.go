package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

var withOneAccess = `
resource "spacelift_idp_group_mapping" "test" {
  name = "%s"
  policy {
    space_id = "root"
    role     = "ADMIN"
  }
}
`

var withTwoAccesses = `
resource "spacelift_idp_group_mapping" "test" {
  name = "%s"
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

func TestIdpGroupMappingResource(t *testing.T) {
	const resourceName = "spacelift_idp_group_mapping.test"

	t.Run("creates and updates a user group mapping without an error", func(t *testing.T) {
		oldName := "old name"
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(withOneAccess, oldName),
				Check: Resource(
					resourceName,
					Attribute("name", Equals(oldName)),
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
					SetContains("policy", "root"),
					SetContains("policy", "ADMIN"),
					SetContains("policy", "legacy"),
					SetContains("policy", "READ"),
				),
			},
			{
				Config: fmt.Sprintf(withOneAccess, randomID),
				Check: Resource(
					resourceName,
					SetContains("policy", "root"),
					SetContains("policy", "ADMIN"),
					SetDoesNotContain("policy", "legacy"),
					SetDoesNotContain("policy", "READ"),
				),
			},
		})
	})

}
