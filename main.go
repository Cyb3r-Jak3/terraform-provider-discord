package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/lucky3028/discord-terraform/discord"
)

var (
	version string = "dev"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: discord.Provider(version),
		Debug:        debugMode,
	})
}
