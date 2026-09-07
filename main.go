package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/kislerdm/terraform-provider-neon/provider"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "0.0.0-alpha.0"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	legacyServer, err := tf5to6server.UpgradeServer(context.Background(), func() tfprotov5.ProviderServer {
		return provider.New(version).GRPCProvider()
	})
	if err != nil {
		log.Fatal(err)
	}

	muxServer, err := tf6muxserver.NewMuxServer(context.Background(),
		func() tfprotov6.ProviderServer {
			return legacyServer
		},
		providerserver.NewProtocol6(provider.NewFramework(version)),
	)
	if err != nil {
		log.Fatal(err)
	}

	opts := &plugin.ServeOpts{
		Debug: debugMode,

		ProviderAddr: "registry.terraform.io/" + provider.Name,

		GRPCProviderV6Func: func() tfprotov6.ProviderServer {
			return muxServer
		},
	}

	plugin.Serve(opts)
}
