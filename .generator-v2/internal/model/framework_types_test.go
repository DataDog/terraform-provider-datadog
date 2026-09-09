package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// FrameworkType is the context-blind type table: one schema node in, its two
// framework type strings out. New protocol-v6 schemas use nested attributes for
// object containers so callers never need a compatibility-only block form.
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
		Entry("an object becomes SingleNestedAttribute / types.Object",
			&Schema{Kind: SchemaKindObject},
			"schema.SingleNestedAttribute", "types.Object"),
		Entry("an array of primitive becomes ListAttribute / types.List",
			&Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}},
			"schema.ListAttribute", "types.List"),
		Entry("an array of array becomes ListAttribute / types.List",
			arrSchema(arrSchema(primSchema("string"))),
			"schema.ListAttribute", "types.List"),
		Entry("an array of map becomes ListAttribute / types.List",
			arrSchema(mapSchema(primSchema("string"))),
			"schema.ListAttribute", "types.List"),
		Entry("an array of object becomes ListNestedAttribute / types.List",
			&Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindObject}},
			"schema.ListNestedAttribute", "types.List"),
		Entry("a map of primitive becomes MapAttribute / types.Map",
			&Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"}},
			"schema.MapAttribute", "types.Map"),
		Entry("a map of array becomes MapAttribute / types.Map",
			mapSchema(arrSchema(primSchema("string"))),
			"schema.MapAttribute", "types.Map"),
		Entry("a map of map becomes MapAttribute / types.Map",
			mapSchema(mapSchema(primSchema("string"))),
			"schema.MapAttribute", "types.Map"),
		Entry("a map of object becomes MapNestedAttribute / types.Map",
			&Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindObject}},
			"schema.MapNestedAttribute", "types.Map"),
		Entry("an array of oneOf nests its element's variant attributes, like an array of object",
			&Schema{Kind: SchemaKindArray, Items: &Schema{Kind: SchemaKindOneOf}},
			"schema.ListNestedAttribute", "types.List"),
		Entry("a map of oneOf nests its value's variant attributes, like a map of object",
			&Schema{Kind: SchemaKindMap, Items: &Schema{Kind: SchemaKindOneOf}},
			"schema.MapNestedAttribute", "types.Map"),
	)

	DescribeTable("returns an error naming the offender for an unrepresentable node",
		func(s *Schema, wantSubstr string) {
			_, _, err := FrameworkType(s)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(wantSubstr))
		},
		Entry("a oneOf has no direct framework equivalent",
			&Schema{Kind: SchemaKindOneOf}, "one_of"),
		Entry("a ref_cycle has no framework equivalent",
			&Schema{Kind: SchemaKindRefCycle}, "ref_cycle"),
		Entry("an unsupported node has no framework equivalent",
			&Schema{Kind: SchemaKindUnsupported}, "unsupported"),
		// The reason is the actionable half of the message: a cycle names the
		// chain that produced it. Reached through a list or a map the node is not
		// the one being mapped, so its explanation has to be carried across.
		Entry("a cyclic node carries its own explanation",
			&Schema{Kind: SchemaKindRefCycle, UnsupportedReason: "circular $ref: A -> B -> A"},
			"circular $ref: A -> B -> A"),
		Entry("a cyclic array element carries its explanation",
			&Schema{Kind: SchemaKindArray, Items: &Schema{
				Kind: SchemaKindRefCycle, UnsupportedReason: "circular $ref: A -> B -> A"}},
			"circular $ref: A -> B -> A"),
		Entry("a cyclic map value carries its explanation",
			&Schema{Kind: SchemaKindMap, Items: &Schema{
				Kind: SchemaKindRefCycle, UnsupportedReason: "circular $ref: A -> B -> A"}},
			"circular $ref: A -> B -> A"),
		Entry("a primitive with an unrecognized type names that type",
			&Schema{Kind: SchemaKindPrimitive, Type: "decimal"}, "decimal"),
		Entry("a primitive with an empty type cannot be mapped",
			&Schema{Kind: SchemaKindPrimitive, Type: ""}, "primitive type"),
		Entry("an array with nil items has no element type to map",
			&Schema{Kind: SchemaKindArray}, "nil items"),
	)
})

// ElementType represents any list/map chain ending in a primitive; object
// collections carry their shape in Children instead.
var _ = Describe("ElementType", func() {

	DescribeTable("maps a primitive-terminal element/value schema to its framework attr.Type",
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
		Entry("an array element becomes a recursive ListType",
			arrSchema(primSchema("string")), "types.ListType{ElemType: types.StringType}"),
		Entry("a map element becomes a recursive MapType",
			mapSchema(primSchema("string")), "types.MapType{ElemType: types.StringType}"),
		Entry("a deep map-map-list element preserves every level",
			mapSchema(mapSchema(arrSchema(primSchema("string")))),
			"types.MapType{ElemType: types.MapType{ElemType: types.ListType{ElemType: types.StringType}}}"),
	)

	DescribeTable("errors for elements that do not terminate in a primitive",
		func(elem *Schema) {
			_, err := ElementType(elem)
			Expect(err).To(HaveOccurred())
		},
		Entry("an object element has no scalar element type",
			&Schema{Kind: SchemaKindObject}),
		Entry("an array ending in unsupported JSON has no element type",
			arrSchema(&Schema{Kind: SchemaKindUnsupported})),
		Entry("a map ending in an object uses nested form",
			mapSchema(objSchema(nil))),
	)
})
