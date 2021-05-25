package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModuleData(t *testing.T) {
	testSteps(t, []resource.TestStep{{
		Config: `
			resource "spacelift_module" "test" {
				administrative  = true
				branch          = "master"
				description     = "description"
				labels          = ["one", "two"]
				repository      = "terraform-bacon-tasty"
				shared_accounts = ["foo-subdomain", "bar-subdomain"]
			}

			data "spacelift_module" "test" {
				module_id = spacelift_module.test.id
			}
		`,
		Check: Resource(
			"data.spacelift_module.test",
			Attribute("id", Equals("terraform-bacon-tasty")),
			Attribute("administrative", Equals("true")),
			Attribute("branch", Equals("master")),
			Attribute("description", Equals("description")),
			SetEquals("labels", "one", "two"),
			Attribute("name", Equals("tasty")),
			AttributeNotPresent("project_root"),
			Attribute("repository", Equals("terraform-bacon-tasty")),
			SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
			Attribute("terraform_provider", Equals("bacon")),
		),
	}})

	testSteps(t, []resource.TestStep{{
		Config: `
			resource "spacelift_module" "test" {
				administrative     = true
				branch             = "master"
				description        = "description"
				labels             = ["one", "two"]
				name               = "my-module"
				project_root       = "project-root"
				repository         = "terraform-bacon-tasty"
				shared_accounts    = ["foo-subdomain", "bar-subdomain"]
				terraform_provider = "papaya"
			}

			data "spacelift_module" "test" {
				module_id = spacelift_module.test.id
			}
		`,
		Check: Resource(
			"data.spacelift_module.test",
			Attribute("id", Equals("my-module")),
			Attribute("administrative", Equals("true")),
			Attribute("branch", Equals("master")),
			Attribute("description", Equals("description")),
			SetEquals("labels", "one", "two"),
			Attribute("name", Equals("my-module")),
			Attribute("project_root", Equals("project-root")),
			Attribute("repository", Equals("terraform-bacon-tasty")),
			SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
			Attribute("terraform_provider", Equals("papaya")),
		),
	}})
}
