package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModuleResource(t *testing.T) {
	t.Run("with GitHub", func(t *testing.T) {
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
					administrative  = true
					branch          = "master"
					description     = "%s"
					labels          = ["one", "two"]
					repository      = "terraform-bacon-tasty"
					shared_accounts = ["foo-subdomain", "bar-subdomain"]
				}
			`, description)
		}

		testSteps(t, []resource.TestStep{
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
					SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
				),
			},
			{
				Config: config("new description"),
				Check:  Resource("spacelift_module.test", Attribute("description", Equals("new description"))),
			},
		})
	})
}
