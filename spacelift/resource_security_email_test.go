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

func Test_resourceSecurityEmail(t *testing.T) {
	t.Parallel()
	const resourceName = "spacelift_security_email.test"

	t.Run("creates and updates a security email without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		emailAddress := fmt.Sprintf("%s@example.com", randomID)
		emailAddress2 := fmt.Sprintf("%s@example2.com", randomID)

		testSteps(t, []resource.TestStep{
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
