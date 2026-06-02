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

// TrackingFieldSchemaID is the schema's canonical $id — the URL it is
// registered and compiled under by the validator.
const TrackingFieldSchemaID = "https://datadog.github.io/terraform-provider-datadog/contracts/tracking-field.schema.json"
