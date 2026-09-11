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
