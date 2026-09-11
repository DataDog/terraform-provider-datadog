package parser

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("ResolveOperationGroups", func() {
	// op builds a bare operation; groupedOp adds tracking metadata declaring a
	// group. Both stay minimal — resolution reads only OperationId and Tracking.
	op := func(operationId string) *model.Operation {
		return &model.Operation{Path: "/" + operationId, Method: "GET", OperationId: operationId}
	}
	groupedOp := func(operationId string, group *model.OperationGroup) *model.Operation {
		o := op(operationId)
		o.Tracking = &model.TrackingFieldMetadata{
			ArtifactKind: model.ArtifactKindResource,
			ArtifactName: "thing",
			Group:        group,
		}
		return o
	}

	Context("when every declared operationId exists", func() {
		It("wires each CRUD role to the operation it names", func() {
			create := groupedOp("CreateThing", &model.OperationGroup{
				Create: "CreateThing",
				Read:   "GetThing",
				Update: "UpdateThing",
				Delete: "DeleteThing",
			})
			read, update, del := op("GetThing"), op("UpdateThing"), op("DeleteThing")
			spec := &model.Spec{Operations: []*model.Operation{create, read, update, del}}

			ResolveOperationGroups(spec)

			g := create.ResolvedGroup
			Expect(g).NotTo(BeNil())
			Expect(g.Create).To(BeIdenticalTo(create))
			Expect(g.Read).To(BeIdenticalTo(read))
			Expect(g.Update).To(BeIdenticalTo(update))
			Expect(g.Delete).To(BeIdenticalTo(del))
			Expect(g.Search).To(BeNil())
			Expect(g.Unresolved).To(BeEmpty())
		})

		It("resolves a reference to an operation declared later in the spec", func() {
			read := groupedOp("GetThing", &model.OperationGroup{Read: "GetThing", Search: "ListThings"})
			list := op("ListThings")
			spec := &model.Spec{Operations: []*model.Operation{read, list}}

			ResolveOperationGroups(spec)

			Expect(read.ResolvedGroup.Search).To(BeIdenticalTo(list))
			Expect(read.ResolvedGroup.Op(model.GroupRoleSearch)).To(BeIdenticalTo(list))
		})

		It("resolves a search-only group to the annotated operation itself", func() {
			list := groupedOp("ListThings", &model.OperationGroup{Search: "ListThings"})
			spec := &model.Spec{Operations: []*model.Operation{list}}

			ResolveOperationGroups(spec)

			Expect(list.ResolvedGroup.Search).To(BeIdenticalTo(list))
			Expect(list.ResolvedGroup.Read).To(BeNil())
		})
	})

	Context("when a declared operationId matches no operation", func() {
		It("leaves the role nil and records the reference", func() {
			create := groupedOp("CreateThing", &model.OperationGroup{
				Create: "CreateThing",
				Update: "UpdateThingTypo",
			})
			spec := &model.Spec{Operations: []*model.Operation{create}}

			ResolveOperationGroups(spec)

			g := create.ResolvedGroup
			Expect(g.Create).To(BeIdenticalTo(create))
			Expect(g.Update).To(BeNil())
			Expect(g.Unresolved).To(Equal([]model.GroupReference{
				{Role: model.GroupRoleUpdate, OperationId: "UpdateThingTypo"},
			}))
		})

		It("records several dangling references in create/read/search/update/delete order", func() {
			create := groupedOp("CreateThing", &model.OperationGroup{
				Delete: "DeleteTypo",
				Update: "UpdateTypo",
				Read:   "ReadTypo",
				Create: "CreateTypo",
			})
			spec := &model.Spec{Operations: []*model.Operation{create}}

			ResolveOperationGroups(spec)

			Expect(create.ResolvedGroup.Unresolved).To(Equal([]model.GroupReference{
				{Role: model.GroupRoleCreate, OperationId: "CreateTypo"},
				{Role: model.GroupRoleRead, OperationId: "ReadTypo"},
				{Role: model.GroupRoleUpdate, OperationId: "UpdateTypo"},
				{Role: model.GroupRoleDelete, OperationId: "DeleteTypo"},
			}))
		})

		It("does not fail the run, so unrelated artifacts still resolve", func() {
			broken := groupedOp("CreateBroken", &model.OperationGroup{Read: "GoneMissing"})
			healthy := groupedOp("GetHealthy", &model.OperationGroup{Read: "GetHealthy"})
			healthy.Tracking.ArtifactName = "healthy"
			spec := &model.Spec{Operations: []*model.Operation{broken, healthy}}

			ResolveOperationGroups(spec)

			Expect(broken.ResolvedGroup.Read).To(BeNil())
			Expect(healthy.ResolvedGroup.Read).To(BeIdenticalTo(healthy))
			Expect(healthy.ResolvedGroup.Unresolved).To(BeEmpty())
		})
	})

	Context("when an operation declares no group", func() {
		It("leaves ResolvedGroup nil for tracked and untracked operations alike", func() {
			untracked := op("GetHealth")
			groupless := op("GetThing")
			groupless.Tracking = &model.TrackingFieldMetadata{
				ArtifactKind: model.ArtifactKindDataSource,
				ArtifactName: "thing",
			}
			spec := &model.Spec{Operations: []*model.Operation{untracked, groupless}}

			ResolveOperationGroups(spec)

			Expect(untracked.ResolvedGroup).To(BeNil())
			Expect(groupless.ResolvedGroup).To(BeNil())
			// A nil group answers for every role without a guard at the call site.
			Expect(groupless.ResolvedGroup.Op(model.GroupRoleSearch)).To(BeNil())
			Expect(groupless.ResolvedGroup.Op(model.GroupRoleCreate)).To(BeNil())
		})
	})

	It("tolerates a nil spec and nil operations", func() {
		Expect(func() { ResolveOperationGroups(nil) }).NotTo(Panic())
		spec := &model.Spec{Operations: []*model.Operation{nil, op("GetThing")}}
		Expect(func() { ResolveOperationGroups(spec) }).NotTo(Panic())
	})

	It("is idempotent: a second pass reproduces the first result", func() {
		build := func() *model.Spec {
			create := groupedOp("CreateThing", &model.OperationGroup{
				Create: "CreateThing", Read: "GetThing", Update: "Typo",
			})
			return &model.Spec{Operations: []*model.Operation{create, op("GetThing")}}
		}
		once, twice := build(), build()
		ResolveOperationGroups(once)
		ResolveOperationGroups(twice)
		ResolveOperationGroups(twice)
		Expect(twice).To(Equal(once))
	})

	Context("through LoadSpec", func() {
		It("resolves a full CRUD group from an annotated spec", func() {
			spec := loadSpecMust("schema_normalize_crud.yaml")
			g := opByID(spec, "CreateThing").ResolvedGroup

			Expect(g).NotTo(BeNil())
			Expect(g.Create).To(BeIdenticalTo(opByID(spec, "CreateThing")))
			Expect(g.Read).To(BeIdenticalTo(opByID(spec, "GetThing")))
			Expect(g.Update).To(BeIdenticalTo(opByID(spec, "UpdateThing")))
			Expect(g.Delete).To(BeIdenticalTo(opByID(spec, "DeleteThing")))
			Expect(g.Unresolved).To(BeEmpty())
		})

		It("leaves Update nil when the annotation omits it", func() {
			spec := loadSpecMust("schema_normalize_missing_update.yaml")
			g := opByID(spec, "CreateThing").ResolvedGroup

			Expect(g.Update).To(BeNil())
			Expect(g.Delete).To(BeIdenticalTo(opByID(spec, "DeleteThing")))
			// An omitted role is not a dangling reference.
			Expect(g.Unresolved).To(BeEmpty())
		})

		It("records dangling references without failing the load", func() {
			spec := loadSpecMust("group_unresolved.yaml")
			g := opByID(spec, "CreateThing").ResolvedGroup

			Expect(g.Read).To(BeIdenticalTo(opByID(spec, "GetThing")))
			Expect(g.Update).To(BeNil())
			Expect(g.Delete).To(BeNil())
			Expect(g.Unresolved).To(Equal([]model.GroupReference{
				{Role: model.GroupRoleUpdate, OperationId: "PatchThingTypo"},
				{Role: model.GroupRoleDelete, OperationId: "DestroyThingTypo"},
			}))
		})

		It("normalizes the bodies of every operation the group resolved to", func() {
			spec := loadSpecMust("schema_normalize_crud.yaml")
			for _, id := range []string{"CreateThing", "GetThing", "UpdateThing"} {
				Expect(opByID(spec, id).ResponseSchema).NotTo(BeNil(), "response schema of %s", id)
			}
			// The annotated operation's own request body is normalized too.
			Expect(opByID(spec, "CreateThing").RequestSchema).NotTo(BeNil())
		})
	})
})

var _ = Describe("ResolvedGroup.Operations", func() {
	first, second := &model.Operation{OperationId: "First"}, &model.Operation{OperationId: "Second"}

	DescribeTable("returns the resolved operations in create/read/search/update/delete order",
		func(g *model.ResolvedGroup, want []*model.Operation) {
			Expect(g.Operations()).To(Equal(want))
		},
		Entry("distinct roles", &model.ResolvedGroup{Create: first, Delete: second},
			[]*model.Operation{first, second}),
		Entry("an operation filling two roles appears once", &model.ResolvedGroup{Read: first, Search: first, Delete: second},
			[]*model.Operation{first, second}),
	)

	It("is nil-safe and skips unfilled roles", func() {
		Expect((*model.ResolvedGroup)(nil).Operations()).To(BeEmpty())
		Expect((&model.ResolvedGroup{}).Operations()).To(BeEmpty())
	})
})

var _ = Describe("ResolvedGroup.Op", func() {
	group := &model.ResolvedGroup{
		Create: &model.Operation{OperationId: "CreateThing"},
		Read:   &model.Operation{OperationId: "GetThing"},
		Search: &model.Operation{OperationId: "ListThings"},
		Update: &model.Operation{OperationId: "UpdateThing"},
		Delete: &model.Operation{OperationId: "DeleteThing"},
	}

	DescribeTable("returns the operation resolved for each role",
		func(role model.GroupRole, wantOperationId string) {
			Expect(group.Op(role).OperationId).To(Equal(wantOperationId))
		},
		Entry("create", model.GroupRoleCreate, "CreateThing"),
		Entry("read", model.GroupRoleRead, "GetThing"),
		Entry("search", model.GroupRoleSearch, "ListThings"),
		Entry("update", model.GroupRoleUpdate, "UpdateThing"),
		Entry("delete", model.GroupRoleDelete, "DeleteThing"),
	)

	It("returns nil for an unfilled role, an unknown role, and a nil group", func() {
		Expect((&model.ResolvedGroup{}).Op(model.GroupRoleCreate)).To(BeNil())
		Expect(group.Op(model.GroupRole("publish"))).To(BeNil())
		Expect((*model.ResolvedGroup)(nil).Op(model.GroupRoleRead)).To(BeNil())
	})
})
