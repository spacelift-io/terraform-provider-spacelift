package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAIIntegrationData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	// Created inline so the test does not depend on what exists in the account.
	config := fmt.Sprintf(`
		resource "spacelift_ai_integration" "test" {
			name        = "test-ai-data-%s"
			description = "read back by the data sources"
			labels      = ["data-source-test"]
			models      = ["gemini-2.5-pro"]
			space_id    = "%s"
			enabled     = false

			google {
				api_key = "%s"
			}
		}

		data "spacelift_ai_integration" "test" {
			integration_id = spacelift_ai_integration.test.id
		}

		data "spacelift_ai_integrations" "test" {
			labels     = ["data-source-test"]
			depends_on = [spacelift_ai_integration.test]
		}
	`, randomID, testConfig.AI.Space, testConfig.AI.APIKey)

	testSteps(t, []resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeTestCheckFunc(
				Resource(
					"data.spacelift_ai_integration.test",
					Attribute("name", Equals("test-ai-data-"+randomID)),
					Attribute("description", Equals("read back by the data sources")),
					Attribute("ai_provider", Equals(aiProviderGoogle)),
					Attribute("models.0", Equals("gemini-2.5-pro")),
					Attribute("space_id", Equals(testConfig.AI.Space)),
					Attribute("enabled", Equals("false")),
					Attribute("is_spacelift_provided", Equals("false")),
					// Only the matching provider block is populated; the others
					// are null, so they emit no keys at all.
					Attribute("google.#", Equals("1")),
					AttributeNotPresent("anthropic.#"),
					AttributeNotPresent("bedrock.#"),
					AttributeNotPresent("openai.#"),
					SetEquals("labels", "data-source-test"),
				),
				Resource(
					"data.spacelift_ai_integrations.test",
					Nested("integrations",
						CheckInList(
							Attribute("name", Equals("test-ai-data-"+randomID)),
							Attribute("ai_provider", Equals(aiProviderGoogle)),
							Attribute("enabled", Equals("false")),
						),
					),
				),
			),
		},
	})
}
