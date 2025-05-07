package spacelift

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSavedFilterResource(t *testing.T) {
	t.Parallel()
	const resourceName = "spacelift_saved_filter.test"

	t.Run("creates and updates a filter", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(filterType string) string {
			return `
				resource "spacelift_saved_filter" "test" {
					name = "My first filter ` + randomID + `"
					type = "` + filterType + `"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}
			`
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("stacks"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("data", Contains("activeFilters")),
					Attribute("type", Equals("stacks")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("contexts"),
				Check: Resource(
					resourceName,
					Attribute("type", Equals("contexts")),
				),
			},
		})
	})

	t.Run("unexpected type", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: `
					resource "spacelift_saved_filter" "test" {
						name = "My first filter ` + randomID + `"
						type = "whatever"
						is_public = true
						data = jsonencode({
							"key": "activeFilters",
							"value": jsonencode({})
						})
					}
				`,
				ExpectError: regexp.MustCompile(`expected type to be one of \["stacks" "blueprints" "contexts" "webhooks"\], got whatever`),
			},
		})
	})
}
