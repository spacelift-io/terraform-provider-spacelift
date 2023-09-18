package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
)

var commit = "dev"
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		ProviderAddr: "spacelift.io/spacelift-io/spacelift",
		Debug:        debug,
		ProviderFunc: spacelift.Provider(commit, version),
	})
}
