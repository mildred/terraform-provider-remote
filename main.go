package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/mildred/terraform-provider-remote/remote"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: remote.Provider})
}
