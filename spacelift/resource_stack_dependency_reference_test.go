package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackDependencyReferenceResource(t *testing.T) {
	const resourceName = "spacelift_stack_dependency_reference.test"

	t.Run("creates, updates and deletes stack dependency reference", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		configWithoutReference := func() string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test1" {
					branch     = "master"
					repository = "demo"
					name       = "my-first-stack-%s"
				}

				resource "spacelift_stack" "test2" {
					branch     = "master"
					repository = "demo"
					name       = "my-second-stack-%s"
				}

				resource "spacelift_stack_dependency" "test" {
					stack_id = spacelift_stack.test1.id
					depends_on_stack_id = spacelift_stack.test2.id
				}`, randomID, randomID)
		}

		configWithReference := func(outputName, inputName string) string {
			return configWithoutReference() + fmt.Sprintf(`
				resource "spacelift_stack_dependency_reference" "test" {
					stack_dependency_id = spacelift_stack_dependency.test.id
					output_name = "%s"
					input_name = "%s"
				}`, outputName, inputName)
		}

		testSteps(t, []resource.TestStep{
			{ // creates reference
				Config: configWithReference("output_abc", "input_123"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("output_name", Equals("output_abc")),
					Attribute("input_name", Equals("input_123")),
				),
			},
			{ // updates input_name
				Config: configWithReference("output_abc", "input_456"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("output_name", Equals("output_abc")),
					Attribute("input_name", Equals("input_456")),
				),
			},
			{ // updates output_name
				Config: configWithReference("output_xyz", "input_456"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("output_name", Equals("output_xyz")),
					Attribute("input_name", Equals("input_456")),
				),
			},
			{ // deletes reference
				Config: configWithoutReference(),
				Check: func(state *terraform.State) error {
					if len(state.Modules) == 0 {
						return errors.New("no modules present")
					}

					_, ok := state.Modules[0].Resources[resourceName]
					if ok {
						return errors.Errorf("resource %s not found", resourceName)
					}
					return nil
				},
			},
			{ // re-create reference
				Config: configWithReference("output_final", "input_final"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("output_name", Equals("output_final")),
					Attribute("input_name", Equals("input_final")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		})
	})
}
