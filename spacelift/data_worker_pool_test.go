package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWorkerPoolData(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My first worker pool %s"
			}

			data "spacelift_worker_pool" "test" {
				worker_pool_id = spacelift_worker_pool.test.id
			}
		`, randomID),
		Check: Resource(
			"data.spacelift_worker_pool.test",
			Attribute("id", IsNotEmpty()),
			Attribute("config", IsNotEmpty()),
			Attribute("name", StartsWith("My first worker pool")),
		),
	}})
}
