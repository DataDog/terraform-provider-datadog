package main

import (
	"os"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/cli"
)

// Version is stamped at link time via -ldflags "-X main.Version=<tag>".
var Version = "dev"

func main() {
	os.Exit(cli.Execute(Version))
}
