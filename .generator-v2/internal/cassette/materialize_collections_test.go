package cassette

import (
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestMaterializedCollectionsRetainSanitizedAggregate(t *testing.T) {
	for _, kind := range []model.SchemaKind{model.SchemaKindArray, model.SchemaKindMap} {
		t.Run(string(kind), func(t *testing.T) {
			var value any = []any{"secret-value"}
			if kind == model.SchemaKindMap {
				value = map[string]any{"key": "secret-value"}
			}
			set, err := MaterializeSet(SelectedSet{Key: SetKey{Role: SetRoleRequest}, Candidate: &model.ExampleCandidate{Value: value}}, &model.Schema{Kind: kind, Items: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Sensitive: true}})
			if err != nil {
				t.Fatal(err)
			}
			for _, leaf := range set.Values {
				if leaf.Path == "(root)" {
					if strings.Contains(fmt.Sprint(leaf.Value), "secret-value") {
						t.Fatal("aggregate retained an unsanitized item")
					}
					return
				}
			}
			t.Fatal("collection aggregate missing")
		})
	}
}
