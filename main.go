package main

import (
	"context"
	"flag"
	"github.com/Cyb3r-Jak3/discord-terraform/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"log"
)

var (
	version = "dev"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	opts := providerserver.ServeOpts{
		// Also update the tfplugindocs generate command to either remove the
		// -provider-name flag or set its value to the updated provider name.
		Address: "registry.terraform.io/Cyb3r-Jak3/discord",
		Debug:   debugMode,
	}
	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
