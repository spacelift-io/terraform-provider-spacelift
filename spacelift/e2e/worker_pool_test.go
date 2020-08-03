package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type WorkerPoolTest struct {
	ResourceTest
}

func (e *WorkerPoolTest) TestLifecycle_OK() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($description:String!$name:String!){workerPoolCreate(name: $name, description: $description){config,workerPool{id,name,description}}}","variables":{"description":"bar","name":"foo"}}`,
		`{"data":{"workerPoolCreate":{"config":"secret","workerPool":{"id":"babys-first-worker-pool","name":"foo","description":"bar"}}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){workerPool(id: $id){id,name,description}}","variables":{"id":"babys-first-worker-pool"}}`,
		`{"data":{"workerPool":{"id":"babys-first-pool","name":"foo","description":"bar"}}}`,
		6,
	)

	e.posts(
		`{"query":"mutation($id:ID!){workerPoolDelete(id: $id){id,name,description}}","variables":{"id":"babys-first-worker-pool"}}`,
		`{"data":{"workerPoolDelete":{"id":"babys-first-worker-pool"}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_worker_pool" "worker_pool" {
  name = "foo"
  description = "bar"
}

data "spacelift_worker_pool" "worker_pool" {
  worker_pool_id = spacelift_worker_pool.worker_pool.id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_worker_pool.worker_pool", "id", "babys-first-worker-pool"),
				resource.TestCheckResourceAttr("spacelift_worker_pool.worker_pool", "name", "foo"),
				resource.TestCheckResourceAttr("spacelift_worker_pool.worker_pool", "description", "bar"),
				resource.TestCheckResourceAttr("spacelift_worker_pool.worker_pool", "config", "secret"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_worker_pool.worker_pool", "id", "babys-first-worker-pool"),
				resource.TestCheckResourceAttr("data.spacelift_worker_pool.worker_pool", "name", "foo"),
				resource.TestCheckResourceAttr("data.spacelift_worker_pool.worker_pool", "description", "bar"),
			),
		},
	})
}

func TestWorkerPool(t *testing.T) {
	suite.Run(t, new(WorkerPoolTest))
}
