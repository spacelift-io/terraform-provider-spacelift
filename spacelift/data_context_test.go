package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextData(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("retrieves context data without an error", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			Providers: map[string]terraform.ResourceProvider{
				"spacelift": testProvider(),
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
						resource "spacelift_context" "test" {
							name        = "Provider test context %s"
							description = "description"
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
					),
				},
			},
		})
	})
}
