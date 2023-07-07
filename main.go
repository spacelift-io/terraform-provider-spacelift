package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/framework"
)

var (
	commit  = "dev"
	version = "dev"
)

func main() {
	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")
	flag.Parse()

	ctx := context.Background()
	providers := []func() tfprotov5.ProviderServer{
		spacelift.Provider(commit, version)().GRPCProvider,
		providerserver.NewProtocol5(
			framework.New(commit, version)(),
		),
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt
	if *debugFlag {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	err = tf5server.Serve("registry.terraform.io/spacelift-io/spacelift", muxServer.ProviderServer, serveOpts...)
	if err != nil {
		log.Fatal(err)
	}
}
