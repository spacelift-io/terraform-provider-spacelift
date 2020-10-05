package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPolicyData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates and updates a policy", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					body = <<EOF
					package spacelift

					deny["boom"] { true }
					EOF
					type = "TERRAFORM_PLAN"
				}

				data "spacelift_policy" "test" {
					policy_id = spacelift_policy.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_policy.test",
				Attribute("id", StartsWith("my-first-policy")),
				Attribute("body", Contains("boom")),
				Attribute("type", Equals("TERRAFORM_PLAN")),
			),
		}})
	})
}
