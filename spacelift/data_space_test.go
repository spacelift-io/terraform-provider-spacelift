package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSpaceData(t *testing.T) {
	t.Run("creates and reads a space", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_space" "test" {
					name = "My first space %s"
					inherit_entities = true
					parent_space_id = "root"
					description = "some valid description"
				}

				data "spacelift_space" "test" {
					policy_id = spacelift_policy.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_space.test",
				Attribute("id", StartsWith("my-first-space")),
				Attribute("parent_space_id", Equals("root")),
				Attribute("description", Equals("some valid description")),
			),
		}})
	})
}
