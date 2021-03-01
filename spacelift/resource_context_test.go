package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextResource(t *testing.T) {
	t.Parallel()

	const resourceName = "spacelift_context.test"

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

		const resourceName = "spacelift_context.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-context-")),
					Attribute("name", StartsWith("Provider test context")),
					Attribute("description", Equals("old description")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description"),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("new description")),
				),
			},
		})
	})
}
