package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextData(t *testing.T) {
	t.Run("retrieves context data without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					description = "description"
					labels      = ["one", "two"]
					space_id    = "root"
				}

				data "spacelift_context" "test" {
					context_id = spacelift_context.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_context.test",
				Attribute("id", StartsWith("provider-test-context-")),
				Attribute("name", StartsWith("Provider test context")),
				Attribute("description", Equals("description")),
				Attribute("space_id", Equals("root")),
				SetEquals("labels", "one", "two"),
			),
		}})
	})
}
