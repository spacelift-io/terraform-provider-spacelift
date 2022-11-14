package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStacksData(t *testing.T) {
	t.Run("reads the stacks collection", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		resourceName := "spacelift_stack.test"
		datasourceName := "data.spacelift_stacks.test"

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					name = "My stack %s"

					administrative = true
					branch         = "master"
					labels         = ["bacon", "cabbage"]
					project_root   = "root"
					repository     = "demo"
				}

				data "spacelift_stacks" "test" {
					depends_on = [spacelift_stack.test]

					administrative {}

					branch {
					  any_of = ["main", "master"]
					}

					locked {
					  equals = false
					}

					labels {
					  any_of = ["bacon"]
					}

					labels {
					  any_of = ["cabbage"]
					}

					name {
					  any_of = ["My stack %s"]
					}

					project_root {
					  any_of = ["root"]
					}

					state {
					  any_of = ["NONE"]
					}

					vendor {
					  any_of = ["Terraform"]
					}
				  }
			`, randomID, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
				CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"stacks", "stack_id"}, resourceName, "id"),
				CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"stacks", "name"}, resourceName, "name"),
			),
		}})
	})
}
