package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPolicyResource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates and updates a policy", func(t *testing.T) {
		config := func(message string) string {
			return fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
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
					"spacelift_policy.test",
					Attribute("id", StartsWith("my-first-policy")),
					Attribute("body", Contains("boom")),
					Attribute("type", Equals("PLAN")),
				),
			},
			{
				Config: config("bang"),
				Check:  Resource("spacelift_policy.test", Attribute("body", Contains("bang"))),
			},
		})
	})
}
