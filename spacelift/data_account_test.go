package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAccountData(t *testing.T) {
	testSteps(t, []resource.TestStep{{
		Config: `data "spacelift_account" "test" {}`,
		Check: Resource(
			"data.spacelift_account.test",
			Attribute("id", Equals("spacelift-account")),
			// We don't know in advance which account the test is going to run
			// against so the only thing we can reliably assume about it is that
			// the name and tier fields are not empty.
			Attribute("name", IsNotEmpty()),
			Attribute("tier", IsNotEmpty()),
		),
	}})
}
