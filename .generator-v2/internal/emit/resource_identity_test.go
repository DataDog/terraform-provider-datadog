package emit

import (
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestCreateIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, typ, format, expression, diagnostic string
		required, readOnly                        bool
	}{
		{name: "required string", typ: "string", required: true, expression: "state.ID.ValueString()"},
		{name: "required integer", typ: "integer", format: "int64", required: true, expression: "parsedBodyId"},
		{name: "required UUID", typ: "string", format: "uuid", required: true, expression: "parsedBodyId"},
		{name: "optional", typ: "string", diagnostic: "optional create data.id is unsupported"},
		{name: "read only", typ: "string", required: true, readOnly: true},
		{name: "unsupported", typ: "string", format: "binary", required: true, diagnostic: "outside scalar-first support"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			op := incidentTypeResourceOperation(true)
			data := op.ResolvedGroup.Create.RequestSchema.Properties["data"]
			id := prim(tc.typ, "Caller-assigned ID.")
			id.Format, id.ReadOnly = tc.format, tc.readOnly
			data.Properties["id"] = id
			if tc.typ == "integer" {
				op.ResolvedGroup.Read.ResponseSchema.Properties["data"].Properties["id"] = prim("integer", "ID")
				op.ResolvedGroup.Update.RequestSchema.Properties["data"].Properties["id"] = prim("integer", "ID")
			}
			if tc.required {
				data.Required = append(data.Required, "id")
			}
			art, err := model.BuildArtifact(op)
			if err != nil {
				t.Fatal(err)
			}
			if art.Lifecycle.Create.RequestIDRequired != tc.required {
				t.Fatal("lost request ID requiredness")
			}
			view, err := BuildResourceView(art)
			if tc.diagnostic != "" {
				if err == nil || !strings.Contains(err.Error(), tc.diagnostic) {
					t.Fatalf("expected %q, got %v", tc.diagnostic, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if view.Create.Envelope.IDExpr != tc.expression {
				t.Fatalf("create ID expression: %q", view.Create.Envelope.IDExpr)
			}
			source, err := RenderResource(view)
			if err != nil {
				t.Fatal(err)
			}
			if !tc.readOnly && !strings.Contains(string(source), "stringplanmodifier.RequiresReplace()") {
				t.Fatal("missing replacement modifier")
			}
			if view.Schema.IncludeResourceID != tc.readOnly {
				t.Fatal("incorrect computed ID schema")
			}
			if tc.typ != "integer" && view.Update.Envelope.IDExpr != "state.ID.ValueString()" {
				t.Fatal("update ID binding changed")
			}
		})
	}
}
