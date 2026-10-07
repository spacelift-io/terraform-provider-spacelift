package spacelift

import "fmt"

// dummyWorkerPoolConfig returns a worker pool with no workers, so a run on a
// stack that uses it stays queued. Leave spaceID empty to use the default space.
func dummyWorkerPoolConfig(nameSuffix, spaceID string) string {
	spaceIDAttribute := ""
	if spaceID != "" {
		spaceIDAttribute = fmt.Sprintf("space_id = %q", spaceID)
	}

	return fmt.Sprintf(`
		resource "spacelift_worker_pool" "test" {
			name = "Let's create a dummy worker pool to avoid running the job %s"
			%s
		}
	`, nameSuffix, spaceIDAttribute)
}
