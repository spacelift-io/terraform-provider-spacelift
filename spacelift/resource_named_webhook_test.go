package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestNamedWebhookResource(t *testing.T) {
	const resourceName = "spacelift_named_webhook.test"

	t.Run("attach a webhook to root space with all fields filled", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint         = "%s"
					space_id         = "root"
					name             = "testing-named-%s"
					labels           = ["1", "2"]
					secret           = "super-secret"
					enabled          = true
					retry_on_failure = false
				}
			`, endpoint, randomID)
		}

		testStepsFramework(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("secret", Equals("")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals(fmt.Sprintf("testing-named-%s", randomID))),
					Attribute("enabled", Equals("true")),
					SetEquals("labels", "1", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("https://cabbage.org"),
				Check:  Resource(resourceName, Attribute("endpoint", Equals("https://cabbage.org"))),
			},
		})
	})

	t.Run("attach a webhook to root space with default values", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint         = "%s"
					space_id         = "root"
					name             = "testing-named-%s"
					enabled          = true
					retry_on_failure = false
				}
			`, endpoint, randomID)
		}

		testStepsFramework(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("secret", Equals("")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals(fmt.Sprintf("testing-named-%s", randomID))),
					Attribute("enabled", Equals("true")),
					AttributeNotPresent("labels"),
				),
			},
		})
	})

	t.Run("attach a webhook to root space with write-only values", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint          = "%s"
					space_id          = "root"
					name              = "testing-named-%s"
					enabled           = true
					secret_wo         = "super-secret"
					secret_wo_version = 1
					retry_on_failure  = false
				}
			`, endpoint, randomID)
		}

		testStepsFramework(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals(fmt.Sprintf("testing-named-%s", randomID))),
					Attribute("enabled", Equals("true")),
					AttributeNotPresent("labels"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "secret_wo_version"},
			},
		})
	})
}

func TestNamedWebhookResourceMigration(t *testing.T) {
	// lastSDKv2Release picks the latest published release, which is the last SDKv2-based version.
	// Update this to a specific version constraint (e.g. "= 1.19.0") once the migration PR is merged.
	const lastSDKv2Release = ">= 1.0"

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testConfig := fmt.Sprintf(`
		resource "spacelift_named_webhook" "migration_test" {
			endpoint         = "https://migration-test.example.com"
			space_id         = "root"
			name             = "migration-test-%s"
			enabled          = true
			retry_on_failure = false
		}
	`, randomID)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// Step 1: create state with the last SDKv2-based release.
				ExternalProviders: map[string]resource.ExternalProvider{
					"spacelift": {
						Source:            "spacelift.io/spacelift-io/spacelift",
						VersionConstraint: lastSDKv2Release,
					},
				},
				Config: testConfig,
			},
			{
				// Step 2: Framework provider reads the existing state — must produce no diff.
				// PlanOnly: true causes the test to fail if there are any planned changes.
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				Config:                   testConfig,
				PlanOnly:                 true,
			},
		},
	})
}
