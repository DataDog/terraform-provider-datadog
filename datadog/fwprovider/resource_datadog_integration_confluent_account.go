package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
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
	_ resource.ResourceWithConfigure   = &integrationConfluentAccountResource{}
	_ resource.ResourceWithImportState = &integrationConfluentAccountResource{}
)

type integrationConfluentAccountResource struct {
	Api  *datadogV2.ConfluentCloudApi
	Auth context.Context
}

type integrationConfluentAccountModel struct {
	ID        types.String `tfsdk:"id"`
	ApiKey    types.String `tfsdk:"api_key"`
	ApiSecret types.String `tfsdk:"api_secret"`
	Tags      types.Set    `tfsdk:"tags"`
}

func NewIntegrationConfluentAccountResource() resource.Resource {
	return &integrationConfluentAccountResource{}
}

func (r *integrationConfluentAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetConfluentCloudApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationConfluentAccountResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_confluent_account"
}

func (r *integrationConfluentAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog IntegrationConfluentAccount resource. This can be used to create and manage Datadog integration_confluent_account.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:    true,
				Description: "The API key associated with your Confluent account.",
			},
			"api_secret": schema.StringAttribute{
				Required:    true,
				Description: "The API secret associated with your Confluent account.",
				Sensitive:   true,
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Description: "A list of strings representing tags. Can be a single key, or key-value pairs separated by a colon.",
				ElementType: types.StringType,
				Validators:  []validator.Set{validators.TagsSetIsNormalized()},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationConfluentAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationConfluentAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationConfluentAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetConfluentAccount(r.Auth, id)
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

func (r *integrationConfluentAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationConfluentAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildIntegrationConfluentAccountRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateConfluentAccount(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationConfluentAccount"))
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

func (r *integrationConfluentAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationConfluentAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildIntegrationConfluentAccountUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateConfluentAccount(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationConfluentAccount"))
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

func (r *integrationConfluentAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationConfluentAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteConfluentAccount(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_confluent_account"))
		return
	}
}

func (r *integrationConfluentAccountResource) updateState(ctx context.Context, state *integrationConfluentAccountModel, resp *datadogV2.ConfluentAccountResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if apiKey, ok := attributes.GetApiKeyOk(); ok {
		state.ApiKey = types.StringValue(*apiKey)
	}

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		state.Tags, _ = types.SetValueFrom(ctx, types.StringType, *tags)
	}
}

func (r *integrationConfluentAccountResource) buildIntegrationConfluentAccountRequestBody(ctx context.Context, state *integrationConfluentAccountModel) (*datadogV2.ConfluentAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewConfluentAccountCreateRequestAttributesWithDefaults()

	attributes.SetApiKey(state.ApiKey.ValueString())
	attributes.SetApiSecret(state.ApiSecret.ValueString())

	if !state.Tags.IsNull() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	req := datadogV2.NewConfluentAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewConfluentAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *integrationConfluentAccountResource) buildIntegrationConfluentAccountUpdateRequestBody(ctx context.Context, state *integrationConfluentAccountModel) (*datadogV2.ConfluentAccountUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewConfluentAccountUpdateRequestAttributesWithDefaults()

	attributes.SetApiKey(state.ApiKey.ValueString())
	attributes.SetApiSecret(state.ApiSecret.ValueString())

	if !state.Tags.IsNull() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	req := datadogV2.NewConfluentAccountUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewConfluentAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
