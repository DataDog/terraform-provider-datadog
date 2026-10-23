package emit

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Example-backed test views
//
// An example-backed test and its cassette are rendered from one scenario, so
// the request the test provokes and the request the cassette expects stay the
// same by construction. This file builds the parts common to every such test:
// its identity, its ownership header, the replay-only guard, and the HCL and
// checks for each step.
//
// The scenario carries values keyed by their API paths, because the cassette
// package must not know Terraform naming. Mapping those to attributes is this
// package's job — the emitter already owns the JSON:API envelope flattening
// that makes `data.attributes.name` into the attribute `name`, and keeping the
// mapping in one place is what stops the test and the cassette drifting.
// ----------------------------------------------------------------------------

// exampleTestView is the render context for an example-backed test.
type exampleTestView struct {
	// Marker is the ownership header. The writer replaces a file carrying it
	// and never one without, so it is what separates regenerating tfgen's own
	// output from overwriting a hand-recorded fixture.
	Marker string
	// FuncName is the Go test function, which doubles as the cassette basename
	// because the replay harness derives the cassette from t.Name().
	FuncName string
	// ConfigFunc is the config-helper function stem; step helpers append an
	// index when there is more than one step.
	ConfigFunc string
	// ResourceType is the provider-prefixed Terraform type, e.g.
	// "datadog_integration_twilio_account".
	ResourceType string
	// Steps are the ordered Terraform steps.
	Steps []exampleStepView
}

// exampleStepView is one apply step's configuration and assertions.
type exampleStepView struct {
	// ConfigFunc is this step's config-helper name.
	ConfigFunc string
	// ConfigBody is the HCL between the helper's backticks, built in Go so its
	// whitespace is exact — gofmt does not touch raw-string contents.
	ConfigBody string
	// ConfigLiteral embeds the HCL safely in Go, including examples containing backticks.
	ConfigLiteral string
	// Checks are the resource.TestCheckResourceAttr expressions asserted after
	// apply, as rendered Go argument source.
	Checks []string
}

// BuildExampleTestView derives the render context for a resource's
// example-backed test from its scenario and schema.
//
// Checks assert state only. A check that called the API would add an
// interaction the scenario did not plan for, and the cassette is the oracle
// here — there is nothing for a live existence check to add.
func BuildExampleTestView(
	scenario *model.GeneratedTestScenario,
	schema SchemaView,
	apiPaths map[string]string,
) (exampleTestView, error) {
	if err := scenario.Validate(); err != nil {
		return exampleTestView{}, err
	}

	resourceType := "datadog_" + scenario.ArtifactName
	view := exampleTestView{
		Marker:       model.GeneratedMarker,
		FuncName:     scenario.TestFuncName,
		ConfigFunc:   testHelperName(scenario.TestFuncName, "Config"),
		ResourceType: resourceType,
	}

	// A Terraform configuration is the complete desired state, while an update
	// request is the API's sparse PATCH delta. Rendering a later step from its
	// delta alone would drop required attributes the first step set, and the
	// plan would fail validation before any interaction was replayed. So each
	// step's configuration is every prior step overlaid with its own values.
	desired := map[string]model.MaterializedValue{}
	for index, step := range scenario.Steps {
		stepView := exampleStepView{ConfigFunc: view.ConfigFunc}
		if len(scenario.Steps) > 1 {
			stepView.ConfigFunc = fmt.Sprintf("%sStep%d", view.ConfigFunc, index+1)
		}
		model.OverlayMaterializedValues(desired, stepValues(step))
		values := maps.Clone(desired)
		stepView.ConfigBody = renderExampleConfig(resourceType, schema, values, apiPaths)
		stepView.ConfigLiteral = "`" + stepView.ConfigBody + "`"
		if strings.ContainsRune(stepView.ConfigBody, '`') {
			stepView.ConfigLiteral = strconv.Quote(stepView.ConfigBody)
		}
		stepView.Checks = renderExampleChecks(scenario.TerraformAddress, schema, values, apiPaths)
		view.Steps = append(view.Steps, stepView)
	}
	return view, nil
}

// testHelperName derives a generated helper's name from the test function,
// lower-casing the leading Test so the helper stays unexported. A name that
// does not follow the convention is prefixed rather than sliced, so this can
// never panic on one.
func testHelperName(testFuncName, suffix string) string {
	if rest, ok := strings.CutPrefix(testFuncName, "Test"); ok {
		return "test" + rest + suffix
	}
	return "testAcc" + testFuncName + suffix
}

// stepValues indexes a step's materialized values by their API path.
func stepValues(step model.ScenarioStep) map[string]model.MaterializedValue {
	out := map[string]model.MaterializedValue{}
	if step.State == nil {
		return out
	}
	for _, value := range step.State.RequestValues {
		out[value.Path] = value
	}
	return out
}

// ----------------------------------------------------------------------------
// Configuration rendering
// ----------------------------------------------------------------------------

// renderExampleConfig writes the HCL for one step by walking the artifact's own
// attribute tree, so the configuration can only contain attributes the
// generated schema actually declares.
func renderExampleConfig(
	resourceType string,
	schema SchemaView,
	values map[string]model.MaterializedValue,
	apiPaths map[string]string,
) string {
	var out strings.Builder
	fmt.Fprintf(&out, "resource %q \"foo\" {\n", resourceType)
	// A generated schema splits top-level leaves from nested objects, so both
	// have to be walked: rendering only Attributes omits every block, which
	// leaves a configuration missing its required arguments.
	writeAttributes(&out, schemaMembers(schema), values, apiPaths, nil, 1)
	out.WriteString("}")
	return out.String()
}

// schemaMembers is a schema's leaves and nested objects together, which is what
// a configuration has to cover. AttrView carries its own children the same way
// at every level, so the walkers below need only this one flattening.
func schemaMembers(schema SchemaView) []AttrView {
	return append(append([]AttrView{}, schema.Attributes...), schema.Blocks...)
}

// nestedMembers is schemaMembers for one block.
func nestedMembers(attr AttrView) []AttrView {
	return append(append([]AttrView{}, attr.Attributes...), attr.Blocks...)
}

// writeAttributes renders one level of the attribute tree. A nested block is
// written only when something inside it has a value: an empty block is a claim
// the scenario never made.
func writeAttributes(
	out *strings.Builder,
	attributes []AttrView,
	values map[string]model.MaterializedValue,
	apiPaths map[string]string,
	prefix []string,
	depth int,
) {
	for _, attr := range attributes {
		path := append(slices.Clone(prefix), attr.TFName)
		if !selectedExampleVariant(attr, path, values, apiPaths) {
			continue
		}
		if attr.Computed && !attr.Optional && !attr.Required {
			continue
		}
		if attr.IsBlock && attr.ListBlock {
			value, ok := lookupValue(values, apiPaths, path)
			if !ok {
				continue
			}
			items, ok := value.Value.([]any)
			if !ok {
				continue
			}
			writeIndent(out, depth)
			fmt.Fprintf(out, "%s = [\n", attr.TFName)
			for i := range items {
				writeIndent(out, depth+1)
				out.WriteString("{\n")
				writeAttributes(out, nestedMembers(attr), values, indexedAPIPaths(apiPaths, path, i), path, depth+2)
				writeIndent(out, depth+1)
				out.WriteString("},\n")
			}
			writeIndent(out, depth)
			out.WriteString("]\n")
			continue
		}
		if attr.IsBlock {
			// Rendered first, then kept only if it produced something: an
			// empty block is a claim the scenario never made. Asking a
			// separate walker whether the block has values would be a second
			// implementation of the skips below, free to drift from them.
			var body strings.Builder
			writeAttributes(&body, nestedMembers(attr), values, apiPaths, path, depth+1)
			if body.Len() == 0 {
				continue
			}
			writeIndent(out, depth)
			fmt.Fprintf(out, "%s = {\n", attr.TFName)
			out.WriteString(body.String())
			writeIndent(out, depth)
			out.WriteString("}\n")
			continue
		}
		if attr.Computed && !attr.Optional && !attr.Required {
			// A purely computed attribute is never configured.
			continue
		}
		// A write-only secret is not configured under its own name: the
		// generated schema replaces it with a <name>_wo attribute carrying the
		// value and a <name>_wo_version rotation trigger, and requires both.
		// The placeholder attribute here is the original, so its value
		// resolves normally and is written under the replacement's name.
		if secret := attr.WriteOnlySecret; secret != nil {
			value, ok := lookupValue(values, apiPaths, path)
			if !ok {
				continue
			}
			writeIndent(out, depth)
			fmt.Fprintf(out, "%s = %s\n", secret.WriteOnlyAttr, hclLiteral(value.Value))
			writeIndent(out, depth)
			fmt.Fprintf(out, "%s = %s\n", secret.TriggerAttr, hclLiteral(writeOnlySecretVersion))
			continue
		}
		value, ok := lookupValue(values, apiPaths, path)
		if !ok {
			continue
		}
		writeIndent(out, depth)
		fmt.Fprintf(out, "%s = %s\n", attr.TFName, hclLiteral(value.Value))
	}
}

func selectedExampleVariant(attr AttrView, path []string, values map[string]model.MaterializedValue, apiPaths map[string]string) bool {
	if !attr.OneOfVariant {
		return true
	}
	value, ok := lookupValue(values, apiPaths, path)
	return ok && value.Variant == attr.TFName
}

// indexedAPIPaths binds a list element's schema paths to its concrete values.
func indexedAPIPaths(paths map[string]string, path []string, index int) map[string]string {
	tfPath := strings.Join(path, ".")
	apiPath := paths[tfPath]
	out := maps.Clone(paths)
	for tfChild, apiChild := range paths {
		if strings.HasPrefix(tfChild, tfPath+".") && strings.HasPrefix(apiChild, apiPath) {
			out[tfChild] = fmt.Sprintf("%s[%d]%s", apiPath, index, strings.TrimPrefix(apiChild, apiPath))
		}
	}
	return out
}

// lookupValue finds the materialized value a Terraform path corresponds to.
//
// The index records which API path each attribute was built from, so this is
// two exact map hits: Terraform path to API path, then API path to value. It
// deliberately is not a suffix match between the two — a normalized name never
// suffix-matches its API spelling, and the value would be dropped from the
// configuration while the cassette still sent it.
func lookupValue(
	values map[string]model.MaterializedValue,
	apiPaths map[string]string,
	path []string,
) (model.MaterializedValue, bool) {
	apiPath, ok := apiPaths[strings.Join(path, ".")]
	if !ok {
		return model.MaterializedValue{}, false
	}
	value, ok := values[apiPath]
	return value, ok
}

// hclLiteral renders a value as HCL. Only the scalar and collection forms
// materialization can produce are handled; anything else is written as a
// quoted string rather than guessed at.
func hclLiteral(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case bool, int, int64, float64:
		// Unquoted scalars render exactly as a check expects them.
		return checkValue(typed)
	case string:
		return string(hclwrite.TokensForValue(cty.StringVal(typed)).Bytes())
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, hclLiteral(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case map[string]any:
		items := make([]string, 0, len(typed))
		for _, key := range slices.Sorted(maps.Keys(typed)) {
			items = append(items, hclLiteral(key)+" = "+hclLiteral(typed[key]))
		}
		return "{" + strings.Join(items, ", ") + "}"
	default:
		return strconv.Quote(checkValue(typed))
	}
}

func writeIndent(out *strings.Builder, depth int) {
	out.WriteString(strings.Repeat("  ", depth))
}

// ----------------------------------------------------------------------------
// Check rendering
// ----------------------------------------------------------------------------

// renderExampleChecks builds one TestCheckResourceAttr per configured leaf,
// plus a check that the resource has an id. Checks read state only — the
// cassette is the oracle, so a live existence check would add an unplanned
// interaction and assert nothing the fixture does not already pin.
func renderExampleChecks(
	address string,
	schema SchemaView,
	values map[string]model.MaterializedValue,
	apiPaths map[string]string,
) []string {
	checks := []string{
		fmt.Sprintf("resource.TestCheckResourceAttrSet(%q, \"id\")", address),
	}
	var leaves []checkLeaf
	collectCheckLeaves(schemaMembers(schema), values, apiPaths, nil, &leaves)
	slices.SortFunc(leaves, func(a, b checkLeaf) int { return strings.Compare(a.path, b.path) })
	for _, leaf := range leaves {
		// A replaced secret is asserted as the placeholder, which is what the
		// configuration sends and therefore what state will hold.
		checks = append(checks, fmt.Sprintf("resource.TestCheckResourceAttr(%q, %q, %q)",
			address, leaf.path, checkValue(leaf.value)))
	}
	return checks
}

// collectCheckLeaves gathers the Terraform paths of every leaf carrying a
// value. Collections are skipped: asserting a list element by index pins an
// ordering the description does not promise.
func collectCheckLeaves(
	attributes []AttrView,
	values map[string]model.MaterializedValue,
	apiPaths map[string]string,
	prefix []string,
	out *[]checkLeaf,
) {
	for _, attr := range attributes {
		path := append(slices.Clone(prefix), attr.TFName)
		if !selectedExampleVariant(attr, path, values, apiPaths) {
			continue
		}
		if attr.ListBlock {
			continue
		}
		if attr.IsBlock {
			collectCheckLeaves(nestedMembers(attr), values, apiPaths, path, out)
			continue
		}
		// Terraform never persists a write-only value, nor its rotation
		// trigger, so asserting either against state would fail.
		if attr.WriteOnlySecret != nil {
			continue
		}
		value, ok := lookupValue(values, apiPaths, path)
		if !ok || value.Value == nil {
			continue
		}
		if _, isList := value.Value.([]any); isList {
			continue
		}
		if _, isMap := value.Value.(map[string]any); isMap {
			continue
		}
		*out = append(*out, checkLeaf{path: strings.Join(path, "."), value: value.Value})
	}
}

// writeOnlySecretVersion is the rotation trigger every generated test writes.
// The trigger's only job is to change when an operator wants the secret
// resent, so a generated fixture pins it: a value derived from anything that
// varies would rewrite the configuration on every run.
const writeOnlySecretVersion = "1"

// checkLeaf is one asserted attribute: its Terraform path and the value the
// configuration set there. Carrying the value avoids resolving it twice.
type checkLeaf struct {
	path  string
	value any
}

// checkValue renders a value the way Terraform stores it in state, which is
// always as a string.
func checkValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(typed)
	}
}
