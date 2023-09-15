package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &integrationAWSLogCollectionResource{}
	_ resource.ResourceWithImportState = &integrationAWSLogCollectionResource{}
)

type integrationAWSLogCollectionResource struct {
	Api  *datadogV1.AWSLogsIntegrationApi
	Auth context.Context
}

type integrationAWSLogCollectionModel struct {
	ID        types.String `tfsdk:"id"`
	AccountID types.String `tfsdk:"account_id"`
	Services  types.List   `tfsdk:"services"`
}

func NewIntegrationAWSLogCollectionResource() resource.Resource {
	return &integrationAWSLogCollectionResource{}
}

func (r *integrationAWSLogCollectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSLogsIntegrationApiV1()
	r.Auth = providerData.Auth
}

func (r *integrationAWSLogCollectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_aws_log_collection"
}

func (r *integrationAWSLogCollectionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog - Amazon Web Services integration log collection resource. This can be used to manage which AWS services logs are collected from for an account.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "Your AWS Account ID without dashes. If your account is a GovCloud or China account, specify the `access_key_id` here.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"services": schema.ListAttribute{
				Required:    true,
				Description: "A list of services to collect logs from. See the [api docs](https://docs.datadoghq.com/api/v1/aws-logs-integration/#get-list-of-aws-log-ready-services) for more details on which services are supported.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationAWSLogCollectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationAWSLogCollectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationAWSLogCollectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	logCollection, diags := r.getAWSLogCollectionAccount(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if logCollection == nil {
		response.State.RemoveResource(ctx)
		return
	}

	r.updateState(ctx, &state, logCollection)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSLogCollectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationAWSLogCollectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	enableLogCollectionServices := r.buildDatadogIntegrationAWSLogCollectionStruct(ctx, &state)
	resp, httpresp, err := r.Api.EnableAWSLogServices(r.Auth, *enableLogCollectionServices)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error enabling log collection services for Amazon Web Services integration account"))
		return
	}
	res := resp.(map[string]interface{})
	if status, ok := res["status"]; ok && status == "error" {
		response.Diagnostics.AddError("error creating aws log collection", fmt.Sprintf("%s", httpresp.Body))
		return
	}

	logCollection, diags := r.getAWSLogCollectionAccount(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if logCollection == nil {
		response.Diagnostics.AddError("error retrieving log collection", "")
		return
	}

	r.updateState(ctx, &state, logCollection)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSLogCollectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationAWSLogCollectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	enableLogCollectionServices := r.buildDatadogIntegrationAWSLogCollectionStruct(ctx, &state)
	_, _, err := r.Api.EnableAWSLogServices(r.Auth, *enableLogCollectionServices)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating log collection services for Amazon Web Services integration account"))
		return
	}

	logCollection, diags := r.getAWSLogCollectionAccount(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if logCollection == nil {
		response.Diagnostics.AddError("error retrieving log collection", "")
		return
	}

	r.updateState(ctx, &state, logCollection)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSLogCollectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationAWSLogCollectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	services := []string{}
	state.Services.ElementsAs(ctx, &services, false)

	deleteLogCollectionServices := datadogV1.NewAWSLogsServicesRequest(state.AccountID.ValueString(), services)
	_, _, err := r.Api.EnableAWSLogServices(r.Auth, *deleteLogCollectionServices)
	if err != nil {
		response.Diagnostics.AddError("error disabling Amazon Web Services log collection", err.Error())
	}

}

func (r *integrationAWSLogCollectionResource) getAWSLogCollectionAccount(ctx context.Context, state *integrationAWSLogCollectionModel) (*datadogV1.AWSLogsListResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	logCollections, _, err := r.Api.ListAWSLogsIntegrations(r.Auth)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error getting log collection for aws integration."))
		return nil, diags
	}
	if err := utils.CheckForUnparsed(logCollections); err != nil {
		diags.AddError("response contains unparsedObject", err.Error())
		return nil, diags
	}

	var logCollection *datadogV1.AWSLogsListResponse
	for _, c := range logCollections {
		if c.GetAccountId() == state.AccountID.ValueString() {
			logCollection = &c
		}
	}

	return logCollection, diags
}

func (r *integrationAWSLogCollectionResource) updateState(ctx context.Context, state *integrationAWSLogCollectionModel, collection *datadogV1.AWSLogsListResponse) {
	state.ID = types.StringValue(collection.GetAccountId())
	state.AccountID = types.StringValue(collection.GetAccountId())
	state.Services, _ = types.ListValueFrom(ctx, types.StringType, collection.GetServices())
}

func (r *integrationAWSLogCollectionResource) buildDatadogIntegrationAWSLogCollectionStruct(ctx context.Context, state *integrationAWSLogCollectionModel) *datadogV1.AWSLogsServicesRequest {
	services := []string{}
	state.Services.ElementsAs(ctx, &services, false)

	enableLogCollectionServices := datadogV1.NewAWSLogsServicesRequest(state.ID.ValueString(), services)

	return enableLogCollectionServices
}
