package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPolicyData(t *testing.T) {
	t.Run("creates and updates a policy", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					labels = ["one", "two"]
					description = "My awesome policy"
					body = <<EOF
					package spacelift
					deny contains "boom" if { true }
					EOF
					type = "PLAN"
					engine_type = "REGO_V1"
				}
				data "spacelift_policy" "test" {
					policy_id = spacelift_policy.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_policy.test",
				Attribute("id", StartsWith("my-first-policy")),
				Attribute("body", Contains("boom")),
				Attribute("type", Equals("PLAN")),
				SetEquals("labels", "one", "two"),
				Attribute("description", Equals("My awesome policy")),
				Attribute("engine_type", Equals("REGO_V1")),
			),
		}})
	})
}

func TestPolicyDataSpace(t *testing.T) {
	t.Run("creates and updates a policy", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					description = "My awesome policy"
					labels = ["one", "two"]
					space_id = "root"
					body = <<EOF
					package spacelift

					deny["boom"] { true }
					EOF
					type = "PLAN"
					engine_type = "REGO_V0"
				}

				data "spacelift_policy" "test" {
					policy_id = spacelift_policy.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_policy.test",
				Attribute("id", StartsWith("my-first-policy")),
				Attribute("body", Contains("boom")),
				Attribute("type", Equals("PLAN")),
				Attribute("space_id", Equals("root")),
				SetEquals("labels", "one", "two"),
				Attribute("description", Equals("My awesome policy")),
				Attribute("engine_type", Equals("REGO_V0")),
			),
		}})
	})
}
