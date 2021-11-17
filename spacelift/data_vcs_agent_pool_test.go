package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVCSAgentPoolData(t *testing.T) {
	t.Run("retrieves VCS agent pool data without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_vcs_agent_pool" "test" {
					name        = "Provider test VCS agent pool %s"
					description = "description"
				}

				data "spacelift_context" "test" {
					vcs_agent_pool_id = spacelift_vcs_agent_pool.test.id
				}
			`, randomID),
			Check: Resource(
				"data.spacelift_vcs_agent_pool.test",
				Attribute("id", IsNotEmpty()),
				Attribute("name", StartsWith("Provider test context")),
				Attribute("description", Equals("Provider test VCS agent pool")),
			),
		}})
	})
}
