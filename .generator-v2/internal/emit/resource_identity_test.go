package emit

import (
	"errors"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestCreateIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, typ, format, expression, diagnostic string
		enumRef                                   string
		enum                                      bool
		required, readOnly                        bool
	}{
		{name: "required string", typ: "string", required: true, expression: "state.Name.ValueString()"},
		{name: "required integer", typ: "integer", format: "int64", required: true, expression: "parsedBodyId"},
		{name: "required UUID", typ: "string", format: "uuid", required: true, expression: "parsedBodyId"},
		{name: "named enum", typ: "string", required: true, enum: true, enumRef: "IdentityKind", expression: "*parsedBodyId"},
		{name: "anonymous enum", typ: "string", required: true, enum: true, diagnostic: "unsupported type or enum"},
		{name: "UUID enum", typ: "string", format: "uuid", required: true, enum: true, enumRef: "IdentityKind", diagnostic: "unsupported type or enum"},
		{name: "numeric enum", typ: "integer", required: true, enum: true, enumRef: "IdentityKind", diagnostic: "unsupported type or enum"},
		{name: "optional", typ: "string", diagnostic: "optional create data.id is unsupported"},
		{name: "read only", typ: "string", required: true, readOnly: true},
		{name: "unsupported", typ: "string", format: "binary", required: true, diagnostic: "outside scalar-first support"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			op := identityResourceOperation()
			data := op.ResolvedGroup.Create.RequestSchema.Properties["data"]
			id := prim(tc.typ, "Caller-assigned ID.")
			id.Format, id.ReadOnly = tc.format, tc.readOnly
			if tc.enum {
				id.Enum = []string{"value"}
				id.RefName = tc.enumRef
			}
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
			if tc.enum && tc.diagnostic == "" {
				want := "datadogV2.NewIdentityKindFromValue(state.Name.ValueString())"
				if view.Create.Envelope.IDPrep == nil || view.Create.Envelope.IDPrep.ParseCall != want {
					t.Fatalf("expected validating enum constructor %s, got %+v", want, view.Create.Envelope.IDPrep)
				}
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
			assignsID := false
			for _, assignment := range view.State.Assignments {
				assignsID = assignsID || assignment.LHS == "state.ID"
			}
			if !assignsID {
				t.Fatalf("response assigns identity = %v, readOnly = %v", assignsID, tc.readOnly)
			}
			if !tc.readOnly && !strings.Contains(string(source), "stringvalidator.LengthAtLeast(1)") {
				t.Fatal("missing nonempty identity validation")
			}
			if !view.Schema.IncludeResourceID {
				t.Fatal("incorrect computed ID schema")
			}
			if tc.typ != "integer" && view.Update.Envelope.IDExpr != "state.ID.ValueString()" {
				t.Fatal("update ID binding changed")
			}
		})
	}
}

func TestCreateIdentityEnvelopeDiagnostic(t *testing.T) {
	op := identityResourceOperation()
	data := op.ResolvedGroup.Create.RequestSchema.Properties["data"]
	data.Properties["id"] = prim("string", "Caller-assigned ID.")
	data.Required = append(data.Required, "id")
	data.RefName = ""
	art, err := model.BuildArtifact(op)
	if err != nil {
		t.Fatal(err)
	}
	_, err = BuildResourceView(art)
	var unsupported *UnsupportedEmitError
	if !errors.As(err, &unsupported) {
		t.Fatalf("expected unsupported envelope, got %v", err)
	}
	if len(unsupported.Nodes) != 1 || !strings.Contains(unsupported.Nodes[0].Reason, "inline") {
		t.Fatalf("expected one actionable envelope diagnostic, got %v", err)
	}
}

func identityResourceOperation() *model.Operation {
	op := incidentTypeResourceOperation(true)
	for _, schema := range []*model.Schema{op.ResolvedGroup.Create.RequestSchema, op.ResolvedGroup.Update.RequestSchema, op.ResolvedGroup.Read.ResponseSchema} {
		attrs := schema.Properties["data"].Properties["attributes"]
		delete(attrs.Properties, "name")
		attrs.Required = nil
	}
	return op
}

func TestCreateIdentityNameCollision(t *testing.T) {
	op := incidentTypeResourceOperation(true)
	data := op.ResolvedGroup.Create.RequestSchema.Properties["data"]
	data.Properties["id"] = prim("string", "ID")
	data.Required = append(data.Required, "id")
	art, err := model.BuildArtifact(op)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = BuildResourceView(art); err == nil || !strings.Contains(err.Error(), "reserves name") {
		t.Fatalf("expected name collision, got %v", err)
	}
}

func TestCreateIdentityRequiresResponseID(t *testing.T) {
	op := identityResourceOperation()
	data := op.ResolvedGroup.Create.RequestSchema.Properties["data"]
	data.Properties["id"] = prim("string", "ID")
	data.Required = append(data.Required, "id")
	delete(op.ResolvedGroup.Read.ResponseSchema.Properties["data"].Properties, "id")
	art, err := model.BuildArtifact(op)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = BuildResourceView(art); err == nil || !strings.Contains(err.Error(), "requires a response data.id") {
		t.Fatalf("expected response ID diagnostic, got %v", err)
	}
}

func TestUpdateIdentityEnum(t *testing.T) {
	op := identityResourceOperation()
	id := op.ResolvedGroup.Update.RequestSchema.Properties["data"].Properties["id"]
	id.RefName, id.Enum = "IdentityKind", []string{"value"}
	art, err := model.BuildArtifact(op)
	if err != nil {
		t.Fatal(err)
	}
	view, err := BuildResourceView(art)
	if err != nil {
		t.Fatal(err)
	}
	envelope := view.Update.Envelope
	if envelope.IDExpr != "*parsedBodyId" || envelope.IDPrep == nil || envelope.IDPrep.ParseCall != "datadogV2.NewIdentityKindFromValue(state.ID.ValueString())" {
		t.Fatalf("update must validate its named enum independently of the string path: %+v", envelope)
	}
}
