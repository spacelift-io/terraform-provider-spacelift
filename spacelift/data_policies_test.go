package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPoliciesData(t *testing.T) {
	t.Run("load all policies", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		datasourceName := "data.spacelift_policies.test"

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "%s"
					body = <<EOF
					package spacelift

					read { true }
					EOF
					type = "PLAN"
				}

				data "spacelift_policies" "test" {
					depends_on = [spacelift_policy.test]
				}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
			),
		}})
	})
}
