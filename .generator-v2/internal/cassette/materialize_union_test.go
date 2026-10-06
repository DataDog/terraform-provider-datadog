package cassette

import (
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func exampleUnionSchema() *model.Schema {
	variant := func(mode string) *model.Schema {
		return &model.Schema{Kind: model.SchemaKindObject, Required: []string{"mode", "secret"}, Properties: map[string]*model.Schema{
			"mode":   {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{mode}},
			"secret": {Kind: model.SchemaKindPrimitive, Type: "string", Sensitive: true},
		}}
	}
	return &model.Schema{Kind: model.SchemaKindOneOf, OneOf: &model.OneOfSpec{Variants: []model.OneOfVariant{
		{TFName: "basic", Schema: variant("basic")}, {TFName: "token", Schema: variant("token")},
	}}}
}

func TestMaterializeUnionSelectsAndSanitizesBranch(t *testing.T) {
	set, err := MaterializeSet(SelectedSet{Key: SetKey{Role: SetRoleRequest}, Candidate: &model.ExampleCandidate{Value: map[string]any{"mode": "token", "secret": "private-example"}}}, exampleUnionSchema())
	if err != nil {
		t.Fatal(err)
	}
	if set.Body.(map[string]any)["secret"] != model.RedactedPlaceholder {
		t.Fatal("union secret was not replaced")
	}
	for _, value := range set.Values {
		if value.Path == "(root)" && value.Variant == "token" {
			return
		}
	}
	t.Fatal("chosen union branch missing")
}

func TestMaterializeUnionRejectsAmbiguousExample(t *testing.T) {
	schema := exampleUnionSchema()
	schema.OneOf.Variants[1].Schema = schema.OneOf.Variants[0].Schema
	_, err := MaterializeSet(SelectedSet{Key: SetKey{Role: SetRoleRequest}, Candidate: &model.ExampleCandidate{Value: map[string]any{"mode": "basic", "secret": "private-example"}}}, schema)
	if err == nil || !strings.Contains(err.Error(), "more than one branch") {
		t.Fatalf("expected ambiguous union, got %v", err)
	}
	if strings.Contains(err.Error(), "private-example") {
		t.Fatal("diagnostic leaked example")
	}
}
