package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPluginResource(t *testing.T) {
	const resourceName = "spacelift_plugin.test"

	t.Run("creates and updates plugin without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(name, labelID string) string {
			return fmt.Sprintf(`
				resource "spacelift_plugin" "test" {
					name               = "Provider test plugin %s"
					plugin_template_id = "infracost"
					stack_label        = "%s"
					labels             = ["test", "terraform"]
					parameters = {
						infracost_api_key = "test"
					}
				}
			`, name+randomID, labelID+randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("foo", "test-label"),
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test plugin foo")),
					Attribute("plugin_template_id", Equals("infracost")),
					Attribute("stack_label", StartsWith("test-label")),
					SetEquals("labels", "test", "terraform"),
					Attribute("id", IsNotEmpty()),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: config("bar", "test-label"),
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test plugin bar")),
				),
			},
		})
	})

	t.Run("creates plugin with parameters", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_plugin" "test" {
				name               = "Provider test plugin %s"
				plugin_template_id = "infracost"
				stack_label        = "test-label-%s"
				parameters = {
					infracost_api_key = "test"
				}
				labels             = ["test"]
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test plugin")),
					SetEquals("labels", "test"),
				),
			},
		})
	})
}

func TestPluginResourceInSpace(t *testing.T) {
	const resourceName = "spacelift_plugin.test"

	t.Run("creates plugin in a space", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_plugin" "test" {
				name               = "Provider test plugin %s"
				plugin_template_id = "infracost"
				stack_label        = "test-label-%s"
				space_id           = "root"
				labels             = ["test"]
				parameters = {
					infracost_api_key = "test"
				}
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test plugin")),
					Attribute("space_id", Equals("root")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
		})
	})
}

func TestPluginResourceMissingParameters(t *testing.T) {
	const resourceName = "spacelift_plugin.test"

	t.Run("missing parameter causes errors", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_plugin" "test" {
				name               = "Provider test plugin %s"
				plugin_template_id = "infracost"
				stack_label        = "test-label-%s"
				space_id           = "root"
				labels             = ["test"]
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("required parameter 'infracost_api_key' is missing"),
			},
		})
	})
}

func TestPluginResourceInvalidParameters(t *testing.T) {
	const resourceName = "spacelift_plugin.test"

	t.Run("unknown parameter causes error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_plugin" "test" {
				name               = "Provider test plugin %s"
				plugin_template_id = "infracost"
				stack_label        = "test-label-%s"
				space_id           = "root"
				labels             = ["test"]
				parameters = {
					infracost_api_key = "test"
					invalid_param     = "value"
				}
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("unknown parameter 'invalid_param': not defined in plugin template"),
			},
		})
	})

	t.Run("typo in parameter name causes error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_plugin" "test" {
				name               = "Provider test plugin %s"
				plugin_template_id = "infracost"
				stack_label        = "test-label-%s"
				space_id           = "root"
				labels             = ["test"]
				parameters = {
					infracost_api_keyy = "test"
				}
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("unknown parameter 'infracost_api_keyy': not defined in plugin template"),
			},
		})
	})
}
