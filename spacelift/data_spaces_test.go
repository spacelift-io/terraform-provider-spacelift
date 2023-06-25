package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSpacesData(t *testing.T) {
	t.Run("load all spaces", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		datasourceName := "data.spacelift_spaces.test"

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_space" "test" {
					name = "%s"
					parent_space_id = "root"
					description = "This is the test space %s."
				}

				data "spacelift_spaces" "test" {
					depends_on = [spacelift_space.test]
				}
			`, randomID, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
			),
		}})
	})
}
