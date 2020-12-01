package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackResource(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with GitHub and no state import", func(t *testing.T) {
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative = true
					autodeploy     = true
					autoretry      = false
					before_init    = ["terraform fmt -check", "tflint"]
					branch         = "master"
					description    = "%s"
					labels         = ["one", "two"]
					name           = "Provider test stack %s"
					project_root   = "root"
					repository     = "demo"
					runner_image   = "custom_image:runner"
				}
			`, description, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					"spacelift_stack.test",
					Attribute("id", StartsWith("provider-test-stack-")),
					Attribute("administrative", Equals("true")),
					Attribute("autodeploy", Equals("true")),
					Attribute("autoretry", Equals("false")),
					SetEquals("before_init", "terraform fmt -check", "tflint"),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				Config: config("new description"),
				Check:  Resource("spacelift_stack.test", Attribute("description", Equals("new description"))),
			},
		})
	})

	t.Run("with private worker pool and autoretry", func(t *testing.T) {
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative = true
					autodeploy     = true
					autoretry      = true
					before_init    = ["terraform fmt -check", "tflint"]
					branch         = "master"
					description    = "%s"
					labels         = ["one", "two"]
					name           = "Provider test stack %s"
					project_root   = "root"
					repository     = "demo"
					runner_image   = "custom_image:runner"
					worker_pool_id = spacelift_worker_pool.test.id
				}

				resource "spacelift_worker_pool" "test" {
					name        = "Autoretryable worker pool."
					description = "test worker pool"
				}
			`, description, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					"spacelift_stack.test",
					Attribute("id", StartsWith("provider-test-stack-")),
					Attribute("administrative", Equals("true")),
					Attribute("autodeploy", Equals("true")),
					Attribute("autoretry", Equals("true")),
					SetEquals("before_init", "terraform fmt -check", "tflint"),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				Config: config("new description"),
				Check:  Resource("spacelift_stack.test", Attribute("description", Equals("new description"))),
			},
		})
	})

	t.Run("with GitHub and vendor-specific configuration", func(t *testing.T) {
		config := func(vendorConfig string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative = true
					autodeploy     = true
					autoretry      = false
					before_init    = ["terraform fmt -check", "tflint"]
					branch         = "master"
					labels         = ["one", "two"]
					name           = "Provider test stack"
					project_root   = "root"
					repository     = "demo"
					runner_image   = "custom_image:runner"
					%s
				}
			`, vendorConfig)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(``),
				Check: Resource(
					"spacelift_stack.test",
					Attribute("id", Equals("provider-test-stack")),
					Attribute("administrative", Equals("true")),
					Attribute("autodeploy", Equals("true")),
					Attribute("autoretry", Equals("false")),
					SetEquals("before_init", "terraform fmt -check", "tflint"),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				Config: config(`pulumi {
						login_url = "s3://bucket"
						stack_name = "mainpl"
					}`),
				Check: Resource(
					"spacelift_stack.test",
					Attribute("id", Equals("provider-test-stack")),
					Attribute("administrative", Equals("true")),
					Attribute("autodeploy", Equals("true")),
					Attribute("autoretry", Equals("false")),
					SetEquals("before_init", "terraform fmt -check", "tflint"),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
					Attribute("pulumi[0].login_url", Equals("s3://bucket")),
					Attribute("pulumi[0].stack_name", Equals("mainpl")),
					Attribute("cloudformation[0]", Equals("null")),
				),
			},
			{
				Config: config(`cloudformation {
						entry_template_file = "main.yaml"
						region = "eu-central-1"
						template_bucket = "s3://bucket"
						stack_name = "maincf"
					}`),
				Check: Resource(
					"spacelift_stack.test",
					Attribute("id", Equals("provider-test-stack")),
					Attribute("administrative", Equals("true")),
					Attribute("autodeploy", Equals("true")),
					Attribute("autoretry", Equals("false")),
					SetEquals("before_init", "terraform fmt -check", "tflint"),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
					Attribute("cloudformation[0].entry_template_file", Equals("main.yaml")),
					Attribute("cloudformation[0].region", Equals("eu-central-1")),
					Attribute("cloudformation[0].template_bucket", Equals("s3://bucket")),
					Attribute("cloudformation[0].stack_name", Equals("maincf")),
					Attribute("pulumi[0]", Equals("null")),
				),
			},
			{
				Config: config(``),
				Check: Resource(
					"spacelift_stack.test",
					Attribute("id", Equals("provider-test-stack")),
					Attribute("administrative", Equals("true")),
					Attribute("autodeploy", Equals("true")),
					Attribute("autoretry", Equals("false")),
					SetEquals("before_init", "terraform fmt -check", "tflint"),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
		})
	})
}
