package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModuleResource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("with GitHub", func(t *testing.T) {
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name            = "github-module-%s"
					administrative  = true
					branch          = "master"
					description     = "%s"
					labels          = ["one", "two"]
					repository      = "terraform-bacon-tasty"
					shared_accounts = ["foo-subdomain", "bar-subdomain"]
				}
			`, randomID, description)
		}

		const resourceName = "spacelift_module.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					"spacelift_module.test",
					Attribute("id", Equals(fmt.Sprintf("github-module-%s", randomID))),
					Attribute("administrative", Equals("true")),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", Equals(fmt.Sprintf("github-module-%s", randomID))),
					AttributeNotPresent("project_root"),
					Attribute("repository", Equals("terraform-bacon-tasty")),
					SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
					Attribute("terraform_provider", Equals("bacon")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description"),
				Check:  Resource("spacelift_module.test", Attribute("description", Equals("new description"))),
			},
		})
	})

	t.Run("project root and custom name", func(t *testing.T) {
		config := func(projectRoot string) string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
                    name               = "project-root-%s"
					administrative     = true
					branch             = "master"
					description        = "description"
					labels             = ["one", "two"]
                    project_root       = "%s"
					repository         = "terraform-bacon-tasty"
					shared_accounts    = ["foo-subdomain", "bar-subdomain"]
                    terraform_provider = "papaya"
				}
			`, randomID, projectRoot)
		}

		const resourceName = "spacelift_module.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("test-root/ab"),
				Check: Resource(
					"spacelift_module.test",
					Attribute("id", Equals(fmt.Sprintf("project-root-%s", randomID))),
					Attribute("administrative", Equals("true")),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", Equals("my-module")),
					Attribute("project_root", Equals("test-root/ab")),
					Attribute("repository", Equals("terraform-bacon-tasty")),
					SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
					Attribute("terraform_provider", Equals("papaya")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("test-root/bc"),
				Check:  Resource("spacelift_module.test", Attribute("project_root", Equals("test-root/bc"))),
			},
		})
	})
}
