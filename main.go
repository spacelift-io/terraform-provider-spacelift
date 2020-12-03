package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: spacelift.Provider,
	})
}
