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

// RunReportSchema is the embedded JSON Schema (draft 2020-12) for the run
// report tfgen generate writes to --report. CI asserts against that report, so
// the schema pins its shape: it is strict throughout, and a new field must be
// added here in the same change that emits it.
//
//go:embed run-report.schema.json
var RunReportSchema []byte
