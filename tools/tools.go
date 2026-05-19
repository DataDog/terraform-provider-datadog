//go:build tools
// +build tools

package tools

//go:generate go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.25.0

import (
	// docs generator
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
