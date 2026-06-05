package parser

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// loadSpecMust loads a fixture via LoadSpec and fails the test on any error.
func loadSpecMust(fixture string, opts ...Option) *model.Spec {
	GinkgoHelper()
	spec, err := LoadSpec(filepath.Join("../testdata/parser", fixture), opts...)
	Expect(err).To(Succeed(), "loading fixture %s", fixture)
	return spec
}

// opByID finds a single operation by operationId or fails the test.
func opByID(spec *model.Spec, operationId string) *model.Operation {
	GinkgoHelper()
	for _, op := range spec.Operations {
		if op.OperationId == operationId {
			return op
		}
	}
	Fail("operation " + operationId + " not found in spec")
	return nil
}

// -------------------------------------------------------------------
//  Kind classification
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas kind classification", func() {

	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("schema_normalize_kinds.yaml")
	})

	DescribeTable("classifies the request body schema kind from structure — not type alone",
		func(operationId string, wantKind model.SchemaKind) {
			op := opByID(spec, operationId)
			Expect(op.RequestSchema).NotTo(BeNil(), "operation %s must have a non-nil RequestSchema", operationId)
			Expect(op.RequestSchema.Kind).To(Equal(wantKind))
		},
		Entry("type:string → primitive", "CreatePrimitive", model.SchemaKindPrimitive),
		Entry("type:object with properties → object", "CreateObject", model.SchemaKindObject),
		Entry("type:array with items → array", "CreateArray", model.SchemaKindArray),
		Entry("additionalProperties without properties → map", "CreateMap", model.SchemaKindMap),
		Entry("oneOf → variant", "CreateVariantOneOf", model.SchemaKindVariant),
		Entry("anyOf → variant", "CreateVariantAnyOf", model.SchemaKindVariant),
	)

	DescribeTable("classifies the response body schema kind from structure — not type alone",
		func(operationId string, wantKind model.SchemaKind) {
			op := opByID(spec, operationId)
			Expect(op.ResponseSchema).NotTo(BeNil(), "operation %s must have a non-nil ResponseSchema", operationId)
			Expect(op.ResponseSchema.Kind).To(Equal(wantKind))
		},
		Entry("type:integer response → primitive", "CreatePrimitive", model.SchemaKindPrimitive),
		Entry("type:object with properties response → object", "CreateObject", model.SchemaKindObject),
	)
})

// -------------------------------------------------------------------
//  Field carrying
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas field carrying", func() {

	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("schema_normalize_kinds.yaml")
	})

	It("carries Type and Format from a primitive request schema", func() {
		op := opByID(spec, "CreatePrimitive")
		Expect(op.RequestSchema.Type).To(Equal("string"))
		Expect(op.RequestSchema.Format).To(Equal("date-time"))
	})

	It("carries Type and Format from a primitive response schema", func() {
		op := opByID(spec, "CreatePrimitive")
		Expect(op.ResponseSchema.Type).To(Equal("integer"))
		Expect(op.ResponseSchema.Format).To(Equal("int64"))
	})

	It("carries Properties and Required from an object schema", func() {
		op := opByID(spec, "CreateObject")
		Expect(op.RequestSchema.Properties).To(HaveKey("name"))
		Expect(op.RequestSchema.Properties).To(HaveKey("count"))
		Expect(op.RequestSchema.Required).To(Equal([]string{"name"}))
	})

	It("sorts the Required slice alphabetically regardless of spec declaration order", func() {
		op := opByID(spec, "CreateObjectMultiRequired")
		Expect(op.RequestSchema.Required).To(Equal([]string{"a_prop", "m_prop", "z_prop"}))
	})

	It("carries the Items element schema for an array schema", func() {
		op := opByID(spec, "CreateArray")
		Expect(op.RequestSchema.Items).NotTo(BeNil())
		Expect(op.RequestSchema.Items.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(op.RequestSchema.Items.Type).To(Equal("string"))
	})

	It("carries the Items value schema for a map schema", func() {
		op := opByID(spec, "CreateMap")
		Expect(op.RequestSchema.Items).NotTo(BeNil())
		Expect(op.RequestSchema.Items.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(op.RequestSchema.Items.Type).To(Equal("string"))
	})

	It("carries Enum values from the schema", func() {
		op := opByID(spec, "CreateEnum")
		Expect(op.RequestSchema.Enum).To(ConsistOf("active", "inactive", "pending"))
	})

	It("carries Sensitive=true when the schema node carries x-datadog-tf-generator.sensitive:true", func() {
		op := opByID(spec, "CreateSensitive")
		Expect(op.RequestSchema.Sensitive).To(BeTrue())
	})

	It("leaves Sensitive=false when no sensitive extension is present on the schema", func() {
		op := opByID(spec, "CreatePrimitive")
		Expect(op.RequestSchema.Sensitive).To(BeFalse())
	})

	It("populates Variants for oneOf schemas and does not drop them", func() {
		op := opByID(spec, "CreateVariantOneOf")
		Expect(op.RequestSchema.Variants).To(HaveLen(2))
	})

	It("populates Variants for anyOf schemas and does not drop them", func() {
		op := opByID(spec, "CreateVariantAnyOf")
		Expect(op.RequestSchema.Variants).To(HaveLen(2))
	})
})

// -------------------------------------------------------------------
//  2xx response selection
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas 2xx response selection", func() {

	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("schema_normalize_2xx.yaml")
	})

	It("skips a bodiless 2xx code and picks the first code that has an application/json body", func() {
		// 201 has no JSON body; 202 does — the 202 schema is selected
		op := opByID(spec, "CreateSkipBodiless")
		Expect(op.ResponseSchema).NotTo(BeNil())
		Expect(op.ResponseSchema.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(op.ResponseSchema.Type).To(Equal("string"))
	})

	It("prefers the numerically lower 2xx code when multiple have JSON bodies", func() {
		// 200 (string/uuid) and 201 (integer) both have JSON; 200 is lower so it wins
		op := opByID(spec, "CreatePreferLower")
		Expect(op.ResponseSchema).NotTo(BeNil())
		Expect(op.ResponseSchema.Type).To(Equal("string"))
		Expect(op.ResponseSchema.Format).To(Equal("uuid"))
	})

	It("leaves ResponseSchema nil when no 2xx code carries an application/json body", func() {
		op := opByID(spec, "CreateOnly204")
		Expect(op.ResponseSchema).To(BeNil())
	})
})

// -------------------------------------------------------------------
//  No-body cases
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas no-body cases", func() {

	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("schema_normalize_kinds.yaml")
	})

	It("leaves RequestSchema nil for a GET operation that carries no request body", func() {
		op := opByID(spec, "GetNoBody")
		Expect(op.RequestSchema).To(BeNil())
	})

	It("leaves ResponseSchema nil for an operation whose only response is a bodiless 204", func() {
		op := opByID(spec, "CreateArray")
		Expect(op.ResponseSchema).To(BeNil())
	})

	It("does not return an error when an operation has neither a request body nor a response body", func() {
		_, err := LoadSpec(filepath.Join("../testdata/parser", "schema_normalize_kinds.yaml"))
		Expect(err).To(Succeed())
	})
})

// -------------------------------------------------------------------
//  $ref resolution
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas $ref resolution", func() {

	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("schema_normalize_refs.yaml")
	})

	It("resolves a direct body $ref to a component and carries its Type and Format", func() {
		op := opByID(spec, "CreateDirectRef")
		Expect(op.RequestSchema).NotTo(BeNil())
		Expect(op.RequestSchema.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(op.RequestSchema.Type).To(Equal("string"))
		Expect(op.RequestSchema.Format).To(Equal("uuid"))
	})

	It("recursively resolves a nested $ref inside a component object's properties", func() {
		op := opByID(spec, "CreateNestedRef")
		Expect(op.RequestSchema).NotTo(BeNil())
		Expect(op.RequestSchema.Kind).To(Equal(model.SchemaKindObject))

		inner, ok := op.RequestSchema.Properties["inner"]
		Expect(ok).To(BeTrue(), "Properties must contain 'inner'")
		Expect(inner.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(inner.Type).To(Equal("string"))
	})
})

// -------------------------------------------------------------------
//  Depth limit → SchemaKindRefCycle
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas depth limit", func() {

	It("classifies a $ref that would exceed --max-depth as SchemaKindRefCycle instead of returning an error", func() {
		// maxDepth=1: body→OuterObject costs depth 1 (OK).
		// OuterObject.properties.inner→MyString would cost depth 2, which exceeds the bound.
		spec, err := LoadSpec(
			filepath.Join("../testdata/parser", "schema_normalize_refs.yaml"),
			WithMaxDepth(1),
		)
		Expect(err).To(Succeed())

		op := opByID(spec, "CreateNestedRef")
		Expect(op.RequestSchema).NotTo(BeNil())
		Expect(op.RequestSchema.Kind).To(Equal(model.SchemaKindObject))

		inner, ok := op.RequestSchema.Properties["inner"]
		Expect(ok).To(BeTrue(), "Properties must contain the depth-limited 'inner' entry")
		Expect(inner.Kind).To(Equal(model.SchemaKindRefCycle),
			"a $ref that exceeds --max-depth must be classified as ref_cycle, not dropped or errored")
	})
})

// -------------------------------------------------------------------
//  Unresolvable $ref
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas unresolvable $ref", func() {

	It("returns a typed *UnresolvableRefError naming the missing $ref target", func() {
		_, err := LoadSpec(filepath.Join("../testdata/parser", "schema_normalize_unresolvable.yaml"))
		Expect(err).To(HaveOccurred())

		var refErr *UnresolvableRefError
		Expect(errors.As(err, &refErr)).To(BeTrue(),
			"expected *UnresolvableRefError, got %T: %v", err, err)
		Expect(refErr.Ref).To(Equal("#/components/schemas/DoesNotExist"))
	})
})

// -------------------------------------------------------------------
//  Full CRUD group resolution
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas CRUD group resolution", func() {

	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("schema_normalize_crud.yaml")
	})

	It("fills RequestSchema on the create operation from its own request body", func() {
		op := opByID(spec, "CreateThing")
		Expect(op.RequestSchema).NotTo(BeNil())
		Expect(op.RequestSchema.Kind).To(Equal(model.SchemaKindObject))
	})

	It("fills ResponseSchema on the create operation from its own 2xx response", func() {
		op := opByID(spec, "CreateThing")
		Expect(op.ResponseSchema).NotTo(BeNil())
		Expect(op.ResponseSchema.Kind).To(Equal(model.SchemaKindObject))
	})

	It("fills ResponseSchema on the read operation from its own 2xx response", func() {
		op := opByID(spec, "GetThing")
		Expect(op.ResponseSchema).NotTo(BeNil())
		Expect(op.ResponseSchema.Kind).To(Equal(model.SchemaKindObject))
	})

	It("fills RequestSchema and ResponseSchema on the update operation from its own body and response", func() {
		op := opByID(spec, "UpdateThing")
		Expect(op.RequestSchema).NotTo(BeNil())
		Expect(op.ResponseSchema).NotTo(BeNil())
	})

	It("leaves RequestSchema and ResponseSchema nil on the delete operation that has a 204 and no body", func() {
		op := opByID(spec, "DeleteThing")
		Expect(op.RequestSchema).To(BeNil())
		Expect(op.ResponseSchema).To(BeNil())
	})
})

// -------------------------------------------------------------------
//  Missing update operationId
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas missing update operationId", func() {

	It("does not return an error when a resource group omits the update operationId", func() {
		_, err := LoadSpec(filepath.Join("../testdata/parser", "schema_normalize_missing_update.yaml"))
		Expect(err).To(Succeed())
	})

	It("still populates schemas for the create, read, and delete operations when update is absent", func() {
		spec := loadSpecMust("schema_normalize_missing_update.yaml")

		create := opByID(spec, "CreateThing")
		Expect(create.RequestSchema).NotTo(BeNil(), "create must have a RequestSchema")
		Expect(create.ResponseSchema).NotTo(BeNil(), "create must have a ResponseSchema")

		read := opByID(spec, "GetThing")
		Expect(read.ResponseSchema).NotTo(BeNil(), "read must have a ResponseSchema")

		del := opByID(spec, "DeleteThing")
		Expect(del.RequestSchema).To(BeNil(), "delete has no request body")
		Expect(del.ResponseSchema).To(BeNil(), "delete has only a 204 response")
	})
})

// -------------------------------------------------------------------
//  Only tracked operations are processed
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas only processes tracked operations", func() {

	It("leaves RequestSchema and ResponseSchema nil on an untracked operation even when it has a body", func() {
		spec := loadSpecMust("schema_normalize_kinds.yaml")
		op := opByID(spec, "GetUntracked")
		Expect(op.RequestSchema).To(BeNil(),
			"untracked operation must not have RequestSchema populated")
		Expect(op.ResponseSchema).To(BeNil(),
			"untracked operation must not have ResponseSchema populated")
	})
})

// -------------------------------------------------------------------
//  Determinism
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas determinism", func() {

	It("produces equal Schema trees across two independent loads of the same spec", func() {
		first := loadSpecMust("schema_normalize_crud.yaml")
		second := loadSpecMust("schema_normalize_crud.yaml")

		for _, op := range first.Operations {
			match := opByID(second, op.OperationId)
			Expect(op.RequestSchema).To(Equal(match.RequestSchema),
				"RequestSchema for %s diverged between runs", op.OperationId)
			Expect(op.ResponseSchema).To(Equal(match.ResponseSchema),
				"ResponseSchema for %s diverged between runs", op.OperationId)
		}
	})

	It("sorts the Required slice so its order does not depend on spec declaration order", func() {
		spec := loadSpecMust("schema_normalize_kinds.yaml")
		op := opByID(spec, "CreateObjectMultiRequired")
		// Spec declares required in order: z_prop, a_prop, m_prop — must come out sorted.
		Expect(op.RequestSchema.Required).To(Equal([]string{"a_prop", "m_prop", "z_prop"}))
	})
})
