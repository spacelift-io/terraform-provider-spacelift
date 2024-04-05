package spacelift

import (
	"fmt"
	"regexp"
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

	t.Run("creates a user without invitation email returns an error", func(t *testing.T) {
		randomUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_user" "test" {
					username = "%s"
					policy {
						space_id = "root"
						role     = "ADMIN"
					}
				}`, randomUsername),
				ExpectError: regexp.MustCompile(`invitation_email is required for new users`),
			},
		})
	})

	t.Run("can edit access list", func(t *testing.T) {
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

	t.Run("cannot change email address", func(t *testing.T) {
		randomUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		exampleEmail := fmt.Sprintf("%s@example.com", randomUsername)

		randomUsername2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		exampleEmail2 := fmt.Sprintf("%s@example.com", randomUsername2)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(userWithOneAccess, exampleEmail, randomUsername),
				Check: Resource(
					resourceName,
					Attribute("invitation_email", Equals(exampleEmail)),
					Attribute("username", Equals(randomUsername)),
					SetContains("policy", "root", "ADMIN"),
					SetDoesNotContain("policy", "legacy"),
				),
			},
			{
				Config:      fmt.Sprintf(userWithOneAccess, exampleEmail2, randomUsername),
				ExpectError: regexp.MustCompile(`invitation_email cannot be changed`),
			},
		})
	})

	t.Run("cannot change username", func(t *testing.T) {
		randomUsername := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		exampleEmail := fmt.Sprintf("%s@example.com", randomUsername)

		randomUsername2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(userWithOneAccess, exampleEmail, randomUsername),
				Check: Resource(
					resourceName,
					Attribute("invitation_email", Equals(exampleEmail)),
					Attribute("username", Equals(randomUsername)),
					SetContains("policy", "root", "ADMIN"),
					SetDoesNotContain("policy", "legacy"),
				),
			},
			{
				Config:      fmt.Sprintf(userWithOneAccess, exampleEmail, randomUsername2),
				ExpectError: regexp.MustCompile(`username cannot be changed`),
			},
		})
	})

}
