package parser

import (
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ResolveOperationGroups fills Operation.ResolvedGroup for every tracked
// operation, replacing the group's declared operationIds with the operations
// they name; a whole-spec pass because a group routinely names an operation
// declared later. Never fails: an operationId matching nothing is recorded in
// ResolvedGroup.Unresolved and leaves that role nil. Idempotent.
func ResolveOperationGroups(spec *model.Spec) {
	if spec == nil {
		return
	}
	byID := indexOperationsByID(spec.Operations)
	for _, op := range spec.Operations {
		if op == nil || op.Tracking == nil || op.Tracking.Group == nil {
			continue
		}
		op.ResolvedGroup = resolveGroup(op.Tracking.Group, byID)
		recordLifecycleRoles(op.ResolvedGroup)
	}
}

// recordLifecycleRoles stamps each resolved operation with the group slots it
// fills. An operation can fill several — a minimal annotation may name the same
// operationId as both create and read — so the roles accumulate rather than
// overwrite, and a repeated resolution never duplicates one.
//
// Downstream this is how a consumer tells a create from a read without
// re-walking the group: cassette selection needs it to know that a path
// parameter appearing only on the read/update/delete paths carries the
// identity the create response minted, and so needs no example of its own.
func recordLifecycleRoles(group *model.ResolvedGroup) {
	if group == nil {
		return
	}
	for _, role := range []model.GroupRole{
		model.GroupRoleCreate,
		model.GroupRoleRead,
		model.GroupRoleSearch,
		model.GroupRoleUpdate,
		model.GroupRoleDelete,
	} {
		target := group.Op(role)
		if target == nil || target.HasLifecycleRole(role) {
			continue
		}
		target.LifecycleRoles = append(target.LifecycleRoles, role)
	}
}

// resolveGroup replaces each operationId declared in decl with the operation it
// names, recording the ones byID does not know. Roles are visited in
// create/read/search/update/delete order so Unresolved is deterministic
// regardless of how the annotation was written.
func resolveGroup(decl *model.OperationGroup, byID map[string]*model.Operation) *model.ResolvedGroup {
	g := &model.ResolvedGroup{}
	resolve := func(role model.GroupRole, id string) *model.Operation {
		if id == "" {
			return nil
		}
		if target := byID[id]; target != nil {
			return target
		}
		g.Unresolved = append(g.Unresolved, model.GroupReference{Role: role, OperationId: id})
		return nil
	}
	g.Create = resolve(model.GroupRoleCreate, decl.Create)
	g.Read = resolve(model.GroupRoleRead, decl.Read)
	g.Search = resolve(model.GroupRoleSearch, decl.Search)
	g.Update = resolve(model.GroupRoleUpdate, decl.Update)
	g.Delete = resolve(model.GroupRoleDelete, decl.Delete)
	return g
}

// indexOperationsByID indexes operations by operationId, skipping unnamed ones
// (an operationId is optional, but a group can only name an operation that has
// one). Duplicate ids are malformed input; the last one wins, made
// deterministic by the (path, method) sort applied before this runs.
func indexOperationsByID(ops []*model.Operation) map[string]*model.Operation {
	byID := make(map[string]*model.Operation, len(ops))
	for _, op := range ops {
		if op == nil || op.OperationId == "" {
			continue
		}
		byID[op.OperationId] = op
	}
	return byID
}
