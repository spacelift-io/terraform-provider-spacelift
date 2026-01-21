package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestTemplateResource(t *testing.T) {
	const resourceName = "spacelift_template.test"

	t.Run("Creates and updates a template", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_template" "test" {
					name        = "test-template-%s"
					space       = "root"
					description = "%s"
					labels      = ["one", "two"]
				}`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("test description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals("test-template-"+randomID)),
					Attribute("space", Equals("root")),
					Attribute("description", Equals("test description")),
					SetEquals("labels", "one", "two"),
					Attribute("ulid", IsNotEmpty()),
					Attribute("created_at", IsNotEmpty()),
					Attribute("updated_at", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("updated description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals("test-template-"+randomID)),
					Attribute("space", Equals("root")),
					Attribute("description", Equals("updated description")),
					SetEquals("labels", "one", "two"),
				),
			},
		})
	})

	t.Run("Creates a template without optional fields", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "spacelift_template" "test" {
				name   = "test-template-%s"
				space  = "root"
			}`, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: config,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals("test-template-"+randomID)),
					Attribute("space", Equals("root")),
					Attribute("description", Equals("")),
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
						name   = "test-template-%s"
						space  = "root"
						labels = ["one", "two"]
					}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "spacelift_template" "test" {
						name   = "test-template-%s"
						space  = "root"
						labels = []
					}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels"),
				),
			},
		})
	})
}
