package spacelift

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestRoleAttachmentResource(t *testing.T) {
	const resourceName = "spacelift_role_attachment.test"

	t.Run("exactly one of API key, IDP group mapping, stack, or user must be set", func(t *testing.T) {
		config := `
			resource "spacelift_role_attachment" "test" {
				api_key_id           = "AAAA"
				idp_group_mapping_id = "BBBB"
				stack_id             = "CCCC"
				user_id              = "DDDD"
				space_id             = "EEEE"
				role_id              = "FFFF"
			}
		`

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("(?s)only one of `api_key_id,idp_group_mapping_id,stack_id,user_id`.*can be specified"),
			},
		})
	})

	t.Run("with an API key", func(t *testing.T) {
		apiKeyID := os.Getenv("SPACELIFT_API_KEY_ID")
		if apiKeyID == "" {
			t.Skip("SPACELIFT_API_KEY_ID environment variable is not set, skipping role attachment tests")
			return
		}

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		configInitial := fmt.Sprintf(`
			resource "spacelift_role" "test" {
				name        = "Test role attachment - initial role %s"
				description = "Test role for attachment"
				actions     = ["SPACE_READ"]
			}

			resource "spacelift_space" "test" {
				name = "Test API Space %s"
				parent_space_id = "root"
			}

			resource "spacelift_role_attachment" "test" {
				api_key_id = "%s"
				role_id    = spacelift_role.test.id
				space_id   = spacelift_space.test.id
			}
		`, randomID, randomID, apiKeyID)

		configUpdated := fmt.Sprintf(`
		    resource "spacelift_role" "another_role" {
				name        = "Test role attachment - another role %s"
				description = "Another role for attachment"
				actions     = ["SPACE_READ", "SPACE_WRITE"]
			}

			resource "spacelift_space" "another_space" {
				name = "Test API Space Another %s"
				parent_space_id = "root"
			}

			resource "spacelift_role_attachment" "test" {
				api_key_id = "%s"
				role_id    = spacelift_role.another_role.id
				space_id   = spacelift_space.another_space.id
			}
		`, randomID, randomID, apiKeyID)

		testSteps(t, []resource.TestStep{
			{
				Config: configInitial,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("api_key_id", Equals(apiKeyID)),
					Attribute("role_id", IsNotEmpty()),
					Attribute("space_id", StartsWith("test-api-space-")),
					AttributeNotPresent("idp_group_mapping_id"),
				),
			},
			{
				Config: configUpdated,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("api_key_id", Equals(apiKeyID)),
					Attribute("role_id", IsNotEmpty()),
					Attribute("space_id", StartsWith("test-api-space-another-")),
					AttributeNotPresent("idp_group_mapping_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})

	t.Run("with an IDP group mapping", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		configInitial := fmt.Sprintf(`
			resource "spacelift_idp_group_mapping" "test" {
				name = "Test IDP Group Mapping %s"
	        }

			resource "spacelift_role" "test" {
				name        = "Test role attachment - initial role %s"
				description = "Test role for attachment"
				actions     = ["SPACE_READ"]
			}

			resource "spacelift_space" "test" {
				name = "Test IDP Space %s"
				parent_space_id = "root"
			}

			resource "spacelift_role_attachment" "test" {
				idp_group_mapping_id = spacelift_idp_group_mapping.test.id
				role_id              = spacelift_role.test.id
				space_id             = spacelift_space.test.id
			}
		`, randomID, randomID, randomID)

		configUpdated := fmt.Sprintf(`
			resource "spacelift_idp_group_mapping" "test" {
				name = "Test IDP Group Mapping %s Updated"
	        }
			
			resource "spacelift_role" "another_role" {
				name        = "Test role attachment - another role %s"
				description = "Another role for attachment"
				actions     = ["SPACE_READ", "SPACE_WRITE"]
			}

			resource "spacelift_space" "another_space" {
				name = "Test IDP Space Another %s"
				parent_space_id = "root"
			}

			resource "spacelift_role_attachment" "test" {
				idp_group_mapping_id = spacelift_idp_group_mapping.test.id
				role_id              = spacelift_role.another_role.id
				space_id             = spacelift_space.another_space.id
			}
		`, randomID, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: configInitial,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("idp_group_mapping_id", IsNotEmpty()),
					Attribute("role_id", IsNotEmpty()),
					Attribute("space_id", IsNotEmpty()),
				),
			},
			{
				Config: configUpdated,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("idp_group_mapping_id", IsNotEmpty()),
					Attribute("role_id", IsNotEmpty()),
					Attribute("space_id", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})

	t.Run("with a user", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		configInitial := fmt.Sprintf(`
			resource "spacelift_role" "test" {
				name        = "Test role attachment - initial role %s"
				description = "Test role for attachment"
				actions     = ["SPACE_READ"]
			}

			resource "spacelift_space" "test" {
				name = "Test User Space %s"
				parent_space_id = "root"
			}

			resource "spacelift_user" "test" {
				username = "%s"
				invitation_email = "%s"
			}

			resource "spacelift_role_attachment" "test" {
				user_id  = spacelift_user.test.id
				role_id  = spacelift_role.test.id
				space_id = spacelift_space.test.id
			}
		`, randomID, randomID, fmt.Sprintf("%s@example.com", randomID), fmt.Sprintf("%s@example.com", randomID))

		configUpdated := fmt.Sprintf(`
			resource "spacelift_role" "another_role" {
				name        = "Test role attachment - another role %s"
				description = "Another role for attachment"
				actions     = ["SPACE_READ", "SPACE_WRITE"]
			}

			resource "spacelift_space" "another_space" {
				name = "Test User Space Another %s"
				parent_space_id = "root"
			}

			resource "spacelift_user" "another_user" {
				username = "%s"
				invitation_email = "%s"
			}

			resource "spacelift_role_attachment" "test" {
				user_id  = spacelift_user.another_user.id
				role_id  = spacelift_role.another_role.id
				space_id = spacelift_space.another_space.id
			}
		`, randomID, randomID, fmt.Sprintf("%s+another@example.com", randomID), fmt.Sprintf("%s+another@example.com", randomID))

		testSteps(t, []resource.TestStep{
			{
				Config: configInitial,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("user_id", IsNotEmpty()),
					Attribute("role_id", IsNotEmpty()),
					Attribute("space_id", StartsWith("test-user-space-")),
					AttributeNotPresent("api_key_id"),
					AttributeNotPresent("idp_group_mapping_id"),
				),
			},
			{
				Config: configUpdated,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("user_id", IsNotEmpty()),
					Attribute("role_id", IsNotEmpty()),
					Attribute("space_id", StartsWith("test-user-space-another-")),
					AttributeNotPresent("api_key_id"),
					AttributeNotPresent("idp_group_mapping_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})
}
