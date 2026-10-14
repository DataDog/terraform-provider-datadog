// Package contracts holds the machine-readable contracts the generator
// validates against. The files are embedded so the tfgen binary is
// self-contained and never depends on its working directory at runtime.
//
// New contracts (e.g. a CLI-flag doc or a run-report schema) are added here as
// another //go:embed directive plus an exported var.
package contracts

import _ "embed"

// FamilyRegistrySchema is the embedded JSON Schema (draft 2020-12) for the
// tag-scoped form of x-datadog-tf-generator: a registry of artifact
// families keyed by family name. parser.DecodeRegistry compiles it once and
// validates the whole registry against it.
//
//go:embed family-registry.schema.json
var FamilyRegistrySchema []byte
