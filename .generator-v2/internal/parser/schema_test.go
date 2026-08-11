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
		Entry("oneOf → one_of", "CreateVariantOneOf", model.SchemaKindOneOf),
		Entry("anyOf → unsupported", "CreateVariantAnyOf", model.SchemaKindUnsupported),
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
//  JSON and unsupported classification
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas JSON classification", func() {

	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("schema_normalize_unsupported.yaml")
	})

	DescribeTable("classifies an unconstrained node as normalized JSON",
		func(operationId string, wantKind model.SchemaKind) {
			op := opByID(spec, operationId)
			Expect(op.RequestSchema).NotTo(BeNil(), "operation %s must have a non-nil RequestSchema", operationId)
			Expect(op.RequestSchema.Kind).To(Equal(wantKind))
		},
		Entry("type:object with no properties", "CreateEmptyObject", model.SchemaKindJSON),
		Entry("a typeless leaf", "CreateUntyped", model.SchemaKindJSON),
		Entry("type:array with no items", "CreateArrayNoItems", model.SchemaKindJSON),
		Entry("additionalProperties:{}", "CreateFreeFormMap", model.SchemaKindJSON),
		Entry("additionalProperties:true", "CreateBoolMap", model.SchemaKindJSON),
	)

	It("collapses a free-form additionalProperties:{} map into one JSON value", func() {
		op := opByID(spec, "CreateFreeFormMap")
		Expect(op.RequestSchema.Kind).To(Equal(model.SchemaKindJSON))
		Expect(op.RequestSchema.Items).To(BeNil())
	})

	It("collapses an additionalProperties:true map into one JSON value", func() {
		op := opByID(spec, "CreateBoolMap")
		Expect(op.RequestSchema.Kind).To(Equal(model.SchemaKindJSON))
		Expect(op.RequestSchema.Items).To(BeNil())
	})

	It("still classifies a concrete scalar leaf as primitive, never unsupported", func() {
		op := opByID(spec, "CreateTypedPrimitive")
		Expect(op.RequestSchema.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(op.RequestSchema.Type).To(Equal("string"))
	})

	DescribeTable("preserves recursively typed collection elements",
		func(operationId string, wantParentKind, wantChildKind model.SchemaKind) {
			op := opByID(spec, operationId)
			Expect(op.RequestSchema.Kind).To(Equal(wantParentKind), "the outer collection keeps its structural kind")
			Expect(op.RequestSchema.Items).NotTo(BeNil())
			Expect(op.RequestSchema.Items.Kind).To(Equal(wantChildKind))
			Expect(op.RequestSchema.Items.Items).NotTo(BeNil())
			Expect(op.RequestSchema.Items.Items.Kind).To(Equal(model.SchemaKindPrimitive))
		},
		Entry("array-of-array", "CreateArrayOfArray", model.SchemaKindArray, model.SchemaKindArray),
		Entry("array-of-map", "CreateArrayOfMap", model.SchemaKindArray, model.SchemaKindMap),
		Entry("map-of-array", "CreateMapOfArray", model.SchemaKindMap, model.SchemaKindArray),
		Entry("map-of-map", "CreateMapOfMap", model.SchemaKindMap, model.SchemaKindMap),
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
		// Spec declares required in order: z_prop, a_prop, m_prop — must come out sorted.
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

	It("carries the Description from the schema", func() {
		op := opByID(spec, "CreatePrimitive")
		Expect(op.RequestSchema.Description).To(Equal("The creation timestamp of the resource."))
	})

	It("carries Sensitive=true when the schema node carries x-datadog-tf-generator.sensitive:true", func() {
		op := opByID(spec, "CreateSensitive")
		Expect(op.RequestSchema.Sensitive).To(BeTrue())
	})

	It("leaves Sensitive=false when no sensitive extension is present on the schema", func() {
		op := opByID(spec, "CreatePrimitive")
		Expect(op.RequestSchema.Sensitive).To(BeFalse())
	})

	It("populates the normalized OneOf model and leaves the legacy Variants bridge empty", func() {
		op := opByID(spec, "CreateVariantOneOf")
		Expect(op.RequestSchema.OneOf).NotTo(BeNil())
		Expect(op.RequestSchema.OneOf.Variants).To(HaveLen(2))
		Expect(op.RequestSchema.Variants).To(BeEmpty())
	})

	It("classifies anyOf as unsupported and never treats it as a oneOf", func() {
		op := opByID(spec, "CreateVariantAnyOf")
		Expect(op.RequestSchema.Kind).To(Equal(model.SchemaKindUnsupported))
		Expect(op.RequestSchema.OneOf).To(BeNil())
		Expect(op.RequestSchema.Variants).To(BeEmpty())
	})
})

// -------------------------------------------------------------------
//  Normalized oneOf metadata
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas oneOf metadata", func() {

	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("schema_normalize_oneof.yaml")
	})

	oneOfFor := func(operationID string) *model.OneOfSpec {
		GinkgoHelper()
		schema := opByID(spec, operationID).RequestSchema
		Expect(schema).NotTo(BeNil())
		Expect(schema.Kind).To(Equal(model.SchemaKindOneOf))
		Expect(schema.OneOf).NotTo(BeNil(), "normalized oneOf must use Schema.OneOf, not legacy Schema.Variants")
		Expect(schema.Variants).To(BeEmpty(), "legacy Schema.Variants must be empty after oneOf normalization")
		return schema.OneOf
	}

	It("normalizes an inline discriminator-less union and separates nullable from its variants", func() {
		union := oneOfFor("CreateInlineOneOf")

		Expect(union.Name).NotTo(BeEmpty())
		Expect(union.Path).To(Equal("request"))
		Expect(union.Optional).To(BeFalse())
		Expect(union.Nullable).To(BeTrue())
		Expect(union.Discriminator).To(BeNil())
		Expect(union.Variants).To(HaveLen(2))
		Expect(union.Variants).To(HaveEach(Not(HaveField("TFName", "null"))))
		Expect(union.Variants).To(ConsistOf(
			SatisfyAll(
				HaveField("TFName", "boolean"),
				HaveField("Schema.Kind", model.SchemaKindPrimitive),
				HaveField("Schema.Type", "boolean"),
				HaveField("ValueWrapped", true),
			),
			SatisfyAll(
				HaveField("TFName", "string"),
				HaveField("Schema.Kind", model.SchemaKindPrimitive),
				HaveField("Schema.Type", "string"),
				HaveField("ValueWrapped", true),
			),
		))
	})

	It("retains component identity, discriminator mapping, and referenced alternatives", func() {
		union := oneOfFor("CreateAnimal")

		Expect(union.Name).To(Equal("AnimalUnion"))
		Expect(union.Discriminator).To(Equal(&model.OneOfDiscriminator{
			PropertyName: "kind",
			Mapping: map[string]string{
				"dog": "#/components/schemas/Dog",
				"cat": "#/components/schemas/Cat",
			},
		}))
		Expect(union.Variants).To(HaveLen(2))
		Expect(union.Variants[0]).To(SatisfyAll(
			HaveField("TFName", "cat"),
			HaveField("RefName", "Cat"),
			HaveField("Schema.Kind", model.SchemaKindObject),
			HaveField("ValueWrapped", false),
		))
		Expect(union.Variants[1]).To(SatisfyAll(
			HaveField("TFName", "dog"),
			HaveField("RefName", "Dog"),
			HaveField("Schema.Kind", model.SchemaKindObject),
			HaveField("ValueWrapped", false),
		))
	})

	It("retains the same normalized metadata for a response oneOf", func() {
		op := opByID(spec, "CreateAnimal")
		Expect(op.ResponseRefName).To(Equal("AnimalUnion"))
		Expect(op.ResponseSchema).NotTo(BeNil())
		Expect(op.ResponseSchema.Kind).To(Equal(model.SchemaKindOneOf))
		Expect(op.ResponseSchema.Variants).To(BeEmpty())

		union := op.ResponseSchema.OneOf
		Expect(union).NotTo(BeNil())
		Expect(union.Name).To(Equal("AnimalUnion"))
		Expect(union.Path).To(Equal("response"))
		Expect(union.Optional).To(BeFalse())
		Expect(union.Variants).To(HaveLen(2))
		Expect(union.Variants).To(ConsistOf(
			SatisfyAll(HaveField("TFName", "cat"), HaveField("RefName", "Cat")),
			SatisfyAll(HaveField("TFName", "dog"), HaveField("RefName", "Dog")),
		))
	})

	It("assigns canonical paths and optionality at property and collection placements", func() {
		root := opByID(spec, "CreateNestedOneOf").RequestSchema
		Expect(root.Kind).To(Equal(model.SchemaKindObject))

		choice := root.Properties["choice"].OneOf
		Expect(choice).NotTo(BeNil())
		Expect(choice.Path).To(Equal("request.choice"))
		Expect(choice.Optional).To(BeFalse())

		optional := root.Properties["optional_choice"].OneOf
		Expect(optional).NotTo(BeNil())
		Expect(optional.Path).To(Equal("request.optional_choice"))
		Expect(optional.Optional).To(BeTrue())

		listItem := root.Properties["choices"].Items.OneOf
		Expect(listItem).NotTo(BeNil())
		Expect(listItem.Path).To(Equal("request.choices[]"))
		Expect(listItem.Optional).To(BeFalse())

		mapValue := root.Properties["choice_map"].Items.OneOf
		Expect(mapValue).NotTo(BeNil())
		Expect(mapValue.Path).To(Equal("request.choice_map{}"))
		Expect(mapValue.Optional).To(BeFalse())
	})

	It("normalizes a oneOf recursively inside another alternative", func() {
		outer := opByID(spec, "CreateNestedOneOf").RequestSchema.Properties["recursive"].OneOf
		Expect(outer).NotTo(BeNil())

		var objectVariant *model.OneOfVariant
		for i := range outer.Variants {
			if outer.Variants[i].Schema.Kind == model.SchemaKindObject {
				objectVariant = &outer.Variants[i]
				break
			}
		}
		Expect(objectVariant).NotTo(BeNil())

		nested := objectVariant.Schema.Properties["nested"].OneOf
		Expect(nested).NotTo(BeNil())
		Expect(nested.Path).To(Equal("request.recursive." + objectVariant.TFName + ".nested"))
		Expect(nested.Variants).To(HaveLen(2))
	})

	It("merges properties and required constraints adjacent to oneOf into every alternative", func() {
		union := oneOfFor("CreateOneOfWithSiblings")
		Expect(union.Variants).To(HaveLen(2))

		for _, variant := range union.Variants {
			Expect(variant.Schema.Kind).To(Equal(model.SchemaKindObject))
			Expect(variant.Schema.Properties).To(HaveKey("shared"))
			Expect(variant.Schema.Required).To(ContainElement("shared"))
		}
		Expect(union.Variants).To(ConsistOf(
			SatisfyAll(
				HaveField("Schema.Properties", HaveKey("alpha")),
				HaveField("Schema.Required", ConsistOf("alpha", "shared")),
			),
			SatisfyAll(
				HaveField("Schema.Properties", HaveKey("beta")),
				HaveField("Schema.Required", ConsistOf("beta", "shared")),
			),
		))
	})

	It("retains a required-only object constraint adjacent to oneOf", func() {
		union := oneOfFor("CreateOneOfWithRequiredOnlySibling")
		Expect(union.Variants).To(HaveLen(2))

		Expect(union.Variants).To(ConsistOf(
			SatisfyAll(
				HaveField("Schema.Properties", HaveKey("kind")),
				HaveField("Schema.Properties", HaveKey("alpha")),
				HaveField("Schema.Required", ConsistOf("alpha", "kind")),
			),
			SatisfyAll(
				HaveField("Schema.Properties", HaveKey("kind")),
				HaveField("Schema.Properties", HaveKey("beta")),
				HaveField("Schema.Required", ConsistOf("beta", "kind")),
			),
		))
	})

	It("normalizes one adjacent allOf and merges it into every oneOf alternative", func() {
		union := oneOfFor("CreateOneOfWithAllOfSibling")
		Expect(union.Variants).To(HaveLen(2))

		for _, variant := range union.Variants {
			Expect(variant.Schema.Kind).To(Equal(model.SchemaKindObject))
			Expect(variant.Schema.Properties).To(HaveKey("shared"))
			Expect(variant.Schema.Required).To(ContainElement("shared"))
		}
	})

	It("does not Cartesian-expand multiple independent unions", func() {
		union := oneOfFor("CreateIndependentUnions")
		Expect(union.Variants).To(HaveLen(2))

		for _, variant := range union.Variants {
			Expect(variant.Schema.Kind).To(Equal(model.SchemaKindUnsupported))
			Expect(variant.Schema.UnsupportedReason).To(ContainSubstring("conflicts with adjacent schema kind"))
		}
	})

	It("retains naming and local metadata for a oneOf alternative using a $ref sibling", func() {
		union := oneOfFor("CreateOneOfRefSiblingAlternative")
		Expect(union.Variants).To(HaveLen(2))

		var alpha *model.OneOfVariant
		for i := range union.Variants {
			if union.Variants[i].RefName == "AlphaVariant" {
				alpha = &union.Variants[i]
				break
			}
		}
		Expect(alpha).NotTo(BeNil())
		Expect(alpha.Schema.Kind).To(Equal(model.SchemaKindObject))
		Expect(alpha.Schema.Description).To(Equal("Local alpha description."))
	})

	It("produces stable names and sorted variants across independent loads", func() {
		first := oneOfFor("CreateInlineOneOf")
		secondSpec := loadSpecMust("schema_normalize_oneof.yaml")
		second := opByID(secondSpec, "CreateInlineOneOf").RequestSchema.OneOf
		Expect(second).NotTo(BeNil())

		Expect(second.Name).To(Equal(first.Name))
		Expect(second.Variants).To(Equal(first.Variants))
		Expect([]string{first.Variants[0].TFName, first.Variants[1].TFName}).To(Equal([]string{"boolean", "string"}))
	})

	It("produces the same sorted names when oneOf alternatives are reordered", func() {
		first := oneOfFor("CreateInlineOneOf")
		reordered := oneOfFor("CreateReorderedInlineOneOf")

		Expect([]string{reordered.Variants[0].TFName, reordered.Variants[1].TFName}).To(
			Equal([]string{first.Variants[0].TFName, first.Variants[1].TFName}),
		)
	})
})

var _ = Describe("NormalizeSchemas oneOf naming failures", func() {

	It("marks an anonymous non-primitive alternative unsupported without aborting spec loading", func() {
		spec, err := LoadSpec(filepath.Join("../testdata/parser", "schema_normalize_oneof_unnamed.yaml"))
		Expect(err).NotTo(HaveOccurred())

		schema := opByID(spec, "CreateUnnamedOneOf").RequestSchema
		Expect(schema.Kind).To(Equal(model.SchemaKindUnsupported))
		Expect(schema.UnsupportedReason).To(Equal(
			`parser: oneOf at "request" alternative 2 has no stable Terraform variant name; use a discriminator mapping key or a named schema reference`,
		))
	})

	It("marks a colliding union unsupported without aborting spec loading", func() {
		spec, err := LoadSpec(filepath.Join("../testdata/parser", "schema_normalize_oneof_collision.yaml"))
		Expect(err).NotTo(HaveOccurred())

		schema := opByID(spec, "CreateCollidingOneOf").RequestSchema
		Expect(schema.Kind).To(Equal(model.SchemaKindUnsupported))
		Expect(schema.UnsupportedReason).To(Equal(
			`parser: oneOf at "request" has colliding variant name "foo_bar"`,
		))

		healthy := opByID(spec, "CreateHealthyAfterCollision").RequestSchema
		Expect(healthy.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(healthy.Type).To(Equal("string"))
	})

	It("does not replace a local naming failure with outer sibling constraints", func() {
		unsupported := &model.Schema{
			Kind:              model.SchemaKindUnsupported,
			UnsupportedReason: `parser: oneOf at "request.choice" has colliding variant name "object"`,
		}
		common := &model.Schema{
			Kind:       model.SchemaKindObject,
			Properties: map[string]*model.Schema{"shared": {Kind: model.SchemaKindPrimitive, Type: "string"}},
		}

		Expect(mergeNormalizedSchemas(unsupported, common)).To(BeIdenticalTo(unsupported))
		Expect(unsupported.UnsupportedReason).To(ContainSubstring("colliding variant name"))
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
//  allOf normalization
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas allOf normalization", func() {

	var request *model.Schema
	var list *model.Operation
	var get *model.Operation

	BeforeEach(func() {
		spec := loadSpecMust("schema_normalize_allof.yaml")
		request = opByID(spec, "CreateAllOfCases").RequestSchema
		list = opByID(spec, "ListAllOfItems")
		get = opByID(spec, "GetAllOfItem")
	})

	It("preserves a local description on an OpenAPI 3.0 $ref sibling while resolving the referenced shape", func() {
		status := schemaProperty(request, "ref_description")
		Expect(status.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(status.Type).To(Equal("string"))
		Expect(status.Format).To(Equal("uuid"))
		Expect(status.Enum).To(Equal([]string{"active", "inactive"}))
		Expect(status.Description).To(Equal("Local status description."))
	})

	It("ignores a non-structural example sibling without losing the referenced description", func() {
		status := schemaProperty(request, "ref_example")
		Expect(status.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(status.Description).To(Equal("Referenced status description."))
	})

	It("treats an authored single-branch allOf as a metadata overlay", func() {
		status := schemaProperty(request, "single_authored")
		Expect(status.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(status.Type).To(Equal("string"))
		Expect(status.Description).To(Equal("Authored single-branch description."))
	})

	It("ignores empty and nullable-only identity branches inside allOf", func() {
		status := schemaProperty(request, "empty_identity")
		Expect(status.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(status.Type).To(Equal("string"))
		Expect(status.Enum).To(Equal([]string{"active", "inactive"}))

		object := schemaProperty(request, "nullable_identity")
		Expect(object.Kind).To(Equal(model.SchemaKindObject))
		Expect(object.Properties).To(HaveKey("alpha"))
		Expect(object.Properties).To(HaveKey("zeta"))
	})

	DescribeTable("merges object branches with deterministic properties and required fields",
		func(property string, wantProperties, wantRequired []string) {
			composed := schemaProperty(request, property)
			Expect(composed.Kind).To(Equal(model.SchemaKindObject))
			Expect(composed.Properties).To(HaveLen(len(wantProperties)))
			for _, name := range wantProperties {
				Expect(composed.Properties).To(HaveKey(name))
			}
			Expect(composed.Required).To(Equal(wantRequired))
		},
		Entry("two referenced objects", "two_objects", []string{"alpha", "beta", "zeta"}, []string{"alpha", "beta", "zeta"}),
		Entry("three referenced objects", "three_objects", []string{"alpha", "beta", "gamma", "zeta"}, []string{"alpha", "beta", "gamma", "zeta"}),
		Entry("a reference and an inline object", "object_with_inline", []string{"alpha", "delta", "zeta"}, []string{"alpha", "delta", "zeta"}),
		Entry("outer required fields", "outer_required", []string{"alpha", "theta", "zeta"}, []string{"alpha", "theta", "zeta"}),
		Entry("a nested composition", "nested_objects", []string{"alpha", "beta", "gamma", "zeta"}, []string{"alpha", "beta", "gamma", "zeta"}),
	)

	It("uses a referenced metadata-only branch as the composed object's description", func() {
		composed := schemaProperty(request, "metadata_ref")
		Expect(composed.Kind).To(Equal(model.SchemaKindObject))
		Expect(composed.Description).To(Equal("Description supplied by a metadata-only base."))
	})

	It("applies description precedence outer schema, annotation branch, then structural branch", func() {
		Expect(schemaProperty(request, "outer_description").Description).To(Equal("Outer description wins."))
		Expect(schemaProperty(request, "annotation_description").Description).To(Equal("Annotation description wins."))
		Expect(schemaProperty(request, "ref_example").Description).To(Equal("Referenced status description."))
	})

	It("ORs the sensitive marker from an annotation-only branch onto the structural result", func() {
		status := schemaProperty(request, "sensitive_overlay")
		Expect(status.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(status.Sensitive).To(BeTrue())
	})

	It("preserves supported outer scalar constraints", func() {
		formatted := schemaProperty(request, "outer_format")
		Expect(formatted.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(formatted.Type).To(Equal("string"))
		Expect(formatted.Format).To(Equal("date-time"))

		enumerated := schemaProperty(request, "outer_enum")
		Expect(enumerated.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(enumerated.Enum).To(Equal([]string{"inactive"}))
	})

	It("preserves the existing reference-depth limit through allOf", func() {
		limited := opByID(loadSpecMust("schema_normalize_allof_depth.yaml", WithMaxDepth(1)), "CreateAllOfDepth").RequestSchema
		Expect(schemaProperty(limited, "depth_limited").Kind).To(Equal(model.SchemaKindDepthExceeded))
	})

	DescribeTable("keeps unsupported intersections local to their schema node and explains why",
		func(property, reason string) {
			unsupported := schemaProperty(request, property)
			Expect(unsupported.Kind).To(Equal(model.SchemaKindUnsupported))
			Expect(unsupported.UnsupportedReason).To(Equal(reason))
		},
		Entry("duplicate property", "duplicate_property", `allOf property "alpha" is declared by branches 1 and 2`),
		Entry("mixed object and primitive", "mixed_kinds", `allOf branch 2 has schema kind "primitive"; multi-branch composition supports objects only`),
		Entry("annotations without structure", "no_structure", "allOf has no structural branches"),
		Entry("outer type conflict", "outer_type_conflict", `allOf object composition conflicts with outer type "string"`),
		Entry("outer and branch properties", "outer_properties", "allOf combined with outer properties is not supported"),
		Entry("constraint-bearing annotation branch", "constrained_annotation", `allOf branch 2 has unsupported schema kind "unsupported"`),
	)

	It("normalizes representative JSON:API responses and retains SDK reference identities through overlays", func() {
		Expect(list.ResponseRefName).To(Equal("AllOfItemsResponse"))
		Expect(list.ItemRefName).To(Equal("AllOfItem"))
		Expect(list.ResponseDataRefName).To(BeEmpty())
		Expect(get.ResponseRefName).To(Equal("AllOfItemResponse"))
		Expect(get.ResponseDataRefName).To(Equal("AllOfItem"))

		data := schemaProperty(list.ResponseSchema, "data")
		Expect(data.Kind).To(Equal(model.SchemaKindArray))
		Expect(data.Items.Kind).To(Equal(model.SchemaKindObject))
		Expect(data.Items.Properties).To(HaveKey("id"))
		Expect(data.Items.Properties).To(HaveKey("type"))
		attributes := schemaProperty(data.Items, "attributes")
		result := schemaProperty(attributes, "result")
		Expect(result.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(result.Description).To(Equal("Local result description."))

		artifact, err := model.BuildArtifact(list)
		Expect(err).NotTo(HaveOccurred())
		Expect(artifact.Schema.Attributes).To(HaveLen(1))
		Expect(artifact.Schema.Attributes[0].Path).To(Equal("response.data"))
	})

	It("loads the whole spec but fails only the artifact containing an unsupported composition", func() {
		bad := opByID(loadSpecMust("schema_normalize_allof.yaml"), "ListBadAllOfItems")
		_, err := model.BuildArtifact(bad)
		var unsupported *model.UnsupportedKindError
		Expect(errors.As(err, &unsupported)).To(BeTrue(), "expected *UnsupportedKindError, got %T: %v", err, err)
		Expect(unsupported.Path).To(Equal("response.data[].composed"))
		Expect(unsupported.Reason).To(Equal(`allOf property "alpha" is declared by branches 1 and 2`))

		artifact, goodErr := model.BuildArtifact(list)
		Expect(goodErr).NotTo(HaveOccurred())
		Expect(artifact.Name).To(Equal("allof_items"))
	})
})

var _ = Describe("raw schema identity traversal", func() {

	It("terminates property lookup through a recursive allOf", func() {
		components := loadComponents("schema_raw_traversal_cycles.yaml")
		normalizer := &schemaNormalizer{components: components, maxDepth: 16}

		property, err := normalizer.findPropertyProxy(componentProxy(components, "PropertyLoop"), "missing")
		Expect(err).To(Succeed())
		Expect(property).To(BeNil())
	})

	It("terminates overlay resolution through a recursive allOf", func() {
		components := loadComponents("schema_raw_traversal_cycles.yaml")
		normalizer := &schemaNormalizer{components: components, maxDepth: 16}

		schema, err := normalizer.resolveOverlayToSchema(componentProxy(components, "OverlayLoop"))
		Expect(err).To(Succeed())
		Expect(schema).To(BeNil())
	})
})

// -------------------------------------------------------------------
//  Depth limit → SchemaKindDepthExceeded (distinct from a $ref cycle)
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas depth limit", func() {

	// depthLimited normalizes the nested-ref fixture at maxDepth=1 and returns the
	// node that the bound stopped at. body→OuterObject costs depth 1 (OK);
	// OuterObject.properties.inner→MyString would cost depth 2, exceeding the bound.
	depthLimited := func() *model.Schema {
		GinkgoHelper()
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
		return inner
	}

	It("classifies a $ref that would exceed --max-depth as depth_exceeded, not as a cycle", func() {
		inner := depthLimited()
		Expect(inner.Kind).To(Equal(model.SchemaKindDepthExceeded),
			"exhausting --max-depth is not a $ref cycle: cycles.go finds those independently of depth, "+
				"so labelling it ref_cycle sends the reader hunting for a cycle that does not exist")
		Expect(inner.Kind).NotTo(Equal(model.SchemaKindRefCycle))
	})

	It("retains an actionable reason naming the bound, the blocked $ref, and the remedy", func() {
		reason := depthLimited().UnsupportedReason
		Expect(reason).To(ContainSubstring("--max-depth=1"), "the reason must name the bound that was hit")
		Expect(reason).To(ContainSubstring("MyString"), "the reason must name the $ref it stopped short of")
		Expect(reason).To(ContainSubstring("not a $ref cycle"), "the reason must rule out a cycle explicitly")
		Expect(reason).To(ContainSubstring("higher --max-depth"), "the reason must state the remedy")
	})

	It("resolves the same node once the bound is raised, proving it was never a cycle", func() {
		spec, err := LoadSpec(
			filepath.Join("../testdata/parser", "schema_normalize_refs.yaml"),
			WithMaxDepth(8),
		)
		Expect(err).To(Succeed())

		inner := opByID(spec, "CreateNestedRef").RequestSchema.Properties["inner"]
		Expect(inner.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(inner.Type).To(Equal("string"))
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
})

// -------------------------------------------------------------------
//  List operation: query params, pagination, item element type
// -------------------------------------------------------------------

var _ = Describe("NormalizeSchemas list operation", func() {

	var list *model.Operation

	BeforeEach(func() {
		list = opByID(loadSpecMust("schema_normalize_list.yaml"), "ListThings")
	})

	It("captures query parameters sorted by name, preserving bracketed names", func() {
		var names []string
		for _, p := range list.QueryParams {
			names = append(names, p.Name)
		}
		Expect(names).To(Equal([]string{
			"filter[keyword]", "filter[me]", "include", "page[number]", "page[size]", "sort",
		}))
	})

	It("retains OpenAPI declaration order for SDK signature derivation", func() {
		want := map[string]int{
			"page[number]": 1, "page[size]": 2, "sort": 3,
			"include": 4, "filter[keyword]": 5, "filter[me]": 6,
		}
		for _, parameter := range list.QueryParams {
			Expect(parameter.DeclarationOrder).To(Equal(want[parameter.Name]), parameter.Name)
		}
	})

	It("resolves a $ref parameter (#/components/parameters) and normalizes its schema", func() {
		page := paramByName(list, "page[number]")
		Expect(page.Schema).NotTo(BeNil())
		Expect(page.Schema.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(page.Schema.Type).To(Equal("integer"))
		Expect(page.Schema.Format).To(Equal("int64"))
	})

	It("normalizes scalar, array, and enum parameter schemas through to type/enum/array", func() {
		Expect(paramByName(list, "filter[keyword]").Schema.Type).To(Equal("string"))
		Expect(paramByName(list, "filter[me]").Schema.Type).To(Equal("boolean"))
		Expect(paramByName(list, "filter[me]").Required).To(BeTrue())
		Expect(paramByName(list, "include").Schema.Kind).To(Equal(model.SchemaKindArray))
		sort := paramByName(list, "sort").Schema
		Expect(sort.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(sort.Enum).To(Equal([]string{"name", "-name"}))
	})

	It("decodes the x-pagination extension", func() {
		Expect(list.Pagination).To(Equal(&model.Pagination{
			LimitParam: "page[size]", PageParam: "page[number]", ResultsPath: "data",
		}))
	})

	It("retains the results-array element $ref as ItemRefName, leaving ResponseDataRefName empty for a list", func() {
		Expect(list.ItemRefName).To(Equal("Thing"))
		Expect(list.ResponseDataRefName).To(BeEmpty())
	})

	It("retains a get-by-id data object $ref as ResponseDataRefName, leaving ItemRefName empty", func() {
		get := opByID(loadSpecMust("schema_normalize_list.yaml"), "GetThing")
		Expect(get.QueryParams).To(BeEmpty())
		Expect(get.PathParams).To(HaveLen(1))
		Expect(get.PathParams[0].Name).To(Equal("thing_id"))
		Expect(get.PathParams[0].Required).To(BeTrue())
		Expect(get.PathParams[0].Schema.Type).To(Equal("string"))
		Expect(get.Pagination).To(BeNil())
		Expect(get.ItemRefName).To(BeEmpty())
		Expect(get.ResponseDataRefName).To(Equal("Thing"))
	})
})

// paramByName returns the named query parameter or fails the test.
func paramByName(op *model.Operation, name string) model.QueryParam {
	GinkgoHelper()
	for _, p := range op.QueryParams {
		if p.Name == name {
			return p
		}
	}
	Fail("query parameter " + name + " not found on " + op.OperationId)
	return model.QueryParam{}
}

// -------------------------------------------------------------------
//  $ref with sibling keywords
// -------------------------------------------------------------------

// libopenapi surfaces a $ref with siblings as a synthetic allOf. Normalization
// must retain the referenced shape while preserving supported local annotation
// metadata such as descriptions.
var _ = Describe("NormalizeSchemas $ref with sibling keywords", func() {

	var props map[string]*model.Schema

	BeforeEach(func() {
		spec := loadSpecMust("schema_normalize_ref_siblings.yaml")
		op := opByID(spec, "GetSiblings")
		Expect(op.ResponseSchema).NotTo(BeNil())
		Expect(op.ResponseSchema.Kind).To(Equal(model.SchemaKindObject))
		props = op.ResponseSchema.Properties
	})

	// wantType is the OpenAPI "type" keyword the referenced schema declares.
	// normalizeSchema records it for every kind, not only primitives, so an object
	// alternative carries "object" here.
	DescribeTable("resolves the reference alongside supported sibling metadata",
		func(property string, wantKind model.SchemaKind, wantType string) {
			got := props[property]
			Expect(got).NotTo(BeNil(), "property %q missing from the normalized object", property)
			Expect(got.Kind).To(Equal(wantKind), "property %q normalized to the wrong kind", property)
			Expect(got.Type).To(Equal(wantType))
		},
		Entry("$ref + example", "named", model.SchemaKindPrimitive, "string"),
		Entry("$ref + description", "described", model.SchemaKindPrimitive, "string"),
		Entry("$ref + example + description", "multi", model.SchemaKindPrimitive, "string"),
		Entry("a bare $ref (control)", "plain", model.SchemaKindPrimitive, "string"),
		Entry("$ref to an object + example", "nested", model.SchemaKindObject, "object"),
	)

	It("keeps the referenced enum when a sibling is present", func() {
		Expect(props["kind"].Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(props["kind"].Enum).To(Equal([]string{"SECRET", "PUBLIC"}))
	})

	It("resolves a collection element that is a $ref with a sibling", func() {
		params := props["params"]
		Expect(params.Kind).To(Equal(model.SchemaKindArray))
		Expect(params.Items.Kind).To(Equal(model.SchemaKindObject))
		Expect(params.Items.Properties).To(HaveKey("name"))
	})

	It("resolves a $ref-with-sibling nested inside a referenced object", func() {
		// UrlParam.name is itself {$ref: TokenName, example: ...} — the shape that
		// blocked datadog_action_connection.
		name := props["nested"].Properties["name"]
		Expect(name).NotTo(BeNil())
		Expect(name.Kind).To(Equal(model.SchemaKindPrimitive))
		Expect(name.Type).To(Equal("string"))
	})
})

// schemaProperty returns a named normalized object property or fails the test.
func schemaProperty(schema *model.Schema, name string) *model.Schema {
	GinkgoHelper()
	Expect(schema).NotTo(BeNil())
	Expect(schema.Kind).To(Equal(model.SchemaKindObject))
	property, ok := schema.Properties[name]
	Expect(ok).To(BeTrue(), "schema property %s not found", name)
	Expect(property).NotTo(BeNil())
	return property
}
