package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestIdpGroupMappingData(t *testing.T) {
	t.Run("reads an existing IdP group mapping by name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := "test-group-" + randomID
		description := "test description " + randomID

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_idp_group_mapping" "test" {
					name        = "%s"
					description = "%s"
					policy {
						space_id = "root"
						role     = "READ"
					}
				}

				data "spacelift_idp_group_mapping" "test" {
					name = spacelift_idp_group_mapping.test.name
				}
			`, name, description),
			Check: Resource(
				"data.spacelift_idp_group_mapping.test",
				Attribute("name", Equals(name)),
				Attribute("description", Equals(description)),
				Attribute("policy.#", Equals("1")),
			),
		}})
	})

	t.Run("returns error when name not found", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
				data "spacelift_idp_group_mapping" "test" {
					name = "non-existent-group"
				}
			`,
			ExpectError: regexp.MustCompile(`could not find IdP group mapping with name "non-existent-group"`),
		}})
	})
}
