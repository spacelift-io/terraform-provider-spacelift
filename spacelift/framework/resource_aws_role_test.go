package framework

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/framework/testhelpers"
)

func TestAWSRoleResource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	const resourceName = "spacelift_aws_role.test"

	config := func(roleARN string) string {
		return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_aws_role" "test" {
					stack_id = spacelift_stack.test.id
					role_arn = "%s"
				}
			`, randomID, roleARN)
	}

	compatibilityProviders := func() map[string]resource.ExternalProvider {
		return map[string]resource.ExternalProvider{
			"spacelift": {
				VersionConstraint: "1.1.7",
				Source:            "spacelift-io/spacelift",
			},
		}
	}

	protoV5ProviderFactories := func() map[string]func() (tfprotov5.ProviderServer, error) {
		return map[string]func() (tfprotov5.ProviderServer, error){
			"spacelift": providerserver.NewProtocol5WithError(New("test", "2")()),
		}
	}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: compatibilityProviders(),
				Config:            config("arn:aws:iam::039653571618:role/empty-test-role"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", testhelpers.IsNotEmpty()),
					resource.TestCheckResourceAttrWith(resourceName, "id", testhelpers.Contains(randomID)),
					resource.TestCheckResourceAttrWith(resourceName, "role_arn", testhelpers.Equals("arn:aws:iam::039653571618:role/empty-test-role")),
					resource.TestCheckResourceAttrWith(resourceName, "generate_credentials_in_worker", testhelpers.Equals("false")),
					resource.TestCheckResourceAttrWith(resourceName, "duration_seconds", testhelpers.IsNotEmpty()),
					resource.TestCheckResourceAttrWith(resourceName, "external_id", testhelpers.IsEmpty()),
					resource.TestCheckNoResourceAttr(resourceName, "module_id"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(),
				Config:                   config("arn:aws:iam::039653571618:role/empty-test-role"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", testhelpers.IsNotEmpty()),
					resource.TestCheckResourceAttrWith(resourceName, "id", testhelpers.Contains(randomID)),
					resource.TestCheckResourceAttrWith(resourceName, "role_arn", testhelpers.Equals("arn:aws:iam::039653571618:role/empty-test-role")),
					resource.TestCheckResourceAttrWith(resourceName, "generate_credentials_in_worker", testhelpers.Equals("false")),
					resource.TestCheckResourceAttrWith(resourceName, "duration_seconds", testhelpers.IsNotEmpty()),
					resource.TestCheckResourceAttrWith(resourceName, "external_id", testhelpers.IsEmpty()),
					resource.TestCheckNoResourceAttr(resourceName, "module_id"),
				),
			},
		},
	})

	//t.Run("with a stack", func(t *testing.T) {
	//
	//	testSteps(t, []resource.TestStep{
	//		{},
	//		{
	//			ResourceName:      resourceName,
	//			ImportState:       true,
	//			ImportStateId:     fmt.Sprintf("stack/test-stack-%s", randomID),
	//			ImportStateVerify: true,
	//		},
	//		{
	//			Config: config("arn:aws:iam::039653571618:role/another-empty-test-role"),
	//			Check: Resource(
	//				resourceName,
	//				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/another-empty-test-role")),
	//			),
	//		},
	//	})
	//})
	//
	//t.Run("with a module", func(t *testing.T) {
	//	const resourceName = "spacelift_aws_role.test"
	//
	//	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	//
	//	testSteps(t, []resource.TestStep{
	//		{
	//			Config: fmt.Sprintf(`
	//			resource "spacelift_module" "test" {
	//                name       = "test-module-%s"
	//				branch     = "master"
	//				repository = "terraform-bacon-tasty"
	//			}
	//
	//			resource "spacelift_aws_role" "test" {
	//				module_id        = spacelift_module.test.id
	//				role_arn         = "arn:aws:iam::039653571618:role/empty-test-role"
	//				duration_seconds = 942
	//			}
	//		`, randomID),
	//			Check: Resource(
	//				resourceName,
	//				Attribute("id", IsNotEmpty()),
	//				Attribute("module_id", Equals(fmt.Sprintf("test-module-%s", randomID))),
	//				Attribute("generate_credentials_in_worker", Equals("false")),
	//				Attribute("duration_seconds", Equals("942")),
	//				Attribute("external_id", IsEmpty()),
	//				AttributeNotPresent("stack_id"),
	//			),
	//		},
	//		{
	//			ResourceName:      resourceName,
	//			ImportState:       true,
	//			ImportStateId:     fmt.Sprintf("module/test-module-%s", randomID),
	//			ImportStateVerify: true,
	//		},
	//	})
	//})
	//
	//t.Run("with generating AWS creds in the worker for stack", func(t *testing.T) {
	//	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	//
	//	testSteps(t, []resource.TestStep{{
	//		Config: fmt.Sprintf(`
	//			resource "spacelift_stack" "test" {
	//				branch     = "master"
	//				repository = "demo"
	//				name       = "Test stack custom AWS %s"
	//			}
	//
	//			resource "spacelift_aws_role" "test" {
	//				stack_id                       = spacelift_stack.test.id
	//				role_arn                       = "custom_role_arn"
	//				generate_credentials_in_worker = true
	//				external_id                    = "external@id"
	//			}
	//		`, randomID),
	//		Check: Resource(
	//			"spacelift_aws_role.test",
	//			Attribute("id", IsNotEmpty()),
	//			Attribute("stack_id", Equals(fmt.Sprintf("test-stack-custom-aws-%s", randomID))),
	//			Attribute("role_arn", Equals("custom_role_arn")),
	//			Attribute("generate_credentials_in_worker", Equals("true")),
	//			Attribute("external_id", Equals("external@id")),
	//			AttributeNotPresent("module_id"),
	//		),
	//	}})
	//})
	//
	//t.Run("with generating AWS creds in the worker for module", func(t *testing.T) {
	//	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	//
	//	testSteps(t, []resource.TestStep{{
	//		Config: fmt.Sprintf(`
	//			resource "spacelift_module" "test" {
	//				name       = "test-module-%s"
	//				branch     = "master"
	//				repository = "terraform-bacon-tasty"
	//			}
	//
	//			resource "spacelift_aws_role" "test" {
	//				module_id                      = spacelift_module.test.id
	//				role_arn                       = "custom_role_arn"
	//				generate_credentials_in_worker = true
	//			}
	//		`, randomID),
	//		Check: Resource(
	//			"spacelift_aws_role.test",
	//			Attribute("id", IsNotEmpty()),
	//			Attribute("module_id", Equals(fmt.Sprintf("test-module-%s", randomID))),
	//			Attribute("generate_credentials_in_worker", Equals("true")),
	//			Attribute("external_id", IsEmpty()),
	//			AttributeNotPresent("stack_id"),
	//		),
	//	}})
	//})
}
