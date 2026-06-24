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

	It("prepends the lookup id and maps state off resp.Data through guarded optional getters", func() {
		view, err := BuildDataSourceView(incidentTypeArtifact())
		Expect(err).NotTo(HaveOccurred())

		Expect(view.Models).To(HaveLen(1))
		var fields []string
		for _, f := range view.Models[0].Fields {
			fields = append(fields, f.GoField)
		}
		Expect(fields).To(Equal([]string{"ID", "Description", "IsDefault", "Name"}))

		Expect(view.State.ParamName).To(Equal("resp"))
		Expect(view.State.ParamType).To(Equal("*datadogV2.IncidentTypeResponse"))
		Expect(view.State.Preamble).To(Equal([]string{"attributes := resp.Data.GetAttributes()"}))
		// Guarded assignments: an absent field stays null rather than a zero value.
		Expect(view.State.Assignments).To(Equal([]StateAssignment{
			{Var: "id", GetterOk: "resp.Data.GetIdOk()", LHS: "state.ID", RHS: "types.StringValue(*id)"},
			{Var: "description", GetterOk: "attributes.GetDescriptionOk()", LHS: "state.Description", RHS: "types.StringValue(*description)"},
			{Var: "isDefault", GetterOk: "attributes.GetIsDefaultOk()", LHS: "state.IsDefault", RHS: "types.BoolValue(*isDefault)"},
			{Var: "name", GetterOk: "attributes.GetNameOk()", LHS: "state.Name", RHS: "types.StringValue(*name)"},
		}))
	})

	It("renders a date-time string via .String(), a named enum via a string() cast, and avoids shadowing state", func() {
		op := incidentTypeOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"].Properties
		attrs["created_at"] = &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time"}
		attrs["state"] = &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"active", "archived"}}
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())

		assign := map[string]StateAssignment{}
		for _, a := range view.State.Assignments {
			assign[a.LHS] = a
		}
		Expect(assign["state.CreatedAt"].RHS).To(Equal("types.StringValue(createdAt.String())"))
		Expect(assign["state.CreatedAt"].GetterOk).To(Equal("attributes.GetCreatedAtOk()"))
		// "state" would shadow the updateState receiver, so its local is suffixed.
		Expect(assign["state.State"].Var).To(Equal("stateValue"))
		Expect(assign["state.State"].RHS).To(Equal("types.StringValue(string(*stateValue))"))
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

var _ = Describe("BuildDataSourceView singular search", func() {
	Context("search only", func() {
		It("binds the list call as Search, derives filters, and computes the id", func() {
			view, err := BuildDataSourceView(mustArtifact(powerpackSearchOperation()))
			Expect(err).NotTo(HaveOccurred())

			Expect(view.ByID).To(BeFalse())
			Expect(view.Searchable).To(BeTrue())
			Expect(view.Search.Method).To(Equal("ListPowerpacks"))
			Expect(view.Search.ItemType).To(Equal("PowerpackData"))
			Expect(view.Search.Paginated).To(BeTrue())

			// The lone filter becomes both an Optional schema attr and a list param.
			Expect(view.Search.Filters).To(Equal([]FilterParamView{
				{StateField: "FilterName", ParamField: "FilterName", ValueExpr: "ValueStringPointer()"},
			}))
			Expect(view.Schema.Blocks).To(BeEmpty(), "singular output has no list/items block")

			// The record reads off the list element by value, through guarded getters.
			Expect(view.State.ParamName).To(Equal("data"))
			Expect(view.State.ParamType).To(Equal("datadogV2.PowerpackData"))
			Expect(view.State.Preamble).To(Equal([]string{"attributes := data.GetAttributes()"}))
		})
	})

	Context("both", func() {
		It("binds the by-id Read and the list Search and makes the id optional+computed", func() {
			view, err := BuildDataSourceView(mustArtifact(datastoreBothOperation()))
			Expect(err).NotTo(HaveOccurred())

			Expect(view.ByID).To(BeTrue())
			Expect(view.Searchable).To(BeTrue())
			Expect(view.Read.Method).To(Equal("GetDatastore"))
			Expect(view.Search.Method).To(Equal("ListDatastores"))
			Expect(view.State.ParamType).To(Equal("datadogV2.DatastoreData"))
		})
	})

	DescribeTable("the emitted Read guards the result count and indexes only the single match",
		func(fixture func() *model.Operation) {
			got, err := RenderDataSource(mustView(fixture()))
			Expect(err).NotTo(HaveOccurred())
			src := string(got)
			Expect(src).To(ContainSubstring(`if len(items) == 0 {`))
			Expect(src).To(ContainSubstring(`response.Diagnostics.AddError("filters returned no results", "")`))
			Expect(src).To(ContainSubstring(`if len(items) > 1 {`))
			Expect(src).To(ContainSubstring(`use more specific search criteria`))
			Expect(src).To(ContainSubstring(`d.updateState(&state, items[0])`))
		},
		Entry("search only", powerpackSearchOperation),
		Entry("both", datastoreBothOperation),
	)

	It("absent fields stay null: every record assignment is a guarded optional getter", func() {
		got, err := RenderDataSource(mustView(datastoreBothOperation()))
		Expect(err).NotTo(HaveOccurred())
		src := string(got)
		Expect(src).To(ContainSubstring(`if name, ok := attributes.GetNameOk(); ok && name != nil {`))
		Expect(src).NotTo(ContainSubstring(`types.StringValue(attributes.GetName())`), "must not write the unguarded zero value")
	})
})

// mustView builds an Artifact and its view from op or fails the test.
func mustView(op *model.Operation) DataSourceView {
	GinkgoHelper()
	view, err := BuildDataSourceView(mustArtifact(op))
	Expect(err).NotTo(HaveOccurred())
	return view
}

// powerpackSearchOperation is a search-only singular data source: the list GET is
// the tracked op, paginated, with one scalar filter. (A representative server-side
// search shape; the real powerpack matches client-side, which is out of scope.)
func powerpackSearchOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/powerpacks",
		Method:          "GET",
		OperationId:     "ListPowerpacks",
		Tag:             "Powerpacks",
		ResponseRefName: "PowerpacksResponse",
		ItemRefName:     "PowerpackData",
		Pagination:      &model.Pagination{LimitParam: "page[limit]", PageParam: "page[offset]", ResultsPath: "data"},
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "powerpack",
			TfDescription: "Use this data source to retrieve information about an existing Datadog Powerpack.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Search: "ListPowerpacks"},
		},
		QueryParams: []model.QueryParam{
			{Name: "filter[name]", Schema: prim("string", ""), Description: "The name of the Powerpack to search for."},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind: model.SchemaKindArray,
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The ID of the Powerpack."),
					"type": prim("string", "Type of widget, must be `powerpack`."),
					"attributes": obj(map[string]*model.Schema{
						"description": prim("string", "Description of the powerpack."),
						"name":        prim("string", "The name of the powerpack."),
					}),
				}),
			},
		}),
	}
}

// datastoreBothOperation is an id-optional singular data source: the tracked op is
// the by-id GET, and SearchOp points at the list GET (no query params, matching the
// real ListDatastores). Its element mirrors data_source_datadog_datastore.go.
func datastoreBothOperation() *model.Operation {
	listOp := &model.Operation{
		Path:            "/api/v2/actions-datastores",
		Method:          "GET",
		OperationId:     "ListDatastores",
		Tag:             "Actions Datastores",
		ResponseRefName: "DatastoreArray",
		ItemRefName:     "DatastoreData",
		ResponseSchema:  obj(map[string]*model.Schema{"data": {Kind: model.SchemaKindArray, Items: datastoreElement()}}),
	}
	return &model.Operation{
		Path:            "/api/v2/actions-datastores/{datastore_id}",
		Method:          "GET",
		OperationId:     "GetDatastore",
		Tag:             "Actions Datastores",
		ResponseRefName: "Datastore",
		SearchOp:        listOp,
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "datastore",
			TfDescription: "Use this data source to retrieve information about an existing Datadog datastore.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetDatastore", Search: "ListDatastores"},
		},
		ResponseSchema: obj(map[string]*model.Schema{"data": datastoreElement()}),
	}
}

// datastoreElement is the JSON:API datastore element ({id,type,attributes}) shared
// by the by-id and list responses, mirroring data_source_datadog_datastore.go.
func datastoreElement() *model.Schema {
	return obj(map[string]*model.Schema{
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
	})
}

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
