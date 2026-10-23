package emit

import (
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestExampleUnionConfigSelectsExactlyOneBranch(t *testing.T) {
	branch := func(mode string) *model.Schema {
		return &model.Schema{Kind: model.SchemaKindObject, Required: []string{"mode"}, Properties: map[string]*model.Schema{
			"mode": {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{mode}},
			"note": {Kind: model.SchemaKindPrimitive, Type: "string"},
		}}
	}
	schema := &model.Schema{Kind: model.SchemaKindOneOf, OneOf: &model.OneOfSpec{Variants: []model.OneOfVariant{
		{TFName: "basic", Schema: branch("basic")}, {TFName: "token", Schema: branch("token")},
	}}}
	scenario := collectionScenario(t, schema, map[string]any{"mode": "basic", "note": "remove-me"}, map[string]any{"mode": "token"})
	viewSchema := SchemaView{Blocks: []AttrView{{TFName: "field", IsBlock: true, Required: true}}}
	paths := map[string]string{"field": "field"}
	for _, name := range []string{"basic", "token"} {
		viewSchema.Blocks[0].Blocks = append(viewSchema.Blocks[0].Blocks, AttrView{TFName: name, OneOfVariant: true, IsBlock: true, Optional: true, Attributes: []AttrView{{TFName: "mode", Required: true}, {TFName: "note", Optional: true}}})
		paths["field."+name] = "field"
		paths["field."+name+".mode"] = "field.mode"
		paths["field."+name+".note"] = "field.note"
	}
	view, err := BuildExampleTestView(scenario, viewSchema, paths)
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range []string{"basic", "token"} {
		other := []string{"token", "basic"}[i]
		config := view.Steps[i].ConfigBody
		if !strings.Contains(config, name+" = {") || strings.Contains(config, other+" = {") {
			t.Fatalf("incorrect union configuration: %s", config)
		}
		for _, check := range view.Steps[i].Checks {
			if strings.Contains(check, "field."+other) {
				t.Fatalf("inactive branch asserted: %s", check)
			}
		}
	}
	if strings.Contains(view.Steps[1].ConfigBody, "remove-me") {
		t.Fatal("previous branch field retained")
	}
}
