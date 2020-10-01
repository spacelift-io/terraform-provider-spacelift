package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModuleResource(t *testing.T) {
	t.Run("with GitHub", func(t *testing.T) {
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
					administrative = true
					branch         = "master"
					description    = "%s"
					labels         = ["one", "two"]
					repository     = "terraform-bacon-tasty"
				}
			`, description)
		}

		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			Providers: map[string]terraform.ResourceProvider{
				"spacelift": testProvider(),
			},
			Steps: []resource.TestStep{
				{
					Config: config("old description"),
					Check: Resource(
						"spacelift_module.test",
						Attribute("id", Equals("terraform-bacon-tasty")),
						Attribute("administrative", Equals("true")),
						Attribute("branch", Equals("master")),
						Attribute("description", Equals("old description")),
						SetEquals("labels", "one", "two"),
						Attribute("repository", Equals("terraform-bacon-tasty")),
					),
				},
				{
					Config: config("new description"),
					Check:  Resource("spacelift_module.test", Attribute("description", Equals("new description"))),
				},
			},
		})
	})
}
