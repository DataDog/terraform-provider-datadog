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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure = &agentlessScanningAzureScanOptionsResource{}
)

type agentlessScanningAzureScanOptionsResource struct {
	Api  *datadogV2.AgentlessScanningApi
	Auth context.Context
}

type agentlessScanningAzureScanOptionsResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	AzureSubscriptionId types.String `tfsdk:"azure_subscription_id"`
	ComplianceHost      types.Bool   `tfsdk:"compliance_host"`
	Function            types.Bool   `tfsdk:"function"`
	VulnContainersOs    types.Bool   `tfsdk:"vuln_containers_os"`
	VulnHostOs          types.Bool   `tfsdk:"vuln_host_os"`
}

func NewAgentlessScanningAzureScanOptionsResource() resource.Resource {
	return &agentlessScanningAzureScanOptionsResource{}
}

func (r *agentlessScanningAzureScanOptionsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAgentlessScanningApiV2()
	r.Auth = providerData.Auth
}

func (r *agentlessScanningAzureScanOptionsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "agentless_scanning_azure_scan_options"
}

func (r *agentlessScanningAzureScanOptionsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Agentless Scanning Azure scan options resource. This can be used to activate and configure Agentless scan options for an Azure subscription.",
		Attributes: map[string]schema.Attribute{
			// Resource ID
			"id": utils.ResourceIDAttribute(),
			"azure_subscription_id": schema.StringAttribute{
				Description: "The Azure subscription ID for which agentless scanning is configured.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid Azure subscription ID (UUID format)",
					),
				},
			},
			"compliance_host": schema.BoolAttribute{
				Description: "Indicates whether host compliance scanning is enabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"function": schema.BoolAttribute{
				Description: "Indicates if scanning of Azure Functions is enabled.",
				Required:    true,
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

func (r *agentlessScanningAzureScanOptionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state agentlessScanningAzureScanOptionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.createOrUpdate(&state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *agentlessScanningAzureScanOptionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state agentlessScanningAzureScanOptionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	subscriptionID := state.AzureSubscriptionId.ValueString()

	azureScanOptionsResponse, httpResp, err := r.Api.GetAzureScanOptions(r.Auth, subscriptionID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Azure scan options", err.Error())
		return
	}

	r.updateStateFromScanOptionsData(&state, *azureScanOptionsResponse.Data)
	// Set the Terraform resource ID to the Azure subscription ID
	state.ID = types.StringValue(state.AzureSubscriptionId.ValueString())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *agentlessScanningAzureScanOptionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state agentlessScanningAzureScanOptionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.createOrUpdate(&state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *agentlessScanningAzureScanOptionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state agentlessScanningAzureScanOptionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	subscriptionID := state.AzureSubscriptionId.ValueString()

	_, err := r.Api.DeleteAzureScanOptions(r.Auth, subscriptionID)
	if err != nil {
		response.Diagnostics.AddError("Error deleting Azure scan options", err.Error())
		return
	}
}

func (r *agentlessScanningAzureScanOptionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Import the Azure subscription ID as both the Terraform resource ID and the azure_subscription_id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
	// Also set the azure_subscription_id to the same value
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("azure_subscription_id"), request.ID)...)
}

func (r *agentlessScanningAzureScanOptionsResource) updateStateFromResponse(state *agentlessScanningAzureScanOptionsResourceModel, resp datadogV2.AzureScanOptions) {
	data := resp.GetData()
	r.updateStateFromScanOptionsData(state, data)
}

func (r *agentlessScanningAzureScanOptionsResource) updateStateFromScanOptionsData(state *agentlessScanningAzureScanOptionsResourceModel, data datadogV2.AzureScanOptionsData) {
	state.AzureSubscriptionId = types.StringValue(data.GetId())

	attributes := data.GetAttributes()
	state.ComplianceHost = types.BoolValue(attributes.GetComplianceHost())
	state.Function = types.BoolValue(attributes.GetFunction())
	state.VulnContainersOs = types.BoolValue(attributes.GetVulnContainersOs())
	state.VulnHostOs = types.BoolValue(attributes.GetVulnHostOs())
}

// createOrUpdate attempts to read existing scan options and either creates or updates them accordingly
func (r *agentlessScanningAzureScanOptionsResource) createOrUpdate(state *agentlessScanningAzureScanOptionsResourceModel, diagnostics *diag.Diagnostics) {
	subscriptionID := state.AzureSubscriptionId.ValueString()

	// Try to read existing scan options
	_, httpResp, err := r.Api.GetAzureScanOptions(r.Auth, subscriptionID)

	// Check if scan options already exist
	scanOptionsExist := err == nil
	if err != nil && httpResp != nil && httpResp.StatusCode != 404 {
		// If error is not a 404, it's an unexpected error
		diagnostics.AddError("Error checking existing Azure scan options", err.Error())
		return
	}

	if scanOptionsExist {
		// Scan options exist, perform update
		updateBody := datadogV2.AzureScanOptionsInputUpdate{
			Data: &datadogV2.AzureScanOptionsInputUpdateData{
				Id:   subscriptionID,
				Type: datadogV2.AZURESCANOPTIONSINPUTUPDATEDATATYPE_AZURE_SCAN_OPTIONS,
				Attributes: &datadogV2.AzureScanOptionsInputUpdateDataAttributes{
					ComplianceHost:   state.ComplianceHost.ValueBoolPointer(),
					Function:         state.Function.ValueBoolPointer(),
					VulnContainersOs: state.VulnContainersOs.ValueBoolPointer(),
					VulnHostOs:       state.VulnHostOs.ValueBoolPointer(),
				},
			},
		}

		azureScanOptionsResponse, res, err := r.Api.UpdateAzureScanOptions(r.Auth, subscriptionID, updateBody)
		if err != nil {
			errorMsg := "Error updating Azure scan options"
			if res != nil {
				errorMsg += fmt.Sprintf(". API response: %s", res.Body)
			}
			diagnostics.AddError(errorMsg, err.Error())
			return
		}

		r.updateStateFromResponse(state, azureScanOptionsResponse)
	} else {
		// Scan options don't exist (404), perform create
		createBody := datadogV2.AzureScanOptions{
			Data: &datadogV2.AzureScanOptionsData{
				Id:   subscriptionID,
				Type: datadogV2.AZURESCANOPTIONSDATATYPE_AZURE_SCAN_OPTIONS,
				Attributes: &datadogV2.AzureScanOptionsDataAttributes{
					ComplianceHost:   state.ComplianceHost.ValueBoolPointer(),
					Function:         state.Function.ValueBoolPointer(),
					VulnContainersOs: state.VulnContainersOs.ValueBoolPointer(),
					VulnHostOs:       state.VulnHostOs.ValueBoolPointer(),
				},
			},
		}

		azureScanOptionsResponse, _, err := r.Api.CreateAzureScanOptions(r.Auth, createBody)
		if err != nil {
			diagnostics.AddError("Error creating Azure scan options", err.Error())
			return
		}

		r.updateStateFromResponse(state, azureScanOptionsResponse)
	}

	// Set the Terraform resource ID to the Azure subscription ID
	state.ID = types.StringValue(subscriptionID)
}
