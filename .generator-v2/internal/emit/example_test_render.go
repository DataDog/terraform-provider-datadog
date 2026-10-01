package emit

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

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
	// TerraformAddress is the address checks assert against.
	TerraformAddress string
	// CassettePath and FreezePath are where the fixture and its time companion
	// live, surfaced in the file's header so a reviewer can find them.
	CassettePath string
	FreezePath   string
	// FreezeTime is the fixed instant the cassette was built against, in the
	// RFC3339Nano form the freeze companion carries.
	FreezeTime string
	// InteractionCount is the number of interactions the cassette holds, named
	// in the header because a replay failure is usually a count mismatch.
	InteractionCount int
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
) (exampleTestView, error) {
	if err := scenario.Validate(); err != nil {
		return exampleTestView{}, err
	}

	resourceType := "datadog_" + scenario.ArtifactName
	view := exampleTestView{
		Marker:           model.GeneratedMarker,
		FuncName:         scenario.TestFuncName,
		ConfigFunc:       testHelperName(scenario.TestFuncName, "Config"),
		ResourceType:     resourceType,
		TerraformAddress: scenario.TerraformAddress,
		CassettePath:     fmt.Sprintf("cassettes/%s.yaml", scenario.CassetteBaseName),
		FreezePath:       fmt.Sprintf("cassettes/%s.freeze", scenario.CassetteBaseName),
		FreezeTime:       scenario.FreezeTime.UTC().Format(freezeTimeLayout),
		InteractionCount: len(scenario.Interactions),
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
		for path, value := range stepValues(step) {
			desired[path] = value
		}
		values := maps.Clone(desired)
		stepView.ConfigBody = renderExampleConfig(resourceType, schema, values)
		stepView.Checks = renderExampleChecks(scenario.TerraformAddress, schema, values)
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

// freezeTimeLayout matches what the provider harness writes to a freeze file.
const freezeTimeLayout = "2006-01-02T15:04:05.999999999Z07:00"

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
func renderExampleConfig(resourceType string, schema SchemaView, values map[string]model.MaterializedValue) string {
	var out strings.Builder
	fmt.Fprintf(&out, "resource %q \"foo\" {\n", resourceType)
	writeAttributes(&out, schema.Attributes, values, nil, 1)
	out.WriteString("}")
	return out.String()
}

// writeAttributes renders one level of the attribute tree. A nested block is
// written only when something inside it has a value: an empty block is a claim
// the scenario never made.
func writeAttributes(
	out *strings.Builder,
	attributes []AttrView,
	values map[string]model.MaterializedValue,
	prefix []string,
	depth int,
) {
	for _, attr := range attributes {
		path := append(slices.Clone(prefix), attr.TFName)
		if attr.IsBlock {
			if !hasValues(attr.Attributes, values) {
				continue
			}
			writeIndent(out, depth)
			fmt.Fprintf(out, "%s = {\n", attr.TFName)
			writeAttributes(out, attr.Attributes, values, path, depth+1)
			writeIndent(out, depth)
			out.WriteString("}\n")
			continue
		}
		if attr.Computed && !attr.Optional && !attr.Required {
			// A purely computed attribute is never configured.
			continue
		}
		value, ok := lookupValue(values, attr)
		if !ok {
			continue
		}
		writeIndent(out, depth)
		fmt.Fprintf(out, "%s = %s\n", attr.TFName, hclLiteral(value.Value))
	}
}

// hasValues reports whether any leaf under a block has a value, so an empty
// block is omitted rather than rendered.
func hasValues(attributes []AttrView, values map[string]model.MaterializedValue) bool {
	for _, attr := range attributes {
		if attr.IsBlock {
			if hasValues(attr.Attributes, values) {
				return true
			}
			continue
		}
		if _, ok := lookupValue(values, attr); ok {
			return true
		}
	}
	return false
}

// lookupValue finds the materialized value an attribute corresponds to.
//
// The attribute carries the API path it was built from, recorded by
// apiPathIndex while both spellings were in hand, so this is an exact lookup.
// It deliberately is not a suffix match: SnakeCase turns `backgroundColor` into
// `background_color` and `twilio-messages-logs` into `twilio_messages_logs`,
// neither of which suffix-matches its API path, and a near-miss there drops the
// value from the configuration while the cassette still sends it.
//
// An empty APIPath means the attribute has no API counterpart, so nothing can
// resolve to it.
func lookupValue(values map[string]model.MaterializedValue, attr AttrView) (model.MaterializedValue, bool) {
	if attr.APIPath == "" {
		return model.MaterializedValue{}, false
	}
	value, ok := values[attr.APIPath]
	return value, ok
}

// hclLiteral renders a value as HCL. Only the scalar and collection forms
// materialization can produce are handled; anything else is written as a
// quoted string rather than guessed at.
func hclLiteral(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case string:
		return strconv.Quote(typed)
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, hclLiteral(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	default:
		return strconv.Quote(fmt.Sprint(typed))
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
func renderExampleChecks(address string, schema SchemaView, values map[string]model.MaterializedValue) []string {
	checks := []string{
		fmt.Sprintf("resource.TestCheckResourceAttrSet(%q, \"id\")", address),
	}
	var leaves []checkLeaf
	collectCheckLeaves(schema.Attributes, values, nil, &leaves)
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
	prefix []string,
	out *[]checkLeaf,
) {
	for _, attr := range attributes {
		path := append(slices.Clone(prefix), attr.TFName)
		if attr.IsBlock {
			collectCheckLeaves(attr.Attributes, values, path, out)
			continue
		}
		value, ok := lookupValue(values, attr)
		if !ok || value.Value == nil {
			continue
		}
		if _, isList := value.Value.([]any); isList {
			continue
		}
		*out = append(*out, checkLeaf{path: strings.Join(path, "."), value: value.Value})
	}
}

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
