package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestIntentProjectResource(t *testing.T) {
	const resourceName = "spacelift_intent_project.test"

	t.Run("creates, updates and clears the TTL", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		withTTL := func(ttl string) string {
			return fmt.Sprintf(`
				resource "spacelift_intent_project" "test" {
					name             = "Provider test intent project %s"
					space_id         = "root"
					description      = "old description"
					labels           = ["one", "two"]
					ttl              = "%s"
					on_expiry_action = "ARCHIVE"
				}
			`, randomID, ttl)
		}

		withoutTTL := fmt.Sprintf(`
			resource "spacelift_intent_project" "test" {
				name             = "Provider test intent project %s"
				space_id         = "root"
				description      = "new description"
				labels           = ["one", "two"]
				on_expiry_action = "ARCHIVE"
			}
		`, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: withTTL("72h"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-intent-project")),
					Attribute("name", StartsWith("Provider test intent project")),
					Attribute("space_id", Equals("root")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("ttl", Equals("72h")),
					Attribute("ttl_seconds", Equals("259200")),
					Attribute("on_expiry_action", Equals("ARCHIVE")),
					Attribute("state", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ttl",
					"expires_at",
					"keep_resources_on_destroy",
				},
			},
			{
				Config: withTTL("168h"),
				Check: Resource(
					resourceName,
					Attribute("ttl", Equals("168h")),
					Attribute("ttl_seconds", Equals("604800")),
				),
			},
			{
				Config: withoutTTL,
				Check: Resource(
					resourceName,
					Attribute("description", Equals("new description")),
					Attribute("ttl", Equals("")),
					Attribute("ttl_seconds", Equals("0")),
				),
			},
		})
	})
}
