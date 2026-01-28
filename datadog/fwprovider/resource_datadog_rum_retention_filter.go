package fwprovider

import (
	"context"
	"errors"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const RumRetentionFilterImportIdDelimiter = ":"

var (
	_ resource.ResourceWithConfigure   = &rumRetentionFilterResource{}
	_ resource.ResourceWithImportState = &rumRetentionFilterResource{}
)

type rumRetentionFilterResource struct {
	Api  *datadogV2.RumRetentionFiltersApi
	Auth context.Context
}

type rumRetentionFilterModel struct {
	ID            types.String  `tfsdk:"id"` // retention_filter_id
	ApplicationID types.String  `tfsdk:"application_id"`
	Name          types.String  `tfsdk:"name"`
	EventType     types.String  `tfsdk:"event_type"`
	SampleRate    types.Float64 `tfsdk:"sample_rate"`
	Query         types.String  `tfsdk:"query"`   // Optional
	Enabled       types.Bool    `tfsdk:"enabled"` // Optional
}

func NewRumRetentionFilterResource() resource.Resource {
	return &rumRetentionFilterResource{}
}

func (r *rumRetentionFilterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRumRetentionFiltersApiV2()
	r.Auth = providerData.Auth
}

func (r *rumRetentionFilterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_retention_filter"
}

func (r *rumRetentionFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RumRetentionFilter resource. This can be used to create and manage Datadog rum_retention_filter.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Description: "RUM application ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of a RUM retention filter.",
				Required:    true,
			},
			"event_type": schema.StringAttribute{
				Description: "The type of RUM events to filter on.",
				Required:    true,
			},
			"sample_rate": schema.Float64Attribute{
				Description: "The sample rate for a RUM retention filter, between 0.1 and 100. Supports one decimal place (for example, 50.5).",
				Required:    true,
			},
			"query": schema.StringAttribute{
				Description: "The Query string for a RUM retention filter.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the retention filter is to be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *rumRetentionFilterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	appId, retentionFilterId, err := ParseRumRetentionFilterImportId(request.ID)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), "")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), retentionFilterId)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("application_id"), appId)...)
}

func (r *rumRetentionFilterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state rumRetentionFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetRetentionFilter(r.Auth, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumRetentionFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumRetentionFilterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state rumRetentionFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	body, diags := r.buildRumRetentionFilterCreateRequestBody(&state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	appId := state.ApplicationID.ValueString()

	resp, _, err := r.Api.CreateRetentionFilter(r.Auth, appId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumRetentionFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumRetentionFilterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state rumRetentionFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildRumRetentionFilterUpdateRequestBody(&state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateRetentionFilter(r.Auth, state.ApplicationID.ValueString(), state.ID.ValueString(), *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumRetentionFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumRetentionFilterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state rumRetentionFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteRetentionFilter(r.Auth, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting rum_metric"))
		return
	}
}

func (r *rumRetentionFilterResource) updateState(state *rumRetentionFilterModel, resp *datadogV2.RumRetentionFilterResponse) {
	data := resp.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())
	state.EventType = types.StringValue(string(attributes.GetEventType()))
	state.Name = types.StringValue(attributes.GetName())
	state.SampleRate = types.Float64Value(attributes.GetSampleRate())
	state.Query = types.StringValue(attributes.GetQuery())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
}

func (r *rumRetentionFilterResource) buildRumRetentionFilterCreateRequestBody(state *rumRetentionFilterModel) (*datadogV2.RumRetentionFilterCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := datadogV2.NewRumRetentionFilterCreateAttributesWithDefaults()
	attributes.SetName(state.Name.ValueString())
	attributes.SetEventType(datadogV2.RumRetentionFilterEventType(state.EventType.ValueString()))
	attributes.SetSampleRate(state.SampleRate.ValueFloat64())

	if !state.Query.IsNull() {
		attributes.SetQuery(state.Query.ValueString())
	}

	if !state.Enabled.IsNull() {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}

	req := datadogV2.NewRumRetentionFilterCreateRequestWithDefaults()
	req.Data = *datadogV2.NewRumRetentionFilterCreateDataWithDefaults()

	req.Data.SetType(datadogV2.RUMRETENTIONFILTERTYPE_RETENTION_FILTERS)
	req.Data.SetAttributes(*attributes)

	req.Data.AdditionalProperties = make(map[string]any)
	req.Data.AdditionalProperties["meta"] = r.composeMeta()

	return req, diags
}

func (r *rumRetentionFilterResource) buildRumRetentionFilterUpdateRequestBody(state *rumRetentionFilterModel) (*datadogV2.RumRetentionFilterUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRumRetentionFilterUpdateAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	attributes.SetEventType(datadogV2.RumRetentionFilterEventType(state.EventType.ValueString()))
	attributes.SetSampleRate(state.SampleRate.ValueFloat64())

	if !state.Query.IsNull() {
		attributes.SetQuery(state.Query.ValueString())
	}

	if !state.Enabled.IsNull() {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}

	req := datadogV2.NewRumRetentionFilterUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewRumRetentionFilterUpdateDataWithDefaults()

	req.Data.SetId(state.ID.ValueString())
	req.Data.SetType(datadogV2.RUMRETENTIONFILTERTYPE_RETENTION_FILTERS)
	req.Data.SetAttributes(*attributes)

	req.Data.AdditionalProperties = make(map[string]any)
	req.Data.AdditionalProperties["meta"] = r.composeMeta()

	return req, diags
}

func (r *rumRetentionFilterResource) composeMeta() map[string]any {
	meta := make(map[string]any)
	meta["source"] = "terraform"
	return meta
}

func ParseRumRetentionFilterImportId(id string) (appId string, retentionFilterId string, err error) {
	result := strings.SplitN(id, RumRetentionFilterImportIdDelimiter, 2)
	if len(result) != 2 {
		return "", "", errors.New("error parsing id into application_id and retention_filter_id")
	}
	return result[0], result[1], nil
}
