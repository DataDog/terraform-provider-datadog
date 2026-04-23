package fwprovider

import (
	"context"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &OrgGroupResource{}
	_ resource.ResourceWithImportState = &OrgGroupResource{}
)

type OrgGroupResource struct {
	API  *datadogV2.OrgGroupsApi
	Auth context.Context
}

type OrgGroupModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	OwnerOrgSite types.String `tfsdk:"owner_org_site"`
	OwnerOrgUuid types.String `tfsdk:"owner_org_uuid"`
}

func NewOrgGroupResource() resource.Resource {
	return &OrgGroupResource{}
}

func (r *OrgGroupResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetOrgGroupsApiV2()
	r.Auth = providerData.Auth
}

func (r *OrgGroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "org_group"
}

func (r *OrgGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Org Group resource. This can be used to create and manage Datadog organization groups.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the org group.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"owner_org_site": schema.StringAttribute{
				Computed:    true,
				Description: "The site of the organization that owns this org group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_org_uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The UUID of the organization that owns this org group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *OrgGroupResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *OrgGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state OrgGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes := datadogV2.NewOrgGroupCreateAttributes(state.Name.ValueString())
	data := datadogV2.NewOrgGroupCreateData(*attributes, datadogV2.ORGGROUPTYPE_ORG_GROUPS)
	body := datadogV2.NewOrgGroupCreateRequest(*data)

	resp, _, err := r.API.CreateOrgGroup(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating org group"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group: response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state OrgGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org group ID must be a valid UUID"))
		return
	}

	resp, httpResp, err := r.API.GetOrgGroup(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving org group"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group: response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state OrgGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org group ID must be a valid UUID"))
		return
	}

	attributes := datadogV2.NewOrgGroupUpdateAttributes(state.Name.ValueString())
	data := datadogV2.NewOrgGroupUpdateData(*attributes, id, datadogV2.ORGGROUPTYPE_ORG_GROUPS)
	body := datadogV2.NewOrgGroupUpdateRequest(*data)

	resp, _, err := r.API.UpdateOrgGroup(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating org group"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group: response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state OrgGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org group ID must be a valid UUID"))
		return
	}

	httpResp, err := r.API.DeleteOrgGroup(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting org group"))
	}
}

func (r *OrgGroupResource) updateState(state *OrgGroupModel, resp *datadogV2.OrgGroupResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId().String())

	attributes := data.GetAttributes()
	state.Name = types.StringValue(attributes.GetName())
	state.OwnerOrgSite = types.StringValue(attributes.GetOwnerOrgSite())
	state.OwnerOrgUuid = types.StringValue(attributes.GetOwnerOrgUuid().String())
}
