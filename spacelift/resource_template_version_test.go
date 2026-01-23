package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestTemplateVersionResource(t *testing.T) {
	const resourceName = "spacelift_template_version.test"

	t.Run("Creates and updates a template version in DRAFT state", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(instructions string) string {
			return fmt.Sprintf(`
				resource "spacelift_template" "test" {
					name  = "test-template-%s"
					space = "root"
				}

				resource "spacelift_template_version" "test" {
					template_id    = spacelift_template.test.id
					version_number = "1.0.0"
					state          = "DRAFT"
					instructions   = "%s"
					labels         = ["one", "two"]
					template       = "not validated for drafts"
				}`, randomID, instructions)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("test instructions"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("version_number", Equals("1.0.0")),
					Attribute("state", Equals("DRAFT")),
					Attribute("instructions", Equals("test instructions")),
					SetEquals("labels", "one", "two"),
					Attribute("template", Equals("not validated for drafts")),
					Attribute("ulid", IsNotEmpty()),
					Attribute("created_at", IsNotEmpty()),
					Attribute("updated_at", IsNotEmpty()),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_id"},
			},
			{
				Config: config("updated instructions"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("version_number", Equals("1.0.0")),
					Attribute("state", Equals("DRAFT")),
					Attribute("instructions", Equals("updated instructions")),
					SetEquals("labels", "one", "two"),
				),
			},
		})
	})

	t.Run("Creates a template version in PUBLISHED state", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		validTemplate := `stacks:\n- name: Blueprint v2 - test upgrade 3\n  key: test\n  vcs:\n    reference: \n      value: master\n      type: branch\n    repository: demo\n    provider: GITHUB\n  vendor:\n    terraform:\n      manage_state: true\n      version: \"1.3.0\"`

		config := fmt.Sprintf(`
			resource "spacelift_template" "test" {
				name  = "test-template-%s"
				space = "root"
			}

			resource "spacelift_template_version" "test" {
				template_id    = spacelift_template.test.id
				version_number = "1.0.0"
				state          = "PUBLISHED"
				instructions   = "test instructions"
				labels         = ["one", "two"]
				template       = "%s"
			}`, randomID, validTemplate)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("version_number", Equals("1.0.0")),
					Attribute("state", Equals("PUBLISHED")),
					Attribute("instructions", Equals("test instructions")),
					SetEquals("labels", "one", "two"),
					Attribute("template", IsNotEmpty()),
					Attribute("ulid", IsNotEmpty()),
					Attribute("published_at", IsNotEmpty()),
				),
			},
		})
	})

	t.Run("Creates a template version without optional fields", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_template" "test" {
				name  = "test-template-%s"
				space = "root"
			}

			resource "spacelift_template_version" "test" {
				template_id    = spacelift_template.test.id
				version_number = "1.0.0"
				state          = "DRAFT"
			}`, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("version_number", Equals("1.0.0")),
					Attribute("state", Equals("DRAFT")),
					Attribute("instructions", Equals("")),
					Attribute("ulid", IsNotEmpty()),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_template" "test" {
						name  = "test-template-%s"
						space = "root"
					}

					resource "spacelift_template_version" "test" {
						template_id    = spacelift_template.test.id
						version_number = "1.0.0"
						state          = "DRAFT"
						labels         = ["one", "two"]
					}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "spacelift_template" "test" {
						name  = "test-template-%s"
						space = "root"
					}

					resource "spacelift_template_version" "test" {
						template_id    = spacelift_template.test.id
						version_number = "1.0.0"
						state          = "DRAFT"
						labels         = []
					}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels"),
				),
			},
		})
	})
}
