package emit

import (
	"strconv"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestExampleStringsRoundTripThroughGoAndHCL(t *testing.T) {
	for _, value := range []string{"${var.secret}", "%{ if true }literal%{ endif }", "quotes \" and slash \\", "control\a\b\v\f\x01\n\r\t", "`code`", "snowman ☃"} {
		t.Run(strconv.Quote(value), func(t *testing.T) {
			scenario := collectionScenario(t, &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}, value)
			view, err := BuildExampleTestView(scenario, SchemaView{Attributes: []AttrView{{TFName: "field", Required: true}}}, map[string]string{"field": "field"})
			if err != nil {
				t.Fatal(err)
			}
			config, err := strconv.Unquote(view.Steps[0].ConfigLiteral)
			if err != nil {
				t.Fatal(err)
			}
			parsed, diagnostics := hclsyntax.ParseConfig([]byte(config), "example.tf", hcl.Pos{Line: 1, Column: 1})
			if diagnostics.HasErrors() {
				t.Fatal(diagnostics.Error())
			}
			expression := parsed.Body.(*hclsyntax.Body).Blocks[0].Body.Attributes["field"].Expr
			actual, diagnostics := expression.Value(nil)
			if diagnostics.HasErrors() {
				t.Fatal(diagnostics.Error())
			}
			if actual.AsString() != value {
				t.Fatalf("round trip = %q, want %q", actual.AsString(), value)
			}
		})
	}
}
