package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPluginData(t *testing.T) {
	const resourceName = "data.spacelift_plugin.test"

	t.Run("reads plugin data without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_plugin" "test" {
				name               = "Provider test plugin %s"
				plugin_template_id = "infracost"
				stack_label        = "test-label-%s"
				labels             = ["test", "data-source"]
                parameters = {
					infracost_api_key = "test"
				}
			}

			data "spacelift_plugin" "test" {
				plugin_id = spacelift_plugin.test.id
			}
		`, randomID, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test plugin")),
					Attribute("plugin_template_id", Equals("infracost")),
					SetEquals("labels", "test", "data-source"),
					Attribute("id", IsNotEmpty()),
				),
			},
		})
	})

	t.Run("throws error when plugin not found", func(t *testing.T) {
		config := `
			data "spacelift_plugin" "test" {
				plugin_id = "this-plugin-does-not-exist-123234lkj23"
			}
		`

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("plugin not found"),
			},
		})
	})
}
