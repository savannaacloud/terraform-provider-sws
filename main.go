// terraform-provider-sws — Savannaa Cloud Terraform provider.
//
// Entrypoint that hands the provider to Terraform via the plugin
// framework. Run as a subprocess by `terraform` itself; nothing in
// here is meant to be invoked directly.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/savannaacloud/terraform-provider-sws/internal/provider"
)

// Version is set at build time by goreleaser (-ldflags "-X main.version=...").
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run provider in debug mode for IDE attach")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/savannaacloud/sws",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
