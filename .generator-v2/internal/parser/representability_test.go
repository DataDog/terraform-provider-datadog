package parser

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// resourceOp builds a tracked resource operation whose group references the given
// CRUD operationIds, with the supplied request/response schemas.
func resourceOp(opID, method, path, artifact string, group *model.OperationGroup, req, resp *model.Schema) *model.Operation {
	return &model.Operation{
		OperationId:    opID,
		Method:         method,
		Path:           path,
		RequestSchema:  req,
		ResponseSchema: resp,
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind: model.ArtifactKindResource,
			ArtifactName: artifact,
			Group:        group,
		},
	}
}

func unrepresentableErr(err error) *UnrepresentableSchemaError {
	GinkgoHelper()
	Expect(err).To(HaveOccurred())
	var ue *UnrepresentableSchemaError
	Expect(errors.As(err, &ue)).To(BeTrue(), "expected *UnrepresentableSchemaError, got %T: %v", err, err)
	return ue
}

var _ = Describe("CheckSchemaRepresentability representable specs", func() {

	It("returns nil when every tracked schema is a primitive, object, array, or map", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				&model.Schema{Kind: model.SchemaKindObject, Properties: map[string]*model.Schema{
					"name":  {Kind: model.SchemaKindPrimitive, Type: "string"},
					"tags":  {Kind: model.SchemaKindArray, Items: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}},
					"attrs": {Kind: model.SchemaKindMap, Items: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}},
				}},
				&model.Schema{Kind: model.SchemaKindPrimitive, Type: "integer"}),
		}}
		Expect(CheckSchemaRepresentability(spec)).To(Succeed())
	})

	It("returns nil for a nil spec", func() {
		Expect(CheckSchemaRepresentability(nil)).To(Succeed())
	})

	It("ignores operations that have no filled schemas", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			{OperationId: "GetUntracked", Method: "GET", Path: "/untracked"},
		}}
		Expect(CheckSchemaRepresentability(spec)).To(Succeed())
	})
})

var _ = Describe("CheckSchemaRepresentability flagging", func() {

	It("flags a oneOf/anyOf request body as a variant at the request root", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				&model.Schema{Kind: model.SchemaKindVariant}, nil),
		}}
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		Expect(ue.Findings).To(HaveLen(1))
		Expect(ue.Findings[0].Kind).To(Equal(model.SchemaKindVariant))
		Expect(ue.Findings[0].SchemaPath).To(Equal("request"))
		Expect(ue.Findings[0].ArtifactName).To(Equal("thing"))
	})

	It("flags a ref_cycle nested inside an object property with its dotted path", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				&model.Schema{Kind: model.SchemaKindObject, Properties: map[string]*model.Schema{
					"spec": {Kind: model.SchemaKindRefCycle},
				}}, nil),
		}}
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		Expect(ue.Findings).To(HaveLen(1))
		Expect(ue.Findings[0].Kind).To(Equal(model.SchemaKindRefCycle))
		Expect(ue.Findings[0].SchemaPath).To(Equal("request.spec"))
	})

	It("flags a variant inside array elements with a [] path segment", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				nil,
				&model.Schema{Kind: model.SchemaKindArray, Items: &model.Schema{Kind: model.SchemaKindVariant}}),
		}}
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		Expect(ue.Findings).To(HaveLen(1))
		Expect(ue.Findings[0].SchemaPath).To(Equal("response[]"))
	})

	It("flags a variant inside map values with a {} path segment", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				&model.Schema{Kind: model.SchemaKindMap, Items: &model.Schema{Kind: model.SchemaKindVariant}}, nil),
		}}
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		Expect(ue.Findings).To(HaveLen(1))
		Expect(ue.Findings[0].SchemaPath).To(Equal("request{}"))
	})

	It("does not descend into a variant, so a variant nested in a variant yields a single finding", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				&model.Schema{Kind: model.SchemaKindVariant, Variants: []*model.Schema{
					{Kind: model.SchemaKindVariant},
					{Kind: model.SchemaKindPrimitive, Type: "string"},
				}}, nil),
		}}
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		Expect(ue.Findings).To(HaveLen(1))
		Expect(ue.Findings[0].SchemaPath).To(Equal("request"))
	})
})

var _ = Describe("CheckSchemaRepresentability artifact attribution", func() {

	It("attributes a finding on an untracked group member back to its artifact", func() {
		// The create op is tracked and names the read op in its group; the read op
		// itself carries no tracking but its response schema is unrepresentable.
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "GetThing"},
				&model.Schema{Kind: model.SchemaKindObject}, nil),
			{
				OperationId:    "GetThing",
				Method:         "GET",
				Path:           "/things/{id}",
				ResponseSchema: &model.Schema{Kind: model.SchemaKindVariant},
			},
		}}
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		Expect(ue.Findings).To(HaveLen(1))
		Expect(ue.Findings[0].OperationId).To(Equal("GetThing"))
		Expect(ue.Findings[0].ArtifactName).To(Equal("thing"),
			"a finding on the untracked read op must still name its owning artifact")
		Expect(ue.Findings[0].SchemaPath).To(Equal("response"))
	})
})

var _ = Describe("CheckSchemaRepresentability batching", func() {

	It("collects every unrepresentable node across operations and bodies in one error", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				&model.Schema{Kind: model.SchemaKindObject, Properties: map[string]*model.Schema{
					"spec": {Kind: model.SchemaKindVariant},
				}},
				&model.Schema{Kind: model.SchemaKindVariant}),
			resourceOp("CreateOther", "POST", "/others", "other",
				&model.OperationGroup{Create: "CreateOther", Read: "CreateOther"},
				&model.Schema{Kind: model.SchemaKindRefCycle}, nil),
		}}
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		Expect(ue.Findings).To(HaveLen(3))
	})

	It("sorts findings by artifact, then operation, then schema path", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				&model.Schema{Kind: model.SchemaKindObject, Properties: map[string]*model.Schema{
					"zeta":  {Kind: model.SchemaKindVariant},
					"alpha": {Kind: model.SchemaKindVariant},
				}},
				&model.Schema{Kind: model.SchemaKindVariant}),
			resourceOp("CreateAaa", "POST", "/aaa", "aaa",
				&model.OperationGroup{Create: "CreateAaa", Read: "CreateAaa"},
				&model.Schema{Kind: model.SchemaKindVariant}, nil),
		}}
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		paths := make([]string, 0, len(ue.Findings))
		artifacts := make([]string, 0, len(ue.Findings))
		for _, f := range ue.Findings {
			artifacts = append(artifacts, f.ArtifactName)
			paths = append(paths, f.SchemaPath)
		}
		// artifact "aaa" sorts before "thing"; within "thing", request.alpha <
		// request.zeta < response.
		Expect(artifacts).To(Equal([]string{"aaa", "thing", "thing", "thing"}))
		Expect(paths).To(Equal([]string{"request", "request.alpha", "request.zeta", "response"}))
	})
})

var _ = Describe("CheckSchemaRepresentability error message", func() {

	It("names the artifact, operation, location, and reason for each finding", func() {
		spec := &model.Spec{Operations: []*model.Operation{
			resourceOp("CreateThing", "POST", "/things", "thing",
				&model.OperationGroup{Create: "CreateThing", Read: "CreateThing"},
				&model.Schema{Kind: model.SchemaKindObject, Properties: map[string]*model.Schema{
					"spec": {Kind: model.SchemaKindVariant},
				}},
				&model.Schema{Kind: model.SchemaKindRefCycle}),
		}}
		msg := unrepresentableErr(CheckSchemaRepresentability(spec)).Error()

		Expect(msg).To(ContainSubstring("2 unrepresentable schema node(s)"))
		Expect(msg).To(ContainSubstring(`resource "thing"`))
		Expect(msg).To(ContainSubstring("CreateThing (POST /things)"))
		Expect(msg).To(ContainSubstring("at request.spec"))
		Expect(msg).To(ContainSubstring("oneOf/anyOf has no Terraform Plugin Framework equivalent (variant)"))
		Expect(msg).To(ContainSubstring("at response"))
		Expect(msg).To(ContainSubstring("$ref is circular or exceeds --max-depth (ref_cycle)"))
	})
})

var _ = Describe("CheckSchemaRepresentability over loaded fixtures", func() {

	It("flags the oneOf and anyOf request bodies normalized from schema_normalize_kinds.yaml", func() {
		spec := loadSpecMust("schema_normalize_kinds.yaml")
		ue := unrepresentableErr(CheckSchemaRepresentability(spec))

		ids := make([]string, 0, len(ue.Findings))
		for _, f := range ue.Findings {
			Expect(f.Kind).To(Equal(model.SchemaKindVariant))
			ids = append(ids, f.OperationId)
		}
		Expect(ids).To(ConsistOf("CreateVariantOneOf", "CreateVariantAnyOf"))
	})

	It("flags the depth-limited ref_cycle normalized from schema_normalize_refs.yaml at --max-depth 1", func() {
		spec, err := LoadSpec(filepath.Join("../testdata/parser", "schema_normalize_refs.yaml"), WithMaxDepth(1))
		Expect(err).To(Succeed())

		ue := unrepresentableErr(CheckSchemaRepresentability(spec))
		Expect(ue.Findings).To(HaveLen(1))
		Expect(ue.Findings[0].OperationId).To(Equal("CreateNestedRef"))
		Expect(ue.Findings[0].Kind).To(Equal(model.SchemaKindRefCycle))
		Expect(ue.Findings[0].SchemaPath).To(Equal("request.inner"))
	})

	It("returns nil for a fully representable loaded spec", func() {
		spec := loadSpecMust("schema_normalize_crud.yaml")
		Expect(CheckSchemaRepresentability(spec)).To(Succeed())
	})
})
