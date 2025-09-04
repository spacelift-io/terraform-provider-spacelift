package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestUserData(t *testing.T) {
	t.Run("basic test", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testEmail := fmt.Sprintf("test-user-%s@example.com", randomID)
		testUsername := fmt.Sprintf("test-user-%s", randomID)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_user" "test" {
				username         = "%s"
				invitation_email = "%s"
				policy {
					space_id = "root"
					role     = "READ"
				}
			}

			data "spacelift_user" "test" {
				username = spacelift_user.test.username
			}
		`, testUsername, testEmail),
			Check: Resource(
				"data.spacelift_user.test",
				Attribute("username", Equals(testUsername)),
				Attribute("invitation_email", Equals(testEmail)),
				Attribute("policy.#", Equals("1")),
			),
		}})
	})

	t.Run("user not found", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
			data "spacelift_user" "test" {
				username = "non-existent-user"
			}
		`,
			ExpectError: regexp.MustCompile("user with username \"non-existent-user\" not found"),
		}})
	})
}
