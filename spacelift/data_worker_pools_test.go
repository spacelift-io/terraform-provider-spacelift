package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWorkerPoolsData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resourceName := "spacelift_worker_pool.test"
	datasourceName := "data.spacelift_worker_pools.test"

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My first worker pool %s"
				space_id = "root"
			}

			data "spacelift_worker_pools" "test" {
				depends_on = [spacelift_worker_pool.test]
			}
		`, randomID), Check: resource.ComposeTestCheckFunc(
			Resource(datasourceName, Attribute("id", IsNotEmpty())),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"worker_pools", "worker_pool_id"}, resourceName, "id"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"worker_pools", "name"}, resourceName, "name"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"worker_pools", "config"}, resourceName, "config"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"worker_pools", "space_id"}, resourceName, "space_id"),
		),
	}})
}
