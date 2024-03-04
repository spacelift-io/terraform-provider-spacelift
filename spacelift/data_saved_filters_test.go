package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestFiltersData(t *testing.T) {
	t.Run("load all saved filters", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		datasourceName := "data.spacelift_saved_filters.all"

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_saved_filter" "test" {
					name = "%s"					
					type = "stacks"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				data "spacelift_saved_filters" "all" {
					depends_on = [spacelift_saved_filter.test]
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
			),
		}})
	})
}
