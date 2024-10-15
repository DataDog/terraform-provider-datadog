//go:build tools
// +build tools

package tools

//go:generate go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

import (
	// docs generator
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
