package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestMountedFileResource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with a context", func(t *testing.T) {
		config := func(writeOnly bool) string {
			return fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name = "My first context %s"
				}

				resource "spacelift_mounted_file" "test" {
					context_id    = spacelift_context.test.id
					content       = base64encode("bacon is tasty")
					relative_path = "bacon.txt"
					write_only    = %t
				}
			`, randomID, writeOnly)
		}

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			Providers: map[string]terraform.ResourceProvider{
				"spacelift": testProvider(),
			},
			Steps: []resource.TestStep{
				{
					Config: config(true),
					Check: Resource(
						"spacelift_mounted_file.test",
						Attribute("id", IsNotEmpty()),
						Attribute("checksum", Equals("fb13e7977b7548a324b598e155b5b5ba3dcca2dad5789abe1411a88fa544be9b")),
						Attribute("context_id", Contains(randomID)),
						Attribute("relative_path", Equals("bacon.txt")),
						Attribute("write_only", Equals("true")),
					),
				},
				{
					Config: config(false),
					Check:  Resource("spacelift_mounted_file.test", Attribute("write_only", Equals("false"))),
				},
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			Providers: map[string]terraform.ResourceProvider{
				"spacelift": testProvider(),
			},
			Steps: []resource.TestStep{
				{
					Config: `
						resource "spacelift_module" "test" {
							branch         = "master"
							repository     = "terraform-bacon-tasty"
						}

						resource "spacelift_mounted_file" "test" {
							module_id     = spacelift_module.test.id
							content       = base64encode("bacon is tasty")
							relative_path = "bacon.txt"
						}
					`,
					Check: Resource(
						"spacelift_mounted_file.test",
						Attribute("module_id", Equals("terraform-bacon-tasty")),
						Attribute("write_only", Equals("true")),
					),
				},
			},
		})
	})

	t.Run("with a stack", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			Providers: map[string]terraform.ResourceProvider{
				"spacelift": testProvider(),
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
						resource "spacelift_stack" "test" {
							branch     = "master"
							repository = "demo"
							name       = "Test stack %s"
						}

						resource "spacelift_mounted_file" "test" {
							stack_id      = spacelift_stack.test.id
							content       = base64encode("bacon is tasty")
							relative_path = "bacon.txt"
						}
					`, randomID),
					Check: Resource(
						"spacelift_mounted_file.test",
						Attribute("stack_id", StartsWith("test-stack-")),
						Attribute("stack_id", Contains(randomID)),
					),
				},
			},
		})
	})
}
