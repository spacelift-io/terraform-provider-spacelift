package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackData(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				administrative      = true
				autodeploy          = true
				autoretry           = false
				before_init         = ["terraform fmt -check", "tflint"]
				before_apply        = ["ls -la", "rm -rf /"]
				branch              = "master"
				description         = "description"
				labels              = ["one", "two"]
				name                = "Test stack %s"
				project_root        = "root"
				repository          = "demo"
				runner_image        = "custom_image:runner"
				terraform_workspace = "bacon"
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
			Attribute("autoretry", Equals("false")),
			Attribute("before_init.#", Equals("2")),
			Attribute("before_init.0", Equals("terraform fmt -check")),
			Attribute("before_init.1", Equals("tflint")),
			Attribute("before_apply.#", Equals("2")),
			Attribute("before_apply.0", Equals("ls -la")),
			Attribute("before_apply.1", Equals("rm -rf /")),
			Attribute("branch", Equals("master")),
			Attribute("description", Equals("description")),
			SetEquals("labels", "one", "two"),
			Attribute("name", StartsWith("Test stack")),
			Attribute("project_root", Equals("root")),
			Attribute("repository", Equals("demo")),
			Attribute("runner_image", Equals("custom_image:runner")),
			Attribute("terraform_workspace", Equals("bacon")),
		),
	}})
}
