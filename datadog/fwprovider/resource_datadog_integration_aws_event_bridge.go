package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &integrationAwsEventBridgeResource{}
	_ resource.ResourceWithImportState = &integrationAwsEventBridgeResource{}
)

type integrationAwsEventBridgeResource struct {
	Api  *datadogV1.AWSIntegrationApi
	Auth context.Context
}

type integrationAwsEventBridgeModel struct {
	ID                 types.String `tfsdk:"id"`
	AccountId          types.String `tfsdk:"account_id"`
	CreateEventBus     types.Bool   `tfsdk:"create_event_bus"`
	EventGeneratorName types.String `tfsdk:"event_generator_name"`
	Region             types.String `tfsdk:"region"`
}

func NewIntegrationAwsEventBridgeResource() resource.Resource {
	return &integrationAwsEventBridgeResource{}
}

func (r *integrationAwsEventBridgeResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV1()
	r.Auth = providerData.Auth
}

func (r *integrationAwsEventBridgeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_aws_event_bridge"
}

func (r *integrationAwsEventBridgeResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog - Amazon Web Services integration EventBridge resource. This can be used to create and manage Event Sources for each Datadog integrated AWS account.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "Your AWS Account ID without dashes.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"create_event_bus": schema.BoolAttribute{
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "True if Datadog should create the event bus in addition to the event source. Requires the `events:CreateEventBus` permission.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"event_generator_name": schema.StringAttribute{
				Description: "The given part of the event source name, which is then combined with an assigned suffix to form the full name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description: "The event source's [AWS region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationAwsEventBridgeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationAwsEventBridgeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationAwsEventBridgeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	resp, httpResp, err := r.Api.ListAWSEventBridgeSources(r.Auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error listing AWS EventBridge Event Sources"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	found := false
	matchedEventHub := integrationAwsEventBridgeModel{}

	if accounts, ok := resp.GetAccountsOk(); ok && len(*accounts) > 0 {
		for _, account := range *accounts {
			if found {
				break
			}
			if eventhubs, ok := account.GetEventHubsOk(); ok && len(*eventhubs) > 0 {
				for _, eventhub := range *eventhubs {
					if *eventhub.Name == state.ID.ValueString() {
						matchedEventHub.ID = types.StringValue(*eventhub.Name)
						matchedEventHub.AccountId = types.StringValue(account.GetAccountId())
						matchedEventHub.Region = types.StringValue(eventhub.GetRegion())
						found = true
						break
					}
				}
			}
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
		response.Diagnostics.AddWarning("Read error", "No matching Event Source found.")
		return
	}

	r.updateStateAfterRead(ctx, &state, &matchedEventHub)
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAwsEventBridgeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationAwsEventBridgeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	utils.IntegrationAwsMutex.Lock()
	defer utils.IntegrationAwsMutex.Unlock()

	body, diags := r.buildIntegrationAwsEventBridgeRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateAWSEventBridgeSource(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error creating AWS EventBridge Event Source"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateStateAfterWrite(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAwsEventBridgeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported", "AWS EventBridge Event Sources cannot be updated")
}

func (r *integrationAwsEventBridgeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationAwsEventBridgeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	utils.IntegrationAwsMutex.Lock()
	defer utils.IntegrationAwsMutex.Unlock()

	req := datadogV1.NewAWSEventBridgeDeleteRequestWithDefaults()

	if !state.AccountId.IsNull() {
		req.SetAccountId(state.AccountId.ValueString())
	}
	if !state.Region.IsNull() {
		req.SetRegion(state.Region.ValueString())
	}
	if !state.ID.IsNull() {
		// EventGeneratorName in DeleteRequest is the constructed full name, stored as ID in state
		req.SetEventGeneratorName(state.ID.ValueString())
	}

	_, httpResp, err := r.Api.DeleteAWSEventBridgeSource(r.Auth, *req)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "AWS EventBridge Event Source not found"))
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error deleting AWS EventBridge Event Source"))
		return
	}
}

func (r *integrationAwsEventBridgeResource) updateStateAfterRead(ctx context.Context, state *integrationAwsEventBridgeModel, resp *integrationAwsEventBridgeModel) {
	state.ID = resp.ID
	state.AccountId = resp.AccountId
	state.Region = resp.Region
}

func (r *integrationAwsEventBridgeResource) updateStateAfterWrite(ctx context.Context, state *integrationAwsEventBridgeModel, resp *datadogV1.AWSEventBridgeCreateResponse) {

	if createEventBus, ok := resp.GetHasBusOk(); ok {
		state.CreateEventBus = types.BoolValue(*createEventBus)
	}

	if eventSourceName, ok := resp.GetEventSourceNameOk(); ok {
		// Use EventSourceName as ID
		state.ID = types.StringValue(*eventSourceName)
	}

	if region, ok := resp.GetRegionOk(); ok {
		state.Region = types.StringValue(*region)
	}
}

func (r *integrationAwsEventBridgeResource) buildIntegrationAwsEventBridgeRequestBody(ctx context.Context, state *integrationAwsEventBridgeModel) (*datadogV1.AWSEventBridgeCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := datadogV1.NewAWSEventBridgeCreateRequestWithDefaults()

	if !state.AccountId.IsNull() {
		req.SetAccountId(state.AccountId.ValueString())
	}
	if !state.CreateEventBus.IsNull() {
		req.SetCreateEventBus(state.CreateEventBus.ValueBool())
	}
	if !state.Region.IsNull() {
		req.SetRegion(state.Region.ValueString())
	}
	if !state.EventGeneratorName.IsNull() {
		req.SetEventGeneratorName(state.EventGeneratorName.ValueString())
	}

	return req, diags
}
