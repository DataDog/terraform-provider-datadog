package emit

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("explicit resource ID parameter output", func() {
	DescribeTable("preserves scalar parsing and SDK argument types for explicit selection",
		func(goType, schemaType, parseExpression, argumentExpression string) {
			op := incidentTypeResourceOperation(true)
			op.Tracking.IdParameter = "incident_type_id"
			for _, role := range []*model.Operation{op.ResolvedGroup.Read, op.ResolvedGroup.Update, op.ResolvedGroup.Delete} {
				role.Path = "/api/v2/incident-types/{incident_type_id}/action"
				argument := &role.SDKBinding.Required[0]
				argument.GoType = goType
				argument.Schema = &model.Schema{Kind: model.SchemaKindPrimitive, Type: schemaType}
				if goType == "IncidentTypeType" {
					argument.Schema.Enum = []string{"incident_types"}
				}
			}
			artifact, err := model.BuildArtifact(op)
			Expect(err).NotTo(HaveOccurred())
			view, err := BuildResourceView(artifact)
			Expect(err).NotTo(HaveOccurred())
			source := string(mustRenderResource(view))
			Expect(strings.Count(source, parseExpression)).To(Equal(3), "each Read, Update, and Delete converts the computed string ID")
			Expect(source).To(ContainSubstring("r.Api.GetIncidentType(r.Auth, " + argumentExpression + ")"))
			Expect(source).To(ContainSubstring("r.Api.DeleteIncidentType(r.Auth, " + argumentExpression + ")"))
			Expect(view.Import.Parts).To(Equal([]string{"id"}))
			Expect(source).NotTo(ContainSubstring(`tfsdk:"incident_type_id"`))
		},
		Entry("int", "int", "integer", "strconv.ParseInt(state.ID.ValueString(), 10, 64)", "int(parsedId)"),
		Entry("int32", "int32", "integer", "strconv.ParseInt(state.ID.ValueString(), 10, 64)", "int32(parsedId)"),
		Entry("int64", "int64", "integer", "strconv.ParseInt(state.ID.ValueString(), 10, 64)", "parsedId"),
		Entry("float32", "float32", "number", "strconv.ParseFloat(state.ID.ValueString(), 64)", "float32(parsedId)"),
		Entry("float64", "float64", "number", "strconv.ParseFloat(state.ID.ValueString(), 64)", "parsedId"),
		Entry("boolean", "bool", "boolean", "strconv.ParseBool(state.ID.ValueString())", "parsedId"),
		Entry("named string enum", "IncidentTypeType", "string", "datadogV2.NewIncidentTypeTypeFromValue(state.ID.ValueString())", "*parsedId"),
	)

	DescribeTable("binds a selected ID before action suffixes without adding a practitioner input",
		func(goType, format string) {
			op := incidentTypeResourceOperation(true)
			op.Tracking.IdParameter = "feature_flag_id"
			for _, role := range []*model.Operation{op.ResolvedGroup.Read, op.ResolvedGroup.Update, op.ResolvedGroup.Delete} {
				role.Path = "/api/v2/feature-flags/{feature_flag_id}/terraform"
				role.SDKBinding.Required[0] = model.SDKArgument{
					Name: "feature_flag_id", GoName: "featureFlagId", GoType: goType, Location: "path",
					Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Format: format},
				}
			}
			op.ResolvedGroup.Delete.Path = "/api/v2/feature-flags/{feature_flag_id}/archive"
			op.ResolvedGroup.Delete.OperationId = "ArchiveFeatureFlag"
			op.ResolvedGroup.Delete.Method = "POST"
			// Archive returns a response, although Delete ignores its contents.
			op.ResolvedGroup.Delete.ResponseSchema = op.ResolvedGroup.Read.ResponseSchema
			op.ResolvedGroup.Delete.ResponseRefName = "IncidentTypeResponse"
			art, err := model.BuildArtifact(op)
			Expect(err).NotTo(HaveOccurred())
			view, err := BuildResourceView(art)
			Expect(err).NotTo(HaveOccurred())
			Expect(view.Import.Parts).To(Equal([]string{"id"}))
			for _, call := range []CRUDCallView{view.Read, view.Update, view.Delete} {
				Expect(call.Arguments).To(HaveLen(1))
				Expect(call.Arguments[0].TFName).To(Equal("id"))
			}
			source := string(mustRenderResource(view))
			Expect(source).NotTo(ContainSubstring(`tfsdk:"feature_flag_id"`))
			Expect(source).NotTo(ContainSubstring(`"feature_flag_id": schema.`))
			Expect(source).To(ContainSubstring(`resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)`))
			if goType == "uuid.UUID" {
				Expect(strings.Count(source, "uuid.Parse(state.ID.ValueString())")).To(Equal(3))
				Expect(source).To(ContainSubstring(`r.Api.ArchiveFeatureFlag(r.Auth, parsedId)`))
			} else {
				Expect(source).To(ContainSubstring(`r.Api.ArchiveFeatureFlag(r.Auth, state.ID.ValueString())`))
			}
		},
		Entry("string IDs", "string", ""),
		Entry("UUID IDs preserve conversion", "uuid.UUID", "uuid"),
	)

	It("keeps a later parent argument and composite import in SDK order", func() {
		op := incidentTypeResourceOperation(true)
		op.Tracking.IdParameter = "incident_type_id"
		parent := model.SDKArgument{
			Name: "project_id", GoName: "projectId", GoType: "string", Location: "path",
			Description: "The project ID.", Schema: prim("string", "The project ID."),
		}
		op.ResolvedGroup.Create.Path = "/api/v2/projects/{project_id}/incident-types"
		op.ResolvedGroup.Create.SDKBinding = &model.SDKOperationBinding{Required: []model.SDKArgument{parent}}
		for _, role := range []*model.Operation{op.ResolvedGroup.Read, op.ResolvedGroup.Update, op.ResolvedGroup.Delete} {
			role.Path = "/api/v2/incident-types/{incident_type_id}/projects/{project_id}"
			role.SDKBinding.Required = append(role.SDKBinding.Required, parent)
		}
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(schemaTree(view), "project_id").Required).To(BeTrue())
		Expect(view.Import.Parts).To(Equal([]string{"project_id", "id"}))
		for _, call := range []CRUDCallView{view.Read, view.Update, view.Delete} {
			Expect(call.Arguments).To(Equal([]SDKArgumentView{
				{Expression: "state.ID.ValueString()", TFName: "id", GoType: "string"},
				{Expression: "state.ProjectId.ValueString()", TFName: "project_id", GoType: "string"},
			}))
		}
		source := string(mustRenderResource(view))
		Expect(source).To(ContainSubstring(`r.Api.DeleteIncidentType(r.Auth, state.ID.ValueString(), state.ProjectId.ValueString())`))
		Expect(source).To(ContainSubstring(`path.Root("project_id"), parts[0]`))
		Expect(source).To(ContainSubstring(`path.Root("id"), parts[1]`))
		Expect(source).NotTo(ContainSubstring(`tfsdk:"incident_type_id"`))
	})
})
