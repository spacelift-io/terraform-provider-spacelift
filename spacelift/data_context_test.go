package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextData(t *testing.T) {
	t.Parallel()

	t.Run("retrieves context data without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					description = "description"
					labels      = ["one", "two"]
					after_apply = ["after_apply"]
					after_destroy = ["after_destroy"]
					after_init = ["after_init"]
					after_perform = ["after_perform"]
					after_plan = ["after_plan"]
					after_run = ["after_run"]
					before_apply = ["before_apply"]
					before_destroy = ["before_destroy"]
					before_init = ["before_init"]
					before_perform = ["before_perform"]
					before_plan = ["before_plan"]
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
				SetEquals("labels", "one", "two"),
				Attribute("after_apply.#", Equals("1")),
				Attribute("after_apply.0", Equals("after_apply")),
				Attribute("after_destroy.#", Equals("1")),
				Attribute("after_destroy.0", Equals("after_destroy")),
				Attribute("after_init.#", Equals("1")),
				Attribute("after_init.0", Equals("after_init")),
				Attribute("after_perform.#", Equals("1")),
				Attribute("after_perform.0", Equals("after_perform")),
				Attribute("after_plan.#", Equals("1")),
				Attribute("after_plan.0", Equals("after_plan")),
				Attribute("after_run.#", Equals("1")),
				Attribute("after_run.0", Equals("after_run")),
				Attribute("before_apply.#", Equals("1")),
				Attribute("before_apply.0", Equals("before_apply")),
				Attribute("before_destroy.#", Equals("1")),
				Attribute("before_destroy.0", Equals("before_destroy")),
				Attribute("before_init.#", Equals("1")),
				Attribute("before_init.0", Equals("before_init")),
				Attribute("before_perform.#", Equals("1")),
				Attribute("before_perform.0", Equals("before_perform")),
				Attribute("before_plan.#", Equals("1")),
				Attribute("before_plan.0", Equals("before_plan")),
			),
		}})
	})
}

func TestContextDataSpace(t *testing.T) {
	t.Parallel()
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
