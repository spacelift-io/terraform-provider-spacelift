package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAWSIntegrationAttachmentResource(t *testing.T) {
	const resourceName = "spacelift_aws_integration_attachment.test"

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_stack" "test" {
						branch     = "master"
						repository = "demo"
						name       = "Test stack %s"
					}

					resource "spacelift_aws_integration" "test" {
						name                           = "test-aws-integration-%s"
						role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
						generate_credentials_in_worker = false
					}

					resource "spacelift_aws_integration_attachment" "test" {
						stack_id       = spacelift_stack.test.id
						integration_id = spacelift_aws_integration.test.id
					}
				`, randomID, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					AttributeNotPresent("module_id"),
					Attribute("stack_id", IsNotEmpty()),
					Attribute("read", Equals("true")),
					Attribute("write", Equals("true")),
					Attribute("attachment_id", IsNotEmpty()),
				),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_module" "test" {
            name       = "test-module-%s"
						branch     = "master"
						repository = "terraform-bacon-tasty"
					}

					resource "spacelift_aws_integration" "test" {
            name                           = "test-aws-integration-%s"
            role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
            generate_credentials_in_worker = false
					}

					resource "spacelift_aws_integration_attachment" "test" {
						module_id = spacelift_module.test.id
						integration_id = spacelift_aws_integration.test.id
					}
				`, randomID, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("module_id", IsNotEmpty()),
					AttributeNotPresent("stack_id"),
					Attribute("read", Equals("true")),
					Attribute("write", Equals("true")),
				),
			},
		})
	})

	t.Run("update attachment", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_stack" "test" {
						branch     = "master"
						repository = "demo"
						name       = "Test stack %s"
					}

					resource "spacelift_aws_integration" "test" {
						name                           = "test-aws-integration-%s"
						role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
						generate_credentials_in_worker = false
					}

					resource "spacelift_aws_integration_attachment" "test" {
						stack_id       = spacelift_stack.test.id
						integration_id = spacelift_aws_integration.test.id
						read           = true
						write          = false
					}
				`, randomID, randomID),
				Check: Resource(
					resourceName,
					Attribute("write", Equals("false")),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "spacelift_stack" "test" {
						branch     = "master"
						repository = "demo"
						name       = "Test stack %s"
					}

					resource "spacelift_aws_integration" "test" {
						name                           = "test-aws-integration-%s"
						role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
						generate_credentials_in_worker = false
					}

					resource "spacelift_aws_integration_attachment" "test" {
						stack_id       = spacelift_stack.test.id
						integration_id = spacelift_aws_integration.test.id
						read           = true
						write          = true
					}
				`, randomID, randomID),
				Check: Resource(
					resourceName,
					Attribute("write", Equals("true")),
				),
			},
		})
	})
}
