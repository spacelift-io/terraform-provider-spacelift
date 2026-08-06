package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

var slackIntegrationWithOneAccess = `
resource "spacelift_slack_channel_integration" "test" {
  integration_name = "%s"
  slack_channel_id = "%s"
  access_rule {
    space_id = "root"
    role     = "ADMIN"
  }
}
`

var slackIntegrationWithTwoAccesses = `
resource "spacelift_slack_channel_integration" "test" {
  integration_name = "%s"
  slack_channel_id = "%s"
  access_rule {
    space_id = "root"
    role     = "ADMIN"
  }
  access_rule {
    space_id = "legacy"
    role     = "READ"
  }
}
`

func TestSlackChannelIntegrationResource(t *testing.T) {
	const resourceName = "spacelift_slack_channel_integration.test"

	t.Run("creates and updates a Slack channel integration without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		channelID := "C" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		oldName := "old name " + randomID
		newName := "new name " + randomID
		newChannelID := "C" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(slackIntegrationWithOneAccess, oldName, channelID),
				Check: Resource(
					resourceName,
					Attribute("integration_name", Equals(oldName)),
					Attribute("slack_channel_id", Equals(channelID)),
					SetContains("access_rule", "root"),
					SetContains("access_rule", "ADMIN"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(slackIntegrationWithOneAccess, newName, newChannelID),
				Check: Resource(
					resourceName,
					Attribute("integration_name", Equals(newName)),
					Attribute("slack_channel_id", Equals(newChannelID)),
				),
			},
		})
	})

	t.Run("rejects creation without access rules", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		channelID := "C" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_slack_channel_integration" "test" {
						integration_name = "%s"
						slack_channel_id = "%s"
					}
				`, randomID, channelID),
				ExpectError: regexp.MustCompile(`access_rule`),
			},
		})
	})

	t.Run("can remove one access", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		channelID := "C" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(slackIntegrationWithTwoAccesses, randomID, channelID),
				Check: Resource(
					resourceName,
					Attribute("integration_name", Equals(randomID)),
					Attribute("slack_channel_id", Equals(channelID)),
					SetContains("access_rule", "root"),
					SetContains("access_rule", "ADMIN"),
					SetContains("access_rule", "legacy"),
					SetContains("access_rule", "READ"),
				),
			},
			{
				Config: fmt.Sprintf(slackIntegrationWithOneAccess, randomID, channelID),
				Check: Resource(
					resourceName,
					SetContains("access_rule", "root"),
					SetContains("access_rule", "ADMIN"),
					SetDoesNotContain("access_rule", "legacy"),
					SetDoesNotContain("access_rule", "READ"),
				),
			},
		})
	})

}
