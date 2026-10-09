package cassette

import (
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestOptionalFallbackContainers(t *testing.T) {
	for _, tc := range []struct {
		name                                                     string
		required, partial, declared, defaulted, union, sensitive bool
		wantError                                                bool
	}{
		{name: "absent object"},
		{name: "absent union", union: true},
		{name: "required object", required: true, wantError: true},
		{name: "partial object", partial: true, wantError: true},
		{name: "explicit empty object", declared: true, wantError: true},
		{name: "defaulted object", defaulted: true},
		{name: "sensitive fallback container", partial: true, sensitive: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			child := &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}
			if tc.defaulted {
				child.Enum = []string{"default"}
			}
			nested := &model.Schema{Kind: model.SchemaKindObject, Sensitive: tc.sensitive, Required: []string{"required"}, Properties: map[string]*model.Schema{
				"required": child,
				"other":    {Kind: model.SchemaKindPrimitive, Type: "string"},
			}}
			if tc.union {
				nested = &model.Schema{Kind: model.SchemaKindOneOf, Variants: []*model.Schema{nested, {Kind: model.SchemaKindPrimitive, Type: "string"}}}
			}
			schema := &model.Schema{Kind: model.SchemaKindObject, Required: []string{"name"}, Properties: map[string]*model.Schema{
				"name": {Kind: model.SchemaKindPrimitive, Type: "string"}, "nested": nested,
			}}
			if tc.required {
				schema.Required = append(schema.Required, "nested")
			}
			fallbacks := []*model.ExampleCandidate{{Value: "example", Location: model.ExampleLocation{PropertyPath: "name"}}}
			if tc.partial {
				fallbacks = append(fallbacks, &model.ExampleCandidate{Value: "private-value", Location: model.ExampleLocation{PropertyPath: "nested.other"}})
			}
			if tc.declared {
				fallbacks = append(fallbacks, &model.ExampleCandidate{Value: map[string]any{}, Location: model.ExampleLocation{PropertyPath: "nested"}})
			}
			got, err := MaterializeSet(SelectedSet{Key: SetKey{Role: SetRoleRequest}, Fallbacks: fallbacks}, schema)
			if tc.wantError {
				if err == nil {
					t.Fatal("expected incomplete example to be rejected")
				}
				if strings.Contains(err.Error(), "private-value") {
					t.Fatal("diagnostic leaked sensitive value")
				}
				if tc.sensitive && !strings.Contains(err.Error(), "no safe replacement") {
					t.Fatalf("expected sensitive rejection: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			body := got.Body.(map[string]any)
			_, present := body["nested"]
			if present != tc.defaulted {
				t.Fatalf("unexpected optional container presence: %#v", body)
			}
		})
	}
}
