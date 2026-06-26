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
		Expect(view.GoName).To(Equal("datadogIncidentType"))
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
		Entry("a non-envelope response root",
			func(op *model.Operation) {
				op.ResponseSchema = obj(map[string]*model.Schema{"name": prim("string", "")})
			},
			"expected a single-member JSON:API envelope"),
		Entry("an id_strategy other than data.id",
			func(op *model.Operation) { op.Tracking.IdStrategy = model.IdStrategyDataAttributesUID },
			"id_strategy"),
		Entry("a missing response type name",
			func(op *model.Operation) { op.ResponseRefName = "" },
			"missing response type name"),
	)

	It("drops a data member outside {id, type, attributes} and records it", func() {
		op := incidentTypeOperation()
		op.ResponseSchema.Properties["data"].Properties["relationships"] =
			obj(map[string]*model.Schema{"created_by": prim("string", "")})
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildDataSourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Dropped).To(ContainElement(ContainSubstring("relationships")))
	})
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

// teamSingularOperation is the team GET-by-id as a parser-shaped Operation: a
// JSON:API envelope whose attributes carry scalar leaves plus two string arrays
// (visible_modules/hidden_modules), exercising collection-of-primitive hoisting.
func teamSingularOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/team/{team_id}",
		Method:          "GET",
		OperationId:     "GetTeam",
		Tag:             "Teams",
		ResponseRefName: "TeamResponse",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "team",
			TfDescription: "Use this data source to retrieve information about an existing Datadog team.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetTeam"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The team's identifier."),
				"type": prim("string", "Team resource type."),
				"attributes": obj(map[string]*model.Schema{
					"handle":          prim("string", "The team's handle."),
					"name":            prim("string", "The name of the team."),
					"visible_modules": {Kind: model.SchemaKindArray, Description: "Collection of visible modules for the team.", Items: prim("string", "String identifier of the module.")},
					"hidden_modules":  {Kind: model.SchemaKindArray, Description: "Collection of hidden modules for the team.", Items: prim("string", "String identifier of the module.")},
				}),
			}),
		}),
	}
}

// costBudgetOperation is the cost budget GET-by-id as a parser-shaped Operation: a
// JSON:API envelope whose attributes carry a name plus an entries array of objects,
// each holding scalars and a nested tag_filters array of objects — exercising
// recursive array-of-object hoisting.
func costBudgetOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/cost/budget/{budget_id}",
		Method:          "GET",
		OperationId:     "GetBudget",
		Tag:             "Cloud Cost Management",
		ResponseRefName: "BudgetWithEntries",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "cost_budget",
			TfDescription: "Use this data source to retrieve information about an existing Datadog cost budget.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetBudget"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The budget's identifier."),
				"type": prim("string", "Budget resource type."),
				"attributes": obj(map[string]*model.Schema{
					"name": prim("string", "The name of the budget."),
					"entries": {Kind: model.SchemaKindArray, Description: "The list of monthly budget entries.", Items: obj(map[string]*model.Schema{
						"amount": prim("number", "The budgeted amount for this entry."),
						"month":  prim("integer", "The month this budget entry applies to."),
						"tag_filters": {Kind: model.SchemaKindArray, Description: "The list of tag filters scoping this entry.", Items: obj(map[string]*model.Schema{
							"tag_key":   prim("string", "The tag key to filter on."),
							"tag_value": prim("string", "The tag value to filter on."),
						})},
					})},
				}),
			}),
		}),
	}
}

var _ = Describe("BuildDataSourceView singular nested arrays", func() {
	It("hoists an object array into a ListNestedBlock and recurses into nested object arrays", func() {
		view, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())

		Expect(view.Schema.Blocks).To(HaveLen(1))
		entries := view.Schema.Blocks[0]
		Expect(entries.TFName).To(Equal("entries"))
		Expect(entries.ListBlock).To(BeTrue())

		var entryAttrs []string
		for _, a := range entries.Attributes {
			entryAttrs = append(entryAttrs, a.TFName)
		}
		Expect(entryAttrs).To(Equal([]string{"amount", "month"}))
		Expect(entries.Blocks).To(HaveLen(1))
		Expect(entries.Blocks[0].TFName).To(Equal("tag_filters"))
		Expect(entries.Blocks[0].ListBlock).To(BeTrue())
	})

	It("generates a nested model struct per object level, parent first", func() {
		view, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())

		var names []string
		for _, m := range view.Models {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"datadogCostBudgetDataSourceModel", "EntriesModel", "TagFiltersModel"}))
	})

	It("maps each element through a guarded loop, recursing for nested arrays", func() {
		view, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())

		Expect(view.State.Lists).To(HaveLen(1))
		entries := view.State.Lists[0]
		Expect(entries.Kind).To(Equal("object"))
		Expect(entries.LHS).To(Equal("state.Entries"))
		Expect(entries.GetterOk).To(Equal("attributes.GetEntriesOk()"))
		Expect(entries.LoopVar).To(Equal("entriesItem"))
		Expect(entries.ElemVar).To(Equal("entriesModel"))
		Expect(entries.ElemStruct).To(Equal("EntriesModel"))
		Expect(entries.Scalars).To(ContainElement(StateAssignment{
			Var: "amount", GetterOk: "entriesItem.GetAmountOk()",
			LHS: "entriesModel.Amount", RHS: "types.Float64Value(*amount)",
		}))

		Expect(entries.Lists).To(HaveLen(1))
		tagFilters := entries.Lists[0]
		Expect(tagFilters.Kind).To(Equal("object"))
		Expect(tagFilters.LHS).To(Equal("entriesModel.TagFilters"))
		Expect(tagFilters.GetterOk).To(Equal("entriesItem.GetTagFiltersOk()"))
		Expect(tagFilters.LoopVar).To(Equal("tagFiltersItem"))
		Expect(tagFilters.ElemStruct).To(Equal("TagFiltersModel"))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(costBudgetOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})
})

var _ = Describe("BuildDataSourceView singular arrays", func() {
	It("hoists a string array under attributes into a ListAttribute carrying its element type", func() {
		view, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())

		attrs := map[string]AttrView{}
		for _, a := range view.Schema.Attributes {
			attrs[a.TFName] = a
		}
		Expect(attrs["visible_modules"].TFType).To(Equal("schema.ListAttribute"))
		Expect(attrs["visible_modules"].ElementType).To(Equal("types.StringType"))
		Expect(attrs["visible_modules"].Computed).To(BeTrue())
		Expect(view.Schema.Blocks).To(BeEmpty(), "a collection-of-primitive is a leaf attribute, not a block")
	})

	It("declares the list field as a types.List in the model", func() {
		view, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())

		goTypes := map[string]string{}
		for _, f := range view.Models[0].Fields {
			goTypes[f.TFName] = f.GoType
		}
		Expect(goTypes["visible_modules"]).To(Equal("types.List"))
		Expect(goTypes["hidden_modules"]).To(Equal("types.List"))
	})

	It("maps each list through a guarded ListValueFrom assignment", func() {
		view, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())

		// Sorted by attribute name, both string arrays become guarded primitive lists.
		Expect(view.State.Lists).To(Equal([]ListAssignment{
			{Kind: "primitive", LHS: "state.HiddenModules", GetterOk: "attributes.GetHiddenModulesOk()", Var: "hiddenModules", ElementType: "types.StringType"},
			{Kind: "primitive", LHS: "state.VisibleModules", GetterOk: "attributes.GetVisibleModulesOk()", Var: "visibleModules", ElementType: "types.StringType"},
		}))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(teamSingularOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})
})

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
						"description":     prim("string", "Free-form markdown description/content for the team's homepage."),
						"handle":          prim("string", "The team's handle."),
						"hidden_modules":  {Kind: model.SchemaKindArray, Description: "Collection of hidden modules for the team.", Items: prim("string", "String identifier of the module.")},
						"link_count":      prim("integer", "The number of links belonging to the team."),
						"name":            prim("string", "The name of the team."),
						"summary":         prim("string", "A brief summary of the team, derived from the `description`."),
						"user_count":      prim("integer", "The number of users belonging to the team."),
						"visible_modules": {Kind: model.SchemaKindArray, Description: "Collection of visible modules for the team.", Items: prim("string", "String identifier of the module.")},
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

// pluralNestedOperation is a synthetic plural list whose item attributes carry an
// object array (parts), each part an object with scalars — exercising the
// array-of-object item path that walks an element struct inside buildPluralView.
func pluralNestedOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/widgets",
		Method:          "GET",
		OperationId:     "ListWidgets",
		Tag:             "Widgets",
		ResponseRefName: "WidgetsResponse",
		ItemRefName:     "Widget",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "widgets",
			Cardinality:   model.CardinalityPlural,
			TfDescription: "Use this data source to retrieve information about existing widgets.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "ListWidgets"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind:        model.SchemaKindArray,
				Description: "List of widgets",
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The widget's identifier."),
					"type": prim("string", "Widget resource type."),
					"attributes": obj(map[string]*model.Schema{
						"name": prim("string", "The name of the widget."),
						"parts": {Kind: model.SchemaKindArray, Description: "The parts that make up the widget.", Items: obj(map[string]*model.Schema{
							"label":    prim("string", "The label of the part."),
							"quantity": prim("integer", "How many of this part the widget uses."),
						})},
					}),
				}),
			},
		}),
	}
}

var _ = Describe("BuildDataSourceView plural nested arrays", func() {
	It("renders an object array in an item as a ListNestedBlock with a generated element struct", func() {
		view, err := BuildDataSourceView(mustArtifact(pluralNestedOperation()))
		Expect(err).NotTo(HaveOccurred())

		items := view.Schema.Blocks[0]
		Expect(items.TFName).To(Equal("widgets"))
		Expect(items.Blocks).To(HaveLen(1))
		Expect(items.Blocks[0].TFName).To(Equal("parts"))
		Expect(items.Blocks[0].ListBlock).To(BeTrue())

		var names []string
		for _, m := range view.Models {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"datadogWidgetsDataSourceModel", "WidgetModel", "PartsModel"}))
	})

	It("maps the object array off item.Attributes into the item accumulator", func() {
		view, err := BuildDataSourceView(mustArtifact(pluralNestedOperation()))
		Expect(err).NotTo(HaveOccurred())

		Expect(view.State.ItemLists).To(HaveLen(1))
		parts := view.State.ItemLists[0]
		Expect(parts.Kind).To(Equal("object"))
		Expect(parts.LHS).To(Equal("r.Parts"))
		Expect(parts.GetterOk).To(Equal("item.Attributes.GetPartsOk()"))
		Expect(parts.LoopVar).To(Equal("partsItem"))
		Expect(parts.ElemStruct).To(Equal("PartsModel"))
		Expect(parts.Scalars).To(ContainElement(StateAssignment{
			Var: "label", GetterOk: "partsItem.GetLabelOk()",
			LHS: "partsModel.Label", RHS: "types.StringValue(*label)",
		}))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(pluralNestedOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(pluralNestedOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})
})

// retentionFilterOperation is the apm retention filter GET-by-id as a parser-shaped
// Operation: a JSON:API envelope whose attributes carry scalars plus a nested filter
// object — itself holding a scalar, a string array, and a nested metadata object —
// exercising bare-object hoisting, recursion, and composition with arrays.
func retentionFilterOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/apm/config/retention-filters/{filter_id}",
		Method:          "GET",
		OperationId:     "GetApmRetentionFilter",
		Tag:             "APM Retention Filters",
		ResponseRefName: "RetentionFilterResponse",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "apm_retention_filter",
			TfDescription: "Use this data source to retrieve information about an existing APM retention filter.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "GetApmRetentionFilter"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The retention filter's ID."),
				"type": prim("string", "Retention filter resource type."),
				"attributes": obj(map[string]*model.Schema{
					"enabled": prim("boolean", "Whether the retention filter is active."),
					"name":    prim("string", "The name of the retention filter."),
					"filter": obj(map[string]*model.Schema{
						"query": prim("string", "The search query defining the filter."),
						"tags":  {Kind: model.SchemaKindArray, Description: "Tags scoping the filter.", Items: prim("string", "A tag identifier.")},
						"metadata": obj(map[string]*model.Schema{
							"created_by": prim("string", "Handle of the user who created the filter."),
						}),
					}),
				}),
			}),
		}),
	}
}

var _ = Describe("BuildDataSourceView singular nested objects", func() {
	It("hoists a bare object under attributes into a SingleNestedBlock", func() {
		view, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())

		blocks := map[string]AttrView{}
		for _, b := range view.Schema.Blocks {
			blocks[b.TFName] = b
		}
		Expect(blocks).To(HaveKey("filter"))
		Expect(blocks["filter"].ListBlock).To(BeFalse(), "a bare object is a SingleNestedBlock, not a list block")
	})

	It("generates one model struct per object level, parent first", func() {
		view, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())

		var names []string
		for _, m := range view.Models {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"datadogApmRetentionFilterDataSourceModel", "FilterModel", "MetadataModel"}))
	})

	It("maps the object through a guarded assignment, recursing into the nested object", func() {
		view, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())

		var filter ListAssignment
		for _, l := range view.State.Lists {
			if l.LHS == "state.Filter" {
				filter = l
			}
		}
		Expect(filter.Kind).To(Equal("object_single"))
		Expect(filter.GetterOk).To(Equal("attributes.GetFilterOk()"))
		Expect(filter.Var).To(Equal("filter"))
		Expect(filter.ElemVar).To(Equal("filterModel"))
		Expect(filter.ElemStruct).To(Equal("FilterModel"))
		Expect(filter.Scalars).To(ContainElement(StateAssignment{
			Var: "query", GetterOk: "filter.GetQueryOk()",
			LHS: "filterModel.Query", RHS: "types.StringValue(*query)",
		}))

		var metadata ListAssignment
		for _, l := range filter.Lists {
			if l.Kind == "object_single" {
				metadata = l
			}
		}
		Expect(metadata.LHS).To(Equal("filterModel.Metadata"))
		Expect(metadata.GetterOk).To(Equal("filter.GetMetadataOk()"))
		Expect(metadata.ElemStruct).To(Equal("MetadataModel"))
	})

	It("renders the guarded object block and the recursive assignment in updateState", func() {
		got, err := RenderDataSource(mustView(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())
		src := string(got)
		Expect(src).To(ContainSubstring("if filter, ok := attributes.GetFilterOk(); ok && filter != nil {"))
		Expect(src).To(ContainSubstring("state.Filter = filterModel"))
		Expect(src).To(ContainSubstring("if metadata, ok := filter.GetMetadataOk(); ok && metadata != nil {"))
		Expect(src).To(ContainSubstring("filterModel.Metadata = metadataModel"))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(retentionFilterOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})
})

// pluralObjectOperation is a synthetic plural list whose item attributes carry a
// bare object (spec) of scalars — exercising the single-object item path that walks
// an element struct inside buildPluralView.
func pluralObjectOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/gizmos",
		Method:          "GET",
		OperationId:     "ListGizmos",
		Tag:             "Gizmos",
		ResponseRefName: "GizmosResponse",
		ItemRefName:     "Gizmo",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind:  model.ArtifactKindDataSource,
			ArtifactName:  "gizmos",
			Cardinality:   model.CardinalityPlural,
			TfDescription: "Use this data source to retrieve information about existing gizmos.",
			IdStrategy:    model.IdStrategyDataID,
			Group:         &model.OperationGroup{Read: "ListGizmos"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": {
				Kind:        model.SchemaKindArray,
				Description: "List of gizmos",
				Items: obj(map[string]*model.Schema{
					"id":   prim("string", "The gizmo's identifier."),
					"type": prim("string", "Gizmo resource type."),
					"attributes": obj(map[string]*model.Schema{
						"name": prim("string", "The name of the gizmo."),
						"spec": obj(map[string]*model.Schema{
							"shape": prim("string", "The shape of the gizmo."),
							"size":  prim("integer", "The number of segments."),
						}),
					}),
				}),
			},
		}),
	}
}

var _ = Describe("BuildDataSourceView plural nested objects", func() {
	It("renders a bare object in an item as a SingleNestedBlock with a generated struct", func() {
		view, err := BuildDataSourceView(mustArtifact(pluralObjectOperation()))
		Expect(err).NotTo(HaveOccurred())

		items := view.Schema.Blocks[0]
		Expect(items.TFName).To(Equal("gizmos"))
		var spec AttrView
		for _, b := range items.Blocks {
			if b.TFName == "spec" {
				spec = b
			}
		}
		Expect(spec.TFName).To(Equal("spec"))
		Expect(spec.ListBlock).To(BeFalse())

		var names []string
		for _, m := range view.Models {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"datadogGizmosDataSourceModel", "GizmoModel", "SpecModel"}))
	})

	It("maps the object off item.Attributes into the item accumulator", func() {
		view, err := BuildDataSourceView(mustArtifact(pluralObjectOperation()))
		Expect(err).NotTo(HaveOccurred())

		Expect(view.State.ItemLists).To(HaveLen(1))
		spec := view.State.ItemLists[0]
		Expect(spec.Kind).To(Equal("object_single"))
		Expect(spec.LHS).To(Equal("r.Spec"))
		Expect(spec.GetterOk).To(Equal("item.Attributes.GetSpecOk()"))
		Expect(spec.Var).To(Equal("spec"))
		Expect(spec.ElemStruct).To(Equal("SpecModel"))
		Expect(spec.Scalars).To(ContainElement(StateAssignment{
			Var: "shape", GetterOk: "spec.GetShapeOk()",
			LHS: "specModel.Shape", RHS: "types.StringValue(*shape)",
		}))
	})

	It("produces a deeply-equal view across two runs", func() {
		first, err := BuildDataSourceView(mustArtifact(pluralObjectOperation()))
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildDataSourceView(mustArtifact(pluralObjectOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(Equal(second))
	})
})
