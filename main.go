package main

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var commit = "dev"
var version = "dev"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: spacelift.Provider(commit, version),
	})
}
