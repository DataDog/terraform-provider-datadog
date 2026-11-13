package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-only resource inputs", func() {
	DescribeTable("checks nested structure separately from immutable leaf values",
		func(withUpdate bool) {
			create := objSchema(map[string]*Schema{
				"scope":     objSchema(map[string]*Schema{"name": primSchema("string")}),
				"selectors": arrSchema(objSchema(map[string]*Schema{"name": primSchema("string")})),
				"capacity":  primSchema("integer"),
			})
			read := CloneSchema(create)
			read.Properties["scope"].Properties["output"] = primSchema("string")
			read.Properties["selectors"].Items.Properties["output"] = primSchema("string")
			group := &ResolvedGroup{Create: &Operation{RequestSchema: create}, Read: &Operation{ResponseSchema: read}}
			if withUpdate {
				group.Update = &Operation{RequestSchema: objSchema(map[string]*Schema{"capacity": primSchema("integer")})}
			}
			merged, _, err := MergeResourceSchema(group)
			Expect(err).NotTo(HaveOccurred())
			tree, _, err := BuildResourceTree(merged, !withUpdate)
			Expect(err).NotTo(HaveOccurred())
			Expect(attrByPath(tree, "resource.scope").PlanModifiers).To(ContainElement(PlanModifierSpec{Name: "planmodifiers.RequiresReplaceObjectStructure"}))
			Expect(attrByPath(tree, "resource.selectors").PlanModifiers).To(ContainElement(PlanModifierSpec{Name: "planmodifiers.RequiresReplaceListStructure"}))
			for _, path := range []string{"resource.scope.name", "resource.selectors[].name"} {
				Expect(attrByPath(tree, path).PlanModifiers).To(ContainElement(RequiresReplaceSpec("types.String")))
			}
			for _, path := range []string{"resource.scope.output", "resource.selectors[].output"} {
				Expect(attrByPath(tree, path).PlanModifiers).To(BeEmpty())
			}
		},
		Entry("with mutable sibling inputs", true),
		Entry("without an Update operation", false),
	)

	It("replaces immutable inputs while retaining mutable siblings and server outputs", func() {
		create := objSchema(map[string]*Schema{
			"window": primSchema("string"),
			"scope": objSchema(map[string]*Schema{
				"team": primSchema("string"),
				"name": primSchema("string"),
			}),
			"tags": mapSchema(primSchema("string")),
			"selectors": arrSchema(objSchema(map[string]*Schema{
				"dimension": primSchema("string"),
			})),
			"capacity": primSchema("integer"),
		})
		create.Required = []string{"window", "capacity"}
		update := objSchema(map[string]*Schema{
			"capacity": primSchema("integer"),
			"scope": objSchema(map[string]*Schema{
				"name": primSchema("string"),
			}),
		})
		read := CloneSchema(create)
		read.Properties["monitor_id"] = primSchema("integer")
		merged, _, err := MergeResourceSchema(&ResolvedGroup{
			Create: &Operation{RequestSchema: create},
			Update: &Operation{RequestSchema: update},
			Read:   &Operation{ResponseSchema: read},
		})
		Expect(err).NotTo(HaveOccurred())
		tree, _, err := BuildResourceTree(merged, false)
		Expect(err).NotTo(HaveOccurred())
		for path, modifier := range map[string]string{
			"resource.window":     "stringplanmodifier.RequiresReplace",
			"resource.scope.team": "stringplanmodifier.RequiresReplace",
			"resource.tags":       "mapplanmodifier.RequiresReplace",
			"resource.selectors":  "planmodifiers.RequiresReplaceListStructure",
		} {
			Expect(attrByPath(tree, path).PlanModifiers).To(ContainElement(PlanModifierSpec{Name: modifier}), path)
		}
		for _, path := range []string{"resource.capacity", "resource.scope", "resource.scope.name", "resource.monitor_id"} {
			for _, modifier := range attrByPath(tree, path).PlanModifiers {
				Expect(modifier.Name).NotTo(ContainSubstring("RequiresReplace"), path)
			}
		}
	})

	It("ignores read-only request properties when inferring lifecycle presence", func() {
		readOnly := func() *Schema {
			s := primSchema("string")
			s.ReadOnly = true
			return s
		}
		create := objSchema(map[string]*Schema{"immutable": primSchema("string"), "output": readOnly(), "request_only": readOnly()})
		update := objSchema(map[string]*Schema{"immutable": readOnly(), "output": readOnly()})
		merged, _, err := MergeResourceSchema(&ResolvedGroup{
			Create: &Operation{RequestSchema: create},
			Update: &Operation{RequestSchema: update},
			Read: &Operation{ResponseSchema: objSchema(map[string]*Schema{
				"immutable": primSchema("string"), "output": primSchema("string"),
			})},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(merged.Properties).NotTo(HaveKey("request_only"))
		Expect(merged.Properties["immutable"].Provenance.CreateOnly).To(BeTrue())
		Expect(merged.Properties["output"].Provenance).To(Equal(&SchemaProvenance{InResponse: true}))
		Expect(create.Properties["output"].ReadOnly).To(BeTrue(), "merging must not mutate its inputs")
	})
})
