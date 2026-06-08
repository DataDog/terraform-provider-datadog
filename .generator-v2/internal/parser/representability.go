package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// RepresentabilityFinding records one schema node that cannot be expressed as a
// Terraform Plugin Framework attribute. SchemaPath is the dotted location within
// the body, rooted at "request" or "response"; array elements append "[]" and
// map values append "{}".
type RepresentabilityFinding struct {
	ArtifactKind model.ArtifactKind
	ArtifactName string
	OperationId  string
	Method       string
	Path         string // operation path, e.g. /api/v2/things/{id}
	SchemaPath   string // e.g. response.data.config, request.items[]
	Kind         model.SchemaKind
}

// UnrepresentableSchemaError aggregates every unrepresentable node found across
// the spec into one error, sorted deterministically so the message is stable
// regardless of input order.
type UnrepresentableSchemaError struct {
	Findings []RepresentabilityFinding
}

func (e *UnrepresentableSchemaError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "parser: %d unrepresentable schema node(s):", len(e.Findings))
	for _, f := range e.Findings {
		fmt.Fprintf(&b, "\n  %s %q: operation %s (%s %s) at %s — %s",
			f.ArtifactKind, f.ArtifactName, f.OperationId, f.Method, f.Path, f.SchemaPath, representabilityReason(f.Kind))
	}
	return b.String()
}

// CheckSchemaRepresentability reports every request/response schema node that
// cannot become a Terraform Plugin Framework attribute — variant (oneOf/anyOf),
// ref_cycle, and unsupported (no representable type or structure) nodes. It reads
// the model.Schema trees from NormalizeSchemas, so it runs after that pass and
// before BuildAttributeTree. Findings are aggregated into one
// *UnrepresentableSchemaError (nil when everything is representable).
func CheckSchemaRepresentability(spec *model.Spec) error {
	if spec == nil {
		return nil
	}

	// operationId → owning artifact, so a finding on an untracked group member
	// (e.g. the read op) still names its artifact. First match wins.
	artifactOf := make(map[string]artifactRef)
	for _, op := range spec.Operations {
		if op == nil || op.Tracking == nil {
			continue
		}
		ref := artifactRef{kind: op.Tracking.ArtifactKind, name: op.Tracking.ArtifactName}
		for _, id := range crudOperationIds(op.Tracking) {
			if _, seen := artifactOf[id]; !seen {
				artifactOf[id] = ref
			}
		}
	}

	var findings []RepresentabilityFinding
	for _, op := range spec.Operations {
		if op == nil || (op.RequestSchema == nil && op.ResponseSchema == nil) {
			continue
		}
		ref := artifactOf[op.OperationId]
		meta := findingMeta{kind: ref.kind, name: ref.name, opID: op.OperationId, method: op.Method, path: op.Path}
		collectUnrepresentable(op.RequestSchema, "request", meta, &findings)
		collectUnrepresentable(op.ResponseSchema, "response", meta, &findings)
	}

	if len(findings) == 0 {
		return nil
	}
	sortFindings(findings)
	return &UnrepresentableSchemaError{Findings: findings}
}

type artifactRef struct {
	kind model.ArtifactKind
	name string
}

type findingMeta struct {
	kind   model.ArtifactKind
	name   string
	opID   string
	method string
	path   string
}

// collectUnrepresentable walks s depth-first, appending a finding for each
// variant/ref_cycle/unsupported node. An unrepresentable node is terminal — its
// children are not descended, since the whole node is already rejected.
func collectUnrepresentable(s *model.Schema, schemaPath string, meta findingMeta, out *[]RepresentabilityFinding) {
	if s == nil {
		return
	}
	switch s.Kind {
	case model.SchemaKindVariant, model.SchemaKindRefCycle, model.SchemaKindUnsupported:
		*out = append(*out, RepresentabilityFinding{
			ArtifactKind: meta.kind,
			ArtifactName: meta.name,
			OperationId:  meta.opID,
			Method:       meta.method,
			Path:         meta.path,
			SchemaPath:   schemaPath,
			Kind:         s.Kind,
		})
	case model.SchemaKindObject:
		for _, key := range sortedSchemaKeys(s.Properties) {
			collectUnrepresentable(s.Properties[key], schemaPath+"."+key, meta, out)
		}
	case model.SchemaKindArray:
		collectUnrepresentable(s.Items, schemaPath+"[]", meta, out)
	case model.SchemaKindMap:
		collectUnrepresentable(s.Items, schemaPath+"{}", meta, out)
	case model.SchemaKindPrimitive:
		// representable leaf
	}
}

// sortFindings orders findings by artifact, then operation, then position, so the
// aggregated error message is deterministic.
func sortFindings(f []RepresentabilityFinding) {
	sort.Slice(f, func(i, j int) bool {
		a, b := f[i], f[j]
		switch {
		case a.ArtifactName != b.ArtifactName:
			return a.ArtifactName < b.ArtifactName
		case a.ArtifactKind != b.ArtifactKind:
			return a.ArtifactKind < b.ArtifactKind
		case a.OperationId != b.OperationId:
			return a.OperationId < b.OperationId
		case a.SchemaPath != b.SchemaPath:
			return a.SchemaPath < b.SchemaPath
		default:
			return a.Kind < b.Kind
		}
	})
}

func sortedSchemaKeys(m map[string]*model.Schema) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func representabilityReason(k model.SchemaKind) string {
	switch k {
	case model.SchemaKindVariant:
		return "oneOf/anyOf has no Terraform Plugin Framework equivalent (variant)"
	case model.SchemaKindRefCycle:
		return "$ref is circular or exceeds --max-depth (ref_cycle)"
	case model.SchemaKindUnsupported:
		return "no Terraform-representable type or structure (unsupported)"
	default:
		return string(k)
	}
}
