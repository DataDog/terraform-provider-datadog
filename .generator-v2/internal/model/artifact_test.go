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
