package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: datadog.Provider})
}
