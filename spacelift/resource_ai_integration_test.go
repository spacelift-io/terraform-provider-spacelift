package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

// A provider has to be registered in three places, and each mistake fails
// quietly: missing from aiProviderBlockNames escapes ExactlyOneOf and allows
// two providers at once, missing from aiProviderBlocks makes its integrations
// unreadable.
func TestAIProviderBlockWiring(t *testing.T) {
	t.Parallel()

	// The nested ConflictsWith/RequiredWith/AtLeastOneOf paths are strings, so
	// a typo only surfaces at runtime.
	if err := resourceAIIntegration().InternalValidate(nil, true); err != nil {
		t.Fatalf("resource schema is invalid: %v", err)
	}

	resourceSchema := resourceAIIntegration().Schema

	blocksByName := map[string]bool{}
	for provider, block := range aiProviderBlocks {
		if provider == "" || block == "" {
			t.Errorf("aiProviderBlocks has an empty entry: %q -> %q", provider, block)
		}
		blocksByName[block] = true
	}

	if len(blocksByName) != len(aiProviderBlockNames) {
		t.Errorf(
			"aiProviderBlocks has %d distinct blocks but aiProviderBlockNames lists %d",
			len(blocksByName), len(aiProviderBlockNames),
		)
	}

	for _, name := range aiProviderBlockNames {
		if !blocksByName[name] {
			t.Errorf("block %q is listed in aiProviderBlockNames but maps to no provider", name)
		}

		block, ok := resourceSchema[name]
		if !ok {
			t.Errorf("block %q is registered but missing from the resource schema", name)
			continue
		}

		if block.MaxItems != 1 {
			t.Errorf("block %q must set MaxItems: 1, got %d", name, block.MaxItems)
		}

		// ExactlyOneOf is what makes "no provider" and "two providers" both
		// plan-time errors, so every block must carry the full list.
		if len(block.ExactlyOneOf) != len(aiProviderBlockNames) {
			t.Errorf(
				"block %q must set ExactlyOneOf to every provider block, got %v",
				name, block.ExactlyOneOf,
			)
			continue
		}

		for index, expected := range aiProviderBlockNames {
			if block.ExactlyOneOf[index] != expected {
				t.Errorf(
					"block %q has ExactlyOneOf %v, expected %v",
					name, block.ExactlyOneOf, aiProviderBlockNames,
				)
				break
			}
		}
	}
}

func TestAIIntegrationResource(t *testing.T) {
	t.Parallel()

	const resourceName = "spacelift_ai_integration.test"

	t.Run("creates and updates a Google integration", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string, models string, enabled bool) string {
			return fmt.Sprintf(`
				resource "spacelift_ai_integration" "test" {
					name        = "test-ai-integration-%s"
					description = "%s"
					labels      = ["one", "two"]
					models      = %s
					space_id    = "%s"
					enabled     = %t

					google {
						api_key = "%s"
					}
				}
			`, randomID, description, models, testConfig.AI.Space, enabled, testConfig.AI.APIKey)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("initial description", `["gemini-2.5-pro"]`, true),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals("test-ai-integration-"+randomID)),
					// Derived from the block that is set, never written directly.
					Attribute("ai_provider", Equals("Google")),
					Attribute("description", Equals("initial description")),
					Attribute("models.#", Equals("1")),
					Attribute("models.0", Equals("gemini-2.5-pro")),
					Attribute("space_id", Equals(testConfig.AI.Space)),
					Attribute("enabled", Equals("true")),
					Attribute("is_spacelift_provided", Equals("false")),
					Attribute("google.#", Equals("1")),
					// The API never returns the key, so it is always blank in state.
					Attribute("google.0.api_key", IsEmpty()),
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: config("updated description", `["gemini-2.5-pro", "gemini-2.5-flash"]`, true),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("updated description")),
					Attribute("models.#", Equals("2")),
					Attribute("models.1", Equals("gemini-2.5-flash")),
				),
			},
			{
				// enabled goes through aiIntegrationToggle rather than the update
				// mutation, so exercise it on its own.
				Config: config("updated description", `["gemini-2.5-pro", "gemini-2.5-flash"]`, false),
				Check: Resource(
					resourceName,
					Attribute("enabled", Equals("false")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// The API never returns the key, so it cannot survive an import
				// verification.
				ImportStateVerifyIgnore: []string{"google.0.api_key"},
			},
		})
	})

	// A mis-wired block only fails for its own provider, so cover each one.
	for _, testCase := range []struct {
		block    string
		provider string
	}{
		{block: "anthropic", provider: aiProviderAnthropic},
		{block: "google", provider: aiProviderGoogle},
		{block: "openai", provider: aiProviderOpenAI},
	} {
		t.Run("creates a "+testCase.block+" integration and defaults its models", func(t *testing.T) {
			randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

			// base_url is deliberately left out: Spacelift validates it on create
			// by calling the provider's model listing endpoint, so it cannot be
			// exercised without a real, reachable gateway.
			testSteps(t, []resource.TestStep{
				{
					Config: fmt.Sprintf(`
						resource "spacelift_ai_integration" "test" {
							name     = "test-ai-integration-%s"
							space_id = "%s"
							enabled  = false

							%s {
								api_key = "%s"
							}
						}
					`, randomID, testConfig.AI.Space, testCase.block, testConfig.AI.APIKey),
					Check: Resource(
						resourceName,
						Attribute("ai_provider", Equals(testCase.provider)),
						Attribute("enabled", Equals("false")),
						Attribute(testCase.block+".#", Equals("1")),
						Attribute(testCase.block+".0.base_url", IsEmpty()),
						Attribute("models.#", NotEquals("0")),
					),
				},
			})
		})
	}

	t.Run("creates and updates a Bedrock integration", func(t *testing.T) {
		// Unlike the API-key providers, Bedrock cannot run against a placeholder:
		// Spacelift assumes the AWS integration's role and validates the
		// inference profiles against AWS before saving.
		if testConfig.AI.Bedrock.IntegrationID == "" || testConfig.AI.Bedrock.Profile == "" {
			t.Skip("SPACELIFT_PROVIDER_TEST_AI_BEDROCK_INTEGRATIONID and _PROFILE are not set, skipping Bedrock tests")
		}

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(enabled bool) string {
			return fmt.Sprintf(`
				resource "spacelift_ai_integration" "test" {
					name     = "test-ai-integration-%s"
					space_id = "%s"
					enabled  = %t

					bedrock {
						integration_id = "%s"
						region         = "%s"
						profiles       = ["%s"]
					}
				}
			`,
				randomID, testConfig.AI.Space, enabled,
				testConfig.AI.Bedrock.IntegrationID,
				testConfig.AI.Bedrock.Region,
				testConfig.AI.Bedrock.Profile,
			)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(false),
				Check: Resource(
					resourceName,
					Attribute("ai_provider", Equals(aiProviderBedrock)),
					Attribute("enabled", Equals("false")),
					Attribute("bedrock.0.integration_id", Equals(testConfig.AI.Bedrock.IntegrationID)),
					Attribute("bedrock.0.region", Equals(testConfig.AI.Bedrock.Region)),
					Attribute("bedrock.0.profiles.0", Equals(testConfig.AI.Bedrock.Profile)),
					// Resolved by the API from the attached AWS integration.
					Attribute("bedrock.0.integration_name", IsNotEmpty()),
					// Bedrock takes its models from the profiles rather than
					// accepting them directly.
					Attribute("models.0", Equals(testConfig.AI.Bedrock.Profile)),
				),
			},
			{
				Config: config(true),
				Check: Resource(
					resourceName,
					Attribute("enabled", Equals("true")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})

	t.Run("rejects models on a Bedrock integration", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_ai_integration" "test" {
						name     = "test-ai-integration-%s"
						space_id = "%s"
						models   = ["some-model"]

						bedrock {
							integration_id = "some-integration"
							region         = "us-east-1"
							profiles       = ["some-profile"]
						}
					}
				`, randomID, testConfig.AI.Space),
				ExpectError: regexp.MustCompile(`"models": conflicts with bedrock`),
			},
		})
	})

	t.Run("creates and rotates a write-only API key", func(t *testing.T) {
		// Unlike api_key, this is read from the raw config by cty path.
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(version string) string {
			return fmt.Sprintf(`
				resource "spacelift_ai_integration" "test" {
					name     = "test-ai-integration-%s"
					space_id = "%s"
					enabled  = false

					google {
						api_key_wo         = "%s"
						api_key_wo_version = "%s"
					}
				}
			`, randomID, testConfig.AI.Space, testConfig.AI.APIKey, version)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("1"),
				Check: Resource(
					resourceName,
					Attribute("ai_provider", Equals(aiProviderGoogle)),
					Attribute("google.0.api_key_wo_version", Equals("1")),
					Attribute("google.0.api_key", IsEmpty()),
				),
			},
			{
				Config: config("2"),
				Check: Resource(
					resourceName,
					Attribute("google.0.api_key_wo_version", Equals("2")),
					Attribute("google.0.api_key", IsEmpty()),
				),
			},
		})
	})

	t.Run("replaces the integration when the provider block changes", func(t *testing.T) {
		// aiIntegrationUpdate accepts a provider argument but the API rejects
		// changing it, so swapping the block must force a replacement.
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(block string) string {
			return fmt.Sprintf(`
				resource "spacelift_ai_integration" "test" {
					name     = "test-ai-integration-%s"
					space_id = "%s"
					enabled  = false

					%s {
						api_key = "%s"
					}
				}
			`, randomID, testConfig.AI.Space, block, testConfig.AI.APIKey)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("anthropic"),
				Check: Resource(
					resourceName,
					Attribute("ai_provider", Equals(aiProviderAnthropic)),
					Attribute("anthropic.#", Equals("1")),
					Attribute("openai.#", Equals("0")),
				),
			},
			{
				Config: config("openai"),
				Check: Resource(
					resourceName,
					Attribute("ai_provider", Equals(aiProviderOpenAI)),
					Attribute("openai.#", Equals("1")),
					Attribute("anthropic.#", Equals("0")),
				),
			},
		})
	})

	t.Run("rejects a provider block with no API key", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_ai_integration" "test" {
						name     = "test-ai-integration-%s"
						space_id = "%s"

						google {}
					}
				`, randomID, testConfig.AI.Space),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
		})
	})

	t.Run("rejects more than one provider block", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_ai_integration" "test" {
						name     = "test-ai-integration-%s"
						space_id = "%s"

						google {
							api_key = "one"
						}

						openai {
							api_key = "two"
						}
					}
				`, randomID, testConfig.AI.Space),
				ExpectError: regexp.MustCompile("only one of `anthropic,bedrock,google,openai` can be specified"),
			},
		})
	})

	t.Run("rejects a configuration with no provider block", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_ai_integration" "test" {
						name     = "test-ai-integration-%s"
						space_id = "%s"
					}
				`, randomID, testConfig.AI.Space),
				ExpectError: regexp.MustCompile("one of `anthropic,bedrock,google,openai` must be specified"),
			},
		})
	})
}
