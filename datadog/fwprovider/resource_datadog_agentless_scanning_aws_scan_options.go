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
	_ resource.ResourceWithConfigure = &agentlessScanningAwsScanOptionsResource{}
)

// boolPtr converts a bool to *bool
func boolPtr(b bool) *bool {
	return &b
}

type agentlessScanningAwsScanOptionsResource struct {
	Api  *datadogV2.AgentlessScanningApi
	Auth context.Context
}

type agentlessScanningAwsScanOptionsResourceModel struct {
	ID               types.String `tfsdk:"id"`
	AwsAccountId     types.String `tfsdk:"aws_account_id"`
	Lambda           types.Bool   `tfsdk:"lambda"`
	SensitiveData    types.Bool   `tfsdk:"sensitive_data"`
	VulnContainersOs types.Bool   `tfsdk:"vuln_containers_os"`
	VulnHostOs       types.Bool   `tfsdk:"vuln_host_os"`
}

func NewAgentlessScanningAwsScanOptionsResource() resource.Resource {
	return &agentlessScanningAwsScanOptionsResource{}
}

func (r *agentlessScanningAwsScanOptionsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAgentlessScanningApiV2()
	r.Auth = providerData.Auth
}

func (r *agentlessScanningAwsScanOptionsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "agentless_scanning_aws_scan_options"
}

func (r *agentlessScanningAwsScanOptionsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Agentless Scanning AWS scan options resource. This can be used to activate and configure Agentless scan options for an AWS account.",
		Attributes: map[string]schema.Attribute{
			// Resource ID
			"id": utils.ResourceIDAttribute(),
			"aws_account_id": schema.StringAttribute{
				Description: "The AWS account ID for which agentless scanning is configured.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9]{12}$`),
						"must be a valid AWS account ID",
					),
				},
			},
			"lambda": schema.BoolAttribute{
				Description: "Indicates if scanning of Lambda functions is enabled.",
				Required:    true,
			},
			"sensitive_data": schema.BoolAttribute{
				Description: "Indicates if scanning for sensitive data is enabled.",
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

func (r *agentlessScanningAwsScanOptionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state agentlessScanningAwsScanOptionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.AwsScanOptionsCreateRequest{
		Data: datadogV2.AwsScanOptionsCreateData{
			Id:   state.AwsAccountId.ValueString(),
			Type: datadogV2.AWSSCANOPTIONSTYPE_AWS_SCAN_OPTIONS,
			Attributes: datadogV2.AwsScanOptionsCreateAttributes{
				Lambda:           state.Lambda.ValueBool(),
				SensitiveData:    state.SensitiveData.ValueBool(),
				VulnContainersOs: state.VulnContainersOs.ValueBool(),
				VulnHostOs:       state.VulnHostOs.ValueBool(),
			},
		},
	}

	awsScanOptionsResponse, _, err := r.Api.CreateAwsScanOptions(r.Auth, body)
	if err != nil {
		response.Diagnostics.AddError("Error creating AWS scan options", err.Error())
		return
	}

	r.updateStateFromResponse(&state, awsScanOptionsResponse)
	// Set the Terraform resource ID to the AWS account ID
	state.ID = types.StringValue(state.AwsAccountId.ValueString())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *agentlessScanningAwsScanOptionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state agentlessScanningAwsScanOptionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountID := state.AwsAccountId.ValueString()

	// List all AWS scan options and find the one matching our account ID
	awsScanOptionsListResponse, _, err := r.Api.ListAwsScanOptions(r.Auth)
	if err != nil {
		response.Diagnostics.AddError("Error reading AWS scan options", err.Error())
		return
	}

	var foundScanOptions *datadogV2.AwsScanOptionsData
	for _, scanOption := range awsScanOptionsListResponse.GetData() {
		if scanOption.GetId() == accountID {
			foundScanOptions = &scanOption
			break
		}
	}

	if foundScanOptions == nil {
		// Resource doesn't exist, remove from state
		response.State.RemoveResource(ctx)
		return
	}

	r.updateStateFromScanOptionsData(&state, *foundScanOptions)
	// Set the Terraform resource ID to the AWS account ID
	state.ID = types.StringValue(state.AwsAccountId.ValueString())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *agentlessScanningAwsScanOptionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state agentlessScanningAwsScanOptionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountID := state.AwsAccountId.ValueString()

	body := datadogV2.AwsScanOptionsUpdateRequest{
		Data: datadogV2.AwsScanOptionsUpdateData{
			Id:   state.AwsAccountId.ValueString(),
			Type: datadogV2.AWSSCANOPTIONSTYPE_AWS_SCAN_OPTIONS,
			Attributes: datadogV2.AwsScanOptionsUpdateAttributes{
				Lambda:           boolPtr(state.Lambda.ValueBool()),
				SensitiveData:    boolPtr(state.SensitiveData.ValueBool()),
				VulnContainersOs: boolPtr(state.VulnContainersOs.ValueBool()),
				VulnHostOs:       boolPtr(state.VulnHostOs.ValueBool()),
			},
		},
	}

	res, err := r.Api.UpdateAwsScanOptions(r.Auth, accountID, body)
	if err != nil {
		errorMsg := "Error updating AWS scan options"
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

func (r *agentlessScanningAwsScanOptionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state agentlessScanningAwsScanOptionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountID := state.AwsAccountId.ValueString()

	_, err := r.Api.DeleteAwsScanOptions(r.Auth, accountID)
	if err != nil {
		response.Diagnostics.AddError("Error deleting AWS scan options", err.Error())
		return
	}
}

func (r *agentlessScanningAwsScanOptionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Import the AWS account ID as both the Terraform resource ID and the aws_account_id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
	// Also set the aws_account_id to the same value
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("aws_account_id"), request.ID)...)
}

func (r *agentlessScanningAwsScanOptionsResource) updateStateFromResponse(state *agentlessScanningAwsScanOptionsResourceModel, resp datadogV2.AwsScanOptionsResponse) {
	data := resp.GetData()
	r.updateStateFromScanOptionsData(state, data)
}

func (r *agentlessScanningAwsScanOptionsResource) updateStateFromScanOptionsData(state *agentlessScanningAwsScanOptionsResourceModel, data datadogV2.AwsScanOptionsData) {
	state.AwsAccountId = types.StringValue(data.GetId())

	attributes := data.GetAttributes()
	state.Lambda = types.BoolValue(attributes.GetLambda())
	state.SensitiveData = types.BoolValue(attributes.GetSensitiveData())
	state.VulnContainersOs = types.BoolValue(attributes.GetVulnContainersOs())
	state.VulnHostOs = types.BoolValue(attributes.GetVulnHostOs())
}
