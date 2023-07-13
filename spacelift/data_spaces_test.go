package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSpacesData(t *testing.T) {
	t.Run("load all spaces", func(t *testing.T) {
		datasourceName := "data.spacelift_spaces.test"

		testSteps(t, []resource.TestStep{{
			// Should find at least the legacy and root spaces.
			Config: fmt.Sprintf(`
				data "spacelift_spaces" "test" {
				}
			`),
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
			),
		}})
	})
}
