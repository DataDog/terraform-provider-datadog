package fwprovider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure = &agentlessScanningGcpScanOptionsResource{}
)

type agentlessScanningGcpScanOptionsResource struct {
	Api  *datadogV2.AgentlessScanningApi
	Auth context.Context
}

type agentlessScanningGcpScanOptionsResourceModel struct {
	ID               types.String `tfsdk:"id"`
	GcpProjectId     types.String `tfsdk:"gcp_project_id"`
	VulnContainersOs types.Bool   `tfsdk:"vuln_containers_os"`
	VulnHostOs       types.Bool   `tfsdk:"vuln_host_os"`
}

func NewAgentlessScanningGcpScanOptionsResource() resource.Resource {
	return &agentlessScanningGcpScanOptionsResource{}
}

func (r *agentlessScanningGcpScanOptionsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAgentlessScanningApiV2()
	r.Auth = providerData.Auth
}

func (r *agentlessScanningGcpScanOptionsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "agentless_scanning_gcp_scan_options"
}

func (r *agentlessScanningGcpScanOptionsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Agentless Scanning GCP scan options resource. This can be used to activate and configure Agentless scan options for a GCP project.",
		Attributes: map[string]schema.Attribute{
			// Resource ID
			"id": utils.ResourceIDAttribute(),
			"gcp_project_id": schema.StringAttribute{
				Description: "The GCP project ID for which agentless scanning is configured.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z]([a-z0-9-]{4,28}[a-z0-9])?$`),
						"must be a valid GCP project ID: 6–30 characters, start with a lowercase letter, and include only lowercase letters, digits, or hyphens.",
					),
				},
			},
			"vuln_containers_os": schema.BoolAttribute{
				Description: "Indicates if scanning for vulnerabilities in containers is enabled.",
				Required:    true,
			},
			"vuln_host_os": schema.BoolAttribute{
				Description: "Indicates if scanning for vulnerabilities in hosts is enabled.",
				Required:    true,
			},
		},
	}
}

func (r *agentlessScanningGcpScanOptionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state agentlessScanningGcpScanOptionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.GcpScanOptions{
		Data: &datadogV2.GcpScanOptionsData{
			Id:   state.GcpProjectId.ValueString(),
			Type: datadogV2.GCPSCANOPTIONSDATATYPE_GCP_SCAN_OPTIONS,
			Attributes: &datadogV2.GcpScanOptionsDataAttributes{
				VulnContainersOs: boolPtr(state.VulnContainersOs.ValueBool()),
				VulnHostOs:       boolPtr(state.VulnHostOs.ValueBool()),
			},
		},
	}

	gcpScanOptionsResponse, _, err := r.Api.CreateGcpScanOptions(r.Auth, body)
	if err != nil {
		response.Diagnostics.AddError("Error creating GCP scan options", err.Error())
		return
	}

	r.updateStateFromResponse(&state, gcpScanOptionsResponse)
	// Set the Terraform resource ID to the GCP project ID
	state.ID = types.StringValue(state.GcpProjectId.ValueString())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *agentlessScanningGcpScanOptionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state agentlessScanningGcpScanOptionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	projectID := state.GcpProjectId.ValueString()

	gcpScanOptionsResponse, _, err := r.Api.GetGcpScanOptions(r.Auth, projectID)
	if err != nil {
		response.Diagnostics.AddError("Error reading GCP scan options", err.Error())
		return
	}

	r.updateStateFromScanOptionsData(&state, *gcpScanOptionsResponse.Data)
	// Set the Terraform resource ID to the GCP project ID
	state.ID = types.StringValue(state.GcpProjectId.ValueString())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *agentlessScanningGcpScanOptionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state agentlessScanningGcpScanOptionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	projectID := state.GcpProjectId.ValueString()

	body := datadogV2.GcpScanOptionsInputUpdate{
		Data: &datadogV2.GcpScanOptionsInputUpdateData{
			Id:   state.GcpProjectId.ValueString(),
			Type: datadogV2.GCPSCANOPTIONSINPUTUPDATEDATATYPE_GCP_SCAN_OPTIONS,
			Attributes: &datadogV2.GcpScanOptionsInputUpdateDataAttributes{
				VulnContainersOs: boolPtr(state.VulnContainersOs.ValueBool()),
				VulnHostOs:       boolPtr(state.VulnHostOs.ValueBool()),
			},
		},
	}

	_, res, err := r.Api.UpdateGcpScanOptions(r.Auth, projectID, body)
	if err != nil {
		errorMsg := "Error updating GCP scan options"
		if res != nil {
			errorMsg += fmt.Sprintf(". API response: %s", res.Body)
		}
		response.Diagnostics.AddError(errorMsg, err.Error())
		return
	}

	// After update, we need to read the current state since the API doesn't return the updated object
	readReq := resource.ReadRequest{State: response.State}
	readResp := resource.ReadResponse{State: response.State, Diagnostics: diag.Diagnostics{}}

	// Set the state with current values before reading
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.Read(ctx, readReq, &readResp)
	response.Diagnostics.Append(readResp.Diagnostics...)
	response.State = readResp.State
}

func (r *agentlessScanningGcpScanOptionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state agentlessScanningGcpScanOptionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	projectID := state.GcpProjectId.ValueString()

	_, err := r.Api.DeleteGcpScanOptions(r.Auth, projectID)
	if err != nil {
		response.Diagnostics.AddError("Error deleting GCP scan options", err.Error())
		return
	}
}

func (r *agentlessScanningGcpScanOptionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Import the GCP project ID as both the Terraform resource ID and the gcp_project_id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
	// Also set the gcp_project_id to the same value
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("gcp_project_id"), request.ID)...)
}

func (r *agentlessScanningGcpScanOptionsResource) updateStateFromResponse(state *agentlessScanningGcpScanOptionsResourceModel, resp datadogV2.GcpScanOptions) {
	data := resp.GetData()
	r.updateStateFromScanOptionsData(state, data)
}

func (r *agentlessScanningGcpScanOptionsResource) updateStateFromScanOptionsData(state *agentlessScanningGcpScanOptionsResourceModel, data datadogV2.GcpScanOptionsData) {
	state.GcpProjectId = types.StringValue(data.GetId())

	attributes := data.GetAttributes()
	state.VulnContainersOs = types.BoolValue(attributes.GetVulnContainersOs())
	state.VulnHostOs = types.BoolValue(attributes.GetVulnHostOs())
}
