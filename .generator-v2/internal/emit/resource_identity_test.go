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
		{name: "integer response", typ: "integer", format: "int64", required: true, diagnostic: "non-string response IDs are unsupported"},
		{name: "required UUID", typ: "string", format: "uuid", required: true, expression: "parsedBodyId"},
		{name: "named enum", typ: "string", required: true, enum: true, enumRef: "IdentityKind", expression: "*parsedBodyId"},
		{name: "anonymous enum", typ: "string", required: true, enum: true, diagnostic: "unsupported type or enum"},
		{name: "UUID enum", typ: "string", format: "uuid", required: true, enum: true, enumRef: "IdentityKind", diagnostic: "unsupported type or enum"},
		{name: "numeric enum response", typ: "integer", required: true, enum: true, enumRef: "IdentityKind", diagnostic: "non-string response IDs are unsupported"},
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
			if !tc.readOnly && !strings.Contains(string(source), "stringplanmodifier.RequiresReplaceIf(") {
				t.Fatal("missing replacement modifier")
			}
			assignsID := false
			for _, assignment := range view.State.Assignments {
				if assignment.LHS == "state.ID" {
					assignsID = true
					want := "types.StringValue(*id)"
					if tc.format == "uuid" {
						want = "types.StringValue(id.String())"
					}
					if tc.enum {
						want = "types.StringValue(string(*id))"
					}
					if assignment.RHS != want {
						t.Fatalf("response ID assignment: %s", assignment.RHS)
					}
				}
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

// Request conversion is independent of the string-only response identity guard.
func TestIntegerRequestIdentityConversion(t *testing.T) {
	for _, source := range []string{"state.Name", "state.ID"} {
		envelope := &RequestEnvelopeView{}
		call := &model.SDKCall{RequestDeclaresID: true, RequestIDGoType: "int64"}
		if reason := setRequestBodyID(envelope, call, nil, source); reason != "" {
			t.Fatal(reason)
		}
		if envelope.IDExpr != "parsedBodyId" || envelope.IDPrep == nil || envelope.IDPrep.ParseCall != "strconv.ParseInt("+source+".ValueString(), 10, 64)" {
			t.Fatalf("integer request conversion for %s: %+v", source, envelope)
		}
	}
}

func TestResponseIdentityTypes(t *testing.T) {
	for _, kind := range []string{"resource", "data source"} {
		for _, tc := range []struct {
			name, typ, format, want string
			enum                    bool
		}{
			{name: "string", typ: "string", want: "types.StringValue(*id)"},
			{name: "UUID", typ: "string", format: "uuid", want: "types.StringValue(id.String())"},
			{name: "enum", typ: "string", enum: true, want: "types.StringValue(string(*id))"},
			{name: "integer", typ: "integer"},
			{name: "number", typ: "number"},
			{name: "boolean", typ: "boolean"},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				op := incidentTypeOperation()
				if kind == "resource" {
					op = identityResourceOperation()
				}
				schema := op.ResponseSchema
				if kind == "resource" {
					schema = op.ResolvedGroup.Read.ResponseSchema
					// Keep the body schema consistent so the test reaches emission.
					delete(op.ResolvedGroup.Update.RequestSchema.Properties["data"].Properties, "id")
				}
				id := prim(tc.typ, "Response identity")
				id.Format = tc.format
				if tc.enum {
					id.Enum = []string{"value"}
					id.RefName = "IdentityKind"
				}
				schema.Properties["data"].Properties["id"] = id
				schema.Properties["data"].Properties["attributes"].Properties["count"] = prim("integer", "An ordinary numeric payload value")
				art, err := model.BuildArtifact(op)
				if err != nil {
					t.Fatal(err)
				}
				var assignments []StateAssignment
				var models []ModelStructView
				if kind == "resource" {
					view, viewErr := BuildResourceView(art)
					err = viewErr
					assignments, models = view.State.Assignments, view.Models
				} else {
					view, viewErr := BuildDataSourceView(art)
					err = viewErr
					assignments, models = view.State.Assignments, view.Models
				}
				if tc.want == "" {
					var unsupported *UnsupportedEmitError
					if !errors.As(err, &unsupported) || !strings.Contains(err.Error(), "non-string response IDs are unsupported") {
						t.Fatalf("expected response identity diagnostic, got %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if models[0].Fields[0].GoType != "types.String" {
					t.Fatal("identity field must remain a string")
				}
				foundID, foundCount := false, false
				for _, a := range assignments {
					if a.LHS == "state.ID" {
						foundID = true
						if a.RHS != tc.want {
							t.Fatalf("ID expression %q, want %q", a.RHS, tc.want)
						}
					}
					if a.LHS == "state.Count" {
						foundCount = true
						if a.RHS != "types.Int64Value(int64(*count))" {
							t.Fatalf("numeric attribute conversion changed: %q", a.RHS)
						}
					}
				}
				if !foundID || !foundCount {
					t.Fatalf("missing identity or ordinary numeric mapping: %+v", assignments)
				}
			})
		}
	}
}

func TestPluralResultIDsRetainNumericType(t *testing.T) {
	op := pluralNestedOperation()
	op.ResponseSchema.Properties["data"].Items.Properties["id"] = prim("integer", "Result item ID")
	art, err := model.BuildArtifact(op)
	if err != nil {
		t.Fatal(err)
	}
	view, err := BuildDataSourceView(art)
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range view.State.ItemFields {
		if assignment.LHS == "ID" {
			if assignment.RHS != "types.Int64Value(int64(item.GetId()))" {
				t.Fatalf("numeric result ID conversion changed: %s", assignment.RHS)
			}
			return
		}
	}
	t.Fatal("missing result item ID mapping")
}

func TestCallerAssignedIdentityWithoutUpdate(t *testing.T) {
	op := identityResourceOperation()
	op.ResolvedGroup.Update = nil
	data := op.ResolvedGroup.Create.RequestSchema.Properties["data"]
	data.Properties["id"] = prim("string", "Caller-assigned ID.")
	data.Required = append(data.Required, "id")
	art, err := model.BuildArtifact(op)
	if err != nil {
		t.Fatal(err)
	}
	view, err := BuildResourceView(art)
	if err != nil {
		t.Fatal(err)
	}
	source, err := RenderResource(view)
	if err != nil {
		t.Fatal(err)
	}
	if !view.UpdateUnsupported {
		t.Fatal("expected resource without update")
	}
	if !strings.Contains(string(source), "stringplanmodifier.RequiresReplace()") || strings.Contains(string(source), "stringplanmodifier.RequiresReplaceIf(") {
		t.Fatal("resources without update must always replace on name changes")
	}
	if !strings.Contains(string(source), `response.State.SetAttribute(ctx, path.Root("name"), state.ID)`) {
		t.Fatal("resources without update must initialize imported names from the API ID")
	}
}
