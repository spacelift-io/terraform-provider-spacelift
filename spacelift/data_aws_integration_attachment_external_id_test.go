package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAWSIntegrationAttachmentExternalIDData(t *testing.T) {
	t.Parallel()
	const resourceName = "data.spacelift_aws_integration_attachment_external_id.test"

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch     = "master"
				repository = "demo"
				name       = "Test stack %s"
			}

			resource "spacelift_aws_integration" "test" {
                name                           = "test-aws-integration-%s"
                generate_credentials_in_worker = false
                role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
			}

			data "spacelift_aws_integration_attachment_external_id" "test" {
				stack_id       = spacelift_stack.test.id
				integration_id = spacelift_aws_integration.test.id
				read           = true
				write          = true
			}
			`, randomID, randomID),
			Check: Resource(
				resourceName,
				Attribute("id", IsNotEmpty()),
				AttributeNotPresent("module_id"),
				Attribute("stack_id", IsNotEmpty()),
				Attribute("read", Equals("true")),
				Attribute("write", Equals("true")),
				Attribute("external_id", IsNotEmpty()),
				Attribute("assume_role_policy_statement", IsNotEmpty()),
			),
		}})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name       = "test-module-%s"
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}

				resource "spacelift_aws_integration" "test" {
				    name                           = "test-aws-integration-%s"
				    generate_credentials_in_worker = false
				    role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
				}

				data "spacelift_aws_integration_attachment_external_id" "test" {
					module_id      = spacelift_module.test.id
					integration_id = spacelift_aws_integration.test.id
					read           = true
					write          = true
				}
			`, randomID, randomID),
			Check: Resource(
				resourceName,
				Attribute("id", IsNotEmpty()),
				Attribute("module_id", IsNotEmpty()),
				AttributeNotPresent("stack_id"),
				Attribute("read", Equals("true")),
				Attribute("write", Equals("true")),
				Attribute("external_id", IsNotEmpty()),
				Attribute("assume_role_policy_statement", IsNotEmpty()),
			),
		}})
	})

	t.Run("read and write not set", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_stack" "test" {
				branch     = "master"
				repository = "demo"
				name       = "Test stack %s"
			}

			resource "spacelift_aws_integration" "test" {
                name                           = "test-aws-integration-%s"
                generate_credentials_in_worker = false
                role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
			}

			data "spacelift_aws_integration_attachment_external_id" "test" {
				stack_id       = spacelift_stack.test.id
				integration_id = spacelift_aws_integration.test.id
				read = false
				write = false
			}
			`, randomID, randomID),
			ExpectError: regexp.MustCompile(`at least one of either 'read' or 'write' must be true`),
		}})
	})
}
