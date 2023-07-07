package spacelift

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/framework"
)

func TestMuxServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"spacelift": func() (tfprotov5.ProviderServer, error) {
				ctx := context.Background()
				providers := []func() tfprotov5.ProviderServer{
					providerserver.NewProtocol5(framework.New("test", "bacon")()),
					spacelift.Provider("test", "bacon")().GRPCProvider,
				}

				muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

				if err != nil {
					return nil, err
				}

				return muxServer.ProviderServer(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `data "spacelift_account" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.spacelift_account.test", "id", "spacelift-account"),
				),
			},
		},
	})
}
