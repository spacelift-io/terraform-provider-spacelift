package spacelift

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestIPsData(t *testing.T) {
	testSteps(t, []resource.TestStep{{
		Config: `data "spacelift_ips" "test" {}`,
		Check: Resource(
			"data.spacelift_ips.test",
			SetEquals("ips", strings.Split(os.Getenv("SPACELIFT_PROVIDER_TEST_IPS"), ";")...),
		),
	}})
}
