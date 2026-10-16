package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSpacesData(t *testing.T) {
	t.Parallel()

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

func TestComputeSpacePaths(t *testing.T) {
	rootParent := "root"
	parentA := "space-a"
	parentB := "space-b"

	spaces := []structs.Space{
		{
			ID:          "root",
			Name:        "root",
			ParentSpace: nil,
		},
		{
			ID:          "space-a",
			Name:        "Engineering",
			ParentSpace: &rootParent,
		},
		{
			ID:          "space-b",
			Name:        "Backend",
			ParentSpace: &parentA,
		},
		{
			ID:          "space-c",
			Name:        "Payments",
			ParentSpace: &parentB,
		},
		{
			ID:          "space-orphan",
			Name:        "Orphan",
			ParentSpace: nil,
		},
	}

	paths := computeSpacePaths(spaces)

	expected := map[string]string{
		"root":         "root",
		"space-a":      "root/Engineering",
		"space-b":      "root/Engineering/Backend",
		"space-c":      "root/Engineering/Backend/Payments",
		"space-orphan": "Orphan",
	}

	for id, want := range expected {
		if got := paths[id]; got != want {
			t.Errorf("expected path for space %s to be %q, got %q", id, want, got)
		}
	}
}
