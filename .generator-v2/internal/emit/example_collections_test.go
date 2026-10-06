package emit

import (
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/cassette"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func collectionScenario(t *testing.T, schema *model.Schema, values ...any) *model.GeneratedTestScenario {
	t.Helper()
	scenario := &model.GeneratedTestScenario{ArtifactName: "widget", TestFuncName: "TestAccWidget", TerraformAddress: "datadog_widget.foo"}
	for _, value := range values {
		set, err := cassette.MaterializeSet(cassette.SelectedSet{Key: cassette.SetKey{Role: cassette.SetRoleRequest}, Candidate: &model.ExampleCandidate{Value: map[string]any{"field": value}}}, &model.Schema{Kind: model.SchemaKindObject, Required: []string{"field"}, Properties: map[string]*model.Schema{"field": schema}})
		if err != nil {
			t.Fatal(err)
		}
		scenario.Steps = append(scenario.Steps, model.ScenarioStep{State: &model.MaterializedConfiguration{RequestValues: set.Values}})
	}
	return scenario
}

func TestExampleCollectionLiterals(t *testing.T) {
	for _, tc := range []struct {
		name  string
		kind  model.SchemaKind
		value any
		want  string
	}{
		{"array", model.SchemaKindArray, []any{"alpha", "beta"}, `field = ["alpha", "beta"]`},
		{"empty array", model.SchemaKindArray, []any{}, `field = []`},
		{"map", model.SchemaKindMap, map[string]any{"team": "core", "env": "test"}, `field = {"env" = "test", "team" = "core"}`},
		{"empty map", model.SchemaKindMap, map[string]any{}, `field = {}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scenario := collectionScenario(t, &model.Schema{Kind: tc.kind, Items: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}}, tc.value)
			view, err := BuildExampleTestView(scenario, SchemaView{Attributes: []AttrView{{TFName: "field", Required: true}}}, map[string]string{"field": "field"})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(view.Steps[0].ConfigBody, tc.want) {
				t.Fatalf("missing collection: %s", view.Steps[0].ConfigBody)
			}
			if len(view.Steps[0].Checks) != 1 {
				t.Fatalf("collection incorrectly asserted as scalar: %v", view.Steps[0].Checks)
			}
		})
	}
}

func TestExampleNestedListReplacement(t *testing.T) {
	str := &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}
	schema := &model.Schema{Kind: model.SchemaKindArray, Items: &model.Schema{Kind: model.SchemaKindObject, Required: []string{"itemName"}, Properties: map[string]*model.Schema{"itemName": str, "note": str}}}
	scenario := collectionScenario(t, schema,
		[]any{map[string]any{"itemName": "first", "note": "remove-me"}, map[string]any{"itemName": "second"}},
		[]any{map[string]any{"itemName": "changed"}}, []any{})
	view, err := BuildExampleTestView(scenario, SchemaView{Blocks: []AttrView{{TFName: "field", IsBlock: true, ListBlock: true, Required: true, Attributes: []AttrView{{TFName: "item_name", Required: true}, {TFName: "note", Optional: true}}}}}, map[string]string{"field": "field", "field.item_name": "field.itemName", "field.note": "field.note"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(view.Steps[0].ConfigBody, `item_name = "second"`) {
		t.Fatal(view.Steps[0].ConfigBody)
	}
	if !strings.Contains(view.Steps[1].ConfigBody, `item_name = "changed"`) || strings.Contains(view.Steps[1].ConfigBody, "remove-me") || strings.Contains(view.Steps[1].ConfigBody, "second") {
		t.Fatal(view.Steps[1].ConfigBody)
	}
	if strings.Contains(view.Steps[2].ConfigBody, "item_name") {
		t.Fatal(view.Steps[2].ConfigBody)
	}
}
