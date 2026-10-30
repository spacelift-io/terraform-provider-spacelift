package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSlackChannelIntegrationData(t *testing.T) {
	t.Run("reads an existing Slack channel integration by ID", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		integrationName := "test-integration-" + randomID
		channelID := "C" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_slack_channel_integration" "test" {
					integration_name = "%s"
					slack_channel_id = "%s"
					access_rule {
						space_id = "root"
						role     = "READ"
					}
				}

				data "spacelift_slack_channel_integration" "test" {
					integration_id = spacelift_slack_channel_integration.test.id
				}
			`, integrationName, channelID),
			Check: Resource(
				"data.spacelift_slack_channel_integration.test",
				Attribute("integration_name", Equals(integrationName)),
				Attribute("slack_channel_id", Equals(channelID)),
				Attribute("access_rule.#", Equals("1")),
			),
		}})
	})

	t.Run("returns error when ID not found", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_slack_channel_integration" "test" {
					integration_id = "non-existent-id"
				}
			`,
			ExpectError: regexp.MustCompile(`could not find Slack channel integration with ID "non-existent-id"`),
		}})
	})
}
