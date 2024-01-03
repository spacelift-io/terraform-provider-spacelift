package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSavedFilterResource(t *testing.T) {
	const resourceName = "spacelift_saved_filter.test"

	t.Run("creates and updates a filter", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(filterType string) string {
			return fmt.Sprintf(`
				resource "spacelift_saved_filter" "test" {
					name = "My first filter %s"					
					type = "%s"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}
			`, randomID, filterType)
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
}
