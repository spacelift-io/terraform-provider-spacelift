//nolint:unused
package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testAccProviderFactories returns the SDKv2 provider factory for use in the
// ProviderFactories test field. Each test gets its own provider instance:
// schema.Provider is mutated when Terraform configures it, so sharing one
// across parallel tests races on its internal state.
func testAccProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"spacelift": func() (*schema.Provider, error) {
			return Provider("commit", "version")(), nil
		},
	}
}

func testSteps(t *testing.T, steps []resource.TestStep) {
	t.Parallel()
	t.Helper()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviderFactories(),
		Steps:             steps,
	})
}

func testStepsSequential(t *testing.T, steps []resource.TestStep) {
	t.Helper()

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviderFactories(),
		Steps:             steps,
	})
}
