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
	// Kind selects the noun the message uses. Which roles are mandatory depends
	// on the artifact's shape, so the error has to say which shape it judged.
	Kind ArtifactKind
	// Role is the group role that did not resolve.
	Role GroupRole
	// OperationId is the unresolved reference when the annotation named one, and
	// empty when the role was never declared. The two are different author
	// errors and read differently.
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
// nothing. Callers name the roles their shape cannot do without, which is a
// property of the shape rather than of any one sub-builder — so this runs before
// the schema and the lifecycle are built, not inside either.
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
// how a missing update role was handled. It assumes Create, Read and Delete are
// already known to resolve; buildResourceArtifact enforces that before calling.
//
// Update is the one role whose absence is a lifecycle decision rather than a
// precondition, which is why it is settled here. An undeclared update degrades:
// the resource is still generated, and UpdateUnsupported tells the schema builder
// to force replacement on every practitioner-settable attribute (FR-034d, T123).
// A declared-but-dangling update fails instead — treating a typo as an absence
// would silently make every subsequent change destructive.
func buildResourceLifecycle(op *Operation) (*LifecycleBindings, []Diagnostic, error) {
	g := op.ResolvedGroup
	bindings := &LifecycleBindings{
		Create: bodyCall(g.Create, false),
		Read:   sdkCall(g.Read, true),
		// A delete takes the terminal path id and sends no body; a 204 response
		// leaves GoResponseType empty.
		Delete:     sdkCall(g.Delete, true),
		IdStrategy: op.Tracking.IdStrategy,
	}

	var diags []Diagnostic
	switch dangling := g.UnresolvedId(GroupRoleUpdate); {
	case g.Update != nil:
		bindings.Update = bodyCall(g.Update, true)
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

	return bindings, diags, nil
}

// bodyCall resolves the SDK binding for an operation that sends a request body,
// which a read or a delete never does. aliasTerminalID is false for a create,
// which has no id yet, and true for an update, whose path names the record.
//
// It also reads the four facts the request mapper needs about this body's
// JSON:API envelope, all off this operation's own RequestSchema rather than off
// the merged tree: the merge reconciles RefName toward the Read response
// (FR-034c), and a resource's Create and Update data components are routinely
// distinct types, so per-role is the only correct source.
func bodyCall(op *Operation, aliasTerminalID bool) *SDKCall {
	call := sdkCall(op, aliasTerminalID)
	call.GoRequestType = op.RequestRefName

	data := envelopeData(op.RequestSchema)
	if data == nil {
		return call
	}
	// An empty component name means the body left that level inline, so the SDK
	// generated nothing to construct it from; emit decides what that costs,
	// since a body with settable attributes cannot be built without one but a
	// body with none can (T138).
	call.GoRequestDataType = data.RefName
	if attributes := data.Properties["attributes"]; attributes != nil {
		call.GoRequestAttributesType = attributes.RefName
	}
	// A Terraform id is always a string, so the declared type is what decides
	// whether it can be sent verbatim or has to be parsed first (T140).
	if id := data.Properties["id"]; id != nil {
		call.RequestDeclaresID = true
		call.RequestIDGoType, _ = SDKScalarGoType(id)
	}
	// SDKDefaulted reproduces the SDK generator's own predicate rather than
	// reading the generated source (FR-005a): model_simple.j2 assigns a property
	// in WithDefaults() exactly when it declares a default, is neither object nor
	// array, and is not readOnly. A discriminator is always a string, so only the
	// default and readOnly halves can vary (T143).
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
