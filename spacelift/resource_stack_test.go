package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackResource(t *testing.T) {
	const resourceName = "spacelift_stack.test"

	t.Parallel()

	t.Run("with GitHub and no state import", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative = true
					after_apply    = ["ls -la", "rm -rf /"]
					after_destroy  = ["echo 'after_destroy'"]
					after_init     = ["terraform fmt -check", "tflint"]
					after_perform  = ["echo 'after_perform'"]
					after_plan     = ["echo 'after_plan'"]
					autodeploy     = true
					autoretry      = false
					before_apply   = ["ls -la", "rm -rf /"]
					before_destroy = ["echo 'before_destroy'"]
					before_init    = ["terraform fmt -check", "tflint"]
					before_perform = ["echo 'before_perform'"]
					before_plan    = ["echo 'before_plan'"]
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

		const resourceName = "spacelift_stack.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack-")),
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
					Attribute("administrative", Equals("true")),
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
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description"),
				Check:  Resource(resourceName, Attribute("description", Equals("new description"))),
			},
		})
	})

	t.Run("with private worker pool and autoretry", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative       = true
					after_apply          = ["ls -la", "rm -rf /"]
					after_destroy        = ["echo 'after_destroy'"]
					after_init           = ["terraform fmt -check", "tflint"]
					after_perform        = ["echo 'after_perform'"]
					after_plan           = ["echo 'after_plan'"]
					autodeploy           = true
					autoretry            = true
					before_apply         = ["ls -la", "rm -rf /"]
					before_destroy       = ["echo 'before_destroy'"]
					before_init          = ["terraform fmt -check", "tflint"]
					before_perform       = ["echo 'before_perform'"]
					before_plan          = ["echo 'before_plan'"]
					branch               = "master"
					description          = "%s"
					enable_local_preview = true
					labels               = ["one", "two"]
					name                 = "Provider test stack %s"
					project_root         = "root"
					repository           = "demo"
					runner_image         = "custom_image:runner"
					worker_pool_id       = spacelift_worker_pool.test.id
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
					resourceName,
					Attribute("id", StartsWith("provider-test-stack-")),
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
					Attribute("autoretry", Equals("true")),
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
					Attribute("description", Equals("old description")),
					Attribute("enable_local_preview", Equals("true")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description"),
				Check:  Resource(resourceName, Attribute("description", Equals("new description"))),
			},
		})
	})

	t.Run("with GitHub and vendor-specific configuration", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(vendorConfig string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative = true
					after_apply    = ["ls -la", "rm -rf /"]
					after_destroy  = ["echo 'after_destroy'"]
					after_init     = ["terraform fmt -check", "tflint"]
					after_perform  = ["echo 'after_perform'"]
					after_plan     = ["echo 'after_plan'"]
					autodeploy     = true
					autoretry      = false
					before_apply   = ["ls -la", "rm -rf /"]
					before_destroy = ["echo 'before_destroy'"]
					before_init    = ["terraform fmt -check", "tflint"]
					before_perform = ["echo 'before_perform'"]
					before_plan    = ["echo 'before_plan'"]
					branch         = "master"
					labels         = ["one", "two"]
					name           = "Provider test stack %s"
					project_root   = "root"
					repository     = "demo"
					runner_image   = "custom_image:runner"
					%s
				}
			`, randomID, vendorConfig)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(``),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
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
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config(`pulumi {
						login_url = "s3://bucket"
						stack_name = "mainpl"
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
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
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
					Attribute("pulumi.0.login_url", Equals("s3://bucket")),
					Attribute("pulumi.0.stack_name", Equals("mainpl")),
					Attribute("cloudformation.#", Equals("0")),
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
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
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
					Attribute("administrative", Equals("true")),
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
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
					Attribute("cloudformation.0.entry_template_file", Equals("main.yaml")),
					Attribute("cloudformation.0.region", Equals("eu-central-1")),
					Attribute("cloudformation.0.template_bucket", Equals("s3://bucket")),
					Attribute("cloudformation.0.stack_name", Equals("maincf")),
					Attribute("pulumi.#", Equals("0")),
				),
			},
			{
				Config: config(``),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
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
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
					Attribute("cloudformation.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
				),
			},
		})
	})

	t.Run("unsetting fields", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		before := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				administrative      = true
				after_apply         = ["ls -la", "rm -rf /"]
				after_destroy       = ["echo 'after_destroy'"]
				after_init          = ["terraform fmt -check", "tflint"]
				after_perform       = ["echo 'after_perform'"]
				after_plan          = ["echo 'after_plan'"]
				autodeploy          = true
				before_apply        = ["ls -la", "rm -rf /"]
				before_destroy      = ["echo 'before_destroy'"]
				before_init         = ["terraform fmt -check", "tflint"]
				before_perform      = ["echo 'before_perform'"]
				before_plan         = ["echo 'before_plan'"]
				branch              = "master"
				description         = "bacon"
				labels              = ["one", "two"]
				name                = "Provider test stack %s"
				project_root        = "root"
				repository          = "demo"
				runner_image        = "custom_image:runner"
				terraform_version   = "0.12.5"
				terraform_workspace = "bacon"
			}
		`, randomID)

		after := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch         = "master"
				name           = "Provider test stack %s"
				repository     = "demo"
			}
		`, randomID)

		testSteps(t, []resource.TestStep{
			{
				Config: before,
				Check: Resource(
					resourceName,
					Attribute("administrative", Equals("true")),
					Attribute("after_apply.#", Equals("2")),
					Attribute("after_destroy.#", Equals("1")),
					Attribute("after_init.#", Equals("2")),
					Attribute("after_perform.#", Equals("1")),
					Attribute("after_plan.#", Equals("1")),
					Attribute("autodeploy", Equals("true")),
					Attribute("before_apply.#", Equals("2")),
					Attribute("before_destroy.#", Equals("1")),
					Attribute("before_init.#", Equals("2")),
					Attribute("before_perform.#", Equals("1")),
					Attribute("before_plan.#", Equals("1")),
					Attribute("description", Equals("bacon")),
					SetEquals("labels", "one", "two"),
					Attribute("project_root", Equals("root")),
					Attribute("runner_image", Equals("custom_image:runner")),
					Attribute("terraform_version", Equals("0.12.5")),
					Attribute("terraform_workspace", Equals("bacon")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: after,
				Check: Resource(
					resourceName,
					Attribute("administrative", Equals("false")),
					Attribute("after_apply.#", Equals("0")),
					Attribute("after_destroy.#", Equals("0")),
					Attribute("after_init.#", Equals("0")),
					Attribute("after_perform.#", Equals("0")),
					Attribute("after_plan.#", Equals("0")),
					Attribute("autodeploy", Equals("false")),
					Attribute("before_apply.#", Equals("0")),
					Attribute("before_destroy.#", Equals("0")),
					Attribute("before_init.#", Equals("0")),
					Attribute("before_perform.#", Equals("0")),
					Attribute("before_plan.#", Equals("0")),
					Attribute("description", IsEmpty()),
					Attribute("labels.#", Equals("0")),
					Attribute("project_root", IsEmpty()),
					Attribute("runner_image", IsEmpty()),
					Attribute("terraform_version", Equals("0.12.5")),
					Attribute("terraform_workspace", IsEmpty()),
				),
			},
		})
	})
}
