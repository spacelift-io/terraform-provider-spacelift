package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextResource(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates and updates contexts without an error", func(t *testing.T) {
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					description = "%s"
				}
			`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					"spacelift_context.test",
					Attribute("id", StartsWith("provider-test-context-")),
					Attribute("name", StartsWith("Provider test context")),
					Attribute("description", Equals("old description")),
				),
			},
			{
				Config: config("new description"),
				Check: Resource(
					"spacelift_context.test",
					Attribute("description", Equals("new description")),
				),
			},
		})
	})
}
