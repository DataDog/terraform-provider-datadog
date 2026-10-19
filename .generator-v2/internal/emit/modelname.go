package emit

import (
	"fmt"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// modelNamer is the single authority for the Go struct name of every model a
// generated artifact declares. Every name is built the same way:
//
//	<artifact base><stem>Model
//
// Both halves are load-bearing, and each closes a distinct collision scope.
//
// The **artifact base** (dsGoName, e.g. "datadogActionConnection") closes the
// cross-artifact scope. Every generated data source lands in package fwprovider, so
// two artifacts that share an OpenAPI component would otherwise each declare the
// same struct and neither could be compiled alongside the other. That is not
// hypothetical: 54 of the components in the mini-OAS corpus appear in more than one
// slice (GoogleMeetConfigurationReference and IncidentTypeAttributes in three
// each). Prefixing also makes nested models unexported, which is the provider's own
// convention for data-source models — hostListMetadataModel, rowModel — so this
// moves toward the hand-written set rather than away from it.
//
// The **stem** closes the within-file scope. It prefers the OpenAPI component name
// that supplied the object schema (Attribute.ModelRefName) and falls back to the
// enclosing stem plus the property name when that schema was inline. This is the
// rule the SDK generator's child_models() applies — get_name(schema) or
// alternative_name, restarting at every $ref — and the rule oneOf envelope naming
// already applies through OneOfSpec.Name. Deriving from the property name alone is
// what produced two different TagFiltersModel structs in one file and made
// integration_aws_external_id uncompilable.
//
// A component-derived stem also makes two uses of one component in one artifact
// converge on a single struct, which dedupeModels then collapses.
type modelNamer struct {
	// base is the artifact's Go name stem, already lowerCamel (dsGoName).
	base string
}

// qualify scopes an unqualified model name to the artifact. It is the one place an
// artifact prefix is applied, so the model layer's OneOfEnvelope.GoModel can stay
// artifact-agnostic: the model layer does not know which artifact is rendering a
// reusable union, and emit does.
func (n modelNamer) qualify(unqualified string) string {
	return n.base + upperFirst(unqualified)
}

// nested returns the qualified struct name for the model backing a, plus the stem
// its own children accumulate from. stem is the enclosing model's stem ("" at the
// artifact root).
func (n modelNamer) nested(stem string, a *model.Attribute) (name, childStem string) {
	childStem = nestedStem(stem, a)
	return n.qualify(childStem + "Model"), childStem
}

// nestedStem derives the stem of a's model: the component that supplied its object
// schema when there is one — restarting the accumulation, exactly as child_models()
// prefers a $ref name over the parent-derived one — else the enclosing stem plus
// this attribute's property name.
//
// The two branches are spelled differently on purpose. A component name is already
// a Go-style identifier (it is what the SDK names its own struct), so it is taken
// verbatim, which is also what the oneOf envelope path does with OneOfSpec.Name.
// Running it through model.SdkName would mangle acronyms and do so inconsistently —
// AWSLogSourceTagFilter becomes AwsLogSourceTagFilter while XRayServicesList
// survives — for no gain. A property name, by contrast, arrives as snake_case and
// must be converted.
func nestedStem(stem string, a *model.Attribute) string {
	if a.ModelRefName != "" {
		return a.ModelRefName
	}
	return stem + model.SdkName(tfNameOf(a.Path))
}

// stemOf recovers the naming stem from an unqualified model name produced by the
// model layer (model.oneOfModelName appends "Model" to the envelope or variant
// identity). Children of that struct accumulate inline names from the stem, so they
// must not inherit the suffix.
func stemOf(unqualifiedModel string) string {
	return strings.TrimSuffix(unqualifiedModel, "Model")
}

// dedupeModels collapses model structs that share a name, and fails when two of
// them disagree about their fields.
//
// Convergence is expected and correct: one OpenAPI component reached from two
// properties in the same artifact yields one stem, so it must yield one declaration
// rather than a redeclaration. Divergence means the naming rule produced one name
// for two shapes, which is the defect this guards against — fail the artifact
// naming both, because the alternatives are emitting code that cannot compile or
// silently dropping one shape's fields, and the silent option is not acceptable.
func dedupeModels(models []ModelStructView) ([]ModelStructView, []UnsupportedNode) {
	var out []ModelStructView
	seen := make(map[string]ModelStructView, len(models))
	var conflicts []UnsupportedNode

	for _, m := range models {
		prev, ok := seen[m.Name]
		if !ok {
			seen[m.Name] = m
			out = append(out, m)
			continue
		}
		if fieldsEqual(prev.Fields, m.Fields) {
			continue
		}
		conflicts = append(conflicts, UnsupportedNode{
			Path: m.Name,
			Reason: fmt.Sprintf(
				"two different object shapes both derive the model struct %q (%s vs %s); "+
					"the generated file would redeclare the type. Give the schemas distinct "+
					"OpenAPI components so each names its own model",
				m.Name, fieldList(prev.Fields), fieldList(m.Fields)),
		})
	}

	sort.SliceStable(conflicts, func(i, j int) bool { return conflicts[i].Path < conflicts[j].Path })
	return out, conflicts
}

func fieldsEqual(a, b []ModelFieldView) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].GoField != b[i].GoField || a[i].GoType != b[i].GoType || a[i].TFName != b[i].TFName {
			return false
		}
	}
	return true
}

// fieldList renders a model's fields for a collision diagnostic, so the message
// shows which two shapes collided rather than only that they did.
func fieldList(fields []ModelFieldView) string {
	if len(fields) == 0 {
		return "{}"
	}
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = f.GoField + " " + f.GoType
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
