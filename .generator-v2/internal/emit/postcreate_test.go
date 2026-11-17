package emit

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func addUpdateOnlyEnabled(op *model.Operation) {
	operationAttributes(op.ResolvedGroup.Update.RequestSchema).Properties["enabled"] = prim("boolean", "Whether this resource is enabled.")
	operationAttributes(op.ResolvedGroup.Read.ResponseSchema).Properties["enabled"] = prim("boolean", "Whether this resource is enabled.")
}

var _ = Describe("post-create initialization", func() {
	It("uses Update only for known configured update-only settings that differ from the Create response", func() {
		op := incidentTypeResourceOperation(true)
		addUpdateOnlyEnabled(op)
		artifact, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildResourceView(artifact)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Initialize).NotTo(BeNil())
		Expect(view.Initialize.Fields).To(Equal([]CreateInitializationField{{
			GoField:          "Enabled",
			Condition:        "!configured.Enabled.IsNull() && !configured.Enabled.IsUnknown() && !configured.Enabled.Equal(state.Enabled)",
			ResponseNullExpr: "types.BoolNull()",
		}}))

		source := string(mustRenderResource(view))
		create := source[strings.Index(source, "Resource) Create("):strings.Index(source, "Resource) Read(")]
		Expect(create).To(ContainSubstring("if initialize {"))
		Expect(create).To(ContainSubstring("state.Enabled = configured.Enabled"))
		Expect(create).To(ContainSubstring("state.Enabled = types.BoolNull()"))
		Expect(strings.Index(create, "state.Enabled = types.BoolNull()")).To(BeNumerically("<", strings.Index(create, "r.updateState(ctx, &state, &resp)")), "only an actual Create response value may suppress initialization")
		Expect(strings.Count(create, "r.Api.UpdateIncidentType(")).To(Equal(1))
		Expect(strings.Count(create, "SetInternalNote(")).To(Equal(1), "create-only setters must not enter the initialization PATCH")
		Expect(strings.Index(create, "response.State.Set(ctx, &state)")).To(BeNumerically("<", strings.Index(create, "r.Api.UpdateIncidentType(")), "retain the created API ID even if initialization fails")
		Expect(create).To(ContainSubstring("bodyData.SetId(state.ID.ValueString())"))
		Expect(create).To(ContainSubstring("error initializing incident_type"))
	})

	It("does not alter resources without an update-only field", func() {
		for _, withUpdate := range []bool{false, true} {
			artifact, err := model.BuildArtifact(incidentTypeResourceOperation(withUpdate))
			Expect(err).NotTo(HaveOccurred())
			view, err := BuildResourceView(artifact)
			Expect(err).NotTo(HaveOccurred())
			Expect(view.Initialize).To(BeNil())
			Expect(string(mustRenderResource(view))).NotTo(ContainSubstring("var configured "))
		}
	})

	It("applies configured request-only fields even though the initial state retains their planned value", func() {
		op := incidentTypeResourceOperation(true)
		operationAttributes(op.ResolvedGroup.Update.RequestSchema).Properties["reason"] = prim("string", "Reason for updating.")
		artifact, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildResourceView(artifact)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Initialize.Fields).To(Equal([]CreateInitializationField{{
			GoField: "Reason", Condition: "!configured.Reason.IsNull() && !configured.Reason.IsUnknown()",
		}}))
	})

	It("resolves existing mutable secret inputs through Create's configuration during initialization", func() {
		op := rootWriteOnlyOperation(true, true)
		addUpdateOnlyEnabled(op)
		_, view := buildResourceForWriteOnly(op)
		Expect(view.Initialize.Call.WriteOnlySecrets).To(HaveLen(1))
		Expect(view.Initialize.Call.WriteOnlySecrets[0].GetterCall).To(Equal("GetSecretForCreate(ctx, &request.Config)"))
		source := string(mustRenderResource(view))
		create := source[strings.Index(source, "Resource) Create("):strings.Index(source, "Resource) Read(")]
		Expect(create).NotTo(ContainSubstring("GetSecretForUpdate("))
	})

	It("rejects update-only descendants of a shared parent rather than silently omitting configuration on create", func() {
		op := widgetResourceOperation()
		settings := operationAttributes(op.ResolvedGroup.Update.RequestSchema).Properties["settings"]
		settings.Properties["enabled"] = prim("boolean", "Whether settings are enabled.")
		artifact, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		_, err = BuildResourceView(artifact)
		Expect(err).To(MatchError(ContainSubstring("unsupported post-create initialization")))
	})

	It("ignores fields that are read-only in Update", func() {
		op := incidentTypeResourceOperation(true)
		addUpdateOnlyEnabled(op)
		operationAttributes(op.ResolvedGroup.Update.RequestSchema).Properties["enabled"].ReadOnly = true
		artifact, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildResourceView(artifact)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Initialize).To(BeNil())
	})
})
