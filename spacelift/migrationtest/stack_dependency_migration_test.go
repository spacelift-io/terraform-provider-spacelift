// Package migrationtest holds the no-drift verification tests for resources moved from
// terraform-plugin-sdk/v2 to terraform-plugin-framework.
//
// It is a separate package on purpose. These tests need ConfigPlanChecks and
// plancheck.ExpectEmptyPlan, which live only in terraform-plugin-testing/helper/resource,
// and that package and terraform-plugin-sdk/v2/helper/resource both register a global
// -sweep flag at init. Linking both into one test binary panics with "flag redefined:
// sweep" before any test runs, which would take out the whole spacelift package. So
// nothing here may import the SDKv2 helper/resource, directly or transitively.
package migrationtest

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
)

// lastSDKv2Release is the last published release in which spacelift_stack_dependency was
// still SDKv2-based. Pinned exactly: ">=" would resolve to the newest release, which after
// later migrations is itself Framework-based, so the test would compare the Framework
// implementation against itself and prove nothing.
const lastSDKv2Release = "= 1.52.4"

// muxedProviderFactories mirrors main.go. The migration configs below create
// spacelift_stack resources, which are still SDKv2, so a Framework-only factory cannot
// serve them. Built here rather than reused from the spacelift package because that
// package's test helpers are in _test.go files and unreachable from here.
func muxedProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"spacelift": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()

			upgraded, err := tf5to6server.UpgradeServer(ctx, spacelift.Provider("commit", "version")().GRPCProvider)
			if err != nil {
				return nil, err
			}

			muxServer, err := tf6muxserver.NewMuxServer(ctx,
				func() tfprotov6.ProviderServer { return upgraded },
				providerserver.NewProtocol6(spacelift.NewFrameworkProvider("commit", "version")),
			)
			if err != nil {
				return nil, err
			}

			return muxServer.ProviderServer(), nil
		},
	}
}

// TestStackDependencyMigration proves the Plugin Framework implementation of
// spacelift_stack_dependency produces state identical to the last SDKv2 release: step 1
// stands up state with the published release, step 2 re-applies the same config through
// the muxed provider and asserts the post-apply-post-refresh plan is empty.
//
// Requires TF_ACC=1 and credentials — it downloads a real provider release.
func TestStackDependencyMigration(t *testing.T) {
	t.Parallel()

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
		resource "spacelift_stack" "test1" {
			branch     = "master"
			repository = "demo"
			name       = "migration-first-stack-%s"
		}

		resource "spacelift_stack" "test2" {
			branch     = "master"
			repository = "demo"
			name       = "migration-second-stack-%s"
		}

		resource "spacelift_stack_dependency" "test" {
			stack_id            = spacelift_stack.test1.id
			depends_on_stack_id = spacelift_stack.test2.id
		}
	`, randomID, randomID)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"spacelift": {
						Source:            "spacelift.io/spacelift-io/spacelift",
						VersionConstraint: lastSDKv2Release,
					},
				},
				Config: config,
			},
			{
				ProtoV6ProviderFactories: muxedProviderFactories(),
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
