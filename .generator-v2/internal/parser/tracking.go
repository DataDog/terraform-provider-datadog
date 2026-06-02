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

// TODO: will have to be changed later once global cli flags are wired. Default value should be set there.
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
	c := jsonschema.NewCompiler()
	if err := c.AddResource(contracts.TrackingFieldSchemaID, doc); err != nil {
		return nil, fmt.Errorf("parser: registering tracking-field schema: %w", err)
	}
	sch, err := c.Compile(contracts.TrackingFieldSchemaID)
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

	// Decode once into a generic value to validate against the JSON schema —
	// the single source of truth for the extension's shape.
	var raw any
	if err := node.Decode(&raw); err != nil {
		return nil, wrap(fmt.Errorf("decoding extension node: %w", err))
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
	if meta.Skip {
		return nil, nil
	}
	if meta.IdStrategy == "" {
		meta.IdStrategy = model.IdStrategyDataID // schema default "data.id"
	}
	return &meta, nil
}
