package model

import (
	"fmt"
	"slices"
)

// MissingRoleError reports an artifact whose tracking group does not resolve a
// role its shape cannot do without. It names the role, and the operationId when
// the annotation declared one, so a typo reads differently from an omission.
type MissingRoleError struct {
	// Artifact is the Terraform artifact name from the tracking field.
	Artifact string
	// Kind selects the noun the message uses, and says which shape was judged,
	// since which roles are mandatory depends on it.
	Kind ArtifactKind
	// Role is the group role that did not resolve.
	Role GroupRole
	// OperationId is the unresolved reference when the annotation named one, and
	// empty when the role was never declared — two different author errors.
	OperationId string
}

func (e *MissingRoleError) Error() string {
	cause := "is not declared"
	if e.OperationId != "" {
		cause = fmt.Sprintf("names operationId %q, which no operation in the spec declares", e.OperationId)
	}
	noun := "resource"
	if e.Kind == ArtifactKindDataSource {
		noun = "data source"
	}
	return fmt.Sprintf("model: %s %q: group.%s %s; a %s cannot be generated without that operation",
		noun, e.Artifact, e.Role, cause, noun)
}

// requireResolvedRoles fails op's artifact unless every listed role resolves to
// an operation, whether it was omitted or named an operationId that matched
// nothing. Callers name the roles their shape cannot do without.
func requireResolvedRoles(op *Operation, roles ...GroupRole) error {
	for _, role := range roles {
		if op.ResolvedGroup.Op(role) != nil {
			continue
		}
		return &MissingRoleError{
			Artifact:    op.Tracking.ArtifactName,
			Kind:        op.Tracking.ArtifactKind,
			Role:        role,
			OperationId: op.ResolvedGroup.UnresolvedId(role),
		}
	}
	return nil
}

// buildResourceLifecycle resolves the CRUD SDK calls for a resource and reports
// how a missing update role was handled; it assumes Create, Read and Delete
// already resolve. An undeclared update degrades: the resource is still
// generated, and UpdateUnsupported tells the schema builder to force
// replacement on every practitioner-settable attribute. A declared-but-dangling
// update fails instead — treating a typo as an absence would silently make
// every subsequent change destructive.
func buildResourceLifecycle(op *Operation) (*LifecycleBindings, []Diagnostic, error) {
	g := op.ResolvedGroup
	// Explicit selection replaces positional inference on these artifact-local
	// calls. Shared operation bindings remain unchanged, including for siblings
	// that use the same endpoints without an explicit mapping.
	aliasTerminalID := op.Tracking.IdParameter == ""
	bindings := &LifecycleBindings{
		Create: bodyCall(g.Create, false),
		Read:   sdkCall(g.Read, aliasTerminalID),
		// Delete sends no request body. Its response may have a body, including
		// an unnamed one.
		Delete:     sdkCall(g.Delete, aliasTerminalID),
		IdStrategy: op.Tracking.IdStrategy,
	}

	var diags []Diagnostic
	switch dangling := g.UnresolvedId(GroupRoleUpdate); {
	case g.Update != nil:
		bindings.Update = bodyCall(g.Update, aliasTerminalID)
	case dangling != "":
		return nil, nil, &MissingRoleError{
			Artifact:    op.Tracking.ArtifactName,
			Kind:        op.Tracking.ArtifactKind,
			Role:        GroupRoleUpdate,
			OperationId: dangling,
		}
	default:
		bindings.UpdateUnsupported = true
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Message: fmt.Sprintf(
				"resource %q: group.update is not declared, so no endpoint can modify this resource in place; every practitioner-settable attribute forces replacement, and any change to one destroys and recreates the resource",
				op.Tracking.ArtifactName),
		})
	}
	if op.Tracking.IdParameter != "" {
		if err := bindResourceIDParameter(op, bindings); err != nil {
			return nil, nil, err
		}
	}

	return bindings, diags, nil
}

// IDParameterError reports an invalid explicit resource-ID binding. Selection
// failures never fall back to terminal-path inference.
type IDParameterError struct {
	Artifact    string
	Role        GroupRole
	OperationId string
	Parameter   string
	Reason      string
}

func (e *IDParameterError) Error() string {
	return fmt.Sprintf("model: resource %q: group.%s operation %q: id_parameter %q: %s",
		e.Artifact, e.Role, e.OperationId, e.Parameter, e.Reason)
}

// bindResourceIDParameter runs before schema construction, so the selected
// argument contributes neither a required input nor an extra import component.
func bindResourceIDParameter(op *Operation, bindings *LifecycleBindings) error {
	parameter := op.Tracking.IdParameter
	for _, entry := range []struct {
		role GroupRole
		call *SDKCall
	}{
		{GroupRoleCreate, bindings.Create},
		{GroupRoleRead, bindings.Read},
		{GroupRoleUpdate, bindings.Update},
		{GroupRoleDelete, bindings.Delete},
	} {
		if entry.call == nil {
			continue // An undeclared Update is supported through replacement.
		}
		failure := func(reason string) error {
			return &IDParameterError{
				Artifact: op.Tracking.ArtifactName, Role: entry.role,
				OperationId: op.ResolvedGroup.Op(entry.role).OperationId, Parameter: parameter, Reason: reason,
			}
		}
		selected := -1
		for i, argument := range entry.call.Arguments {
			if argument.Name != parameter || argument.Location != "path" {
				continue
			}
			if entry.role == GroupRoleCreate {
				return failure("Create requires this path parameter before response data.id exists; choose a resource identity that is not required during Create")
			}
			if selected >= 0 {
				return failure("the operation has multiple matching path arguments; the binding must be unambiguous")
			}
			selected = i
			if !supportedIDParameter(argument) {
				return failure(fmt.Sprintf("path argument type %q has no supported conversion from Terraform's string id; use a supported scalar or UUID path parameter", argument.GoType))
			}
		}
		if entry.role != GroupRoleCreate && selected < 0 {
			return failure("the operation has no required SDK path argument with this exact OpenAPI name; check id_parameter and the lifecycle operation")
		}
		for i, argument := range entry.call.Arguments {
			if i != selected && argument.TFName == "id" {
				return failure(fmt.Sprintf("another argument %q (%s) also claims Terraform's id; rename or remove the conflicting input", argument.Name, argument.Location))
			}
		}
		for _, argument := range entry.call.OptionalArguments {
			if argument.TFName == "id" {
				return failure(fmt.Sprintf("another optional argument %q (%s) also claims Terraform's id; rename or remove the conflicting input", argument.Name, argument.Location))
			}
		}
		if selected >= 0 {
			entry.call.Arguments[selected].TFName = "id"
		}
	}
	return nil
}

// supportedIDParameter matches the conversions the emitter can perform from a
// computed types.String ID. Named enums additionally need a string schema:
// their SDK validating constructor receives that string unchanged.
func supportedIDParameter(argument SDKArgument) bool {
	if argument.Schema == nil || argument.Schema.Kind != SchemaKindPrimitive {
		return false
	}
	switch argument.GoType {
	case "string", "uuid.UUID", "int", "int32", "int64", "float32", "float64", "bool":
		return true
	default:
		underlying, supported := SDKScalarGoType(argument.Schema)
		return supported && underlying == "string" && len(argument.Schema.Enum) > 0 && IsBareSDKTypeName(argument.GoType)
	}
}

// bodyCall resolves the SDK binding for an operation that sends a request body,
// which a read or a delete never does. aliasTerminalID is false for a create,
// whose path does not name the new record; updates enable it only without an
// explicit ID parameter. The
// envelope facts it reads come off this operation's own RequestSchema, never the
// merged tree: a resource's Create and Update components are routinely distinct
// types declaring different fields, so per-role is the only correct source.
func bodyCall(op *Operation, aliasTerminalID bool) *SDKCall {
	call := sdkCall(op, aliasTerminalID)
	call.GoRequestType = op.RequestRefName

	data := envelopeData(op.RequestSchema)
	if data == nil {
		return call
	}
	// An empty component name means the body left that level inline, so the SDK
	// generated nothing to construct it from; what that costs is decided
	// downstream, since a body with settable attributes cannot be built without
	// one but a body with none can.
	call.GoRequestDataType = data.RefName
	// The attributes node is kept whole, not just its component name, so the
	// request mapper can narrow the merged tree against it.
	if attributes := data.Properties["attributes"]; attributes != nil {
		call.GoRequestAttributesType = attributes.RefName
		call.RequestAttributesSchema = attributes
	}
	// A Terraform id is always a string, so the declared type is what decides
	// whether it can be sent verbatim or has to be parsed first.
	if id := data.Properties["id"]; id != nil {
		call.RequestDeclaresID = true
		call.RequestIDRequired = slices.Contains(data.Required, "id")
		call.RequestIDReadOnly = id.ReadOnly
		call.RequestIDGoType, _ = SDKScalarGoType(id)
		if len(id.Enum) > 0 {
			scalarType := call.RequestIDGoType
			call.RequestIDGoType = ""
			if scalarType == "string" && IsBareSDKTypeName(id.RefName) {
				call.RequestIDGoType = id.RefName
				call.RequestIDEnum = true
			}
		}
	}
	// SDKDefaulted reproduces the SDK generator's own predicate rather than
	// reading the generated source: model_simple.j2 assigns a property in
	// WithDefaults() exactly when it declares a default, is neither object nor
	// array, and is not readOnly. A discriminator is always a string, so only the
	// default and readOnly halves can vary.
	if discriminator := data.Properties["type"]; discriminator != nil {
		call.RequestDiscriminator = &RequestDiscriminator{
			GoType:       discriminator.RefName,
			Values:       slices.Clone(discriminator.Enum),
			SDKDefaulted: discriminator.HasDefault && !discriminator.ReadOnly,
		}
	}
	return call
}

// envelopeData resolves a body's JSON:API "data" member, the level every
// envelope fact hangs off. Nil for an absent body or one that is not enveloped.
func envelopeData(s *Schema) *Schema {
	if s == nil {
		return nil
	}
	return s.Properties["data"]
}
