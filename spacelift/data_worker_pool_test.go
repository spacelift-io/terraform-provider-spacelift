package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWorkerPoolData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My first worker pool %s"
				labels = ["label1", "label2"]
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
			SetEquals("labels", "label1", "label2"),
		),
	}})
}

func TestWorkerPoolDataSpace(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My first worker pool %s"
				space_id = "root"
				labels = ["label1", "label2"]
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
			Attribute("space_id", Equals("root")),
			SetEquals("labels", "label1", "label2"),
		),
	}})
}

func TestWorkerPoolDataDriftDetection(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	resourceName := "spacelift_worker_pool.test"
	singleDataSourceName := "data.spacelift_worker_pool.test"
	listDataSourceName := "data.spacelift_worker_pools.test"

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name                      = "Worker pool with drift limit %s"
				drift_detection_run_limit = 7
				labels                    = ["test"]
			}

			data "spacelift_worker_pool" "test" {
				worker_pool_id = spacelift_worker_pool.test.id
			}

			data "spacelift_worker_pools" "test" {
				depends_on = [spacelift_worker_pool.test]
			}
		`, randomID),
		Check: resource.ComposeTestCheckFunc(
			// Test single data source
			Resource(
				singleDataSourceName,
				Attribute("id", IsNotEmpty()),
				Attribute("name", Equals(fmt.Sprintf("Worker pool with drift limit %s", randomID))),
				Attribute("drift_detection_run_limit", Equals("7")),
				SetEquals("labels", "test"),
			),
			// Test list data source
			Resource(listDataSourceName, Attribute("id", IsNotEmpty())),
			CheckIfResourceNestedAttributeContainsResourceAttribute(listDataSourceName, []string{"worker_pools", "worker_pool_id"}, resourceName, "id"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(listDataSourceName, []string{"worker_pools", "name"}, resourceName, "name"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(listDataSourceName, []string{"worker_pools", "drift_detection_run_limit"}, resourceName, "drift_detection_run_limit"),
		),
	}})
}
