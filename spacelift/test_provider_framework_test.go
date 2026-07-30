// Helpers here are used selectively as resources migrate to the Plugin Framework, so
// some have no callers at any given point in the migration.
//
//nolint:unused
package spacelift

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// testAccProtoV6ProviderFactories returns the Plugin Framework provider factory
// for use in ProtoV6ProviderFactories test fields.
func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"spacelift": providerserver.NewProtocol6WithError(NewFrameworkProvider("commit", "version")),
	}
}

// testAccProtoV6MuxProviderFactories returns a muxed provider factory serving both
// the Plugin Framework and SDKv2 providers, mirroring main.go. Needed for configs that
// reference a migrated resource alongside one still implemented in SDKv2.
func testAccProtoV6MuxProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"spacelift": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()

			upgraded, err := tf5to6server.UpgradeServer(ctx, Provider("commit", "version")().GRPCProvider)
			if err != nil {
				return nil, err
			}

			muxServer, err := tf6muxserver.NewMuxServer(ctx,
				func() tfprotov6.ProviderServer { return upgraded },
				providerserver.NewProtocol6(NewFrameworkProvider("commit", "version")),
			)
			if err != nil {
				return nil, err
			}

			return muxServer.ProviderServer(), nil
		},
	}
}

// testStepsFramework runs acceptance tests against the Plugin Framework provider. Use it
// once every resource in a config has been migrated; it is stricter than testStepsMux
// because the SDKv2 provider is not available as a fallback.
//
// IsUnitTest matches testSteps for the reason given on testStepsMux.
func testStepsFramework(t *testing.T, steps []resource.TestStep) {
	t.Parallel()
	t.Helper()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps:                    steps,
	})
}

// testStepsFrameworkSequential is the non-parallel variant of testStepsFramework.
func testStepsFrameworkSequential(t *testing.T, steps []resource.TestStep) {
	t.Helper()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps:                    steps,
	})
}

// testStepsMux runs acceptance tests against the muxed provider. Use it when a config
// mixes resources that have been migrated to the Framework with ones still on SDKv2.
//
// IsUnitTest mirrors testSteps: CI runs `go test ./...` with credentials but without
// TF_ACC, so a helper replacing testSteps has to keep bypassing that check or the tests
// it serves quietly stop running.
func testStepsMux(t *testing.T, steps []resource.TestStep) {
	t.Parallel()
	t.Helper()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testAccProtoV6MuxProviderFactories(),
		Steps:                    steps,
	})
}

// testStepsMuxSequential is the non-parallel variant of testStepsMux.
func testStepsMuxSequential(t *testing.T, steps []resource.TestStep) {
	t.Helper()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testAccProtoV6MuxProviderFactories(),
		Steps:                    steps,
	})
}
