package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// FrameworkType is the context-blind type table: one schema node in, its two
// framework type strings out. Objects always map to the block form here — the
// builder rewrites to attribute form, not this function (see schema_test.go).
var _ = Describe("FrameworkType", func() {

	DescribeTable("maps a representable schema node to its framework type strings",
		func(s *Schema, wantTf, wantGo string) {
			tf, goType, err := FrameworkType(s)
			Expect(err).NotTo(HaveOccurred())
			Expect(tf).To(Equal(wantTf))
			Expect(goType).To(Equal(wantGo))
		},
		Entry("a string primitive becomes StringAttribute / types.String",
			&Schema{Kind: SchemaKindPrimitive, Type: "string"},
			"schema.StringAttribute", "types.String"),
		Entry("an integer primitive becomes Int64Attribute / types.Int64",
			&Schema{Kind: SchemaKindPrimitive, Type: "integer"},
			"schema.Int64Attribute", "types.Int64"),
		Entry("an int32 integer still becomes Int64Attribute (format ignored)",
			&Schema{Kind: SchemaKindPrimitive, Type: "integer", Format: "int32"},
			"schema.Int64Attribute", "types.Int64"),
		Entry("an int64 integer still becomes Int64Attribute (format ignored)",
			&Schema{Kind: SchemaKindPrimitive, Type: "integer", Format: "int64"},
			"schema.Int64Attribute", "types.Int64"),
		Entry("a number primitive becomes Float64Attribute / types.Float64",
			&Schema{Kind: SchemaKindPrimitive, Type: "number"},
			"schema.Float64Attribute", "types.Float64"),
		Entry("a double number still becomes Float64Attribute (format ignored)",
			&Schema{Kind: SchemaKindPrimitive, Type: "number", Format: "double"},
			"schema.Float64Attribute", "types.Float64"),
		Entry("a boolean primitive becomes BoolAttribute / types.Bool",
			&Schema{Kind: SchemaKindPrimitive, Type: "boolean"},
			"schema.BoolAttribute", "types.Bool"),
		Entry("an object becomes SingleNestedBlock / types.Object (block-default, context-blind)",
			&Schema{Kind: SchemaKindObject},
			"schema.SingleNestedBlock", "types.Object"),
		Entry("an array of primitive becomes ListAttribute / types.List",
			&Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}},
			"schema.ListAttribute", "types.List"),
		Entry("an array of object becomes ListNestedBlock / types.List",
			&Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindObject}},
			"schema.ListNestedBlock", "types.List"),
		Entry("a map of primitive becomes MapAttribute / types.Map",
			&Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}},
			"schema.MapAttribute", "types.Map"),
		Entry("a map of object becomes MapNestedAttribute / types.Map",
			&Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindObject}},
			"schema.MapNestedAttribute", "types.Map"),
	)

	DescribeTable("returns an error naming the offender for an unrepresentable node",
		func(s *Schema, wantSubstr string) {
			_, _, err := FrameworkType(s)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(wantSubstr))
		},
		Entry("a variant (oneOf/anyOf) has no framework equivalent",
			&Schema{Kind: SchemaKindVariant}, "variant"),
		Entry("a ref_cycle has no framework equivalent",
			&Schema{Kind: SchemaKindRefCycle}, "ref_cycle"),
		Entry("an unsupported node has no framework equivalent",
			&Schema{Kind: SchemaKindUnsupported}, "unsupported"),
		Entry("a primitive with an unrecognized type names that type",
			&Schema{Kind: SchemaKindPrimitive, Type: "decimal"}, "decimal"),
		Entry("a primitive with an empty type cannot be mapped",
			&Schema{Kind: SchemaKindPrimitive, Type: ""}, "primitive type"),
		Entry("an array with nil items has no element type to map",
			&Schema{Kind: SchemaKindArray}, "nil items"),
		Entry("an array of array is deferred — error names the element kind",
			&Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}}},
			`"array"`),
		Entry("an array of map is deferred — error names the element kind",
			&Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}}},
			`"map"`),
		Entry("a map of map is deferred — error names the value kind",
			&Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}}},
			`"map"`),
	)
})

// ElementType is only meaningful for collection-of-primitive nodes; nested
// (object) collections carry their shape in Children, not an element type.
var _ = Describe("ElementType", func() {

	DescribeTable("maps a primitive element/value schema to its framework attr.Type",
		func(elem *Schema, want string) {
			got, err := ElementType(elem)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("a string element becomes types.StringType",
			&Schema{Kind: SchemaKindPrimitive, Type: "string"}, "types.StringType"),
		Entry("an integer element becomes types.Int64Type",
			&Schema{Kind: SchemaKindPrimitive, Type: "integer"}, "types.Int64Type"),
		Entry("a number element becomes types.Float64Type",
			&Schema{Kind: SchemaKindPrimitive, Type: "number"}, "types.Float64Type"),
		Entry("a boolean element becomes types.BoolType",
			&Schema{Kind: SchemaKindPrimitive, Type: "boolean"}, "types.BoolType"),
	)

	DescribeTable("errors for a non-primitive element (those use nested forms, not ElementType)",
		func(elem *Schema) {
			_, err := ElementType(elem)
			Expect(err).To(HaveOccurred())
		},
		Entry("an object element has no scalar element type",
			&Schema{Kind: SchemaKindObject}),
		Entry("an array element has no scalar element type",
			&Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}}),
		Entry("a map element has no scalar element type",
			&Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}}),
	)
})
