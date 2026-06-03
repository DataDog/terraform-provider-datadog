package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/contracts"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// DefaultTrackingFieldName is the OpenAPI vendor-extension key the generator
// decodes by default. It can be overridden via WithTrackingFieldName (the
// --tracking-field flag), which is reserved for generator-internal fixture
// tests.
//
// TODO: once the global CLI flags are wired, the default should be set there
// (on the --tracking-field flag) rather than here.
const DefaultTrackingFieldName = "x-datadog-tf-generator"

// compiledSchema compiles the embedded tracking-field schema exactly once for
// the life of the process. Compilation reads only the embedded bytes, so a
// failure indicates a bug in the committed contract rather than anything
// attributable to the spec under test. Compiled schemas are safe for
// concurrent Validate calls.
var compiledSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(contracts.TrackingFieldSchema))
	if err != nil {
		return nil, fmt.Errorf("parser: parsing embedded tracking-field schema: %w", err)
	}
	// Register and compile under the schema's own $id, so the JSON file is the
	// single source of truth for that identifier.
	obj, ok := doc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("parser: embedded tracking-field schema is not a JSON object")
	}
	id, ok := obj["$id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("parser: embedded tracking-field schema has no $id")
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(id, doc); err != nil {
		return nil, fmt.Errorf("parser: registering tracking-field schema: %w", err)
	}
	sch, err := c.Compile(id)
	if err != nil {
		return nil, fmt.Errorf("parser: compiling tracking-field schema: %w", err)
	}
	return sch, nil
})

// TrackingError reports a malformed tracking-field extension. It names the
// OpenAPI location (method + path + operationId) so the spec author can find
// the offending operation, and wraps the underlying validation or decode
// failure for errors.As / Unwrap.
type TrackingError struct {
	Path        string
	Method      string
	OperationId string
	Cause       error
}

func (e *TrackingError) Error() string {
	return fmt.Sprintf("parser: invalid tracking-field extension on %s %s (operationId %q): %v",
		e.Method, e.Path, e.OperationId, e.Cause)
}

func (e *TrackingError) Unwrap() error { return e.Cause }

// DecodeTracking decodes and validates the tracking-field extension (named by
// extensionName) on op. It returns (nil, nil) when op is out of scope: it
// carries no extensions, the extension is absent, or the extension sets
// skip:true (a documented opt-out equivalent to removing the annotation). A
// malformed extension yields a *TrackingError naming the OpenAPI location at
// path/method.
//
// op alone does not know its own OpenAPI path, so path and method are supplied
// by the caller (LoadSpec) for the error message.
func DecodeTracking(op *v3.Operation, path, method, extensionName string) (*model.TrackingFieldMetadata, error) {
	if op == nil || op.Extensions == nil {
		return nil, nil
	}
	node := op.Extensions.GetOrZero(extensionName)
	if node == nil {
		return nil, nil // absent → out of scope
	}

	wrap := func(cause error) error {
		return &TrackingError{Path: path, Method: method, OperationId: op.OperationId, Cause: cause}
	}

	// Decode once into a generic value. It must stay `any`: the validator works
	// on JSON-native values (map[string]any/…), not Go structs, and the model
	// declares no yaml tags — so decoding straight into the struct would both
	// fail validation and silently drop the snake_case keys.
	var raw any
	if err := node.Decode(&raw); err != nil {
		return nil, wrap(fmt.Errorf("decoding extension node: %w", err))
	}

	// Honor an explicit skip:true before validating: it is a documented opt-out
	// equivalent to removing the extension, so a skipped operation is out of
	// scope and must not surface shape errors. Only a real boolean true counts;
	// any other value falls through to validation (a non-bool skip is invalid).
	if m, ok := raw.(map[string]any); ok {
		if skip, _ := m["skip"].(bool); skip {
			return nil, nil
		}
	}

	sch, err := compiledSchema()
	if err != nil {
		return nil, err // committed-contract bug, not author-attributable
	}
	if err := sch.Validate(raw); err != nil {
		return nil, wrap(err)
	}

	// Shape is valid: project onto the typed metadata. The model carries JSON
	// tags (the snake_case keys, and how it later round-trips into the run
	// report), so bridge YAML→JSON rather than decoding the node directly —
	// go.yaml.in/yaml honors yaml tags, which the model does not declare.
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, wrap(fmt.Errorf("re-encoding extension: %w", err))
	}
	var meta model.TrackingFieldMetadata
	if err := json.Unmarshal(jsonBytes, &meta); err != nil {
		return nil, wrap(fmt.Errorf("decoding extension into metadata: %w", err))
	}
	if meta.IdStrategy == "" {
		meta.IdStrategy = model.IdStrategyDataID // schema default "data.id"
	}
	return &meta, nil
}
