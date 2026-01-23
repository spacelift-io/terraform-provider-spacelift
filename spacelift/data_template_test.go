package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestTemplateData(t *testing.T) {
	t.Run("creates and reads a template", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_template" "test" {
					name        = "test-template-%s"
					space       = "root"
					description = "test description"
					labels      = ["label1", "label2"]
				}

				data "spacelift_template" "test" {
					template_id = spacelift_template.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_template.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", Equals("test-template-"+randomID)),
				Attribute("space", Equals("root")),
				Attribute("description", Equals("test description")),
				SetEquals("labels", "label1", "label2"),
				Attribute("ulid", IsNotEmpty()),
				Attribute("created_at", IsNotEmpty()),
				Attribute("updated_at", IsNotEmpty()),
			),
		}})
	})

	t.Run("reads a template without optional fields", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_template" "test" {
					name   = "test-template-%s"
					space  = "root"
					labels = []
				}

				data "spacelift_template" "test" {
					template_id = spacelift_template.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_template.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", Equals("test-template-"+randomID)),
				Attribute("space", Equals("root")),
				Attribute("description", Equals("")),
				Attribute("labels.#", Equals("0")),
			),
		}})
	})
}
