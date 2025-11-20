package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPluginTemplateResource(t *testing.T) {
	const resourceName = "spacelift_plugin_template.test"

	t.Run("creates plugin template without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		// Simple manifest for testing
		manifest := `
manifest_version: 1.0.0
name: Test Plugin
description: A test plugin template
author: "spacelift"
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
				description = "Test plugin template"
				manifest    = <<-EOT
%s
				EOT
				labels      = ["test", "terraform"]
			}
		`, randomID, manifest)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test plugin template")),
					Attribute("description", Equals("Test plugin template")),
					SetEquals("labels", "test", "terraform"),
					Attribute("id", IsNotEmpty()),
					Attribute("manifest", IsNotEmpty()),
					Attribute("is_global", Equals("false")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})

	t.Run("creates plugin template with parameters", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		manifest := `
manifest_version: 1.0.0
name: Test Plugin With Params
description: A test plugin with parameters
author: "spacelift"
version: 1.0.0
parameters:
  - id: api_key
    name: API Key
    type: string
    description: API key for authentication
    required: true
    sensitive: true
  - id: optional_param
    name: Optional Parameter
    type: string
    description: An optional parameter
    required: false
    default: "default_value"

hooks:
  after_plan:
    - run: echo "Running with params"
`

		config := fmt.Sprintf(`
			resource "spacelift_plugin_template" "test" {
				name        = "Provider test plugin template with params %s"
				description = "Test plugin template with parameters"
				manifest    = <<-EOT
%s
				EOT
				labels      = ["test"]
			}
		`, randomID, manifest)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider test plugin template with params")),
					Attribute("parameters.#", Equals("2")),
				),
			},
		})
	})
}

func TestPluginTemplateResourceMinimal(t *testing.T) {
	const resourceName = "spacelift_plugin_template.test"

	t.Run("creates minimal plugin template", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		manifest := `
manifest_version: 1.0.0
name: Minimal Test Plugin
description: Minimal plugin template
author: "spacelift"
version: 1.0.0

hooks:
  before_init:
    - run: echo "Minimal hook"
`

		config := fmt.Sprintf(`
			resource "spacelift_plugin_template" "test" {
				name     = "Provider minimal test plugin template %s"
				manifest = <<-EOT
%s
				EOT
			}
		`, randomID, manifest)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("name", StartsWith("Provider minimal test plugin template")),
					Attribute("id", IsNotEmpty()),
				),
			},
		})
	})
}

func TestPluginTemplateResourceMissingAttributes(t *testing.T) {
	const resourceName = "spacelift_plugin_template.test"

	t.Run("creates minimal plugin template", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		manifest := `
manifest_version: 1.0.0
name: Minimal Test Plugin
description: Minimal plugin template

hooks:
  before_init:
    - run: echo "Minimal hook"
`

		config := fmt.Sprintf(`
			resource "spacelift_plugin_template" "test" {
				name     = "Provider minimal missing attributes test plugin template %s"
				manifest = <<-EOT
%s
				EOT
			}
		`, randomID, manifest)

		testSteps(t, []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("could not create plugin template: invalid manifest: missing properties: 'version', 'author'"),
			},
		})
	})
}
