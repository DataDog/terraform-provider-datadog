package model

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("buildResourceLifecycle", func() {
	It("resolves all four CRUD calls, with a request type only where a body is sent", func() {
		art, err := BuildArtifact(incidentTypeResourceOp())
		Expect(err).NotTo(HaveOccurred())

		Expect(art.Kind).To(Equal(ArtifactKindResource))
		Expect(art.SourceFile).To(Equal("datadog/fwprovider/resource_datadog_incident_type.go"))
		Expect(art.Cardinality).To(BeEmpty(), "cardinality describes a data source, not a resource")

		Expect(art.Lifecycle.Create).To(Equal(&SDKCall{
			GoPackage: "datadogV2", GoApiStruct: "IncidentsApi", GoMethod: "CreateIncidentType",
			GoRequestType: "IncidentTypeCreateRequest", GoResponseType: "IncidentTypeResponse",
		}))
		Expect(art.Lifecycle.Read).To(Equal(&SDKCall{
			GoPackage: "datadogV2", GoApiStruct: "IncidentsApi", GoMethod: "GetIncidentType",
			GoResponseType: "IncidentTypeResponse",
		}))
		Expect(art.Lifecycle.Update).To(Equal(&SDKCall{
			GoPackage: "datadogV2", GoApiStruct: "IncidentsApi", GoMethod: "UpdateIncidentType",
			GoRequestType: "IncidentTypeUpdateRequest", GoResponseType: "IncidentTypeResponse",
		}))
		// Deep equality pins GoRequestType to "" here: a delete sends no body.
		Expect(art.Lifecycle.Delete).To(Equal(&SDKCall{
			GoPackage: "datadogV2", GoApiStruct: "IncidentsApi", GoMethod: "DeleteIncidentType",
		}))
		Expect(art.Lifecycle.IdStrategy).To(Equal(IdStrategyDataID))
		Expect(art.Lifecycle.UpdateUnsupported).To(BeFalse())
		Expect(art.Diagnostics).To(BeEmpty())
	})

	It("leaves Schema nil until the request/response merge lands", func() {
		art, err := BuildArtifact(incidentTypeResourceOp())
		Expect(err).NotTo(HaveOccurred())
		Expect(art.Schema).To(BeNil())
	})

	DescribeTable("a role the resource cannot do without fails the artifact by name",
		func(clear func(*ResolvedGroup), wantRole GroupRole, wantMessage string) {
			op := incidentTypeResourceOp()
			clear(op.ResolvedGroup)

			art, err := BuildArtifact(op)
			Expect(art).To(BeNil())
			Expect(err).To(HaveOccurred())

			var missing *MissingRoleError
			Expect(errors.As(err, &missing)).To(BeTrue(), "want a *MissingRoleError, got %T", err)
			Expect(missing.Role).To(Equal(wantRole))
			Expect(missing.Artifact).To(Equal("incident_type"))
			Expect(missing.Kind).To(Equal(ArtifactKindResource))
			Expect(err.Error()).To(Equal(wantMessage))
		},
		Entry("create omitted", func(g *ResolvedGroup) { g.Create = nil }, GroupRoleCreate,
			`model: resource "incident_type": group.create is not declared; a resource cannot be generated without that operation`),
		Entry("read omitted", func(g *ResolvedGroup) { g.Read = nil }, GroupRoleRead,
			`model: resource "incident_type": group.read is not declared; a resource cannot be generated without that operation`),
		Entry("delete omitted", func(g *ResolvedGroup) { g.Delete = nil }, GroupRoleDelete,
			`model: resource "incident_type": group.delete is not declared; a resource cannot be generated without that operation`),
		Entry("create dangling", func(g *ResolvedGroup) {
			g.Create = nil
			g.Unresolved = []GroupReference{{Role: GroupRoleCreate, OperationId: "CreateIncidentTypeTypo"}}
		}, GroupRoleCreate,
			`model: resource "incident_type": group.create names operationId "CreateIncidentTypeTypo", `+
				`which no operation in the spec declares; a resource cannot be generated without that operation`),
		Entry("update dangling — a typo must not silently become forced replacement", func(g *ResolvedGroup) {
			g.Update = nil
			g.Unresolved = []GroupReference{{Role: GroupRoleUpdate, OperationId: "UpdateIncidentTypeTypo"}}
		}, GroupRoleUpdate,
			`model: resource "incident_type": group.update names operationId "UpdateIncidentTypeTypo", `+
				`which no operation in the spec declares; a resource cannot be generated without that operation`),
	)

	It("degrades to forced replacement when update is simply undeclared", func() {
		op := incidentTypeResourceOp()
		op.ResolvedGroup.Update = nil

		art, err := BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		Expect(art.Lifecycle.Update).To(BeNil())
		Expect(art.Lifecycle.UpdateUnsupported).To(BeTrue())
		Expect(art.Diagnostics).To(ContainElement(Diagnostic{
			Severity: SeverityWarning,
			Message: `resource "incident_type": group.update is not declared, so no endpoint can modify ` +
				`this resource in place; every practitioner-settable attribute forces replacement, and any ` +
				`change to one destroys and recreates the resource`,
		}))
	})

	It("fails a groupless resource on create rather than panicking", func() {
		op := incidentTypeResourceOp()
		op.ResolvedGroup = nil

		_, err := BuildArtifact(op)
		Expect(err).To(MatchError(ContainSubstring("group.create is not declared")))
	})
})

// incidentTypeResourceOp is the annotated Create operation of a full-CRUD
// incident-type resource, with its group already resolved as the parser would
// leave it.
func incidentTypeResourceOp() *Operation {
	create := &Operation{
		Path: "/api/v2/incidents/config/types", Method: "POST",
		OperationId: "CreateIncidentType", Tag: "Incidents",
		RequestRefName: "IncidentTypeCreateRequest", ResponseRefName: "IncidentTypeResponse",
	}
	read := &Operation{
		Path: "/api/v2/incidents/config/types/{incident_type_id}", Method: "GET",
		OperationId: "GetIncidentType", Tag: "Incidents",
		ResponseRefName: "IncidentTypeResponse",
	}
	update := &Operation{
		Path: "/api/v2/incidents/config/types/{incident_type_id}", Method: "PATCH",
		OperationId: "UpdateIncidentType", Tag: "Incidents",
		RequestRefName: "IncidentTypeUpdateRequest", ResponseRefName: "IncidentTypeResponse",
	}
	del := &Operation{
		Path: "/api/v2/incidents/config/types/{incident_type_id}", Method: "DELETE",
		OperationId: "DeleteIncidentType", Tag: "Incidents",
	}
	create.Tracking = &TrackingFieldMetadata{
		ArtifactKind:  ArtifactKindResource,
		ArtifactName:  "incident_type",
		TfDescription: "Provides a Datadog incident type resource.",
		IdStrategy:    IdStrategyDataID,
		Group: &OperationGroup{
			Create: "CreateIncidentType", Read: "GetIncidentType",
			Update: "UpdateIncidentType", Delete: "DeleteIncidentType",
		},
	}
	create.ResolvedGroup = &ResolvedGroup{Create: create, Read: read, Update: update, Delete: del}
	return create
}
