package parser

import (
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ResolveOperationGroups wires every tracked operation's declared group to the
// operations it names, filling Operation.ResolvedGroup: create/read/update/
// delete for a resource lifecycle, read (by-id) and/or search (list) for a data
// source. An operation carrying no tracking field, or tracking metadata with no
// group, is left with a nil ResolvedGroup.
//
// Resolution is a whole-spec pass rather than a per-operation one because a
// group routinely references an operation that appears later in the document,
// and one that carries no tracking field of its own (only the annotated
// operation does).
//
// It never fails the run. An operationId matching no operation in the spec is
// recorded in ResolvedGroup.Unresolved and leaves that role nil, so the
// artifact and lifecycle builders can fail exactly the artifact that depends on
// it while unrelated artifacts continue (FR-012). Calling it twice on the same
// spec produces the same result.
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
	resolve := func(role model.GroupRole, id string, dst **model.Operation) {
		if id == "" {
			return
		}
		if target := byID[id]; target != nil {
			*dst = target
			return
		}
		g.Unresolved = append(g.Unresolved, model.GroupReference{Role: role, OperationId: id})
	}
	resolve(model.GroupRoleCreate, decl.Create, &g.Create)
	resolve(model.GroupRoleRead, decl.Read, &g.Read)
	resolve(model.GroupRoleSearch, decl.Search, &g.Search)
	resolve(model.GroupRoleUpdate, decl.Update, &g.Update)
	resolve(model.GroupRoleDelete, decl.Delete, &g.Delete)
	return g
}

// indexOperationsByID indexes operations by operationId, skipping the unnamed
// ones (an operationId is optional in OpenAPI, but a group can only reference an
// operation that has one). OpenAPI requires operationIds to be unique, so a
// duplicate is a malformed document; the last of a duplicate pair wins, which
// LoadSpec's (path, method) sort makes deterministic.
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
