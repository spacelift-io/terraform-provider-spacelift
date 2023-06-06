package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackResource(t *testing.T) {
	const resourceName = "spacelift_stack.test"

	t.Run("with GitHub and no state import", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string, protectFromDeletion bool) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative        = true
					after_apply           = ["ls -la", "rm -rf /"]
					after_destroy         = ["echo 'after_destroy'"]
					after_init            = ["terraform fmt -check", "tflint"]
					after_perform         = ["echo 'after_perform'"]
					after_plan            = ["echo 'after_plan'"]
					after_run             = ["echo 'after_run'"]
					autodeploy            = true
					autoretry             = false
					before_apply          = ["ls -la", "rm -rf /"]
					before_destroy        = ["echo 'before_destroy'"]
					before_init           = ["terraform fmt -check", "tflint"]
					before_perform        = ["echo 'before_perform'"]
					before_plan           = ["echo 'before_plan'"]
					branch                = "master"
					description           = "%s"
					import_state          = "{}"
					labels                = ["one", "two"]
					name                  = "Provider test stack %s"
					project_root          = "root"
					protect_from_deletion = %t
					repository            = "demo"
					runner_image          = "custom_image:runner"
				}
			`, description, randomID, protectFromDeletion)
		}

		const resourceName = "spacelift_stack.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description", true),
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
					Attribute("after_run.#", Equals("1")),
					Attribute("after_run.0", Equals("echo 'after_run'")),
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
					Attribute("github_action_deploy", Equals("true")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("protect_from_deletion", Equals("true")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"import_state"},
			},
			{
				Config: config("new description", false),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("new description")),
					Attribute("protect_from_deletion", Equals("false")),
				),
			},
		})
	})

	t.Run("with private worker pool, custom slug and autoretry", func(t *testing.T) {
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
					github_action_deploy = false
					labels               = ["one", "two"]
					name                 = "Provider test stack %s"
					project_root         = "root"
					repository           = "demo"
					runner_image         = "custom_image:runner"
					slug                 = "custom-slug-%s"
					worker_pool_id       = spacelift_worker_pool.test.id
				}
				resource "spacelift_worker_pool" "test" {
					name        = "Autoretryable worker pool."
					description = "test worker pool"
				}
			`, description, randomID, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("custom-slug-")),
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
					Attribute("github_action_deploy", Equals("false")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"slug"},
			},
			{
				Config: config("new description"),
				Check:  Resource(resourceName, Attribute("description", Equals("new description"))),
			},
		})
	})

	t.Run("unsetting fields", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		before := fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				administrative               = true
				after_apply                  = ["ls -la", "rm -rf /"]
				after_destroy                = ["echo 'after_destroy'"]
				after_init                   = ["terraform fmt -check", "tflint"]
				after_perform                = ["echo 'after_perform'"]
				after_plan                   = ["echo 'after_plan'"]
				autodeploy                   = true
				before_apply                 = ["ls -la", "rm -rf /"]
				before_destroy               = ["echo 'before_destroy'"]
				before_init                  = ["terraform fmt -check", "tflint"]
				before_perform               = ["echo 'before_perform'"]
				before_plan                  = ["echo 'before_plan'"]
				branch                       = "master"
				description                  = "bacon"
				labels                       = ["one", "two"]
				name                         = "Provider test stack %s"
				project_root                 = "root"
				repository                   = "demo"
				runner_image                 = "custom_image:runner"
				terraform_version            = "1.0.1"
				terraform_workspace          = "bacon"
				terraform_smart_sanitization = true
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
					Attribute("terraform_version", Equals("1.0.1")),
					Attribute("terraform_workspace", Equals("bacon")),
					Attribute("terraform_smart_sanitization", Equals("true")),
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
					Attribute("terraform_version", Equals("1.0.1")),
					Attribute("terraform_workspace", IsEmpty()),
					Attribute("terraform_smart_sanitization", Equals("false")),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_stack" "test" {
					name                  = "Provider test stack %s"
					branch                = "master"
					labels                = ["one", "two"]
					repository            = "demo"
				}`, randomID),
				Check: Resource(
					"spacelift_stack.test",
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_stack" "test" {
					name                  = "labelled-module-%s"
					branch                = "master"
					labels                = []
					repository            = "terraform-bacon-tasty"
				}`, randomID),
				Check: Resource(
					"spacelift_stack.test",
					SetEquals("labels"),
				),
			},
		})
	})

	t.Run("when error returned, it is explained properly", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				administrative               = true
				branch                       = "master"
				description                  = "bacon"
				name                         = "Provider test stack %s"
				project_root                 = "root"
				repository                   = "demo"
				terraform_version            = "1.0.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1"
			}
		`, randomID),
			ExpectError: regexp.MustCompile(`could not create stack: stack has 1 error: terraform: invalid Terraform version constraints: Could not parse Range`),
		}})
	})

	t.Run("external state access", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: `resource "spacelift_stack" "test" {
						name                            = "External state access test"
						project_root                    = "root"
						repository                      = "demo"
						branch                          = "master"
						administrative                  = false
						manage_state                    = true
					 }`,
				Check: Resource(
					"spacelift_stack.test",
					Attribute("terraform_external_state_access", Equals("false")),
				),
			},
			{
				Config: `resource "spacelift_stack" "test" {
						name                            = "External state access test"
						project_root                    = "root"
						repository                      = "demo"
						branch                          = "master"
						administrative                  = false
						manage_state                    = true
						terraform_external_state_access = true
					 }`,
				Check: Resource(
					"spacelift_stack.test",
					Attribute("terraform_external_state_access", Equals("true")),
				),
			},
			{
				Config: `resource "spacelift_stack" "test" {
						name                            = "External state access test"
						project_root                    = "root"
						repository                      = "demo"
						branch                          = "master"
						administrative                  = false
						manage_state                    = true
						terraform_external_state_access = true
					 }`,
				Check: Resource(
					"spacelift_stack.test",
					Attribute("terraform_external_state_access", Equals("true")),
				),
			},
			{
				Config: `resource "spacelift_stack" "test" {
						name                            = "External state access test"
						project_root                    = "root"
						repository                      = "demo"
						branch                          = "master"
						administrative                  = false
						manage_state                    = false
						terraform_external_state_access = true
					 }`,
				ExpectError: regexp.MustCompile(`"terraform_external_state_access" requires "manage_state" to be true`),
			},
		})
	})

	t.Run("with GitHub and Pulumi configuration", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: getConfig(`pulumi {
						login_url = "s3://bucket"
						stack_name = "mainpl"
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("pulumi.0.login_url", Equals("s3://bucket")),
					Attribute("pulumi.0.stack_name", Equals("mainpl")),
					Attribute("ansible.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
					Attribute("terragrunt.#", Equals("0")),
				),
			},
		})
	})

	t.Run("with raw Git link", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: getConfig(`raw_git {
						namespace = "bacon"
						url = "https://github.com/spacelift-io/onboarding.git"
				}`),
				Check: Resource(
					resourceName,
					Attribute("raw_git.#", Equals("1")),
					Attribute("raw_git.0.namespace", Equals("bacon")),
					Attribute("raw_git.0.url", Equals("https://github.com/spacelift-io/onboarding.git")),
				),
			},
		})
	})

	t.Run("with GitHub and CloudFormation configuration", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: getConfig(`cloudformation {
						entry_template_file = "main.yaml"
						region = "eu-central-1"
						template_bucket = "s3://bucket"
						stack_name = "maincf"
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("cloudformation.0.entry_template_file", Equals("main.yaml")),
					Attribute("cloudformation.0.region", Equals("eu-central-1")),
					Attribute("cloudformation.0.template_bucket", Equals("s3://bucket")),
					Attribute("cloudformation.0.stack_name", Equals("maincf")),
					Attribute("ansible.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
					Attribute("terragrunt.#", Equals("0")),
				),
			},
		})
	})

	t.Run("with GitHub and Kubernetes configuration", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: getConfig(`kubernetes {}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("kubernetes.0.namespace", Equals("")),
					Attribute("kubernetes.0.kubectl_version", Equals("1.23.5")),
					Attribute("ansible.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
					Attribute("terragrunt.#", Equals("0")),
				),
			},
			{
				Config: getConfig(`kubernetes {
						namespace = "myapp-prod"
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("kubernetes.0.namespace", Equals("myapp-prod")),
					Attribute("kubernetes.0.kubectl_version", Equals("1.23.5")),
					Attribute("ansible.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
				),
			},
			{
				Config: getConfig(`kubernetes {
						kubectl_version = "1.2.3"
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("kubernetes.0.namespace", Equals("")),
					Attribute("kubernetes.0.kubectl_version", Equals("1.2.3")),
					Attribute("ansible.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
				),
			},
		})
	})

	t.Run("with GitHub and Ansible configuration", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: getConfig(`ansible {
						playbook = "main.yml"
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("ansible.0.playbook", Equals("main.yml")),
					Attribute("cloudformation.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("terragrunt.#", Equals("0")),
				),
			},
		})
	})

	t.Run("with GitHub and Terragrunt configuration", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: getConfig(`terragrunt {
						terragrunt_version = "0.45.0"
						terraform_version = "1.4.0"
						use_run_all = false
						use_smart_sanitization = true
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),

					Attribute("terragrunt.0.terragrunt_version", Equals("0.45.0")),
					Attribute("terragrunt.0.terraform_version", Equals("1.4.0")),
					Attribute("terragrunt.0.use_run_all", Equals("false")),
					Attribute("terragrunt.0.use_smart_sanitization", Equals("true")),

					Attribute("cloudformation.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("ansible.#", Equals("0")),
				),
			},
		})
	})

	t.Run("with GitHub and no vendor-specific configuration", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: getConfig(``),
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
				Config: getConfig(``),
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
					Attribute("ansible.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
				),
			},
		})
	})

}

func TestStackResourceSpace(t *testing.T) {
	const resourceName = "spacelift_stack.test"

	t.Run("with GitHub and no state import", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string, protectFromDeletion bool) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative        = true
					after_apply           = ["ls -la", "rm -rf /"]
					after_destroy         = ["echo 'after_destroy'"]
					after_init            = ["terraform fmt -check", "tflint"]
					after_perform         = ["echo 'after_perform'"]
					after_plan            = ["echo 'after_plan'"]
					autodeploy            = true
					autoretry             = false
					before_apply          = ["ls -la", "rm -rf /"]
					before_destroy        = ["echo 'before_destroy'"]
					before_init           = ["terraform fmt -check", "tflint"]
					before_perform        = ["echo 'before_perform'"]
					before_plan           = ["echo 'before_plan'"]
					branch                = "master"
					description           = "%s"
					import_state          = "{}"
					labels                = ["one", "two"]
					name                  = "Provider test stack %s"
					project_root          = "root"
					protect_from_deletion = %t
					repository            = "demo"
					runner_image          = "custom_image:runner"
					space_id              = "root"
				}
			`, description, randomID, protectFromDeletion)
		}

		const resourceName = "spacelift_stack.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description", true),
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
					Attribute("github_action_deploy", Equals("true")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("protect_from_deletion", Equals("true")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
					Attribute("space_id", Equals("root")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"import_state"},
			},
			{
				Config: config("new description", false),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("new description")),
					Attribute("protect_from_deletion", Equals("false")),
				),
			},
		})
	})

	t.Run("with private worker pool, custom slug and autoretry", func(t *testing.T) {
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
					github_action_deploy = false
					labels               = ["one", "two"]
					name                 = "Provider test stack %s"
					project_root         = "root"
					repository           = "demo"
					runner_image         = "custom_image:runner"
					slug                 = "custom-slug-%s"
					worker_pool_id       = spacelift_worker_pool.test.id
				}

				resource "spacelift_worker_pool" "test" {
					name        = "Autoretryable worker pool."
					description = "test worker pool"
				}
			`, description, randomID, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("custom-slug-")),
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
					Attribute("github_action_deploy", Equals("false")),
					SetEquals("labels", "one", "two"),
					Attribute("name", StartsWith("Provider test stack")),
					Attribute("project_root", Equals("root")),
					Attribute("repository", Equals("demo")),
					Attribute("runner_image", Equals("custom_image:runner")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"slug"},
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
					Attribute("pulumi.0.login_url", Equals("s3://bucket")),
					Attribute("pulumi.0.stack_name", Equals("mainpl")),
					Attribute("ansible.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
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
					Attribute("cloudformation.0.entry_template_file", Equals("main.yaml")),
					Attribute("cloudformation.0.region", Equals("eu-central-1")),
					Attribute("cloudformation.0.template_bucket", Equals("s3://bucket")),
					Attribute("cloudformation.0.stack_name", Equals("maincf")),
					Attribute("ansible.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
				),
			},
			{
				Config: config(`kubernetes {}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("kubernetes.0.namespace", Equals("")),
					Attribute("ansible.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
				),
			},
			{
				Config: config(`kubernetes {
						namespace = "myapp-prod"
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("kubernetes.0.namespace", Equals("myapp-prod")),
					Attribute("ansible.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
				),
			},
			{
				Config: config(`ansible {
						playbook = "main.yml"
					}`),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-stack")),
					Attribute("ansible.0.playbook", Equals("main.yml")),
					Attribute("cloudformation.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
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
					Attribute("ansible.#", Equals("0")),
					Attribute("cloudformation.#", Equals("0")),
					Attribute("pulumi.#", Equals("0")),
					Attribute("kubernetes.#", Equals("0")),
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

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_stack" "test" {
					name                  = "Provider test stack %s"
					branch                = "master"
					labels                = ["one", "two"]
					repository            = "demo"
				}`, randomID),
				Check: Resource(
					"spacelift_stack.test",
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_stack" "test" {
					name                  = "labelled-module-%s"
					branch                = "master"
					labels                = []
					repository            = "terraform-bacon-tasty"
				}`, randomID),
				Check: Resource(
					"spacelift_stack.test",
					SetEquals("labels"),
				),
			},
		})
	})

	t.Run("importing non-existent resource", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		stackID := fmt.Sprintf("non-existent-stack-%s", resourceName)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch                = "master"
					name                  = "Provider test stack %s"
					project_root          = "root"
					repository            = "demo"
				}
			`, randomID),
				ImportState:   true,
				ResourceName:  "spacelift_stack.test",
				ImportStateId: stackID,
				ExpectError:   regexp.MustCompile(fmt.Sprintf(`stack with ID %q does not exist \(or you may not have access to it\)`, stackID)),
			},
		})
	})
}

// getConfig returns a stack config with injected vendor config
func getConfig(vendorConfig string) string {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
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
