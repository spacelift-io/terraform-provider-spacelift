package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestTaskResource(t *testing.T) {
	t.Parallel()

	t.Run("on a new stack", func(t *testing.T) {
		const resourceName = "spacelift_task.test"

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		randomIDwp := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name        = "Let's create a dummy worker pool to avoid running the job %s"
				}

				resource "spacelift_stack" "test" {
					name           = "Test stack %s"
					repository     = "demo"
					branch         = "master"
					worker_pool_id = spacelift_worker_pool.test.id
				}

				resource "spacelift_task" "test" {
					stack_id = spacelift_stack.test.id
					command = "ls -lah"
					keepers = { "bacon" = "tasty" }
				}
			`, randomIDwp, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
				),
			},
		})
	})
}

func TestTaskResourceWait(t *testing.T) {

	t.Run("on a new stack", func(t *testing.T) {
		const resourceName = "spacelift_task.test"

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		randomIDwp := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_worker_pool" "test" {
						name        = "Let's create a dummy worker pool to avoid running the job %s"
					}

					resource "spacelift_stack" "test" {
						name           = "Test stack %s"
						repository     = "demo"
						branch         = "master"
						worker_pool_id = spacelift_worker_pool.test.id
					}

					resource "spacelift_task" "test" {
						stack_id = spacelift_stack.test.id
						command = "ls -lah"
						keepers = { "bacon" = "tasty" }

						timeouts {
							create = "10s"
						}

						wait {
							disabled            = false
							continue_on_timeout = true
						}
					}`, randomIDwp, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
				),
			},
		})
	})

	t.Run("timed out run", func(t *testing.T) {
		const resourceName = "spacelift_task.test"

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		randomIDwp := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_worker_pool" "test" {
						name        = "Let's create a dummy worker pool to avoid running the job %s"
					}

					resource "spacelift_stack" "test" {
						name           = "Test stack %s"
						repository     = "demo"
						branch         = "master"
						worker_pool_id = spacelift_worker_pool.test.id
					}

					resource "spacelift_task" "test" {
						stack_id = spacelift_stack.test.id
						command = "ls -lah"
						keepers = { "bacon" = "tasty" }

						timeouts {
							create = "10s"
						}

						wait {
							disabled            = false
							continue_on_timeout = false
						}
					}`, randomIDwp, randomID),
				ExpectError: regexp.MustCompile("run [0-9A-Z]* on stack test-stack-[a-z0-9]* has timed out"),
			},
		})
	})

	t.Run("finished with autodeploy", func(t *testing.T) {
		const resourceName = "spacelift_task.test"

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_stack" "test" {
						name           = "Test stack %s"
						repository     = "demo"
						branch         = "master"
						autodeploy     = true
					}

					resource "spacelift_task" "test" {
						stack_id = spacelift_stack.test.id
						command = "ls -lah"
						keepers = { "bacon" = "tasty" }

						timeouts {
							create = "180s"
						}

						wait {
							disabled            = false
						}
					}`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
				),
			},
		})
	})
}
