package parser

import (
	"bytes"
	"fmt"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"regexp"
)

// DefaultTrackingFieldName is the tag-scoped family registry extension key.
const DefaultTrackingFieldName = "x-datadog-tf-generator"

// compileEmbedded compiles one embedded contract, registering it under the
// schema's own $id so the JSON file stays the single source of truth for that
// identifier. name appears in errors only. Compiled schemas are safe for
// concurrent Validate calls.
func compileEmbedded(raw []byte, name string) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("parser: parsing embedded %s schema: %w", name, err)
	}
	obj, ok := doc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("parser: embedded %s schema is not a JSON object", name)
	}
	id, ok := obj["$id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("parser: embedded %s schema has no $id", name)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(id, doc); err != nil {
		return nil, fmt.Errorf("parser: registering %s schema: %w", name, err)
	}
	sch, err := c.Compile(id)
	if err != nil {
		return nil, fmt.Errorf("parser: compiling %s schema: %w", name, err)
	}
	return sch, nil
}

// artifactNamePattern and maxArtifactNameLen mirror the artifact_name
// constraints the family registry contract states, for the one name the generator derives
// rather than reads: a pluralized family name is checked against them before it
// becomes a Terraform address.
var artifactNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

const maxArtifactNameLen = 64
