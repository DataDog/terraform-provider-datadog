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

	It("renders a date-time string via .String() and a named enum via a string() cast", func() {
		op := incidentTypeOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"].Properties
		attrs["created_at"] = &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time"}
		attrs["state"] = &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"active", "archived"}}
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())

		rhs := map[string]string{}
		for _, a := range view.State.Assignments {
			rhs[a.LHS] = a.RHS
		}
		Expect(rhs["state.CreatedAt"]).To(Equal("types.StringValue(attributes.GetCreatedAt().String())"))
		Expect(rhs["state.State"]).To(Equal("types.StringValue(string(attributes.GetState()))"))
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

var _ = Describe("BuildDataSourceView plural", func() {
	It("builds the teams plural view end-to-end, matching the golden-backing fixture", func() {
		art, err := model.BuildArtifact(teamsOperation())
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view).To(Equal(pluralFixture()))
	})

	It("drops array and enum query params from the filter set", func() {
		art, err := model.BuildArtifact(teamsOperation())
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())

		var names []string
		for _, f := range view.Schema.Attributes {
			names = append(names, f.TFName)
		}
		Expect(names).To(Equal([]string{"filter_keyword", "filter_me"}))
	})

	It("hashes a fixed seed when an endpoint has no filters", func() {
		op := teamsOperation()
		op.QueryParams = nil
		op.Pagination = nil
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Read.Filters).To(BeEmpty())
		Expect(view.Read.OptionalParamsType).To(BeEmpty())
		Expect(view.Schema.Attributes).To(BeEmpty())
	})

	It("produces a deeply-equal plural view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(teamsOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(teamsOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})

	DescribeTable("fail-slows unsupported item-element nodes into one UnsupportedEmitError",
		func(mutate func(*model.Operation), wantReason string) {
			op := teamsOperation()
			mutate(op)
			view, err := BuildDataSourceView(mustArtifact(op))
			var uerr *UnsupportedEmitError
			Expect(errors.As(err, &uerr)).To(BeTrue(), "expected an UnsupportedEmitError, got %v", err)
			Expect(uerr.Error()).To(ContainSubstring(wantReason))
			Expect(view).To(Equal(DataSourceView{}), "no view should be produced on failure")
		},
		Entry("a nested object under item attributes",
			func(op *model.Operation) {
				teamAttrs(op)["nested"] = obj(map[string]*model.Schema{"x": prim("string", "")})
			},
			"nesting under item attributes is not supported"),
		Entry("a missing item element type",
			func(op *model.Operation) { op.ItemRefName = "" },
			"missing list item type"),
	)
})

// mustArtifact builds an Artifact from op or fails the test.
func mustArtifact(op *model.Operation) *model.Artifact {
	GinkgoHelper()
	art, err := model.BuildArtifact(op)
	Expect(err).NotTo(HaveOccurred())
	return art
}

// teamAttrs returns the item-element attributes map of a teams operation, for
// mutation in fail-slow cases.
func teamAttrs(op *model.Operation) map[string]*model.Schema {
	return op.ResponseSchema.Properties["data"].Items.Properties["attributes"].Properties
}

// teamsOperation is the teams list GET as a parser-shaped Operation: a paginated
// JSON:API collection whose response carries metadata siblings (meta/links/
// included) the builder must drop, plus array and enum query params it must drop
// from the filter set. Descriptions mirror the golden-backing pluralFixture.
func teamsOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/team",
		Method:          "GET",
		OperationId:     "ListTeams",
		Tag:             "Teams",
		ResponseRefName: "TeamsResponse",
		ItemRefName:     "Team",
		Pagination:      &model.Pagination{LimitParam: "page[size]", PageParam: "page[number]", ResultsPath: "data"},
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "teams",
			Cardinality:   model.CardinalityPlural,
			TfDescription: "Use this data source to retrieve information about existing teams for use in other resources.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "ListTeams"},
		},
		QueryParams: []model.QueryParam{
			{Name: "filter[keyword]", Schema: prim("string", ""), Description: "Search query. Can be team name, team handle, or email of team member."},
			{Name: "filter[me]", Schema: prim("boolean", ""), Description: "When true, only returns teams the current user belongs to."},
			{Name: "include", Schema: &model.Schema{Kind: model.SchemaKindArray, Items: prim("string", "")}},
			{Name: "page[number]", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "integer", Format: "int64"}},
			{Name: "page[size]", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "integer", Format: "int64"}},
			{Name: "sort", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"name"}}},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind:        model.SchemaKindArray,
				Description: "List of teams",
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The team's identifier."),
					"type": prim("string", "Team resource type."),
					"attributes": obj(map[string]*model.Schema{
						"description": prim("string", "Free-form markdown description/content for the team's homepage."),
						"handle":      prim("string", "The team's handle."),
						"link_count":  prim("integer", "The number of links belonging to the team."),
						"name":        prim("string", "The name of the team."),
						"summary":     prim("string", "A brief summary of the team, derived from the `description`."),
						"user_count":  prim("integer", "The number of users belonging to the team."),
					}),
				}),
			},
			// Response metadata siblings: the model keeps only the results array.
			"meta":     obj(map[string]*model.Schema{"x": prim("string", "")}),
			"links":    obj(map[string]*model.Schema{"x": prim("string", "")}),
			"included": {Kind: model.SchemaKindArray, Items: obj(map[string]*model.Schema{"x": prim("string", "")})},
		}),
	}
}

// datastoresOperation is the ListDatastores GET: a non-paginated list with no
// query parameters, exercising the no-optional-params call form and the
// zero-filter id seed. Its element attributes include a date-time and an enum,
// mirroring the singular data_source_datadog_datastore.go mapping.
func datastoresOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/actions-datastores",
		Method:          "GET",
		OperationId:     "ListDatastores",
		Tag:             "Actions Datastores",
		ResponseRefName: "DatastoreArray",
		ItemRefName:     "DatastoreData",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "datastores",
			Cardinality:   model.CardinalityPlural,
			TfDescription: "Use this data source to retrieve information about existing Datadog datastores.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "ListDatastores"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind:        model.SchemaKindArray,
				Description: "An array of datastore objects containing their configurations and metadata.",
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The unique identifier of the datastore."),
					"type": prim("string", "The resource type for datastores."),
					"attributes": obj(map[string]*model.Schema{
						"created_at":                      {Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time", Description: "Timestamp when the datastore was created."},
						"creator_user_id":                 prim("integer", "The numeric ID of the user who created the datastore."),
						"creator_user_uuid":               prim("string", "The UUID of the user who created the datastore."),
						"description":                     prim("string", "A human-readable description about the datastore."),
						"modified_at":                     {Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time", Description: "Timestamp when the datastore was last modified."},
						"name":                            prim("string", "The display name of the datastore."),
						"org_id":                          prim("integer", "The ID of the organization that owns this datastore."),
						"primary_column_name":             prim("string", "The name of the primary key column for this datastore."),
						"primary_key_generation_strategy": {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"none", "uuid"}, Description: "Strategy for generating primary keys."},
					}),
				}),
			},
		}),
	}
}
