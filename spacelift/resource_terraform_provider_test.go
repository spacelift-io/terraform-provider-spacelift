package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestTerraformProviderResource(t *testing.T) {
	const resourceName = "spacelift_terraform_provider.test"
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := func(public bool) string {
		return fmt.Sprintf(`
			resource "spacelift_terraform_provider" "test" {
				type     = "%s"
				space_id = "root"
				labels   = ["one", "two"]
				public   = %t
			}
		`, randomID, public)
	}

	testSteps(t, []resource.TestStep{
		{
			Config: config(true),
			Check: Resource(
				resourceName,
				Attribute("id", Equals(randomID)),
				Attribute("space_id", Equals("root")),
				SetEquals("labels", "one", "two"),
				Attribute("public", Equals("true")),
			),
		},
		{
			ResourceName:      resourceName,
			ImportState:       true,
			ImportStateVerify: true,
		},
		{
			Config: config(false),
			Check: Resource(
				resourceName,
				Attribute("public", Equals("false")),
			),
		},
	})
}
