package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				administrative      = true
				after_apply         = ["ls -la", "rm -rf /"]
				after_destroy       = ["echo 'after_destroy'"]
				after_init          = ["terraform fmt -check", "tflint"]
				after_perform       = ["echo 'after_perform'"]
				after_plan          = ["echo 'after_plan'"]
				autodeploy          = true
				autoretry           = false
				before_apply        = ["ls -la", "rm -rf /"]
				before_destroy      = ["echo 'before_destroy'"]
				before_init         = ["terraform fmt -check", "tflint"]
				before_perform      = ["echo 'before_perform'"]
				before_plan         = ["echo 'before_plan'"]
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
			Attribute("after_apply.#", Equals("2")),
			Attribute("after_apply.0", Equals("ls -la")),
			Attribute("after_apply.1", Equals("rm -rf /")),
			Attribute("after_destroy.#", Equals("1")),
			Attribute("after_destroy.0", Equals("echo 'after_destroy'")),
			Attribute("after_init.#", Equals("2")),
			Attribute("after_init.0", Equals("terraform fmt -check")),
			Attribute("after_init.1", Equals("tflint")),
			Attribute("after_perform.#", Equals("1")),
			Attribute("after_perform.0", Equals("echo 'after_perform'")),
			Attribute("after_plan.#", Equals("1")),
			Attribute("after_plan.0", Equals("echo 'after_plan'")),
			Attribute("autodeploy", Equals("true")),
			Attribute("autoretry", Equals("false")),
			Attribute("before_apply.#", Equals("2")),
			Attribute("before_apply.0", Equals("ls -la")),
			Attribute("before_apply.1", Equals("rm -rf /")),
			Attribute("before_destroy.#", Equals("1")),
			Attribute("before_destroy.0", Equals("echo 'before_destroy'")),
			Attribute("before_init.#", Equals("2")),
			Attribute("before_init.0", Equals("terraform fmt -check")),
			Attribute("before_init.1", Equals("tflint")),
			Attribute("before_perform.#", Equals("1")),
			Attribute("before_perform.0", Equals("echo 'before_perform'")),
			Attribute("before_plan.#", Equals("1")),
			Attribute("before_plan.0", Equals("echo 'before_plan'")),
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
