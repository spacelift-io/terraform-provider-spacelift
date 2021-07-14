package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBitbucketDataCenterIntegrationData(t *testing.T) {
	t.Parallel()

	testSteps(t, []resource.TestStep{{
		Config: `
			data "spacelift_bitbucket_datacenter_integration" "test" {}
		`,
		Check: Resource(
			"data.spacelift_bitbucket_datacenter_integration.test",
			Attribute("api_host", Equals("private://bitbucket-datacenter-agent-pool")),
			Attribute("webhook_secret", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_BITBUCKET_DATACENTER_WEBHOOK_SECRET"))),
			Attribute("user_facing_host", Equals("http://localhost:7990")),
		),
	}})
}
