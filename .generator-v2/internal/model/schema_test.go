package model

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ---------------------------------------------------------------------------
//  Schema construction helpers — keep the table cases readable. Set Enum /
//  Sensitive / Description / Format on the returned node inline where a case
//  needs them.
// ---------------------------------------------------------------------------

func primSchema(typ string) *Schema { return &Schema{Kind: SchemaKindPrimitive, Type: typ} }
func objSchema(props map[string]*Schema) *Schema {
	return &Schema{Kind: SchemaKindObject, Properties: props}
}
func arrSchema(item *Schema) *Schema { return &Schema{Kind: SchemaKindArray, Items: item} }
func mapSchema(val *Schema) *Schema  { return &Schema{Kind: SchemaKindMap, Items: val} }
func oneOfSchema(variants ...*Schema) *Schema {
	return &Schema{Kind: SchemaKindVariant, Variants: variants}
}

// richSchema is a response schema exercising one of every shape, including the
// block→attribute switch (cfg is a map<object{settings:object}>). A fresh tree
// is built each call so callers never share mutable state.
func richSchema() *Schema {
	status := primSchema("string")
	status.Enum = []string{"ok", "warn", "alert"}
	return objSchema(map[string]*Schema{
		"id":      primSchema("string"),
		"status":  status,
		"tags":    arrSchema(primSchema("string")),
		"options": objSchema(map[string]*Schema{"notify": primSchema("boolean")}),
		"items":   arrSchema(objSchema(map[string]*Schema{"name": primSchema("string")})),
		"meta":    mapSchema(primSchema("string")),
		"cfg": mapSchema(objSchema(map[string]*Schema{
			"settings": objSchema(map[string]*Schema{"x": primSchema("string")}),
		})),
	})
}

// attrByPath descends the tree (into Children) for the attribute at path, failing
// the spec if it is absent.
func attrByPath(tree *AttributeTree, path string) *Attribute {
	GinkgoHelper()
	var find func(attrs []*Attribute) *Attribute
	find = func(attrs []*Attribute) *Attribute {
		for _, a := range attrs {
			if a.Path == path {
				return a
			}
			if got := find(a.Children); got != nil {
				return got
			}
		}
		return nil
	}
	got := find(tree.Attributes)
	Expect(got).NotTo(BeNil(), "no attribute at path %q", path)
	return got
}

// allAttrs flattens every attribute in the tree, parents before children.
func allAttrs(tree *AttributeTree) []*Attribute {
	var out []*Attribute
	var walk func(attrs []*Attribute)
	walk = func(attrs []*Attribute) {
		for _, a := range attrs {
			out = append(out, a)
			walk(a.Children)
		}
	}
	walk(tree.Attributes)
	return out
}

// pathsOf is the Path of each attribute, in slice order.
func pathsOf(attrs []*Attribute) []string {
	out := make([]string, len(attrs))
	for i, a := range attrs {
		out[i] = a.Path
	}
	return out
}

// ---------------------------------------------------------------------------
//  Path root & overall shape
// ---------------------------------------------------------------------------

var _ = Describe("BuildResponseTree / BuildRequestTree path root and shape", func() {

	It("roots a response object's property paths at response.", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"name": primSchema("string")}))
		Expect(err).NotTo(HaveOccurred())
		Expect(tree.Attributes).To(HaveLen(1))
		Expect(tree.Attributes[0].Path).To(Equal("response.name"))
	})

	It("roots the same schema at request. through the other entry point (shared core)", func() {
		tree, _, err := BuildRequestTree(objSchema(map[string]*Schema{"name": primSchema("string")}))
		Expect(err).NotTo(HaveOccurred())
		Expect(tree.Attributes).To(HaveLen(1))
		Expect(tree.Attributes[0].Path).To(Equal("request.name"))
	})

	It("explodes a root object's properties into top-level attributes without wrapping the object", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"a": primSchema("string"),
			"b": primSchema("integer"),
		}))
		Expect(err).NotTo(HaveOccurred())
		Expect(pathsOf(tree.Attributes)).To(Equal([]string{"response.a", "response.b"}))
		Expect(tree.Attributes[0].TfType).To(Equal("schema.StringAttribute"))
	})

	It("builds a root array-of-object as one attribute at response with [] element children", func() {
		tree, _, err := BuildResponseTree(arrSchema(objSchema(map[string]*Schema{"name": primSchema("string")})))
		Expect(err).NotTo(HaveOccurred())
		Expect(tree.Attributes).To(HaveLen(1))
		root := tree.Attributes[0]
		Expect(root.Path).To(Equal("response"))
		Expect(root.TfType).To(Equal("schema.ListNestedBlock"))
		Expect(root.GoType).To(Equal("types.List"))
		Expect(root.Children).To(HaveLen(1))
		Expect(root.Children[0].Path).To(Equal("response[].name"))
	})

	It("builds a root primitive as one attribute at response", func() {
		tree, _, err := BuildResponseTree(primSchema("boolean"))
		Expect(err).NotTo(HaveOccurred())
		Expect(tree.Attributes).To(HaveLen(1))
		Expect(tree.Attributes[0].Path).To(Equal("response"))
		Expect(tree.Attributes[0].TfType).To(Equal("schema.BoolAttribute"))
	})

	It("returns an empty tree for a nil schema (no body → no attributes)", func() {
		tree, _, err := BuildResponseTree(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(tree.Attributes).To(BeEmpty())
	})
})

// ---------------------------------------------------------------------------
//  Flags, description, defaults
// ---------------------------------------------------------------------------

var _ = Describe("BuildResponseTree flags, description and defaults", func() {

	It("marks every attribute Computed and never Required or Optional (data-source MVP)", func() {
		tree, _, err := BuildResponseTree(richSchema())
		Expect(err).NotTo(HaveOccurred())
		for _, a := range allAttrs(tree) {
			Expect(a.Computed).To(BeTrue(), "attr %s must be Computed", a.Path)
			Expect(a.Required).To(BeFalse(), "attr %s must not be Required", a.Path)
			Expect(a.Optional).To(BeFalse(), "attr %s must not be Optional", a.Path)
		}
	})

	It("copies Sensitive from the schema node", func() {
		secret := primSchema("string")
		secret.Sensitive = true
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"secret": secret,
			"plain":  primSchema("string"),
		}))
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.secret").Sensitive).To(BeTrue())
		Expect(attrByPath(tree, "response.plain").Sensitive).To(BeFalse())
	})

	It("copies Description from the schema node", func() {
		thing := primSchema("string")
		thing.Description = "the thing"
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"thing": thing}))
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.thing").Description).To(Equal("the thing"))
	})

	It("leaves Default nil on every attribute", func() {
		tree, _, err := BuildResponseTree(richSchema())
		Expect(err).NotTo(HaveOccurred())
		for _, a := range allAttrs(tree) {
			Expect(a.Default).To(BeNil(), "attr %s must have a nil Default", a.Path)
		}
	})

	It("maps an int32 integer property to Int64Attribute (format ignored)", func() {
		i32 := primSchema("integer")
		i32.Format = "int32"
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"n": i32}))
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.n").TfType).To(Equal("schema.Int64Attribute"))
	})
})

// ---------------------------------------------------------------------------
//  Type delegation & composites
// ---------------------------------------------------------------------------

var _ = Describe("BuildResponseTree type delegation and composites", func() {

	It("builds array<string> as a ListAttribute leaf carrying ElementType", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"tags": arrSchema(primSchema("string"))}))
		Expect(err).NotTo(HaveOccurred())
		tags := attrByPath(tree, "response.tags")
		Expect(tags.TfType).To(Equal("schema.ListAttribute"))
		Expect(tags.GoType).To(Equal("types.List"))
		Expect(tags.ElementType).To(Equal("types.StringType"))
		Expect(tags.Children).To(BeEmpty())
	})

	It("builds array<object> as a ListNestedBlock with [] element children and no ElementType", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"items": arrSchema(objSchema(map[string]*Schema{"name": primSchema("string")})),
		}))
		Expect(err).NotTo(HaveOccurred())
		items := attrByPath(tree, "response.items")
		Expect(items.TfType).To(Equal("schema.ListNestedBlock"))
		Expect(items.GoType).To(Equal("types.List"))
		Expect(items.ElementType).To(Equal(""))
		Expect(pathsOf(items.Children)).To(Equal([]string{"response.items[].name"}))
	})

	It("builds map<string> as a MapAttribute leaf carrying ElementType", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"labels": mapSchema(primSchema("string"))}))
		Expect(err).NotTo(HaveOccurred())
		labels := attrByPath(tree, "response.labels")
		Expect(labels.TfType).To(Equal("schema.MapAttribute"))
		Expect(labels.GoType).To(Equal("types.Map"))
		Expect(labels.ElementType).To(Equal("types.StringType"))
		Expect(labels.Children).To(BeEmpty())
	})

	It("builds map<object> as a MapNestedAttribute with {} value children", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"configs": mapSchema(objSchema(map[string]*Schema{"x": primSchema("string")})),
		}))
		Expect(err).NotTo(HaveOccurred())
		configs := attrByPath(tree, "response.configs")
		Expect(configs.TfType).To(Equal("schema.MapNestedAttribute"))
		Expect(configs.GoType).To(Equal("types.Map"))
		Expect(pathsOf(configs.Children)).To(Equal([]string{"response.configs{}.x"}))
	})

	It("builds a nested object (no map ancestor) as a SingleNestedBlock with .key children", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"options": objSchema(map[string]*Schema{"notify": primSchema("boolean")}),
		}))
		Expect(err).NotTo(HaveOccurred())
		options := attrByPath(tree, "response.options")
		Expect(options.TfType).To(Equal("schema.SingleNestedBlock"))
		Expect(options.GoType).To(Equal("types.Object"))
		Expect(pathsOf(options.Children)).To(Equal([]string{"response.options.notify"}))
		Expect(options.Children[0].TfType).To(Equal("schema.BoolAttribute"))
	})

	It("surfaces a FrameworkType error for a deferred composite property (array-of-array)", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"matrix": arrSchema(arrSchema(primSchema("string"))),
		}))
		Expect(tree).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("array"))
	})
})

// ---------------------------------------------------------------------------
//  Nesting-form context switch (block ↔ attribute)
// ---------------------------------------------------------------------------

var _ = Describe("BuildResponseTree nesting-form context switch", func() {

	It("rewrites object and array-of-object fields to attribute form inside a map value", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"cfg": mapSchema(objSchema(map[string]*Schema{
				"settings": objSchema(map[string]*Schema{"x": primSchema("string")}),
				"hist":     arrSchema(objSchema(map[string]*Schema{"y": primSchema("string")})),
			})),
		}))
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.cfg{}.settings").TfType).To(Equal("schema.SingleNestedAttribute"))
		Expect(attrByPath(tree, "response.cfg{}.hist").TfType).To(Equal("schema.ListNestedAttribute"))
		// primitive leaves are identical in both worlds, and the switch propagates deeper
		Expect(attrByPath(tree, "response.cfg{}.settings.x").TfType).To(Equal("schema.StringAttribute"))
	})

	It("keeps nested objects in block form outside any map (object-in-object, object-in-list)", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"outer": objSchema(map[string]*Schema{
				"inner": objSchema(map[string]*Schema{"v": primSchema("string")}),
			}),
			"list": arrSchema(objSchema(map[string]*Schema{
				"elem": objSchema(map[string]*Schema{"w": primSchema("string")}),
			})),
		}))
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.outer").TfType).To(Equal("schema.SingleNestedBlock"))
		Expect(attrByPath(tree, "response.outer.inner").TfType).To(Equal("schema.SingleNestedBlock"))
		Expect(attrByPath(tree, "response.list").TfType).To(Equal("schema.ListNestedBlock"))
		Expect(attrByPath(tree, "response.list[].elem").TfType).To(Equal("schema.SingleNestedBlock"))
	})
})

// ---------------------------------------------------------------------------
//  Enums → validators
// ---------------------------------------------------------------------------

var _ = Describe("BuildResponseTree enums to validators", func() {

	It("turns a string enum into a stringvalidator.OneOf validator with Go-quoted args", func() {
		status := primSchema("string")
		status.Enum = []string{"active", "inactive", "pending"}
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"status": status}))
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.status").Validators).To(Equal([]ValidatorSpec{
			{Name: "stringvalidator.OneOf", Args: []string{`"active"`, `"inactive"`, `"pending"`}},
		}))
	})

	It("adds no validator to a property without an enum", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"name": primSchema("string")}))
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.name").Validators).To(BeEmpty())
	})

	It("adds no validator to a non-string (integer) enum property", func() {
		level := primSchema("integer")
		level.Enum = []string{"1", "2", "3"}
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"level": level}))
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.level").Validators).To(BeEmpty())
	})
})

// ---------------------------------------------------------------------------
//  Determinism & ordering
// ---------------------------------------------------------------------------

var _ = Describe("BuildResponseTree determinism and ordering", func() {

	It("sorts top-level attributes and all Children ascending by Path", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"zeta":  primSchema("string"),
			"alpha": primSchema("string"),
			"mid": objSchema(map[string]*Schema{
				"z_child": primSchema("string"),
				"a_child": primSchema("string"),
			}),
		}))
		Expect(err).NotTo(HaveOccurred())
		Expect(pathsOf(tree.Attributes)).To(Equal([]string{"response.alpha", "response.mid", "response.zeta"}))
		Expect(pathsOf(attrByPath(tree, "response.mid").Children)).To(Equal([]string{"response.mid.a_child", "response.mid.z_child"}))
	})

	It("produces deeply-equal trees across two builds of the same input", func() {
		s := richSchema()
		first, _, err := BuildResponseTree(s)
		Expect(err).NotTo(HaveOccurred())
		second, _, err := BuildResponseTree(s)
		Expect(err).NotTo(HaveOccurred())
		Expect(reflect.DeepEqual(first, second)).To(BeTrue(), "two builds of the same schema must be deeply equal")
	})
})

// ---------------------------------------------------------------------------
//  Defensive guard (kinds the representability gate should have rejected)
// ---------------------------------------------------------------------------

var _ = Describe("BuildResponseTree defensive guard", func() {

	DescribeTable("returns a typed *UnsupportedKindError naming the node when a non-representable kind reaches the builder",
		func(badKind SchemaKind) {
			tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
				"self": {Kind: badKind},
			}))
			Expect(tree).To(BeNil())
			var uke *UnsupportedKindError
			Expect(errors.As(err, &uke)).To(BeTrue(), "expected *UnsupportedKindError, got %T: %v", err, err)
			Expect(uke.Path).To(Equal("response.self"))
			Expect(uke.Kind).To(Equal(badKind))
			// Error() names the path and kind, and flags the broken invariant.
			Expect(uke.Error()).To(SatisfyAll(
				ContainSubstring(`"response.self"`),
				ContainSubstring(string(badKind)),
				ContainSubstring("not representable"),
			))
		},
		Entry("a ref_cycle node", SchemaKindRefCycle),
		Entry("an unsupported node", SchemaKindUnsupported),
	)

	DescribeTable("propagates the guard error with the full nested path to the offending node",
		func(schema *Schema, wantPath string) {
			tree, _, err := BuildResponseTree(schema)
			Expect(tree).To(BeNil())
			var uke *UnsupportedKindError
			Expect(errors.As(err, &uke)).To(BeTrue(), "expected *UnsupportedKindError, got %T: %v", err, err)
			Expect(uke.Path).To(Equal(wantPath))
			Expect(uke.Kind).To(Equal(SchemaKindRefCycle))
		},
		Entry("inside a nested object",
			objSchema(map[string]*Schema{"outer": objSchema(map[string]*Schema{"bad": {Kind: SchemaKindRefCycle}})}),
			"response.outer.bad"),
		Entry("inside an array-of-object element",
			objSchema(map[string]*Schema{"list": arrSchema(objSchema(map[string]*Schema{"bad": {Kind: SchemaKindRefCycle}}))}),
			"response.list[].bad"),
		Entry("inside a map-of-object value",
			objSchema(map[string]*Schema{"m": mapSchema(objSchema(map[string]*Schema{"bad": {Kind: SchemaKindRefCycle}}))}),
			"response.m{}.bad"),
	)

	It("returns the type-mapping error for a non-object root that cannot be mapped (array-of-array)", func() {
		tree, _, err := BuildResponseTree(arrSchema(arrSchema(primSchema("string"))))
		Expect(tree).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("array"))
	})

	It("builds a well-formed object nested several levels deep without error", func() {
		deep := objSchema(map[string]*Schema{
			"l1": objSchema(map[string]*Schema{
				"l2": objSchema(map[string]*Schema{
					"l3": objSchema(map[string]*Schema{"leaf": primSchema("string")}),
				}),
			}),
		})
		tree, _, err := BuildResponseTree(deep)
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "response.l1.l2.l3.leaf").TfType).To(Equal("schema.StringAttribute"))
	})
})

var _ = Describe("BuildResponseTree retains oneOf envelopes", func() {

	assertRetained := func(schema *Schema, path string) {
		GinkgoHelper()

		tree, diags, err := BuildResponseTree(schema)
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, path)).NotTo(BeNil())
		for _, diag := range diags {
			Expect(diag.Message).NotTo(
				ContainSubstring("dropped"),
				"oneOf at %q must not be silently dropped", path,
			)
		}
	}

	It("retains a root oneOf as an envelope at the response root", func() {
		assertRetained(
			oneOfSchema(primSchema("boolean"), primSchema("string")),
			"response",
		)
	})

	It("retains a oneOf object property as an envelope attribute", func() {
		assertRetained(
			objSchema(map[string]*Schema{
				"choice": oneOfSchema(
					primSchema("string"),
					objSchema(map[string]*Schema{"name": primSchema("string")}),
				),
			}),
			"response.choice",
		)
	})

	It("retains a oneOf nested inside another object", func() {
		assertRetained(
			objSchema(map[string]*Schema{
				"container": objSchema(map[string]*Schema{
					"choice": oneOfSchema(
						objSchema(map[string]*Schema{"enabled": primSchema("boolean")}),
						objSchema(map[string]*Schema{"threshold": primSchema("number")}),
					),
				}),
			}),
			"response.container.choice",
		)
	})

	DescribeTable("retains a collection whose element is a oneOf envelope",
		func(collection *Schema, path string) {
			assertRetained(
				objSchema(map[string]*Schema{"choices": collection}),
				path,
			)
		},
		Entry(
			"array element",
			arrSchema(oneOfSchema(primSchema("boolean"), primSchema("string"))),
			"response.choices",
		),
		Entry(
			"map value",
			mapSchema(oneOfSchema(primSchema("integer"), primSchema("string"))),
			"response.choices",
		),
	)
})

// ---------------------------------------------------------------------------
//  Golden full-tree shape — locks the whole conversion in one assertion
// ---------------------------------------------------------------------------

var _ = Describe("BuildResponseTree golden tree", func() {

	It("builds a schema with every shape into exactly the expected attribute tree", func() {
		tree, _, err := BuildResponseTree(richSchema())
		Expect(err).NotTo(HaveOccurred())

		want := &AttributeTree{Attributes: []*Attribute{
			{Path: "response.cfg", TfType: "schema.MapNestedAttribute", GoType: "types.Map", Computed: true,
				Children: []*Attribute{
					{Path: "response.cfg{}.settings", TfType: "schema.SingleNestedAttribute", GoType: "types.Object", Computed: true,
						Children: []*Attribute{
							{Path: "response.cfg{}.settings.x", TfType: "schema.StringAttribute", GoType: "types.String", Computed: true},
						}},
				}},
			{Path: "response.id", TfType: "schema.StringAttribute", GoType: "types.String", Computed: true},
			{Path: "response.items", TfType: "schema.ListNestedBlock", GoType: "types.List", Computed: true,
				Children: []*Attribute{
					{Path: "response.items[].name", TfType: "schema.StringAttribute", GoType: "types.String", Computed: true},
				}},
			{Path: "response.meta", TfType: "schema.MapAttribute", GoType: "types.Map", ElementType: "types.StringType", Computed: true},
			{Path: "response.options", TfType: "schema.SingleNestedBlock", GoType: "types.Object", Computed: true,
				Children: []*Attribute{
					{Path: "response.options.notify", TfType: "schema.BoolAttribute", GoType: "types.Bool", Computed: true},
				}},
			{Path: "response.status", TfType: "schema.StringAttribute", GoType: "types.String", Computed: true, IsEnum: true,
				Validators: []ValidatorSpec{{Name: "stringvalidator.OneOf", Args: []string{`"ok"`, `"warn"`, `"alert"`}}}},
			{Path: "response.tags", TfType: "schema.ListAttribute", GoType: "types.List", ElementType: "types.StringType", Computed: true},
		}}

		Expect(tree).To(Equal(want))
	})
})
