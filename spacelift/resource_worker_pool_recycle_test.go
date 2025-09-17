package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWorkerPoolRecycleResource(t *testing.T) {
	t.Run("recycles a worker pool", func(t *testing.T) {
		const resourceName = "spacelift_worker_pool_recycle.test"

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name = "Worker pool for recycle test %s"
				}

				resource "spacelift_worker_pool_recycle" "test" {
					worker_pool_id = spacelift_worker_pool.test.id
					keepers = { "trigger" = "now" }
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("worker_pool_id", IsNotEmpty()),
				),
			},
		})
	})
}