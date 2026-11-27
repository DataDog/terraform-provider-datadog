package model

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func explicitIDResourceOp() *Operation {
	op := incidentTypeResourceOp()
	op.Tracking.IdParameter = "incident_type_id"
	parent := SDKArgument{Name: "account_id", GoName: "accountId", GoType: "int64", Location: "path", Schema: primSchema("integer")}
	id := SDKArgument{Name: "incident_type_id", GoName: "incidentTypeId", GoType: "uuid.UUID", Location: "path", Schema: primSchema("string")}
	id.Schema.Format = "uuid"
	op.SDKBinding = &SDKOperationBinding{Required: []SDKArgument{parent}}
	for _, endpoint := range []*Operation{op.ResolvedGroup.Read, op.ResolvedGroup.Update, op.ResolvedGroup.Delete} {
		endpoint.SDKBinding = &SDKOperationBinding{Required: []SDKArgument{id, parent}}
		// The selected identity precedes another parameter, which inference
		// would incorrectly alias to Terraform's id.
		endpoint.Path = "/api/v2/types/{incident_type_id}/accounts/{account_id}"
	}
	return op
}

func TestExplicitIDParameterBindsBeforeSchemaAndPreservesSharedOperations(t *testing.T) {
	op := explicitIDResourceOp()
	op.ResolvedGroup.Delete.Path += "/archive"
	before := append([]SDKArgument(nil), op.ResolvedGroup.Read.SDKBinding.Required...)
	artifact, err := BuildArtifact(op)
	if err != nil {
		t.Fatal(err)
	}
	for _, call := range []*SDKCall{artifact.Lifecycle.Read, artifact.Lifecycle.Update, artifact.Lifecycle.Delete} {
		want := append([]SDKArgument(nil), before...)
		want[0].TFName = "id"
		want[1].TFName = "account_id"
		if !reflect.DeepEqual(call.Arguments, want) {
			t.Errorf("%s arguments = %#v, want %#v", call.GoMethod, call.Arguments, want)
		}
	}
	if !reflect.DeepEqual(op.ResolvedGroup.Read.SDKBinding.Required, before) {
		t.Fatal("binding mutated the shared OpenAPI SDK arguments")
	}
	parentFound := false
	for _, attribute := range artifact.Schema.Attributes {
		if attribute.Path == "incident_type_id" {
			t.Error("selected ID leaked into the required Terraform schema")
		}
		if attribute.Path == "account_id" {
			parentFound = true
			if !attribute.Required || len(attribute.PlanModifiers) == 0 {
				t.Error("parent must remain required and force replacement")
			}
		}
	}
	if !parentFound {
		t.Error("unrelated parent input disappeared")
	}
	// A sibling reuses the same operation group without the annotation: its
	// positional inference and redundant action-route input remain unchanged.
	sibling := *op
	tracking := *op.Tracking
	tracking.IdParameter = ""
	sibling.Tracking = &tracking
	legacy, err := BuildArtifact(&sibling)
	if err != nil {
		t.Fatal(err)
	}
	if legacy.Lifecycle.Read.Arguments[0].TFName != "incident_type_id" || legacy.Lifecycle.Read.Arguments[1].TFName != "id" {
		t.Fatal("explicit binding changed a sibling's positional inference")
	}
	if legacy.Lifecycle.Delete.Arguments[0].TFName != "incident_type_id" {
		t.Fatal("omitting id_parameter changed action-route behavior")
	}
}

func TestExplicitIDParameterFailuresAreContextualAndIsolated(t *testing.T) {
	tests := []struct {
		name   string
		role   GroupRole
		mutate func(*Operation)
		reason string
	}{
		{"missing read parameter", GroupRoleRead, func(op *Operation) {
			op.ResolvedGroup.Read.SDKBinding.Required = op.ResolvedGroup.Read.SDKBinding.Required[1:]
		}, "no required SDK path argument"},
		{"missing update parameter", GroupRoleUpdate, func(op *Operation) { op.ResolvedGroup.Update.SDKBinding.Required = nil }, "no required SDK path argument"},
		{"missing delete parameter", GroupRoleDelete, func(op *Operation) { op.ResolvedGroup.Delete.SDKBinding.Required[0].Name = "INCIDENT_TYPE_ID" }, "exact OpenAPI name"},
		{"query parameter cannot substitute for path", GroupRoleRead, func(op *Operation) { op.ResolvedGroup.Read.SDKBinding.Required[0].Location = "query" }, "no required SDK path argument"},
		{"unresolved SDK binding", GroupRoleRead, func(op *Operation) { op.ResolvedGroup.Read.SDKBinding = nil }, "no required SDK path argument"},
		{"unsupported conversion", GroupRoleDelete, func(op *Operation) { op.ResolvedGroup.Delete.SDKBinding.Required[0].GoType = "time.Time" }, "no supported conversion"},
		{"unsupported schema", GroupRoleUpdate, func(op *Operation) { op.ResolvedGroup.Update.SDKBinding.Required[0].Schema = objSchema(nil) }, "no supported conversion"},
		{"required ID collision", GroupRoleRead, func(op *Operation) { op.ResolvedGroup.Read.SDKBinding.Required[1].Name = "id" }, "also claims Terraform's id"},
		{"optional ID collision", GroupRoleUpdate, func(op *Operation) {
			op.ResolvedGroup.Update.SDKBinding.Optional = []SDKArgument{{Name: "id", Location: "query", GoType: "string", Schema: primSchema("string")}}
		}, "also claims Terraform's id"},
		{"create needs response-derived ID", GroupRoleCreate, func(op *Operation) { op.SDKBinding.Required[0].Name = "incident_type_id" }, "before response data.id exists"},
		{"create has another ID input", GroupRoleCreate, func(op *Operation) { op.SDKBinding.Required[0].Name = "id" }, "also claims Terraform's id"},
		{"ambiguous selected parameter", GroupRoleRead, func(op *Operation) {
			op.ResolvedGroup.Read.SDKBinding.Required = append(op.ResolvedGroup.Read.SDKBinding.Required, op.ResolvedGroup.Read.SDKBinding.Required[0])
		}, "multiple matching path arguments"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := explicitIDResourceOp()
			tt.mutate(op)
			artifact, err := BuildArtifact(op)
			var bindingErr *IDParameterError
			if artifact != nil || !errors.As(err, &bindingErr) {
				t.Fatalf("BuildArtifact = (%v, %v), want nil artifact and IDParameterError", artifact, err)
			}
			if bindingErr.Artifact != "incident_type" || bindingErr.Role != tt.role || bindingErr.Parameter != "incident_type_id" || bindingErr.OperationId != op.ResolvedGroup.Op(tt.role).OperationId {
				t.Fatalf("error lost binding context: %#v", bindingErr)
			}
			for _, context := range []string{"incident_type", string(tt.role), bindingErr.OperationId, "incident_type_id", tt.reason} {
				if !strings.Contains(err.Error(), context) {
					t.Errorf("error %q lacks %q", err.Error(), context)
				}
			}
			if sibling, siblingErr := BuildArtifact(explicitIDResourceOp()); siblingErr != nil || sibling == nil {
				t.Fatalf("unrelated valid artifact failed: %v", siblingErr)
			}
		})
	}
}

func TestExplicitIDParameterSupportsExistingScalarConversions(t *testing.T) {
	for _, goType := range []string{"string", "uuid.UUID", "int", "int32", "int64", "float32", "float64", "bool", "IncidentTypeEnum"} {
		t.Run(goType, func(t *testing.T) {
			op := explicitIDResourceOp()
			for _, endpoint := range []*Operation{op.ResolvedGroup.Read, op.ResolvedGroup.Update, op.ResolvedGroup.Delete} {
				argument := &endpoint.SDKBinding.Required[0]
				argument.GoType = goType
				argument.Schema = primSchema("string")
				switch goType {
				case "uuid.UUID":
					argument.Schema.Format = "uuid"
				case "int", "int32", "int64":
					argument.Schema.Type = "integer"
				case "float32", "float64":
					argument.Schema.Type = "number"
				case "bool":
					argument.Schema.Type = "boolean"
				}
				if goType == "IncidentTypeEnum" {
					argument.Schema.Enum = []string{"first", "second"}
				}
			}
			if _, err := BuildArtifact(op); err != nil {
				t.Fatalf("existing conversion %s rejected: %v", goType, err)
			}
		})
	}
}

func TestExplicitIDParameterAllowsUndeclaredUpdate(t *testing.T) {
	op := explicitIDResourceOp()
	op.ResolvedGroup.Update = nil
	artifact, err := BuildArtifact(op)
	if err != nil {
		t.Fatal(err)
	}
	if !artifact.Lifecycle.UpdateUnsupported || artifact.Lifecycle.Update != nil {
		t.Fatal("missing Update must retain forced replacement behavior")
	}
}

func TestExplicitIDParameterCanBeNamedID(t *testing.T) {
	op := explicitIDResourceOp()
	op.Tracking.IdParameter = "id"
	for _, endpoint := range []*Operation{op.ResolvedGroup.Read, op.ResolvedGroup.Update, op.ResolvedGroup.Delete} {
		endpoint.SDKBinding.Required[0].Name = "id"
		endpoint.Path = "/api/v2/types/{id}/accounts/{account_id}"
	}
	if _, err := BuildArtifact(op); err != nil {
		t.Fatalf("selected id argument must not conflict with itself: %v", err)
	}
}

func TestExplicitIDParameterRejectsEnumConstructorsWithNonStringInput(t *testing.T) {
	for _, format := range []string{"uuid", "date", "date-time", "binary"} {
		t.Run(format, func(t *testing.T) {
			op := explicitIDResourceOp()
			argument := &op.ResolvedGroup.Read.SDKBinding.Required[0]
			argument.GoType = "FormattedEnum"
			argument.Schema = &Schema{Kind: SchemaKindPrimitive, Type: "string", Format: format, Enum: []string{"value"}}
			artifact, err := BuildArtifact(op)
			var contextual *IDParameterError
			if artifact != nil || !errors.As(err, &contextual) || contextual.Role != GroupRoleRead || !strings.Contains(contextual.Reason, "no supported conversion") {
				t.Fatalf("non-string enum constructor input %s must fail explicitly: artifact=%v error=%v", format, artifact, err)
			}
		})
	}
	// Unknown string formats still use a string SDK type, so the validating
	// enum constructor can receive the computed ID unchanged.
	op := explicitIDResourceOp()
	argument := &op.ResolvedGroup.Read.SDKBinding.Required[0]
	argument.GoType = "StringEnum"
	argument.Schema = &Schema{Kind: SchemaKindPrimitive, Type: "string", Format: "custom-string", Enum: []string{"value"}}
	if _, err := BuildArtifact(op); err != nil {
		t.Fatalf("enum constructor taking a string should remain supported: %v", err)
	}
}
