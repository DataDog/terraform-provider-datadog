// Package contracts holds the machine-readable contracts the generator
// validates against. The files are embedded so the tfgen binary is
// self-contained and never depends on its working directory at runtime.
//
// New contracts (e.g. a CLI-flag doc or a run-report schema) are added here as
// another //go:embed directive plus an exported var.
package contracts

import _ "embed"

// TrackingFieldSchema is the embedded JSON Schema (draft 2020-12) for the
// x-datadog-tf-generator OpenAPI vendor extension. parser.DecodeTracking
// compiles it once and validates every extension against it.
//
//go:embed tracking-field.schema.json
var TrackingFieldSchema []byte

// FamilyRegistrySchema is the embedded JSON Schema (draft 2020-12) for the
// tag-scoped form of x-datadog-tf-generator: a registry of artifact
// families keyed by family name. parser.DecodeRegistry compiles it once and
// validates the whole registry against it.
//
// It is a separate contract from TrackingFieldSchema rather than a branch
// inside it because the two validate different locations: this one a Tag Object, that one an operation (legacy) or a Schema Object (the sensitive flag).
//
//go:embed family-registry.schema.json
var FamilyRegistrySchema []byte
