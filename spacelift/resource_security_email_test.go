package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

var securityEmailSimple = `
resource "spacelift_security_email" "test" {
	email = "%s"
}
`

// The security email is account-wide, so two concurrent writers fight over one
// value. Nothing else reads it today, but two CI jobs against the same account
// would collide, so keep it sequential alongside the other account-level settings.
func Test_resourceSecurityEmail(t *testing.T) {
	const resourceName = "spacelift_security_email.test"

	t.Run("creates and updates a security email without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		emailAddress := fmt.Sprintf("%s@example.com", randomID)
		emailAddress2 := fmt.Sprintf("%s@example2.com", randomID)

		testStepsSequential(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(securityEmailSimple, emailAddress),
				Check: Resource(
					resourceName,
					Attribute("email", Equals(emailAddress)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(securityEmailSimple, emailAddress2),
				Check: Resource(
					resourceName,
					Attribute("email", Equals(emailAddress2)),
				),
			},
		})
	})
}
