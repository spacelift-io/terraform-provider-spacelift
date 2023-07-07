package spacelift

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGitlabIntegrationData(t *testing.T) {
	testSteps(t, []resource.TestStep{{
		Config: `
			data "spacelift_gitlab_integration" "test" {}
		`,
		Check: Resource(
			"data.spacelift_gitlab_integration.test",
			Attribute("api_host", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITLAB_API_HOST"))),
			Attribute("webhook_secret", Equals(os.Getenv("SPACELIFT_PROVIDER_TEST_GITLAB_WEBHOOK_SECRET"))),
		),
	}})
}
