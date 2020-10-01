package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestStackResource(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with GitHub and no state import", func(t *testing.T) {
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					administrative = true
					autodeploy     = true
					branch         = "master"
					description    = "%s"
					labels         = ["one", "two"]
					name           = "Provider test stack %s"
					project_root   = "root"
					repository     = "demo"
				}
			`, description, randomID)
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
						"spacelift_stack.test",
						Attribute("id", StartsWith("provider-test-stack-")),
						Attribute("administrative", Equals("true")),
						Attribute("autodeploy", Equals("true")),
						Attribute("branch", Equals("master")),
						Attribute("description", Equals("old description")),
						SetEquals("labels", "one", "two"),
						Attribute("name", StartsWith("Provider test stack")),
						Attribute("project_root", Equals("root")),
						Attribute("repository", Equals("demo")),
					),
				},
				{
					Config: config("new description"),
					Check:  Resource("spacelift_stack.test", Attribute("description", Equals("new description"))),
				},
			},
		})
	})
}
