// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/thrashr888/terraform-provider-registry/internal/provider"
)

// Run "go generate" to generate the documentation
//go:generate go tool tfplugindocs generate

var (
	// version and commit are set by goreleaser at build time (see .goreleaser.yml ldflags).
	version string = "dev"
	commit  string = ""
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/thrashr888/registry",
		Debug:   debug,
	}

	if debug && commit != "" {
		log.Printf("terraform-provider-registry %s (%s)", version, commit)
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
