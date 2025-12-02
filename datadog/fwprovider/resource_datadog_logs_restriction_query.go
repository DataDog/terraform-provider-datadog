package fwprovider

import (
	"context"
	"net/http"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &LogsRestrictionQueryResource{}
	_ resource.ResourceWithImportState = &LogsRestrictionQueryResource{}
)

type LogsRestrictionQueryResource struct {
	API  *datadogV2.LogsRestrictionQueriesApi
	Auth context.Context
}

type LogsRestrictionQueryModel struct {
	ID               types.String      `tfsdk:"id"`
	RestrictionQuery types.String      `tfsdk:"restriction_query"`
	RoleIds          types.Set         `tfsdk:"role_ids"`
	CreatedAt        timetypes.RFC3339 `tfsdk:"created_at"`
	ModifiedAt       timetypes.RFC3339 `tfsdk:"modified_at"`
}

func NewLogsRestrictionQueryResource() resource.Resource {
	return &LogsRestrictionQueryResource{}
}

func (r *LogsRestrictionQueryResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetLogsRestrictionQueriesApiV2()
	r.Auth = providerData.Auth
}

func (r *LogsRestrictionQueryResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "logs_restriction_query"
}

func (r *LogsRestrictionQueryResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Log Restriction Query resource. This can be used to create and manage Datadog Log Restriction Queries.",
		Attributes: map[string]schema.Attribute{
			"restriction_query": schema.StringAttribute{
				Description: "The query that defines the restriction. Only the content matching the query can be returned.",
				Required:    true,
			},
			"role_ids": schema.SetAttribute{
				Description: "An array of role IDs that have access to this restriction query.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation time of the restriction query (in ISO 8601).",
				CustomType:  timetypes.RFC3339Type{},
				Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Time of last restriction query modification (in ISO 8601).",
				CustomType:  timetypes.RFC3339Type{},
				Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
				Computed:    true,
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *LogsRestrictionQueryResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *LogsRestrictionQueryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data LogsRestrictionQueryModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	resp, httpResp, err := r.API.GetRestrictionQuery(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving logs restriction query"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &data, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *LogsRestrictionQueryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data LogsRestrictionQueryModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildCreateRequestBody(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.CreateRestrictionQuery(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating logs restriction query"))
		return
	}
	if err = utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Get the created restriction query ID
	respData := resp.GetData()
	data.ID = types.StringValue(respData.GetId())

	// Handle role assignments
	if !data.RoleIds.IsNull() && !data.RoleIds.IsUnknown() {
		var roleIds []string
		response.Diagnostics.Append(data.RoleIds.ElementsAs(ctx, &roleIds, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		restrictionQueryId := data.ID.ValueString()
		for _, roleId := range roleIds {
			relationshipToRole := datadogV2.NewRelationshipToRoleWithDefaults()
			roleData := datadogV2.NewRelationshipToRoleDataWithDefaults()
			roleData.SetId(roleId)
			relationshipToRole.SetData(*roleData)

			_, err := r.API.AddRoleToRestrictionQuery(r.Auth, restrictionQueryId, *relationshipToRole)
			if err != nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error adding role to logs restriction query"))
				return
			}
		}
	}

	// Read the full state with relationships
	fullResp, _, err := r.API.GetRestrictionQuery(r.Auth, data.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading logs restriction query after creation"))
		return
	}

	r.updateState(ctx, &data, &fullResp)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *LogsRestrictionQueryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan LogsRestrictionQueryModel
	var state LogsRestrictionQueryModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	restrictionQueryId := plan.ID.ValueString()

	// Update the restriction query itself
	body, diags := r.buildUpdateRequestBody(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.UpdateRestrictionQuery(r.Auth, restrictionQueryId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating logs restriction query"))
		return
	}

	if err = utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Handle role changes
	var oldRoleIds []string
	var newRoleIds []string

	if !state.RoleIds.IsNull() && !state.RoleIds.IsUnknown() {
		response.Diagnostics.Append(state.RoleIds.ElementsAs(ctx, &oldRoleIds, false)...)
	}
	if !plan.RoleIds.IsNull() && !plan.RoleIds.IsUnknown() {
		response.Diagnostics.Append(plan.RoleIds.ElementsAs(ctx, &newRoleIds, false)...)
	}
	if response.Diagnostics.HasError() {
		return
	}

	// Find roles to remove and add
	rolesToRemove := difference(oldRoleIds, newRoleIds)
	rolesToAdd := difference(newRoleIds, oldRoleIds)

	// Remove roles
	for _, roleId := range rolesToRemove {
		relationshipToRole := datadogV2.NewRelationshipToRoleWithDefaults()
		roleData := datadogV2.NewRelationshipToRoleDataWithDefaults()
		roleData.SetId(roleId)
		relationshipToRole.SetData(*roleData)

		_, err := r.API.RemoveRoleFromRestrictionQuery(r.Auth, restrictionQueryId, *relationshipToRole)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error removing role from logs restriction query"))
			return
		}
	}

	// Add roles
	for _, roleId := range rolesToAdd {
		relationshipToRole := datadogV2.NewRelationshipToRoleWithDefaults()
		roleData := datadogV2.NewRelationshipToRoleDataWithDefaults()
		roleData.SetId(roleId)
		relationshipToRole.SetData(*roleData)

		_, err := r.API.AddRoleToRestrictionQuery(r.Auth, restrictionQueryId, *relationshipToRole)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error adding role to logs restriction query"))
			return
		}
	}

	// Read the full state with relationships
	fullResp, _, err := r.API.GetRestrictionQuery(r.Auth, restrictionQueryId)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading logs restriction query after update"))
		return
	}

	r.updateState(ctx, &plan, &fullResp)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *LogsRestrictionQueryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data LogsRestrictionQueryModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	httpResp, err := r.API.DeleteRestrictionQuery(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting logs restriction query"))
	}
}

func (r *LogsRestrictionQueryResource) updateState(ctx context.Context, state *LogsRestrictionQueryModel, response *datadogV2.RestrictionQueryWithRelationshipsResponse) {
	data := response.GetData()
	state.ID = types.StringValue(data.GetId())
	attributes := data.GetAttributes()

	if restrictionQuery, ok := attributes.GetRestrictionQueryOk(); ok {
		state.RestrictionQuery = types.StringValue(*restrictionQuery)
	} else {
		state.RestrictionQuery = types.StringValue("")
	}

	// Extract role IDs from relationships
	if relationships, ok := data.GetRelationshipsOk(); ok {
		if roles, ok := relationships.GetRolesOk(); ok {
			if rolesData, ok := roles.GetDataOk(); ok {
				roleIds := make([]string, 0, len(*rolesData))
				for _, role := range *rolesData {
					roleIds = append(roleIds, role.GetId())
				}
				state.RoleIds, _ = types.SetValueFrom(ctx, types.StringType, roleIds)
			}
		}
	}
	if state.RoleIds.IsNull() {
		state.RoleIds, _ = types.SetValue(types.StringType, []attr.Value{})
	}

	if createdAt, ok := attributes.GetCreatedAtOk(); ok && !createdAt.IsZero() {
		state.CreatedAt = timetypes.NewRFC3339TimeValue(*createdAt)
	} else {
		state.CreatedAt = timetypes.NewRFC3339Null()
	}

	if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && !modifiedAt.IsZero() {
		state.ModifiedAt = timetypes.NewRFC3339TimeValue(*modifiedAt)
	} else {
		state.ModifiedAt = timetypes.NewRFC3339Null()
	}
}

func (r *LogsRestrictionQueryResource) buildCreateRequestBody(ctx context.Context, data *LogsRestrictionQueryModel) (*datadogV2.RestrictionQueryCreatePayload, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := datadogV2.NewRestrictionQueryCreateAttributes(data.RestrictionQuery.ValueString())

	createData := datadogV2.NewRestrictionQueryCreateDataWithDefaults()
	createData.SetAttributes(*attributes)

	payload := datadogV2.NewRestrictionQueryCreatePayloadWithDefaults()
	payload.SetData(*createData)

	return payload, diags
}

func (r *LogsRestrictionQueryResource) buildUpdateRequestBody(ctx context.Context, data *LogsRestrictionQueryModel) (*datadogV2.RestrictionQueryUpdatePayload, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := datadogV2.NewRestrictionQueryUpdateAttributesWithDefaults()
	attributes.SetRestrictionQuery(data.RestrictionQuery.ValueString())

	updateData := datadogV2.NewRestrictionQueryUpdateDataWithDefaults()
	updateData.SetAttributes(*attributes)

	payload := datadogV2.NewRestrictionQueryUpdatePayloadWithDefaults()
	payload.SetData(*updateData)

	return payload, diags
}

// Helper function to find the difference between two slices
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
