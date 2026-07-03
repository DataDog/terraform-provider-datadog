package emit

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("BuildDataSourceView", func() {
	It("resolves the SDK call bindings onto the view", func() {
		view, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())
		Expect(view.TypeName).To(Equal("incident_type"))
		Expect(view.GoName).To(Equal("incidentType"))
		Expect(view.Description).To(Equal("Use this data source to retrieve information about an existing incident type."))
		Expect(view.SDKPackage).To(Equal("datadogV2"))
		Expect(view.APIStruct).To(Equal("IncidentsApi"))
		Expect(view.APIAccessor).To(Equal("GetIncidentsApiV2"))
		Expect(view.Read.Method).To(Equal("GetIncidentType"))
		Expect(view.Read.ResponseType).To(Equal("IncidentTypeResponse"))
	})

	It("flattens the envelope: data.attributes.* become top-level computed attributes, sorted, with no nested blocks", func() {
		view, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Schema.Blocks).To(BeEmpty())

		var names []string
		for _, a := range view.Schema.Attributes {
			names = append(names, a.TFName)
			Expect(a.Computed).To(BeTrue(), "attribute %q should be computed", a.TFName)
		}
		Expect(names).To(Equal([]string{"description", "is_default", "name"}))
	})

	It("prepends the lookup id and maps state off resp.Data and the attributes local", func() {
		view, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())

		Expect(view.Models).To(HaveLen(1))
		var fields []string
		for _, f := range view.Models[0].Fields {
			fields = append(fields, f.GoField)
		}
		Expect(fields).To(Equal([]string{"ID", "Description", "IsDefault", "Name"}))

		Expect(view.State.Preamble).To(Equal([]string{"attributes := resp.Data.GetAttributes()"}))
		Expect(view.State.Assignments).To(Equal([]StateAssignment{
			{LHS: "state.ID", RHS: "types.StringValue(resp.Data.GetId())"},
			{LHS: "state.Description", RHS: "types.StringValue(attributes.GetDescription())"},
			{LHS: "state.IsDefault", RHS: "types.BoolValue(attributes.GetIsDefault())"},
			{LHS: "state.Name", RHS: "types.StringValue(attributes.GetName())"},
		}))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})

	DescribeTable("fail-slows anything outside the recognized envelope",
		func(mutate func(*model.Operation), wantReason string) {
			op := incidentTypeOperation()
			mutate(op)
			art, err := model.BuildArtifact(op)
			Expect(err).NotTo(HaveOccurred())

			view, err := BuildDataSourceView(art)
			var uerr *UnsupportedEmitError
			Expect(errors.As(err, &uerr)).To(BeTrue(), "expected an UnsupportedEmitError, got %v", err)
			Expect(uerr.Error()).To(ContainSubstring(wantReason))
			Expect(view).To(Equal(DataSourceView{}), "no view should be produced on failure")
		},
		Entry("a data member outside {id, type, attributes}",
			func(op *model.Operation) {
				op.ResponseSchema.Properties["data"].Properties["relationships"] = prim("string", "")
			},
			"relationships is not part of the recognized"),
		Entry("a non-envelope response root",
			func(op *model.Operation) {
				op.ResponseSchema = obj(map[string]*model.Schema{"name": prim("string", "")})
			},
			"expected a single-member JSON:API envelope"),
		Entry("a nested object under attributes",
			func(op *model.Operation) {
				op.ResponseSchema.Properties["data"].Properties["attributes"].Properties["nested"] =
					obj(map[string]*model.Schema{"x": prim("string", "")})
			},
			"nesting under attributes is not supported"),
		Entry("an id_strategy other than data.id",
			func(op *model.Operation) { op.Tracking.IdStrategy = model.IdStrategyDataAttributesUID },
			"id_strategy"),
		Entry("a missing response type name",
			func(op *model.Operation) { op.ResponseRefName = "" },
			"missing response type name"),
	)
})

// prim and obj build model.Schema nodes for the emit fixtures (the model package
// keeps its own equivalents; these avoid a cross-package test dependency).
func prim(typ, desc string) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindPrimitive, Type: typ, Description: desc}
}

func obj(props map[string]*model.Schema) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindObject, Properties: props}
}

// incidentTypeOperation is the incident_type GET as a parser-shaped Operation: a
// JSON:API envelope ({data:{id,type,attributes:{description,is_default,name}}}).
func incidentTypeOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/incidents/config/types/{incident_type_id}",
		Method:          "GET",
		OperationId:     "GetIncidentType",
		Tag:             "Incidents",
		ResponseRefName: "IncidentTypeResponse",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "incident_type",
			TfDescription: "Use this data source to retrieve information about an existing incident type.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetIncidentType"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The incident type's ID."),
				"type": prim("string", "Incident type resource type."),
				"attributes": obj(map[string]*model.Schema{
					"name":        prim("string", "Name of the incident type."),
					"description": prim("string", "Description of the incident type."),
					"is_default":  prim("boolean", "Whether this incident type is the default type."),
				}),
			}),
		}),
	}
}

// incidentTypeArtifact resolves incidentTypeOperation into an *model.Artifact.
func incidentTypeArtifact() *model.Artifact {
	GinkgoHelper()
	art, err := model.BuildArtifact(incidentTypeOperation())
	Expect(err).NotTo(HaveOccurred())
	return art
}
