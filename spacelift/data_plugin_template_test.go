package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPluginTemplateData(t *testing.T) {
	const resourceName = "data.spacelift_plugin_template.test"

	t.Run("reads plugin template data without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		manifest := `
manifest_version: 1.0.0
name: Test Plugin
description: A test plugin template
author: "test"
version: 1.0.0
parameters:
  - id: test_param
    name: Test Parameter
    type: string
    description: A test parameter
    required: true

hooks:
  before_plan:
    - run: echo "Test hook"
`

		config := fmt.Sprintf(`
			resource "spacelift_plugin_template" "test" {
				name        = "Provider test plugin template %s"
				description = "Test plugin template data source"
				manifest    = <<-EOT
%s
				EOT
				labels      = ["test", "data-source"]
			}

			data "spacelift_plugin_template" "test" {
				plugin_template_id = spacelift_plugin_template.test.id
			}
		`, randomID, manifest)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test plugin template")),
					Attribute("description", Equals("Test plugin template data source")),
					SetEquals("labels", "test", "data-source"),
					Attribute("id", IsNotEmpty()),
					Attribute("manifest", IsNotEmpty()),
					Attribute("is_global", Equals("false")),
				),
			},
		})
	})

	t.Run("throws error when plugin template not found", func(t *testing.T) {
		config := `
			data "spacelift_plugin_template" "test" {
				plugin_template_id = "this-plugin-does-not-exist-123234lkj23"
			}
		`

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("plugin template not found"),
			},
		})
	})
}
