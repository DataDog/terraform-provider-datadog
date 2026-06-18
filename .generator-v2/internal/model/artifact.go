package model

import "fmt"

// BuildArtifact wraps a tracked Operation's response tree into an *Artifact,
// setting only its Name, Kind, SourceFile and Schema. Lifecycle bindings and
// SDK-call resolution stay nil.
func BuildArtifact(op *Operation) (*Artifact, error) {
	if op == nil || op.Tracking == nil {
		return nil, fmt.Errorf("model: BuildArtifact requires a tracked operation")
	}
	schema, err := BuildResponseTree(op.ResponseSchema)
	if err != nil {
		return nil, err
	}
	name := op.Tracking.ArtifactName
	return &Artifact{
		Name:       name,
		Kind:       op.Tracking.ArtifactKind,
		Schema:     schema,
		SourceFile: "datadog/fwprovider/data_source_datadog_" + name + ".go",
	}, nil
}
