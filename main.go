package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func startTfProtoV5(ctx context.Context, debugMode bool) error {
	providers := []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(fwprovider.New()),
		datadog.Provider().GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt

	if debugMode {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	return tf5server.Serve(
		"registry.terraform.io/DataDog/datadog",
		muxServer.ProviderServer,
		serveOpts...,
	)
}

func startTfProtoV6(ctx context.Context, debugMode bool) error {
	legacyProviderV6, err := tf5to6server.UpgradeServer(ctx, datadog.Provider().GRPCProvider)
	if err != nil {
		return err
	}
	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(fwprovider.New()),
		func() tfprotov6.ProviderServer {
			return legacyProviderV6
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt

	if debugMode {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	return tf6server.Serve(
		"registry.terraform.io/DataDog/datadog",
		muxServer.ProviderServer,
		serveOpts...,
	)
}

func main() {
	ctx := context.Background()

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	var err error
	v := os.Getenv("USE_PROTO_V6")
	if len(v) > 0 {
		err = startTfProtoV6(ctx, debugMode)
	} else {
		err = startTfProtoV5(ctx, debugMode)
	}
	if err != nil {
		log.Fatal(err)
	}
}
