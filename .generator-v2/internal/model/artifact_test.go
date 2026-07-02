package model

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildArtifact", func() {
	DescribeTable("tagToClassName capitalizes each word and preserves in-word casing",
		func(tag, want string) { Expect(tagToClassName(tag)).To(Equal(want)) },
		Entry("a single word passes through", "Incidents", "Incidents"),
		Entry("a space-separated tag joins into PascalCase", "org groups", "OrgGroups"),
		Entry("an acronym is preserved, not lower-cased", "APM", "APM"),
		Entry("mixed punctuation becomes word breaks", "cloud-cost.management", "CloudCostManagement"),
	)

	DescribeTable("versionSegment returns the path segment after /api/",
		func(path, want string) { Expect(versionSegment(path)).To(Equal(want)) },
		Entry("a v2 path", "/api/v2/incidents/config/types/{id}", "v2"),
		Entry("a v1 path", "/api/v1/dashboard/{id}", "v1"),
		Entry("a path with no api segment yields empty", "/incidents/types/{id}", ""),
	)

	It("resolves the read SDK call and id strategy from the operation", func() {
		art, err := BuildArtifact(incidentTypeOp())
		Expect(err).NotTo(HaveOccurred())

		Expect(art.Name).To(Equal("incident_type"))
		Expect(art.Kind).To(Equal(ArtifactKindDataSource))
		Expect(art.Description).To(Equal("Use this data source to retrieve information about an existing incident type."))
		Expect(art.SourceFile).To(Equal("datadog/fwprovider/data_source_datadog_incident_type.go"))
		Expect(art.Lifecycle).NotTo(BeNil())
		Expect(art.Lifecycle.Read).To(Equal(&SDKCall{
			GoPackage:      "datadogV2",
			GoApiStruct:    "IncidentsApi",
			GoMethod:       "GetIncidentType",
			GoResponseType: "IncidentTypeResponse",
		}))
		Expect(art.Lifecycle.IdStrategy).To(Equal(IdStrategyDataID))
	})

	It("produces a deeply-equal artifact across two runs", func() {
		first, err := BuildArtifact(incidentTypeOp())
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildArtifact(incidentTypeOp())
		Expect(err).NotTo(HaveOccurred())
		Expect(reflect.DeepEqual(first, second)).To(BeTrue())
	})

	It("requires a tracked operation", func() {
		_, err := BuildArtifact(&Operation{})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("BuildArtifact plural", func() {
	It("marks the artifact plural and resolves the list-call bindings", func() {
		art, err := BuildArtifact(listThingsOp())
		Expect(err).NotTo(HaveOccurred())
		Expect(art.Cardinality).To(Equal(CardinalityPlural))
		Expect(art.Lifecycle.Read.GoMethod).To(Equal("ListThings"))
		Expect(art.Lifecycle.Read.ItemType).To(Equal("Thing"))
		Expect(art.Lifecycle.Read.OptionalParamsType).To(Equal("ListThingsOptionalParameters"))
		Expect(art.Lifecycle.Read.Paginated).To(BeTrue())
	})

	It("builds the schema as scalar filter leaves followed by the results block, order-preserving", func() {
		art, err := BuildArtifact(listThingsOp())
		Expect(err).NotTo(HaveOccurred())

		var names, types []string
		for _, a := range art.Schema.Attributes {
			names = append(names, a.Path)
			types = append(types, a.TfType)
		}
		Expect(names).To(Equal([]string{"filter_keyword", "filter_me", "response.data"}))
		Expect(types).To(Equal([]string{"schema.StringAttribute", "schema.BoolAttribute", "schema.ListNestedBlock"}))

		// Filter leaves are Optional inputs, not Computed.
		Expect(art.Schema.Attributes[0].Optional).To(BeTrue())
		Expect(art.Schema.Attributes[0].Computed).To(BeFalse())
	})

	It("excludes pagination params and drops array/enum params with an info diagnostic", func() {
		art, err := BuildArtifact(listThingsOp())
		Expect(err).NotTo(HaveOccurred())

		// page[number]/page[size] excluded silently; include (array) + sort (enum) dropped + logged.
		var msgs []string
		for _, d := range art.Diagnostics {
			Expect(d.Severity).To(Equal(SeverityInfo))
			msgs = append(msgs, d.Message)
		}
		Expect(msgs).To(HaveLen(2))
		Expect(msgs).To(ContainElement(ContainSubstring(`"include"`)))
		Expect(msgs).To(ContainElement(ContainSubstring(`"sort"`)))
	})

	It("produces a deeply-equal plural artifact across two runs", func() {
		first, err := BuildArtifact(listThingsOp())
		Expect(err).NotTo(HaveOccurred())
		second, err := BuildArtifact(listThingsOp())
		Expect(err).NotTo(HaveOccurred())
		Expect(reflect.DeepEqual(first, second)).To(BeTrue())
	})

	It("leaves OptionalParamsType empty and Paginated false for a no-param, non-paginated list", func() {
		op := listThingsOp()
		op.QueryParams = nil
		op.Pagination = nil
		art, err := BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		Expect(art.Lifecycle.Read.OptionalParamsType).To(BeEmpty())
		Expect(art.Lifecycle.Read.Paginated).To(BeFalse())
	})
})

var _ = Describe("BuildArtifact singular search", func() {
	Context("search only (group.search, no read)", func() {
		It("resolves the list call and leaves the by-id Read unset", func() {
			art, err := BuildArtifact(searchPowerpackOp())
			Expect(err).NotTo(HaveOccurred())

			Expect(art.Cardinality).To(Equal(CardinalitySingular))
			Expect(art.Lifecycle.Read).To(BeNil())
			Expect(art.Lifecycle.Search).To(Equal(&SDKCall{
				GoPackage:          "datadogV2",
				GoApiStruct:        "PowerpacksApi",
				GoMethod:           "ListPowerpacks",
				GoResponseType:     "PowerpacksResponse",
				ItemType:           "Powerpack",
				OptionalParamsType: "ListPowerpacksOptionalParameters",
				Paginated:          true,
			}))
		})

		It("derives multiple Optional filters from the query params, then one flat record (no list block)", func() {
			art, err := BuildArtifact(searchPowerpackOp())
			Expect(err).NotTo(HaveOccurred())

			var names, types []string
			for _, a := range art.Schema.Attributes {
				names = append(names, a.Path)
				types = append(types, a.TfType)
			}
			// Two scalar filters (sorted by name), then the singular record envelope.
			Expect(names).To(Equal([]string{"filter_name", "filter_query", "response.data"}))
			Expect(types).To(Equal([]string{"schema.StringAttribute", "schema.StringAttribute", "schema.SingleNestedBlock"}))

			// Singular output: the record is a SingleNestedBlock, never a list block.
			for _, a := range art.Schema.Attributes {
				Expect(a.TfType).NotTo(Equal("schema.ListNestedBlock"))
			}
			Expect(art.Schema.Attributes[0].Optional).To(BeTrue())
			Expect(art.Schema.Attributes[0].Computed).To(BeFalse())
		})

		It("drops pagination silently and the enum param with an info diagnostic", func() {
			art, err := BuildArtifact(searchPowerpackOp())
			Expect(err).NotTo(HaveOccurred())
			var msgs []string
			for _, d := range art.Diagnostics {
				Expect(d.Severity).To(Equal(SeverityInfo))
				msgs = append(msgs, d.Message)
			}
			Expect(msgs).To(HaveLen(1))
			Expect(msgs[0]).To(ContainSubstring(`"sort"`))
		})

		It("fails when the response declares no result-array element", func() {
			op := searchPowerpackOp()
			op.ResponseSchema = objSchema(map[string]*Schema{"data": objSchema(nil)}) // object, not array
			_, err := BuildArtifact(op)
			Expect(err).To(HaveOccurred())
		})

		It("produces a deeply-equal artifact across two runs", func() {
			first, err := BuildArtifact(searchPowerpackOp())
			Expect(err).NotTo(HaveOccurred())
			second, err := BuildArtifact(searchPowerpackOp())
			Expect(err).NotTo(HaveOccurred())
			Expect(reflect.DeepEqual(first, second)).To(BeTrue())
		})
	})

	Context("both (group.read + group.search)", func() {
		It("binds the by-id Read and the list Search, with the record from the by-id response", func() {
			art, err := BuildArtifact(bothDatastoreOp())
			Expect(err).NotTo(HaveOccurred())

			Expect(art.Cardinality).To(Equal(CardinalitySingular))
			Expect(art.Lifecycle.Read.GoMethod).To(Equal("GetDatastore"))
			Expect(art.Lifecycle.Search.GoMethod).To(Equal("ListDatastores"))
			Expect(art.Lifecycle.Search.ItemType).To(Equal("Datastore"))

			var names []string
			for _, a := range art.Schema.Attributes {
				names = append(names, a.Path)
			}
			// One Optional filter from the list op, then the singular record envelope.
			Expect(names).To(Equal([]string{"filter_keyword", "response.data"}))
			Expect(art.Schema.Attributes[len(names)-1].TfType).To(Equal("schema.SingleNestedBlock"))
		})

		It("fails when group.search names an operation that does not exist", func() {
			op := bothDatastoreOp()
			op.SearchOp = nil // simulate an unknown operationId
			_, err := BuildArtifact(op)
			Expect(err).To(HaveOccurred())
		})

		It("produces a deeply-equal artifact across two runs", func() {
			first, err := BuildArtifact(bothDatastoreOp())
			Expect(err).NotTo(HaveOccurred())
			second, err := BuildArtifact(bothDatastoreOp())
			Expect(err).NotTo(HaveOccurred())
			Expect(reflect.DeepEqual(first, second)).To(BeTrue())
		})

		It("degrades a diverging both to by-id-only and records why", func() {
			art, err := BuildArtifact(bothApiKeyOp())
			Expect(err).NotTo(HaveOccurred())

			// The by-id record (FullAPIKey) diverges from the list element
			// (PartialAPIKey), so search is dropped and id becomes required.
			Expect(art.Lifecycle.Read.GoMethod).To(Equal("GetAPIKey"))
			Expect(art.Lifecycle.Search).To(BeNil())

			var msgs []string
			for _, d := range art.Diagnostics {
				if d.Severity == SeverityInfo {
					msgs = append(msgs, d.Message)
				}
			}
			Expect(msgs).To(ContainElement(ContainSubstring("search lookup dropped")))
		})

		It("degrades to by-id-only when the by-id record shape cannot be confirmed (inline data schema)", func() {
			op := bothDatastoreOp()
			op.ResponseDataRefName = "" // inline by-id data property, no $ref to compare

			art, err := BuildArtifact(op)
			Expect(err).NotTo(HaveOccurred())

			// Shapes cannot be positively confirmed equal, so search is dropped
			// rather than risk a state mapper that reads the wrong fields.
			Expect(art.Lifecycle.Read.GoMethod).To(Equal("GetDatastore"))
			Expect(art.Lifecycle.Search).To(BeNil())

			var msgs []string
			for _, d := range art.Diagnostics {
				if d.Severity == SeverityInfo {
					msgs = append(msgs, d.Message)
				}
			}
			Expect(msgs).To(ContainElement(ContainSubstring("search lookup dropped")))
		})
	})
})

// searchPowerpackOp is a search-only singular data source: the list GET is the
// tracked op (group.search names it), a JSON:API collection with two scalar
// filters, paginated params (excluded), and one enum param (dropped + logged).
func searchPowerpackOp() *Operation {
	return &Operation{
		Path:            "/api/v2/powerpacks",
		Method:          "GET",
		OperationId:     "ListPowerpacks",
		Tag:             "Powerpacks",
		ResponseRefName: "PowerpacksResponse",
		ItemRefName:     "Powerpack",
		Pagination:      &Pagination{LimitParam: "page[size]", PageParam: "page[number]", ResultsPath: "data"},
		Tracking: &TrackingFieldMetadata{
			ArtifactKind: ArtifactKindDataSource,
			ArtifactName: "powerpack",
			IdStrategy:   IdStrategyDataID,
			Group:        &OperationGroup{Search: "ListPowerpacks"},
		},
		QueryParams: []QueryParam{
			{Name: "filter[name]", Schema: primSchema("string"), Description: "Filter by name."},
			{Name: "filter[query]", Schema: primSchema("string"), Description: "Free-text query."},
			{Name: "page[number]", Schema: &Schema{Kind: SchemaKindPrimitive, Type: "integer", Format: "int64"}},
			{Name: "page[size]", Schema: &Schema{Kind: SchemaKindPrimitive, Type: "integer", Format: "int64"}},
			{Name: "sort", Schema: &Schema{Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"name"}}},
		},
		ResponseSchema: objSchema(map[string]*Schema{
			"data": arrSchema(objSchema(map[string]*Schema{
				"id":         primSchema("string"),
				"type":       primSchema("string"),
				"attributes": objSchema(map[string]*Schema{"name": primSchema("string")}),
			})),
		}),
	}
}

// bothDatastoreOp is an id-optional singular data source: the tracked op is the
// by-id GET (group.read), and SearchOp points at the list GET (group.search)
// carrying one scalar filter.
func bothDatastoreOp() *Operation {
	listOp := &Operation{
		Path:            "/api/v2/datastores",
		Method:          "GET",
		OperationId:     "ListDatastores",
		Tag:             "Datastores",
		ResponseRefName: "DatastoreArray",
		ItemRefName:     "Datastore",
		QueryParams: []QueryParam{
			{Name: "filter[keyword]", Schema: primSchema("string"), Description: "Search query."},
		},
		ResponseSchema: objSchema(map[string]*Schema{
			"data": arrSchema(objSchema(map[string]*Schema{
				"id":         primSchema("string"),
				"type":       primSchema("string"),
				"attributes": objSchema(map[string]*Schema{"name": primSchema("string")}),
			})),
		}),
	}
	return &Operation{
		Path:                "/api/v2/datastores/{datastore_id}",
		Method:              "GET",
		OperationId:         "GetDatastore",
		Tag:                 "Datastores",
		ResponseRefName:     "Datastore",
		ResponseDataRefName: "Datastore", // same as the list element → stays "both"
		SearchOp:            listOp,
		Tracking: &TrackingFieldMetadata{
			ArtifactKind: ArtifactKindDataSource,
			ArtifactName: "datastore",
			IdStrategy:   IdStrategyDataID,
			Group:        &OperationGroup{Read: "GetDatastore", Search: "ListDatastores"},
		},
		ResponseSchema: objSchema(map[string]*Schema{
			"data": objSchema(map[string]*Schema{
				"id":         primSchema("string"),
				"type":       primSchema("string"),
				"attributes": objSchema(map[string]*Schema{"name": primSchema("string")}),
			}),
		}),
	}
}

// bothApiKeyOp is a diverging id-optional singular: the by-id GET returns a
// FullAPIKey record while the list GET yields PartialAPIKey elements, so the
// generator degrades it to by-id-only.
func bothApiKeyOp() *Operation {
	listOp := &Operation{
		Path:            "/api/v2/api_keys",
		Method:          "GET",
		OperationId:     "ListAPIKeys",
		Tag:             "KeyManagement",
		ResponseRefName: "APIKeysResponse",
		ItemRefName:     "PartialAPIKey",
		ResponseSchema: objSchema(map[string]*Schema{
			"data": arrSchema(objSchema(map[string]*Schema{
				"id":         primSchema("string"),
				"type":       primSchema("string"),
				"attributes": objSchema(map[string]*Schema{"name": primSchema("string")}),
			})),
		}),
	}
	return &Operation{
		Path:                "/api/v2/api_keys/{api_key_id}",
		Method:              "GET",
		OperationId:         "GetAPIKey",
		Tag:                 "KeyManagement",
		ResponseRefName:     "APIKeyResponse",
		ResponseDataRefName: "FullAPIKey",
		SearchOp:            listOp,
		Tracking: &TrackingFieldMetadata{
			ArtifactKind: ArtifactKindDataSource,
			ArtifactName: "api_key",
			IdStrategy:   IdStrategyDataID,
			Group:        &OperationGroup{Read: "GetAPIKey", Search: "ListAPIKeys"},
		},
		ResponseSchema: objSchema(map[string]*Schema{
			"data": objSchema(map[string]*Schema{
				"id":   primSchema("string"),
				"type": primSchema("string"),
				"attributes": objSchema(map[string]*Schema{
					"name": primSchema("string"),
					"key":  primSchema("string"),
				}),
			}),
		}),
	}
}

// listThingsOp is a plural list GET as a parser-shaped Operation: a JSON:API
// collection ({data:[{id,type,attributes:{name,count}}]}) with paginated query
// parameters, plus an array and an enum param the filter set must drop.
func listThingsOp() *Operation {
	return &Operation{
		Path:            "/api/v2/things",
		Method:          "GET",
		OperationId:     "ListThings",
		Tag:             "Things",
		ResponseRefName: "ThingsResponse",
		ItemRefName:     "Thing",
		Pagination:      &Pagination{LimitParam: "page[size]", PageParam: "page[number]", ResultsPath: "data"},
		Tracking: &TrackingFieldMetadata{
			ArtifactKind: ArtifactKindDataSource,
			ArtifactName: "things",
			Cardinality:  CardinalityPlural,
			IdStrategy:   IdStrategyDataID,
			Group:        &OperationGroup{Read: "ListThings"},
		},
		QueryParams: []QueryParam{
			{Name: "filter[keyword]", Schema: primSchema("string"), Description: "Search query."},
			{Name: "filter[me]", Required: true, Schema: primSchema("boolean"), Description: "Only mine."},
			{Name: "include", Schema: arrSchema(primSchema("string")), Description: "Related resources."},
			{Name: "page[number]", Schema: &Schema{Kind: SchemaKindPrimitive, Type: "integer", Format: "int64"}},
			{Name: "page[size]", Schema: &Schema{Kind: SchemaKindPrimitive, Type: "integer", Format: "int64"}},
			{Name: "sort", Schema: &Schema{Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"name"}}},
		},
		ResponseSchema: objSchema(map[string]*Schema{
			"data": arrSchema(objSchema(map[string]*Schema{
				"id":   primSchema("string"),
				"type": primSchema("string"),
				"attributes": objSchema(map[string]*Schema{
					"name":  primSchema("string"),
					"count": {Kind: SchemaKindPrimitive, Type: "integer", Format: "int64"},
				}),
			})),
		}),
	}
}

// incidentTypeOp is the incident_type GET as a parser-shaped Operation.
func incidentTypeOp() *Operation {
	return &Operation{
		Path:            "/api/v2/incidents/config/types/{incident_type_id}",
		Method:          "GET",
		OperationId:     "GetIncidentType",
		Tag:             "Incidents",
		ResponseRefName: "IncidentTypeResponse",
		Tracking: &TrackingFieldMetadata{
			ArtifactKind:  ArtifactKindDataSource,
			ArtifactName:  "incident_type",
			TfDescription: "Use this data source to retrieve information about an existing incident type.",
			IdStrategy:    IdStrategyDataID,
			Group:         &OperationGroup{Read: "GetIncidentType"},
		},
		ResponseSchema: objSchema(map[string]*Schema{
			"data": objSchema(map[string]*Schema{
				"id":   primSchema("string"),
				"type": primSchema("string"),
				"attributes": objSchema(map[string]*Schema{
					"name":        primSchema("string"),
					"description": primSchema("string"),
					"is_default":  primSchema("boolean"),
				}),
			}),
		}),
	}
}
