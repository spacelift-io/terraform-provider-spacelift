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
		datasourceName := "data.spacelift_spaces.test"

		testSteps(t, []resource.TestStep{{
			// Should find at least root space.
			Config: `
				data "spacelift_spaces" "test" {
				}
			`,
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
			),
		}})
	})

	t.Run("filter by labels", func(t *testing.T) {
		datasourceName := "data.spacelift_spaces.test"
		randomSuffix := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			// Should find at least root space.
			Config: fmt.Sprintf(`
				resource "spacelift_space" "a" {
					name   = "space-a-%s"
					labels = ["one", "%s"]
				}

				resource "spacelift_space" "b" {
					name   = "space-b-%s"
					labels = ["%s"]
				}

				resource "spacelift_space" "c" {
					name   = "space-c-%s"
					labels = ["three", "four"]
				}

				data "spacelift_spaces" "test" {
					labels = ["%s"]

					depends_on = [spacelift_space.a, spacelift_space.b, spacelift_space.c]
				}
			`, randomSuffix, randomSuffix, randomSuffix, randomSuffix, randomSuffix, randomSuffix),
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
				Resource(datasourceName, Attribute("spaces.#", Equals("2"))),
			),
		}})
	})
}
