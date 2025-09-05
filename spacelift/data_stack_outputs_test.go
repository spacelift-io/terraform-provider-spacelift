package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackOutputsData(t *testing.T) {
	t.Run("retrieves stack outputs metadata", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				name         = "Test stack %s"
				branch       = "master"
				repository   = "demo"
			}

			data "spacelift_stack_outputs" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack_outputs.test",
				Attribute("id", StartsWith("test-stack-")),
				Attribute("stack_id", StartsWith("test-stack-")),
				Attribute("outputs.#", Equals("0")),
			),
		}})
	})
}
