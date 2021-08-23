package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBitbucketCloudIntegrationData(t *testing.T) {
	t.Parallel()

	testSteps(t, []resource.TestStep{{
		Config: `
			data "spacelift_bitbucket_cloud_integration" "test" {}
		`,
		Check: Resource(
			"data.spacelift_bitbucket_cloud_integration.test",
			Attribute("username", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_BITBUCKET_CLOUD_USERNAME"))),
		),
	}})
}
