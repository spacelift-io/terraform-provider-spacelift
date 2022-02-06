package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVCSAgentPoolsData(t *testing.T) {
	t.Run("retrieves VCS agent pools data without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		resourceName := "spacelift_vcs_agent_pool.test"
		datasourceName := "data.spacelift_vcs_agent_pools.test"

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_vcs_agent_pool" "test" {
					name        = "provider-test-vcs-agent-%s"
					description = "Provider test VCS agent pool"
				}

				data "spacelift_vcs_agent_pools" "test" {}
			`, randomID),
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
				CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"vcs_agent_pools", "vcs_agent_pool_id"}, resourceName, "id"),
				CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"vcs_agent_pools", "name"}, resourceName, "name"),
				CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"vcs_agent_pools", "description"}, resourceName, "description"),
			),
		}})
	})
}
