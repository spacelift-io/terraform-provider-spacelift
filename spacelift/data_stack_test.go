package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackData(t *testing.T) {
	t.Run("with Terraform stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				administrative                  = true
				after_apply                     = ["ls -la", "rm -rf /"]
				after_destroy                   = ["echo 'after_destroy'"]
				after_init                      = ["terraform fmt -check", "tflint"]
				after_perform                   = ["echo 'after_perform'"]
				after_plan                      = ["echo 'after_plan'"]
				after_run                       = ["echo 'after_run'"]
				autodeploy                      = true
				autoretry                       = false
				before_apply                    = ["ls -la", "rm -rf /"]
				before_destroy                  = ["echo 'before_destroy'"]
				before_init                     = ["terraform fmt -check", "tflint"]
				before_perform                  = ["echo 'before_perform'"]
				before_plan                     = ["echo 'before_plan'"]
				branch                          = "master"
				description                     = "description"
				labels                          = ["one", "two"]
				name                            = "Test stack %s"
				project_root                    = "root"
				additional_project_globs 		= ["/bacon", "/bacon/eggs/*"]
				repository                      = "demo"
				runner_image                    = "custom_image:runner"
				terraform_workspace             = "bacon"
				terraform_smart_sanitization    = true
				terraform_external_state_access = true
				enable_well_known_secret_masking = true
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
				Attribute("after_run.#", Equals("1")),
				Attribute("after_run.0", Equals("echo 'after_run'")),
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
				SetEquals("additional_project_globs", "/bacon", "/bacon/eggs/*"),
				Attribute("repository", Equals("demo")),
				Attribute("runner_image", Equals("custom_image:runner")),
				Attribute("terraform_workspace", Equals("bacon")),
				Attribute("terraform_smart_sanitization", Equals("true")),
				Attribute("terraform_external_state_access", Equals("true")),
				Attribute("enable_well_known_secret_masking", Equals("true")),
			),
		}})
	})

	t.Run("with CloudFormation stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				cloudformation {
					entry_template_file = "main.yaml"
					region = "eu-central-1"
					template_bucket = "s3://bucket"
					stack_name = "maincf"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("cloudformation.0.entry_template_file", Equals("main.yaml")),
				Attribute("cloudformation.0.region", Equals("eu-central-1")),
				Attribute("cloudformation.0.template_bucket", Equals("s3://bucket")),
				Attribute("cloudformation.0.stack_name", Equals("maincf")),
			),
		}})
	})

	t.Run("with Kubernetes stack with no namespace", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				kubernetes {}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("kubernetes.0.namespace", Equals("")),
			),
		}})
	})

	t.Run("with Kubernetes stack with a namespace", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				kubernetes {
					namespace = "app-prod"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("kubernetes.0.namespace", Equals("app-prod")),
			),
		}})
	})

	t.Run("with Kubernetes stack with no kubectl version", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				kubernetes {}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("kubernetes.0.kubectl_version", IsNotEmpty()),
				Attribute("kubernetes.0.kubernetes_workflow_tool", Equals("KUBERNETES")),
			),
		}})
	})

	t.Run("with Kubernetes stack with a kubectl version", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				kubernetes {
					kubectl_version = "1.2.3"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("kubernetes.0.kubectl_version", Equals("1.2.3")),
				Attribute("kubernetes.0.kubernetes_workflow_tool", Equals("KUBERNETES")),
			),
		}})
	})

	t.Run("with Kubernetes stack with a kubernetes workflow tool", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				kubernetes {
					kubernetes_workflow_tool = "CUSTOM"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("kubernetes.0.kubernetes_workflow_tool", Equals("CUSTOM")),
			),
		}})
	})

	t.Run("with Pulumi stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				pulumi {
					login_url = "s3://bucket"
					stack_name = "mainpl"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("pulumi.0.login_url", Equals("s3://bucket")),
				Attribute("pulumi.0.stack_name", Equals("mainpl")),
			),
		}})
	})

	t.Run("with Ansible stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				ansible {
					playbook = "main.yml"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("ansible.0.playbook", Equals("main.yml")),
			),
		}})
	})

	t.Run("with terraform_workflow_tool defaulted", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch       = "master"
					name         = "Provider test stack workflow_tool default %s"
					project_root = "root"
					repository   = "demo"
				}
				data "spacelift_stack" "test" {
					stack_id = spacelift_stack.test.id
				}
			`, randomID),
				Check: Resource(
					"data.spacelift_stack.test",
					Attribute("terraform_workflow_tool", Equals("TERRAFORM_FOSS")),
				),
			},
		})
	})

	t.Run("with terraform_workflow_tool set", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch                  = "master"
					name                    = "Provider test stack workflow_tool default %s"
					project_root            = "root"
					repository              = "demo"
					terraform_workflow_tool = "CUSTOM"
				}
				data "spacelift_stack" "test" {
					stack_id = spacelift_stack.test.id
				}
			`, randomID),
				Check: Resource(
					"data.spacelift_stack.test",
					Attribute("terraform_workflow_tool", Equals("CUSTOM")),
					Attribute("enable_well_known_secret_masking", Equals("false")),
				),
			},
		})
	})

	t.Run("with Terragrunt stack (default tool)", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				terragrunt {
					terraform_version      = "1.5.7"
					terragrunt_version     = "0.55.15"
					use_run_all            = false
					use_smart_sanitization = true
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("terragrunt.0.terraform_version", Equals("1.5.7")),
				Attribute("terragrunt.0.terragrunt_version", Equals("0.55.15")),
				Attribute("terragrunt.0.use_run_all", Equals("false")),
				Attribute("terragrunt.0.use_smart_sanitization", Equals("true")),
				Attribute("terragrunt.0.tool", Equals("TERRAFORM_FOSS")),
			),
		}})
	})

	t.Run("with Terragrunt stack (TERRAFORM_FOSS)", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				terragrunt {
					terraform_version      = "1.5.7"
					terragrunt_version     = "0.55.15"
					use_run_all            = false
					use_smart_sanitization = true
					tool                   = "TERRAFORM_FOSS"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("terragrunt.0.terraform_version", Equals("1.5.7")),
				Attribute("terragrunt.0.terragrunt_version", Equals("0.55.15")),
				Attribute("terragrunt.0.use_run_all", Equals("false")),
				Attribute("terragrunt.0.use_smart_sanitization", Equals("true")),
				Attribute("terragrunt.0.tool", Equals("TERRAFORM_FOSS")),
			),
		}})
	})

	t.Run("with Terragrunt stack (OPEN_TOFU)", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				terragrunt {
					terraform_version      = "1.6.2"
					terragrunt_version     = "0.55.15"
					use_run_all            = false
					use_smart_sanitization = true
					tool                   = "OPEN_TOFU"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("terragrunt.0.terraform_version", Equals("1.6.2")),
				Attribute("terragrunt.0.terragrunt_version", Equals("0.55.15")),
				Attribute("terragrunt.0.use_run_all", Equals("false")),
				Attribute("terragrunt.0.use_smart_sanitization", Equals("true")),
				Attribute("terragrunt.0.tool", Equals("OPEN_TOFU")),
			),
		}})
	})

	t.Run("with Terragrunt stack (MANUALLY_PROVISIONED)", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				terragrunt {
					terragrunt_version     = "0.55.15"
					use_run_all            = false
					use_smart_sanitization = true
					tool                   = "MANUALLY_PROVISIONED"
				}
			}
			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("terragrunt.0.terragrunt_version", Equals("0.55.15")),
				Attribute("terragrunt.0.use_run_all", Equals("false")),
				Attribute("terragrunt.0.use_smart_sanitization", Equals("true")),
				Attribute("terragrunt.0.tool", Equals("MANUALLY_PROVISIONED")),
			),
		}})
	})
}

func TestStackDataSpace(t *testing.T) {
	t.Run("with Terraform stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				administrative               = true
				after_apply                  = ["ls -la", "rm -rf /"]
				after_destroy                = ["echo 'after_destroy'"]
				after_init                   = ["terraform fmt -check", "tflint"]
				after_perform                = ["echo 'after_perform'"]
				after_plan                   = ["echo 'after_plan'"]
				autodeploy                   = true
				autoretry                    = false
				before_apply                 = ["ls -la", "rm -rf /"]
				before_destroy               = ["echo 'before_destroy'"]
				before_init                  = ["terraform fmt -check", "tflint"]
				before_perform               = ["echo 'before_perform'"]
				before_plan                  = ["echo 'before_plan'"]
				branch                       = "master"
				description                  = "description"
				labels                       = ["one", "two"]
				name                         = "Test stack %s"
				project_root                 = "root"
				repository                   = "demo"
				runner_image                 = "custom_image:runner"
				space_id                     = "root"
				terraform_workspace          = "bacon"
				terraform_smart_sanitization = true
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
				Attribute("space_id", Equals("root")),
				Attribute("runner_image", Equals("custom_image:runner")),
				Attribute("terraform_workspace", Equals("bacon")),
				Attribute("terraform_smart_sanitization", Equals("true")),
			),
		}})
	})

	t.Run("with CloudFormation stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				cloudformation {
					entry_template_file = "main.yaml"
					region = "eu-central-1"
					template_bucket = "s3://bucket"
					stack_name = "maincf"
				}
			}

			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("cloudformation.0.entry_template_file", Equals("main.yaml")),
				Attribute("cloudformation.0.region", Equals("eu-central-1")),
				Attribute("cloudformation.0.template_bucket", Equals("s3://bucket")),
				Attribute("cloudformation.0.stack_name", Equals("maincf")),
			),
		}})
	})

	t.Run("with Kubernetes stack with no namespace", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				kubernetes {}
			}

			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("kubernetes.0.namespace", Equals("")),
			),
		}})
	})

	t.Run("with Kubernetes stack with a namespace", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				kubernetes {
					namespace = "app-prod"
				}
			}

			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("kubernetes.0.namespace", Equals("app-prod")),
			),
		}})
	})

	t.Run("with Pulumi stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				pulumi {
					login_url = "s3://bucket"
					stack_name = "mainpl"
				}
			}

			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("pulumi.0.login_url", Equals("s3://bucket")),
				Attribute("pulumi.0.stack_name", Equals("mainpl")),
			),
		}})
	})

	t.Run("with Ansible stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch              = "master"
				name                = "Test stack %s"
				repository          = "demo"
				ansible {
					playbook = "main.yml"
				}
			}

			data "spacelift_stack" "test" {
				stack_id = spacelift_stack.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_stack.test",
				Attribute("ansible.0.playbook", Equals("main.yml")),
			),
		}})
	})
}
