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
					api_host        = "%s"
					user_facing_host = "%s"
					username = "%s"
					access_token = "access-%s"
				}
			`, apiHost, userFacingHost, username, accessToken)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bitbucket.com/new-"+randomChars, "https://bitbucket.com/new-"+randomChars, "userName", randomChars),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("api_host", StartsWith("https://bitbucket.com")),
					Attribute("user_facing_host", StartsWith("https://bitbucket.com")),
					Attribute("username", StartsWith("userName")),
					Attribute("access_token", StartsWith("access-")),
					Attribute("webhook_url", IsNotEmpty()),
					Attribute("webhook_secret", IsNotEmpty()),
				),
			},
			{
				Config: config("https://bitbucket.com/new-"+randomChars, "https://bitbucket.com/new-"+randomChars, "newUserName", randomChars),
				Check: Resource(
					resourceName,
					Attribute("user_facing_host", StartsWith("https://bitbucket.com/new")),
					Attribute("username", Equals("newUserName")),
				),
			},
		})
	})
}
