package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

var defaultAccountRunnerImageBothFields = `
resource "spacelift_default_runner_image" "test" {
	private = "%s"
	public  = "%s"
}
`

var defaultAccountRunnerImagePrivateOnly = `
resource "spacelift_default_runner_image" "test" {
	private = "%s"
}
`

var defaultAccountRunnerImagePublicOnly = `
resource "spacelift_default_runner_image" "test" {
	public = "%s"
}
`

func Test_resourceDefaultAccountRunnerImage(t *testing.T) {
	const resourceName = "spacelift_default_runner_image.test"

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	privateImage := fmt.Sprintf("private-runner:%s", randomID)
	publicImage := fmt.Sprintf("public-runner:%s", randomID)
	privateImage2 := fmt.Sprintf("private-runner:%s-updated", randomID)
	publicImage2 := fmt.Sprintf("public-runner:%s-updated", randomID)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(defaultAccountRunnerImageBothFields, privateImage, publicImage),
			Check: Resource(
				resourceName,
				Attribute("private", Equals(privateImage)),
				Attribute("public", Equals(publicImage)),
			),
		},
		{
			ResourceName:      resourceName,
			ImportState:       true,
			ImportStateVerify: true,
		},
		{
			Config: fmt.Sprintf(defaultAccountRunnerImageBothFields, privateImage2, publicImage2),
			Check: Resource(
				resourceName,
				Attribute("private", Equals(privateImage2)),
				Attribute("public", Equals(publicImage2)),
			),
		},
		{
			Config: fmt.Sprintf(defaultAccountRunnerImagePrivateOnly, privateImage2),
			Check: Resource(
				resourceName,
				Attribute("private", Equals(privateImage2)),
				Attribute("public", Equals("")),
			),
		},
		{
			Config: fmt.Sprintf(defaultAccountRunnerImagePublicOnly, publicImage2),
			Check: Resource(
				resourceName,
				Attribute("public", Equals(publicImage2)),
				Attribute("private", Equals("")),
			),
		},
		{
			Config: fmt.Sprintf(defaultAccountRunnerImageBothFields, privateImage, publicImage),
			Check: Resource(
				resourceName,
				Attribute("private", Equals(privateImage)),
				Attribute("public", Equals(publicImage)),
			),
		},
	})
}
