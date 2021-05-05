package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPolicyResource(t *testing.T) {
	const resourceName = "spacelift_policy.test"

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates and updates a policy", func(t *testing.T) {
		config := func(message string) string {
			return fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					labels = ["one", "two"]
					body = <<EOF
					package spacelift

					deny["%s"] { true }
					EOF
					type = "PLAN"
				}
			`, randomID, message)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("boom"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("my-first-policy")),
					Attribute("body", Contains("boom")),
					Attribute("type", Equals("PLAN")),
					SetEquals("labels", "one", "two"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("bang"),
				Check: Resource(
					resourceName,
					Attribute("body", Contains("bang")),
				),
			},
		})
	})
}
