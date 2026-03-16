//nolint:unused
package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories returns the Plugin Framework provider factory
// for use in ProtoV6ProviderFactories test fields.
func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"spacelift": providerserver.NewProtocol6WithError(NewFrameworkProvider("commit", "version")),
	}
}

// testStepsFramework runs acceptance tests against the Plugin Framework provider.
// Unlike testSteps, TF_ACC=1 is required (these are not unit tests).
func testStepsFramework(t *testing.T, steps []resource.TestStep) {
	t.Parallel()
	t.Helper()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps:                    steps,
	})
}

// testStepsFrameworkSequential is the non-parallel variant of testStepsFramework.
func testStepsFrameworkSequential(t *testing.T, steps []resource.TestStep) {
	t.Helper()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps:                    steps,
	})
}
