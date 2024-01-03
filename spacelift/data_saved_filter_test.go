package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestFilterData(t *testing.T) {
	t.Run("creates and updates a filter", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_saved_filter" "test" {
					name = "My first filter %s"
					type = "stacks"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}
				data "spacelift_saved_filter" "test" {
					filter_id = spacelift_saved_filter.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_saved_filter.test",
				Attribute("id", IsNotEmpty()),
				Attribute("data", Contains("activeFilters")),
				Attribute("type", Equals("stacks")),
				Attribute("is_public", Equals("true")),
			),
		}})
	})
}
