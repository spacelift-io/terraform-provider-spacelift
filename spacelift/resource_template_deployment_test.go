package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestTemplateDeploymentResource(t *testing.T) {
	const deploymentResource = "spacelift_template_deployment.test"
	const stackDatasource = "data.spacelift_stack.test"

	t.Run("Creates and updates a template deployment", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		initialVersion := "1.0.0"
		newVersion := "1.0.1"

		config := func(name, description, version, envName, secret string) string {
			return fmt.Sprintf(`
				resource "spacelift_template" "test" {
					name  = "test-template-%[1]s"
					space = "root"
				}

				resource "spacelift_template_version" "first_version" {
					template_id    = spacelift_template.test.id
					version_number = "%[2]s"
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

				resource "spacelift_template_version" "new_version" {
					template_id    = spacelift_template.test.id
					version_number = "%[3]s"
					state          = "PUBLISHED"
					template       = <<-EOT
stacks:
- name: test-stack-newname-$${{ inputs.env_name }}-%[1]s
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
					template_version_id = spacelift_template_version.%[4]s.id
					space               = "root"
					name                = "%[5]s"
					description         = "%[6]s"

					input {
						id    = "env_name"
						value = "%[7]s"
					}

					input {
						id    = "secret"
						value = "%[8]s"
					}
				}

				data "spacelift_stack" "test" {
				  stack_id = spacelift_template_deployment.test.stacks[0].id
				}
			`, randomID, initialVersion, newVersion, version, name, description, envName, secret)
		}

		var firstTemplateVersionId string
		testSteps(t, []resource.TestStep{
			{
				Config: config("deployment-"+randomID, "test description", "first_version", "production", "foobar"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(deploymentResource, "id", fmt.Sprintf("test-template-%[1]s/deployment-%[1]s", randomID)),
					resource.TestCheckResourceAttr(deploymentResource, "name", "deployment-"+randomID),
					resource.TestCheckResourceAttr(deploymentResource, "space", "root"),
					resource.TestCheckResourceAttr(deploymentResource, "state", "FINISHED"),
					resource.TestCheckResourceAttr(deploymentResource, "description", "test description"),
					resource.TestCheckResourceAttr(deploymentResource, "template_id", "test-template-"+randomID),
					resource.TestCheckResourceAttr(deploymentResource, "deployment_id", "deployment-"+randomID),
					resource.TestCheckResourceAttrWith(deploymentResource, "template_version_id", func(value string) error {
						firstTemplateVersionId = value
						if !strings.HasPrefix(value, "test-template-"+randomID) {
							return fmt.Errorf("expected template version ID to start with %s, got %s", "test-template-"+randomID, value)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith(deploymentResource, "created_at", func(value string) error {
						if value == "" {
							return fmt.Errorf("expected created_at to be set, got empty string")
						}
						return nil
					}),

					resource.TestCheckResourceAttr(deploymentResource, "input.#", "2"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.id", "env_name"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.secret", "false"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.value", "production"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.checksum", "ab8e18ef4ebebeddc0b3152ce9c9006e14fc05242e3fc9ce32246ea6a9543074"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.id", "secret"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.secret", "true"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.value", "foobar"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.checksum", "c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2"),

					resource.TestCheckResourceAttr(deploymentResource, "stacks.#", "1"),
					resource.TestCheckResourceAttr(deploymentResource, "stacks.0.id", "test-stack-production-"+randomID),

					resource.TestCheckResourceAttr(stackDatasource, "name", "test-stack-production-"+randomID),
					resource.TestCheckResourceAttr(stackDatasource, "autodeploy", "true"),
				),
			},
			{
				// Let's update the description and ensure it is correctly updated
				// Also test that we can update the version and the inputs
				Config: config("deployment-"+randomID, "updated description", "new_version", "staging", "foobar"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(deploymentResource, "description", "updated description"),
					resource.TestCheckResourceAttr(deploymentResource, "state", "FINISHED"),
					resource.TestCheckResourceAttrWith(deploymentResource, "template_version_id", func(value string) error {
						if value == firstTemplateVersionId {
							return fmt.Errorf("expected template version ID to change, got %s", value)
						}
						return nil
					}),

					resource.TestCheckResourceAttr(deploymentResource, "input.#", "2"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.id", "env_name"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.secret", "false"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.value", "staging"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.checksum", "e919a75364398a449f860aeadddc57fa0502145a4e63959ddb33c417a48dc0da"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.id", "secret"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.secret", "true"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.value", "foobar"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.checksum", "c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2"),

					resource.TestCheckResourceAttr(deploymentResource, "stacks.#", "1"),
					resource.TestCheckResourceAttr(deploymentResource, "stacks.0.id", "test-stack-production-"+randomID),

					resource.TestCheckResourceAttr(stackDatasource, "name", "test-stack-newname-staging-"+randomID),
				),
			},
			{
				// Test update of a secret input
				Config: config("deployment-"+randomID, "updated description", "new_version", "staging", "barfoo"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(deploymentResource, "state", "FINISHED"),
					resource.TestCheckResourceAttr(deploymentResource, "input.#", "2"),
					// Test non-secret input is unchanged
					resource.TestCheckResourceAttr(deploymentResource, "input.0.id", "env_name"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.secret", "false"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.value", "staging"),
					resource.TestCheckResourceAttr(deploymentResource, "input.0.checksum", "e919a75364398a449f860aeadddc57fa0502145a4e63959ddb33c417a48dc0da"),
					// Test secret input is changed
					resource.TestCheckResourceAttr(deploymentResource, "input.1.id", "secret"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.secret", "true"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.value", "barfoo"),
					resource.TestCheckResourceAttr(deploymentResource, "input.1.checksum", "88ecde925da3c6f8ec3d140683da9d2a422f26c1ae1d9212da1e5a53416dcc88"),
				),
			},
		})
	})
}
