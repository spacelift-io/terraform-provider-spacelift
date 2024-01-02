package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBitbucketDatacenterIntegrationResource(t *testing.T) {
	const resourceName = "spacelift_bitbucket_datacenter_integration.test"

	t.Run("creates and updates a bitbucket datacenter integration without an error", func(t *testing.T) {
		randomChars := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(apiHost, userFacingHost, username, accessToken string) string {
			return fmt.Sprintf(`
				resource "spacelift_bitbucket_datacenter_integration" "test" {
					api_host        = "apiHost-%s"
					user_facing_host = "facingHost-%s"
					username = "user-%s"
					access_token = "access-%s"
				}
			`, apiHost, userFacingHost, username, accessToken)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(randomChars, randomChars, randomChars, randomChars),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("api_host", StartsWith("apiHost-")),
					Attribute("user_facing_host", StartsWith("facingHost-")),
					Attribute("username", StartsWith("user-")),
					Attribute("access_token", StartsWith("access-")),
					Attribute("webhook_url", IsNotEmpty()),
					Attribute("webhook_secret", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config(randomChars, "newUserFacingHost", "newUsername", randomChars),
				Check: Resource(
					resourceName,
					Attribute("user_facing_host", Equals("newUserFacingHost")),
					Attribute("username", Equals("newUsername")),
				),
			},
		})
	})
}
