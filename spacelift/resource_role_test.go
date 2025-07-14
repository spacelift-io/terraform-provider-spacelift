package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestRoleResource(t *testing.T) {
	const resourceName = "spacelift_role.test"

	t.Run("creates and updates roles without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(name, description string, actions []string) string {
			actionsList := ""
			for i, action := range actions {
				if i > 0 {
					actionsList += ", "
				}
				actionsList += fmt.Sprintf(`"%s"`, action)
			}

			return fmt.Sprintf(`
				resource "spacelift_role" "test" {
					name        = "Provider test role %s"
					description = "%s"
					actions     = [%s]
				}
			`, name+randomID, description, actionsList)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("foo", "old description", []string{"SPACE_READ", "SPACE_WRITE"}),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", StartsWith("Provider test role foo")),
					Attribute("description", Equals("old description")),
					SetEquals("actions", "SPACE_READ", "SPACE_WRITE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("bar", "new description", []string{"SPACE_READ", "SPACE_WRITE", "SPACE_ADMIN"}),
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test role bar")),
					Attribute("description", Equals("new description")),
					SetEquals("actions", "SPACE_READ", "SPACE_WRITE", "SPACE_ADMIN"),
				),
			},
		})
	})

	t.Run("can update description", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_role" "test" {
						name        = "Provider test role %s"
						description = "initial description"
						actions     = ["SPACE_READ"]
					}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("initial description")),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "spacelift_role" "test" {
						name        = "Provider test role %s"
						description = "updated description"
						actions     = ["SPACE_READ"]
					}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("updated description")),
				),
			},
		})
	})

	t.Run("can remove description", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_role" "test" {
						name        = "Provider test role %s"
						description = "initial description"
						actions     = ["SPACE_READ"]
					}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("initial description")),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "spacelift_role" "test" {
						name    = "Provider test role %s"
						actions = ["SPACE_READ"]
					}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("")),
				),
			},
		})
	})
}

func TestRoleResourceValidation(t *testing.T) {
	t.Run("fails with invalid action", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_role" "test" {
						name    = "Provider test role %s"
						actions = ["INVALID_ACTION"]
					}
				`, randomID),
				ExpectError: regexp.MustCompile("action INVALID_ACTION is not a valid action. valid actions are:"),
			},
		})
	})

	t.Run("fails with empty name", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: `
					resource "spacelift_role" "test" {
						name    = ""
						actions = ["SPACE_READ"]
					}
				`,
				ExpectError: regexp.MustCompile("must not be an empty string"),
			},
		})
	})

	t.Run("fails with no actions", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_role" "test" {
						name    = "Provider test role %s"
						actions = []
					}
				`, randomID),
				ExpectError: regexp.MustCompile("Attribute actions requires 1 item minimum, but config has only 0 declared"),
			},
		})
	})
}
