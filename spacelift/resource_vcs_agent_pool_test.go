package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestVCSAgentPoolResource(t *testing.T) {
	const resourceName = "spacelift_vcs_agent_pool.test"

	t.Run("creates and updates a VCS agent pool without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_vcs_agent_pool" "test" {
					name        = "Provider test VCS Agent Pool %s"
					description = "%s"
				}
			`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", StartsWith("Provider test VCS Agent Pool")),
					Attribute("description", Equals("old description")),
					Attribute("config", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description"),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("new description")),
				),
			},
		})
	})
}
