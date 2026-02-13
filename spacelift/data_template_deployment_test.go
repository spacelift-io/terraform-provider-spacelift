package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestTemplateDeploymentData(t *testing.T) {
	const datasourceName = "data.spacelift_template_deployment.test"

	t.Run("creates and reads a template deployment", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_template" "test" {
					name  = "test-template-%[1]s"
					space = "root"
				}

				resource "spacelift_template_version" "test" {
					template_id    = spacelift_template.test.id
					version_number = "1.0.0"
					state          = "PUBLISHED"
					template       = <<-EOT
stacks:
- name: test-stack-$${{ inputs.env_name }}-%[1]s
  key: test
  autodeploy: true
  vcs:
    reference:
      value: master
      type: branch
    repository: terraform-bacon-tasty
    provider: GITHUB
  vendor:
    terraform:
      manage_state: true
      version: "1.5.0"
inputs:
- id: env_name
  name: Environment Name
  default: dev
- id: secret
  name: secret
  type: secret
EOT
				}

				resource "spacelift_template_deployment" "test" {
					template_version_id = spacelift_template_version.test.id
					space               = "root"
					name                = "test-deployment-%[1]s"
					description         = "test description"

					input {
						id    = "env_name"
						value = "production"
					}

					input {
						id    = "secret"
						value = "secret_value"
					}
				}

				data "spacelift_template_deployment" "test" {
					template_id   = spacelift_template.test.id
					deployment_id = spacelift_template_deployment.test.deployment_id
				}
			`, randomID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet(datasourceName, "id"),
				resource.TestCheckResourceAttr(datasourceName, "template_id", "test-template-"+randomID),
				resource.TestCheckResourceAttr(datasourceName, "deployment_id", "test-deployment-"+randomID),
				resource.TestCheckResourceAttr(datasourceName, "name", "test-deployment-"+randomID),
				resource.TestCheckResourceAttr(datasourceName, "space", "root"),
				resource.TestCheckResourceAttr(datasourceName, "description", "test description"),
				resource.TestCheckResourceAttrSet(datasourceName, "state"),
				resource.TestCheckResourceAttrSet(datasourceName, "created_at"),
				resource.TestCheckResourceAttrSet(datasourceName, "template_version_id"),
				resource.TestCheckResourceAttr(datasourceName, "input.#", "2"),
				resource.TestCheckResourceAttr(datasourceName, "input.0.id", "env_name"),
				resource.TestCheckResourceAttr(datasourceName, "input.0.value", "production"),
				resource.TestCheckResourceAttr(datasourceName, "input.0.secret", "false"),
				resource.TestCheckResourceAttr(datasourceName, "input.1.id", "secret"),
				resource.TestCheckResourceAttr(datasourceName, "input.1.secret", "true"),
				resource.TestCheckResourceAttr(datasourceName, "input.1.checksum", "28dd40f834818f2e63827ddd50a1d50198b8e5233b9e21956ccedccc1be8a35e"),
				resource.TestCheckResourceAttr(datasourceName, "stacks.#", "1"),
				resource.TestCheckResourceAttr(datasourceName, "stacks.0.id", "test-stack-production-"+randomID),
			),
		}})
	})
}
