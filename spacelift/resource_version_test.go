package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVersionResource(t *testing.T) {
	t.Run("on a new module", func(t *testing.T) {
		const resourceName = "spacelift_version.test"

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name           = "test-version-module-%s"
					administrative = true
					branch         = "module"
					repository     = "terraform-bacon-tasty"
					labels         = ["version-test"]
				}

				resource "spacelift_version" "test" {
					module_id = spacelift_module.test.id 
					keepers = {
						"repository" = spacelift_module.test.repository
					} 
				}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("module_id", Contains(randomID)),
				),
			},
		})
	})
}
