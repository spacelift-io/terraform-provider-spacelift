package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAPIKeyResource(t *testing.T) {
	const resourceName = "spacelift_api_key.test"

	t.Run("creates and updates a SECRET API key", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(name, idpGroup string) string {
			return fmt.Sprintf(`
				resource "spacelift_api_key" "test" {
					name = "%s"
					idp_groups = ["%s"]
				}
			`, name, idpGroup)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(fmt.Sprintf("Test API Key %s", randomID), "developers"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("Test API Key %s", randomID))),
					Attribute("type", Equals("SECRET")),
					Attribute("secret", IsNotEmpty()),
					SetEquals("idp_groups", "developers"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"}, // Secret is not returned in read operations
			},
			{
				Config: config(fmt.Sprintf("Updated API Key %s", randomID), "admins"),
				Check: Resource(
					resourceName,
					Attribute("name", Equals(fmt.Sprintf("Updated API Key %s", randomID))),
					Attribute("type", Equals("SECRET")),
					SetEquals("idp_groups", "admins"),
				),
			},
		})
	})

	t.Run("creates a SECRET API key with access rules", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_space" "test" {
				name = "Test Space %s"
				parent_space_id = "root"
			}

			resource "spacelift_api_key" "test" {
				name = "Test API Key with Access %s"
				idp_groups = ["developers"]
				
				access_rule {
					space_id = spacelift_space.test.id
					role = "READ"
				}
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", Equals(fmt.Sprintf("Test API Key with Access %s", randomID))),
					Attribute("type", Equals("SECRET")),
					Attribute("secret", IsNotEmpty()),
					SetEquals("idp_groups", "developers"),
					Attribute("access_rule.#", Equals("1")),
				),
			},
		})
	})

	t.Run("creates and updates access rules", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		configSingleRule := fmt.Sprintf(`
			resource "spacelift_space" "test" {
				name = "Test Space %s"
				parent_space_id = "root"
			}

			resource "spacelift_api_key" "test" {
				name = "Test API Key Access Update %s"
				idp_groups = ["developers"]
				
				access_rule {
					space_id = spacelift_space.test.id
					role = "READ"
				}
			}
		`, randomID, randomID)

		configMultipleRules := fmt.Sprintf(`
			resource "spacelift_space" "test" {
				name = "Test Space %s"
				parent_space_id = "root"
			}

			resource "spacelift_space" "test2" {
				name = "Test Space 2 %s"
				parent_space_id = "root"
			}

			resource "spacelift_api_key" "test" {
				name = "Test API Key Access Update %s"
				idp_groups = ["developers"]
				
				access_rule {
					space_id = spacelift_space.test.id
					role = "WRITE"
				}
				
				access_rule {
					space_id = spacelift_space.test2.id
					role = "ADMIN"
				}
			}
		`, randomID, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: configSingleRule,
				Check: Resource(
					resourceName,
					Attribute("access_rule.#", Equals("1")),
				),
			},
			{
				Config: configMultipleRules,
				Check: Resource(
					resourceName,
					Attribute("access_rule.#", Equals("2")),
				),
			},
		})
	})
}

func TestAPIKeyResourceOIDC(t *testing.T) {
	const resourceName = "spacelift_api_key.test"

	t.Run("creates and updates an OIDC API key", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(name, issuer, clientID, subject string) string {
			return fmt.Sprintf(`
				resource "spacelift_api_key" "test" {
					name = "%s"
					idp_groups = ["github-users"]
					
					oidc {
						issuer = "%s"
						client_id = "%s"
						subject_expression = "%s"
					}
				}
			`, name, issuer, clientID, subject)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(
					fmt.Sprintf("OIDC API Key %s", randomID),
					"https://token.actions.githubusercontent.com",
					"client123",
					"repo:spacelift-io/terraform-provider-spacelift:ref:refs/heads/main",
				),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("OIDC API Key %s", randomID))),
					Attribute("type", Equals("OIDC")),
					Attribute("oidc.0.issuer", Equals("https://token.actions.githubusercontent.com")),
					Attribute("oidc.0.client_id", Equals("client123")),
					Attribute("oidc.0.subject_expression", Equals("repo:spacelift-io/terraform-provider-spacelift:ref:refs/heads/main")),
					SetEquals("idp_groups", "github-users"),
					// OIDC keys don't have secrets
					AttributeNotPresent("secret"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oidc"}, // OIDC config is not returned in read operations
			},
			{
				Config: config(
					fmt.Sprintf("Updated OIDC API Key %s", randomID),
					"https://token.actions.githubusercontent.com",
					"client456",
					"repo:spacelift-io/terraform-provider-spacelift:ref:refs/heads/develop",
				),
				Check: Resource(
					resourceName,
					Attribute("name", Equals(fmt.Sprintf("Updated OIDC API Key %s", randomID))),
					Attribute("oidc.0.client_id", Equals("client456")),
					Attribute("oidc.0.subject_expression", Equals("repo:spacelift-io/terraform-provider-spacelift:ref:refs/heads/develop")),
				),
			},
		})
	})

	t.Run("creates an OIDC API key with access rules", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_space" "test" {
				name = "Test OIDC Space %s"
				parent_space_id = "root"
			}

			resource "spacelift_api_key" "test" {
				name = "OIDC API Key with Access %s"
				idp_groups = ["github-users"]
				
				oidc {
					issuer = "https://token.actions.githubusercontent.com"
					client_id = "oidc-client"
					subject_expression = "repo:spacelift-io/test:ref:refs/heads/main"
				}
				
				access_rule {
					space_id = spacelift_space.test.id
					role = "WRITE"
				}
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", Equals(fmt.Sprintf("OIDC API Key with Access %s", randomID))),
					Attribute("type", Equals("OIDC")),
					Attribute("access_rule.#", Equals("1")),
					AttributeNotPresent("secret"), // OIDC keys don't have secrets
				),
			},
		})
	})
}

func TestAPIKeyResourceCustomRoles(t *testing.T) {
	const resourceName = "spacelift_api_key.test"

	// This test requires actual API calls and needs test.env file configured
	t.Run("creates API key with custom role", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		// First create a custom role
		config := fmt.Sprintf(`
			resource "spacelift_space" "test" {
				name = "Test Custom Role Space %s"
				parent_space_id = "root"
			}

			resource "spacelift_role" "test" {
				name = "Custom Test Role %s"
				actions = ["SPACE_READ"]
			}

			resource "spacelift_api_key" "test" {
				name = "API Key with Custom Role %s"
				idp_groups = ["developers"]
				
				access_rule {
					space_id = spacelift_space.test.id
					role = spacelift_role.test.name
				}
			}
		`, randomID, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", Equals(fmt.Sprintf("API Key with Custom Role %s", randomID))),
					Attribute("type", Equals("SECRET")),
					Attribute("access_rule.#", Equals("1")),
				),
			},
		})
	})
}

func TestAPIKeyResourceErrors(t *testing.T) {
	t.Run("fails with empty name", func(t *testing.T) {
		config := `
			resource "spacelift_api_key" "test" {
				name = ""
				idp_groups = ["developers"]
			}
		`

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("must not be an empty string"),
			},
		})
	})

	t.Run("creates API key with empty idp_groups", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_api_key" "test" {
				name = "Test API Key %s"
				idp_groups = []
			}
		`, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					"spacelift_api_key.test",
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("Test API Key %s", randomID))),
					Attribute("type", Equals("SECRET")),
					Attribute("secret", IsNotEmpty()),
				),
			},
		})
	})

	t.Run("fails with invalid OIDC configuration", func(t *testing.T) {
		config := `
			resource "spacelift_api_key" "test" {
				name = "Test OIDC Key"
				idp_groups = ["developers"]
				
				oidc {
					issuer = ""
					client_id = "test"
					subject_expression = "test"
				}
			}
		`

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("must not be an empty string"),
			},
		})
	})
}