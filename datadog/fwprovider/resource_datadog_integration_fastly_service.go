package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &integrationFastlyServiceResource{}
	_ resource.ResourceWithImportState = &integrationFastlyServiceResource{}
)

type integrationFastlyServiceResource struct {
	Api  *datadogV2.FastlyIntegrationApi
	Auth context.Context
}

type integrationFastlyServiceModel struct {
	ID        types.String `tfsdk:"id"`
	AccountId types.String `tfsdk:"account_id"`
	ServiceId types.String `tfsdk:"service_id"`
	Tags      types.Set    `tfsdk:"tags"`
}

func NewIntegrationFastlyServiceResource() resource.Resource {
	return &integrationFastlyServiceResource{}
}

func (r *integrationFastlyServiceResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetFastlyIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationFastlyServiceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_fastly_service"
}

func (r *integrationFastlyServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog IntegrationFastlyService resource. This can be used to create and manage Datadog integration_fastly_service.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Optional:    true,
				Description: "Fastly Account id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Description: "A list of tags for the Fastly service.",
				ElementType: types.StringType,
				Validators:  []validator.Set{validators.TagsSetIsNormalized()},
			},
			"service_id": schema.StringAttribute{
				Description: "The ID of the Fastly service.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationFastlyServiceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	accountID, serviceID, err := utils.AccountIDAndServiceIDFromID(request.ID)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), "")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("account_id"), accountID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
}

func (r *integrationFastlyServiceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationFastlyServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountID, serviceID, err := utils.AccountIDAndServiceIDFromID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(err.Error(), "")
		return
	}

	resp, httpResp, err := r.Api.GetFastlyService(r.Auth, accountID, serviceID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving API Key"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationFastlyServiceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationFastlyServiceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountId := state.AccountId.ValueString()

	body, diags := r.buildIntegrationFastlyServiceRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateFastlyService(r.Auth, accountId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationFastlyService"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationFastlyServiceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationFastlyServiceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountID, serviceID, err := utils.AccountIDAndServiceIDFromID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(err.Error(), "")
		return
	}

	body, diags := r.buildIntegrationFastlyServiceRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateFastlyService(r.Auth, accountID, serviceID, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationFastlyService"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationFastlyServiceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationFastlyServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountID, serviceID, err := utils.AccountIDAndServiceIDFromID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(err.Error(), "")
		return
	}

	httpResp, err := r.Api.DeleteFastlyService(r.Auth, accountID, serviceID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_fastly_service"))
		return
	}
}

func (r *integrationFastlyServiceResource) updateState(ctx context.Context, state *integrationFastlyServiceModel, resp *datadogV2.FastlyServiceResponse) {
	state.ID = types.StringValue(fmt.Sprintf("%s:%s", state.AccountId.ValueString(), resp.Data.Id))

	data := resp.GetData()
	attributes := data.GetAttributes()

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		state.Tags, _ = types.SetValueFrom(ctx, types.StringType, *tags)
	}
}

func (r *integrationFastlyServiceResource) buildIntegrationFastlyServiceRequestBody(ctx context.Context, state *integrationFastlyServiceModel) (*datadogV2.FastlyServiceRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewFastlyServiceAttributesWithDefaults()

	if !state.Tags.IsNull() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	req := datadogV2.NewFastlyServiceRequestWithDefaults()
	req.Data = *datadogV2.NewFastlyServiceDataWithDefaults()
	req.Data.SetId(state.ServiceId.ValueString())
	req.Data.SetAttributes(*attributes)

	return req, diags
}
