package internal_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// TestRegistryShapeGeneratesIdenticalSource is the claim that makes migrating
// the spec repositories safe: the document-root family registry is sugar, so a
// family declared that way renders byte-for-byte what the operation-level
// annotation rendered. Anything less and every migrated family would need its
// generated output re-reviewed.
//
// It drives the real incident-type fixture twice — once as committed, once with
// its annotation lifted into a root registry — through the full offline
// pipeline, and diffs the emitted Go source.
//
// The fixture is a good witness precisely because its annotation sits on the
// GET, not the POST: the registry anchors a resource on group.create, so the
// two runs reach BuildArtifact from different operations and must still agree.
func TestRegistryShapeGeneratesIdenticalSource(t *testing.T) {
	committed := filepath.Join("testdata", "fixtures", "resource_incident_type", "openapi.yaml")

	legacy, legacyOp := renderIncidentType(t, committed)
	registry, registryOp := renderIncidentType(t, writeRegistryVariant(t, committed))

	// Guard against a vacuous pass: if the variant still carried the old
	// annotation, or the registry were quietly ignored, the two renderings would
	// match for the wrong reason. The addresses prove which shape each run read,
	// and the methods prove the anchors really do differ.
	if legacyOp.Tracking.Family != "" {
		t.Errorf("committed fixture should be read as an operation-level annotation, got family %q",
			legacyOp.Tracking.Family)
	}
	if registryOp.Tracking.Family != "incident_type" || registryOp.Tracking.Block != parser.BlockResource {
		t.Errorf("variant should be read from the registry, got address (%q, %q)",
			registryOp.Tracking.Family, registryOp.Tracking.Block)
	}
	if legacyOp.Method != "GET" || registryOp.Method != "POST" {
		t.Errorf("anchors should differ (GET annotation vs create-anchored registry), got %s and %s",
			legacyOp.Method, registryOp.Method)
	}
	if !registryOp.Synthetic {
		t.Error("the registry's artifact should be carried by a synthetic operation")
	}

	if legacy != registry {
		t.Errorf("generated source differs between annotation shapes:\n%s",
			firstDifference(legacy, registry))
	}
}

// renderIncidentType loads a spec, finds its tracked incident_type resource and
// renders it, mirroring the pipeline tfgen generate runs.
func renderIncidentType(t *testing.T, specPath string) (string, *model.Operation) {
	t.Helper()
	spec, err := parser.LoadSpec(specPath)
	if err != nil {
		t.Fatalf("load %s: %v", specPath, err)
	}
	var operation *model.Operation
	for _, candidate := range spec.Operations {
		if candidate.Tracking != nil &&
			candidate.Tracking.ArtifactKind == model.ArtifactKindResource &&
			candidate.Tracking.ArtifactName == "incident_type" {
			operation = candidate
			break
		}
	}
	if operation == nil {
		t.Fatalf("%s has no tracked incident_type resource", specPath)
	}
	artifact, err := model.BuildArtifact(operation)
	if err != nil {
		t.Fatalf("build artifact from %s: %v", specPath, err)
	}
	view, err := emit.BuildResourceView(artifact)
	if err != nil {
		t.Fatalf("build view from %s: %v", specPath, err)
	}
	source, err := emit.RenderResource(view)
	if err != nil {
		t.Fatalf("render from %s: %v", specPath, err)
	}
	return string(source), operation
}

// writeRegistryVariant rewrites the fixture into a temp file whose annotation
// has moved from the operation to a document-root registry, and returns its
// path. It fails rather than silently producing an unannotated spec if the
// fixture's shape ever changes.
func writeRegistryVariant(t *testing.T, specPath string) string {
	t.Helper()
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	// Drop the operation-level annotation: its own line, then every line
	// indented deeper than it (the block it owns).
	lines := strings.Split(string(data), "\n")
	var kept []string
	removed := 0
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) != "x-datadog-tf-generator:" {
			kept = append(kept, line)
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		removed++
		for i+1 < len(lines) {
			next := lines[i+1]
			if strings.TrimSpace(next) == "" {
				break
			}
			if len(next)-len(strings.TrimLeft(next, " ")) <= indent {
				break
			}
			i++
		}
	}
	if removed != 1 {
		t.Fatalf("expected exactly one operation-level annotation in the fixture, removed %d", removed)
	}

	// The same four operations, declared as one family at the document root.
	// tf_description is copied verbatim: a difference there would show up in the
	// rendered schema and defeat the comparison.
	variant := strings.Join(kept, "\n") + `
x-datadog-tf-generator:
  incident_type:
    resource:
      tf_description: Provides a Datadog incident type resource.
      group:
        create: CreateIncidentType
        read: GetIncidentType
        update: UpdateIncidentType
        delete: DeleteIncidentType
`

	path := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := os.WriteFile(path, []byte(variant), 0o600); err != nil {
		t.Fatalf("write variant: %v", err)
	}
	return path
}

// firstDifference reports the first differing line with a little context, which
// is far easier to read than two full renderings of a resource.
func firstDifference(a, b string) string {
	aLines, bLines := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := range max(len(aLines), len(bLines)) {
		var x, y string
		if i < len(aLines) {
			x = aLines[i]
		}
		if i < len(bLines) {
			y = bLines[i]
		}
		if x != y {
			return fmt.Sprintf("line %d:\n  operation-level: %s\n  registry:        %s", i+1, x, y)
		}
	}
	return fmt.Sprintf("(no line differs; lengths %d and %d)", len(a), len(b))
}
