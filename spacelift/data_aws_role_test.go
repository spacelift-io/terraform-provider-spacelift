package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAWSRoleData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with a stack", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_stack" "test" {
						branch     = "master"
						repository = "demo"
						name       = "Test stack %s"
					}

					resource "spacelift_aws_role" "test" {
						stack_id = spacelift_stack.test.id
						role_arn = "arn:aws:iam::039653571618:role/empty-test-role"
					}

					data "spacelift_aws_role" "test" {
						stack_id = spacelift_aws_role.test.stack_id
					}
				`, randomID),
				Check: Resource(
					"data.spacelift_aws_role.test",
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
					Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
					Attribute("generate_credentials_in_worker", Equals("false")),
					AttributeNotPresent("module_id"),
				),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				resource "spacelift_module" "test" {
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}
				resource "spacelift_aws_role" "test" {
					module_id = spacelift_module.test.id
					role_arn  = "arn:aws:iam::039653571618:role/empty-test-role"
				}

				data "spacelift_aws_role" "test" {
					module_id = spacelift_aws_role.test.module_id
				}
			`,
			Check: Resource(
				"data.spacelift_aws_role.test",
				Attribute("id", IsNotEmpty()),
				Attribute("module_id", Equals("terraform-bacon-tasty")),
				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
				Attribute("generate_credentials_in_worker", Equals("false")),
				AttributeNotPresent("stack_id"),
			),
		}})
	})

	t.Run("with generating AWS creds in the worker for stack", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_stack" "test" {
						branch     = "master"
						repository = "demo"
						name       = "Test stack %s"
					}

					resource "spacelift_aws_role" "test" {
						stack_id                       = spacelift_stack.test.id
						role_arn                       = "custom_role_arn"
						generate_credentials_in_worker = true
					}

					data "spacelift_aws_role" "test" {
						stack_id = spacelift_aws_role.test.stack_id
					}
				`, randomID),
				Check: Resource(
					"data.spacelift_aws_role.test",
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Contains(randomID)),
					Attribute("role_arn", Equals("custom_role_arn")),
					Attribute("generate_credentials_in_worker", Equals("true")),
					AttributeNotPresent("module_id"),
				),
			},
		})
	})

	t.Run("with generating AWS creds in the worker for module", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				resource "spacelift_module" "test" {
					branch     = "master"
					repository = "terraform-bacon-tasty"
				}
				resource "spacelift_aws_role" "test" {
					module_id                      = spacelift_module.test.id
					role_arn                       = "custom_role_arn"
					generate_credentials_in_worker = true
				}

				data "spacelift_aws_role" "test" {
					module_id = spacelift_aws_role.test.module_id
				}
			`,
			Check: Resource(
				"data.spacelift_aws_role.test",
				Attribute("id", IsNotEmpty()),
				Attribute("module_id", Equals("terraform-bacon-tasty")),
				Attribute("role_arn", Equals("custom_role_arn")),
				Attribute("generate_credentials_in_worker", Equals("true")),
				AttributeNotPresent("stack_id"),
			),
		}})
	})
}
