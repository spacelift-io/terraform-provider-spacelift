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
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oidc", "secret"}, // OIDC and secret config is not returned in read operations
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
