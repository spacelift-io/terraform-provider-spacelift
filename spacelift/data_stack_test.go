package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackData(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				administrative = true
				autodeploy     = true
				autoretry      = true
				branch         = "master"
				description    = "description"
				labels         = ["one", "two"]
				name           = "Test stack %s"
				project_root   = "root"
				repository     = "demo"
			}

			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
		Check: Resource(
			"data.spacelift_stack.test",
			Attribute("id", StartsWith("test-stack-")),
			Attribute("administrative", Equals("true")),
			Attribute("autodeploy", Equals("true")),
			Attribute("autoretry", Equals("true")),
			Attribute("branch", Equals("master")),
			Attribute("description", Equals("description")),
			SetEquals("labels", "one", "two"),
			Attribute("name", StartsWith("Test stack")),
			Attribute("project_root", Equals("root")),
			Attribute("repository", Equals("demo")),
		),
	}})
}
