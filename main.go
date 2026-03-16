package main

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
)

var commit = "dev"
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	ctx := context.Background()

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(spacelift.NewFrameworkProvider(commit, version)),
		func() tfprotov6.ProviderServer {
			v6server, err := tf5to6server.UpgradeServer(
				ctx,
				spacelift.Provider(commit, version)().GRPCProvider,
			)
			if err != nil {
				panic(err)
			}
			return v6server
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		panic(err)
	}

	serveOpts := []tf6server.ServeOpt{}
	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	if err := tf6server.Serve(
		"spacelift.io/spacelift-io/spacelift",
		muxServer.ProviderServer,
		serveOpts...,
	); err != nil {
		panic(err)
	}
}
