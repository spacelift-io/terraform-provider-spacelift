package spacelift

import (
	"testing"

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

		testSteps(t, []resource.TestStep{{
			// Should find at least root space.
			Config: `
				resource "spacelift_space" "a" {
					name   = "space-a"
					labels = ["one", "two"]
				}

				resource "spacelift_space" "b" {
					name   = "space-b"
					labels = ["two"]
				}

				resource "spacelift_space" "c" {
					name   = "space-c"
					labels = ["three", "four"]
				}

				data "spacelift_spaces" "test" {
					labels = ["two"]

					depends_on = [spacelift_space.a, spacelift_space.b, spacelift_space.c]
				}
			`,
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
				Resource(datasourceName, Attribute("spaces.#", Equals("2"))),
			),
		}})
	})
}
