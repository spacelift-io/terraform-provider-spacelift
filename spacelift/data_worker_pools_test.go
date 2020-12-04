package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWorkerPoolsData(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resourceName := "spacelift_worker_pool.test"
	datasourceName := "data.spacelift_worker_pools.test"

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My first worker pool %s"
			}

			data "spacelift_worker_pools" "test" {
				depends_on = [spacelift_worker_pool.test]
			}
		`, randomID), Check: resource.ComposeTestCheckFunc(
			Resource(datasourceName, Attribute("id", IsNotEmpty())),
			// TODO: Check for inclusion
			resource.TestCheckResourceAttrPair(datasourceName, "worker_pools.0.worker_pool_id", resourceName, "id"),
			resource.TestCheckResourceAttrPair(datasourceName, "worker_pools.0.name", resourceName, "name"),
			resource.TestCheckResourceAttrPair(datasourceName, "worker_pools.0.config", resourceName, "config"),
		),
	}})
}
