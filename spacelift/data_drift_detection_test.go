package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestDriftDetectionData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with a stack", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_drift_detection" "test" {
					stack_id     = spacelift_stack.test.id
					reconcile    = true
					schedule     = ["*/3 * * * *", "*/5 * * * *"]
				}

				data "spacelift_drift_detection" "test" {
					stack_id = spacelift_drift_detection.test.stack_id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_drift_detection.test",
				Attribute("id", IsNotEmpty()),
				Attribute("stack_id", IsNotEmpty()),
				Attribute("schedule.#", Equals("2")),
				Attribute("schedule.0", Equals("*/3 * * * *")),
				Attribute("schedule.1", Equals("*/5 * * * *")),
			),
		}})
	})
}
