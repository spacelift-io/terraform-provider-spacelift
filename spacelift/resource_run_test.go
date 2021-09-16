package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestRunResource(t *testing.T) {

	t.Run("on a new stack", func(t *testing.T) {
		const resourceName = "spacelift_run.test"

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name        = "Let's create a dummy worker pool to avoid running the job"
				}

				resource "spacelift_stack" "test" {
					name           = "Test stack %s"
					repository     = "demo"
					branch         = "master"
					worker_pool_id = spacelift_worker_pool.test.id
				}

				resource "spacelift_run" "test" {
					stack_id = spacelift_stack.test.id

					keepers = { "bacon" = "tasty" }
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
				),
			},
		})
	})
}
