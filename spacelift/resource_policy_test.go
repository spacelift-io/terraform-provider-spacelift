package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

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
					type = "TERRAFORM_PLAN"
				}
			`, randomID, message)
		}

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			Providers: map[string]terraform.ResourceProvider{
				"spacelift": testProvider(),
			},
			Steps: []resource.TestStep{
				{
					Config: config("boom"),
					Check: Resource(
						"spacelift_policy.test",
						Attribute("id", StartsWith("my-first-policy")),
						Attribute("body", Contains("boom")),
						Attribute("type", Equals("TERRAFORM_PLAN")),
					),
				},
				{
					Config: config("bang"),
					Check:  Resource("spacelift_policy.test", Attribute("body", Contains("bang"))),
				},
			},
		})
	})
}
