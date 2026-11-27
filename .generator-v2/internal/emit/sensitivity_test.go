package emit

import (
	"bytes"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = ginkgo.Describe("plural result sensitivity", func() {
	ginkgo.It("preserves sensitivity on scalar, collection, object, and union item attributes", func() {
		for _, sensitive := range []bool{false, true} {
			shapes := map[string]*model.Schema{
				"scalar":  prim("string", "A value."),
				"list":    {Kind: model.SchemaKindArray, Items: prim("string", "An item.")},
				"map":     {Kind: model.SchemaKindMap, Items: prim("string", "A value.")},
				"object":  obj(map[string]*model.Schema{"value": prim("string", "A value.")}),
				"objects": {Kind: model.SchemaKindArray, Items: obj(map[string]*model.Schema{"value": prim("string", "A value.")})},
				"union":   unionSchema("response.data[].attributes.secret_value", "SecretValue"),
			}
			for name, shape := range shapes {
				op := pluralNestedOperation()
				shape.Sensitive = sensitive
				op.ResponseSchema.Properties["data"].Sensitive = sensitive
				op.ResponseSchema.Properties["data"].Items.Properties["attributes"].Properties["secret_value"] = shape
				view := unionView(op)
				results := view.Schema.Blocks[0]
				gomega.Expect(results.Sensitive).To(gomega.Equal(sensitive), name+" result container")
				var found bool
				for _, a := range append(append([]AttrView{}, results.Attributes...), results.Blocks...) {
					if a.TFName != "secret_value" {
						gomega.Expect(a.Sensitive).To(gomega.BeFalse(), name+" public sibling "+a.TFName)
						continue
					}
					found = true
					gomega.Expect(a.Sensitive).To(gomega.Equal(sensitive), name)
					gomega.Expect(a.Computed).To(gomega.BeTrue(), name)
					gomega.Expect(a.Required || a.Optional).To(gomega.BeFalse(), name)
					templateName := "schemaAttribute"
					if a.IsBlock {
						templateName = "schemaNestedAttribute"
					}
					var rendered bytes.Buffer
					gomega.Expect(templates.ExecuteTemplate(&rendered, templateName, a)).To(gomega.Succeed())
					gomega.Expect(bytes.Contains(rendered.Bytes(), []byte("Sensitive: true"))).To(gomega.Equal(sensitive), name)
				}
				gomega.Expect(found).To(gomega.BeTrue(), name)
			}
		}
	})
})
