package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModuleData(t *testing.T) {
	testSteps(t, []resource.TestStep{{
		Config: `
			resource "spacelift_module" "test" {
				administrative = true
				branch         = "master"
				description    = "description"
				labels         = ["one", "two"]
				repository     = "terraform-bacon-tasty"
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
			Attribute("repository", Equals("terraform-bacon-tasty")),
		),
	}})
}
