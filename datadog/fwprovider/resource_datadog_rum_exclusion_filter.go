package fwprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
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

const RumExclusionFilterImportIdDelimiter = ":"

var (
	_ resource.ResourceWithConfigure   = &rumExclusionFilterResource{}
	_ resource.ResourceWithImportState = &rumExclusionFilterResource{}
)

// This endpoint is not yet part of the public datadog-api-client-go SDK, so requests
// are issued via utils.SendRequest against the raw path instead of a generated API method.
const (
	rumExclusionFiltersPath  = "/api/v2/rum/applications/%s/retention_filters/exclusion"
	rumExclusionFilterIdPath = "/api/v2/rum/applications/%s/retention_filters/exclusion/%s"
)

type rumExclusionFilterResource struct {
	Api  *datadog.APIClient
	Auth context.Context
}

type rumExclusionFilterModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	EventType     types.String `tfsdk:"event_type"`
	Query         types.String `tfsdk:"query"`
	Enabled       types.Bool   `tfsdk:"enabled"`
}

type rumExclusionFilterAttributes struct {
	Name      string `json:"name,omitempty"`
	Enabled   *bool  `json:"enabled,omitempty"`
	EventType string `json:"event_type,omitempty"`
	Query     string `json:"query,omitempty"`
}

type rumExclusionFilterRequestData struct {
	ID         string                        `json:"id,omitempty"`
	Type       string                        `json:"type"`
	Attributes *rumExclusionFilterAttributes `json:"attributes"`
}

type rumExclusionFilterRequest struct {
	Data rumExclusionFilterRequestData `json:"data"`
}

type rumExclusionFilterResponseData struct {
	ID         string                       `json:"id"`
	Type       string                       `json:"type"`
	Attributes rumExclusionFilterAttributes `json:"attributes"`
}

type rumExclusionFilterResponse struct {
	Data rumExclusionFilterResponseData `json:"data"`
}

func NewRumExclusionFilterResource() resource.Resource {
	return &rumExclusionFilterResource{}
}

func (r *rumExclusionFilterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.HttpClient
	r.Auth = providerData.Auth
}

func (r *rumExclusionFilterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_exclusion_filter"
}

func (r *rumExclusionFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RumExclusionFilter resource. This can be used to create and manage Datadog rum_exclusion_filter.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Description: "RUM application ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the exclusion filter.",
				Required:    true,
			},
			"event_type": schema.StringAttribute{
				Description: "The type of RUM events to filter on.",
				Required:    true,
			},
			"query": schema.StringAttribute{
				Description: "Additional query used to further restrict which RUM events are excluded.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the exclusion filter is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *rumExclusionFilterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	appId, exclusionFilterId, err := ParseRumExclusionFilterImportId(request.ID)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), "")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), exclusionFilterId)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("application_id"), appId)...)
}

func (r *rumExclusionFilterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state rumExclusionFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf(rumExclusionFilterIdPath, state.ApplicationID.ValueString(), state.ID.ValueString())
	respBytes, httpResp, err := utils.SendRequest(r.Auth, r.Api, http.MethodGet, path, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumExclusionFilter"))
		return
	}

	var resp rumExclusionFilterResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		response.Diagnostics.AddError("error parsing RumExclusionFilter response", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumExclusionFilterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state rumExclusionFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildRumExclusionFilterCreateRequestBody(&state)
	path := fmt.Sprintf(rumExclusionFiltersPath, state.ApplicationID.ValueString())

	respBytes, _, err := utils.SendRequest(r.Auth, r.Api, http.MethodPost, path, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating RumExclusionFilter"))
		return
	}

	var resp rumExclusionFilterResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		response.Diagnostics.AddError("error parsing RumExclusionFilter response", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumExclusionFilterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state rumExclusionFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildRumExclusionFilterUpdateRequestBody(&state)
	path := fmt.Sprintf(rumExclusionFilterIdPath, state.ApplicationID.ValueString(), state.ID.ValueString())

	respBytes, _, err := utils.SendRequest(r.Auth, r.Api, http.MethodPatch, path, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating RumExclusionFilter"))
		return
	}

	var resp rumExclusionFilterResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		response.Diagnostics.AddError("error parsing RumExclusionFilter response", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumExclusionFilterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state rumExclusionFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf(rumExclusionFilterIdPath, state.ApplicationID.ValueString(), state.ID.ValueString())
	_, httpResp, err := utils.SendRequest(r.Auth, r.Api, http.MethodDelete, path, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting RumExclusionFilter"))
		return
	}
}

func (r *rumExclusionFilterResource) updateState(state *rumExclusionFilterModel, resp *rumExclusionFilterResponse) {
	data := resp.Data
	attributes := data.Attributes

	state.ID = types.StringValue(data.ID)
	state.Name = types.StringValue(attributes.Name)
	state.EventType = types.StringValue(attributes.EventType)
	state.Query = types.StringValue(attributes.Query)
	if attributes.Enabled != nil {
		state.Enabled = types.BoolValue(*attributes.Enabled)
	}
}

func (r *rumExclusionFilterResource) buildRumExclusionFilterCreateRequestBody(state *rumExclusionFilterModel) *rumExclusionFilterRequest {
	attributes := &rumExclusionFilterAttributes{
		Name:      state.Name.ValueString(),
		EventType: state.EventType.ValueString(),
	}

	if !state.Query.IsNull() {
		attributes.Query = state.Query.ValueString()
	}

	if !state.Enabled.IsNull() {
		enabled := state.Enabled.ValueBool()
		attributes.Enabled = &enabled
	}

	return &rumExclusionFilterRequest{
		Data: rumExclusionFilterRequestData{
			Type:       "exclusion_filters",
			Attributes: attributes,
		},
	}
}

func (r *rumExclusionFilterResource) buildRumExclusionFilterUpdateRequestBody(state *rumExclusionFilterModel) *rumExclusionFilterRequest {
	attributes := &rumExclusionFilterAttributes{
		Name:      state.Name.ValueString(),
		EventType: state.EventType.ValueString(),
	}

	if !state.Query.IsNull() {
		attributes.Query = state.Query.ValueString()
	}

	if !state.Enabled.IsNull() {
		enabled := state.Enabled.ValueBool()
		attributes.Enabled = &enabled
	}

	return &rumExclusionFilterRequest{
		Data: rumExclusionFilterRequestData{
			ID:         state.ID.ValueString(),
			Type:       "exclusion_filters",
			Attributes: attributes,
		},
	}
}

func ParseRumExclusionFilterImportId(id string) (appId string, exclusionFilterId string, err error) {
	result := strings.SplitN(id, RumExclusionFilterImportIdDelimiter, 2)
	if len(result) != 2 {
		return "", "", errors.New("error parsing id into application_id and exclusion_filter_id")
	}
	return result[0], result[1], nil
}
