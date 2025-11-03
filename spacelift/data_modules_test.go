package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModulesData(t *testing.T) {
	t.Run("reads the modules collection", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		resourceName := "spacelift_module.test"
		datasourceName := "data.spacelift_modules.test"

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name = "test-module-%s"

					administrative = true
					branch         = "master"
					labels         = ["bacon", "cabbage"]
					project_root   = "root"
					repository     = "demo"
				}

				data "spacelift_modules" "test" {
					depends_on = [spacelift_module.test]

					administrative {}

					branch {
					  any_of = ["main", "master"]
					}

					labels {
					  any_of = ["bacon"]
					}

					labels {
					  any_of = ["cabbage"]
					}

					name {
					  any_of = ["test-module-%s"]
					}

					project_root {
					  any_of = ["root"]
					}
				  }
			`, randomID, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
				CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"modules", "module_id"}, resourceName, "id"),
				CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"modules", "name"}, resourceName, "name"),
			),
		}})
	})
}
