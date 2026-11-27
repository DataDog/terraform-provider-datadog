package emit

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// Include an extra object/list level to exercise the recursive walker as well
// as the plural builder's special handling of immediate item children.
func pluralDeepPresenceOperation() *model.Operation {
	op := pluralNestedOperation()
	part := op.ResponseSchema.Properties["data"].Items.Properties["attributes"].Properties["parts"].Items
	part.Properties["details"] = obj(map[string]*model.Schema{
		"entries": {Kind: model.SchemaKindArray, Description: "Detail entries.", Items: obj(map[string]*model.Schema{"value": prim("string", "Entry value.")})},
	})
	return op
}

var _ = ginkgo.Describe("nested attribute presence", func() {
	ginkgo.It("marks every plural result container and leaf computed-only", func() {
		for _, op := range []*model.Operation{pluralNestedOperation(), pluralObjectOperation(), pluralDeepPresenceOperation()} {
			view := mustView(op)
			var check func(AttrView, string)
			check = func(a AttrView, parent string) {
				path := parent + "." + a.TFName
				gomega.Expect(a.Computed).To(gomega.BeTrue(), path)
				gomega.Expect(a.Required).To(gomega.BeFalse(), path)
				gomega.Expect(a.Optional).To(gomega.BeFalse(), path)
				for _, child := range append(append([]AttrView{}, a.Attributes...), a.Blocks...) {
					check(child, path)
				}
			}
			for _, a := range view.Schema.Blocks {
				check(a, view.TypeName)
			}
		}
	})
	ginkgo.It("preserves resource nested input presence", func() {
		for _, tc := range []struct {
			name               string
			required, response bool
		}{
			{name: "required", required: true, response: true},
			{name: "optional"},
			{name: "optional and computed", response: true},
		} {
			op := widgetResourceOperation()
			if tc.required {
				attrs := op.ResolvedGroup.Create.RequestSchema.Properties["data"].Properties["attributes"]
				attrs.Required = append(attrs.Required, "settings")
			}
			if !tc.response {
				delete(op.ResolvedGroup.Read.ResponseSchema.Properties["data"].Properties["attributes"].Properties, "settings")
			}
			art, err := model.BuildArtifact(op)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			view, err := BuildResourceView(art)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			attr := attrByPath(schemaTree(view), "settings")
			gomega.Expect(attr.Required).To(gomega.Equal(tc.required), tc.name)
			gomega.Expect(attr.Optional).To(gomega.Equal(!tc.required), tc.name)
			gomega.Expect(attr.Computed).To(gomega.Equal(!tc.required && tc.response), tc.name)
		}
	})

})
