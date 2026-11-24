package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestIPsData(t *testing.T) {
	testSteps(t, []resource.TestStep{{
		Config: `data "spacelift_ips" "test" {}`,
		Check: Resource(
			"data.spacelift_ips.test",
			SetEquals("ips", testConfig.IPs...),
		),
	}})
}
