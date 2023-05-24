package fwprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.Resource = &datadogIntegrationGCPSTSResource{}
)

// datadogIntegrationGCPSTSResource is the resource implementation.
type datadogIntegrationGCPSTSResource struct {
	Auth   context.Context
	GcpApi *datadogV2.GCPIntegrationApi
}

// NewDatadogIntegrationGCPSTSResource is a helper function to simplify the provider implementation.
func NewDatadogIntegrationGCPSTSResource() resource.Resource {
	return &datadogIntegrationGCPSTSResource{}
}

type datadogIntegrationGCPSTSResourceModel struct {
	ServiceAccountEmail types.String `tfsdk:"service_account_email"`
	ID                  types.String `tfsdk:"id"`
	DelegateEmail       types.String `tfsdk:"delegate_email"`
	Automute            types.Bool   `tfsdk:"automute"`
	EnableCspm          types.Bool   `tfsdk:"enable_cspm"`
	HostFilters         types.List   `tfsdk:"host_filters"`
}

// Metadata returns the resource name.
func (r *datadogIntegrationGCPSTSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "integration_gcp_sts"
}

func (r *datadogIntegrationGCPSTSResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError("Unexpected Resource Configure Type", "")
		return
	}

	r.Auth = providerData.Auth
	r.GcpApi = providerData.DatadogApiInstances.GetGCPIntegrationApiV2()
}

// Schema defines the Terraform Resource configuration.
func (r *datadogIntegrationGCPSTSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog - Google Cloud Platform STS integration resource. This can be used to create and manage Datadog Google Cloud Platform STS integrations",
		Attributes: map[string]schema.Attribute{
			"service_account_email": schema.StringAttribute{
				Description: "Your STS-enabled GCP service account.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delegate_email": schema.StringAttribute{
				Description: "Datadog's STS Delegate Email.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Description: "Your STS-enabled GCP service account's unique ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host_filters": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Your Datadog Host Filters.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"automute": schema.BoolAttribute{
				Description: "Enable Automute.",
				Optional:    true,
			},
			"enable_cspm": schema.BoolAttribute{
				Description: "Enable CSPM.",
				Optional:    true,
			},
		},
	}
}

// Create sets the initial Terraform state.
func (r *datadogIntegrationGCPSTSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan datadogIntegrationGCPSTSResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	delegateResponse, _, err := r.GcpApi.MakeGCPSTSDelegate(r.Auth, *datadogV2.NewMakeGCPSTSDelegateOptionalParameters())
	if err != nil {
		resp.Diagnostics.AddError("Error creating GCP Delegate within Datadog",
			"Could not create Delegate Service Account, unexpected error: "+err.Error())
		return
	}

	var hostFilters []string
	delegateInfoResponse := delegateResponse.GetData()
	delegateAttributes := delegateInfoResponse.GetAttributes()
	hostFilterPlanElements := plan.HostFilters.Elements()
	listOfHostFilters, err := attributeListToStringList(ctx, hostFilterPlanElements)
	if err != nil {
		resp.Diagnostics.AddError("Error converting attribute list to strings",
			"Error converting attribute list to strings: "+err.Error())
		return
	}
	if len(listOfHostFilters) == 0 {
		hostFilters = make([]string, 0)
	} else {
		hostFilters = listOfHostFilters
	}

	var enableAutomute bool
	if !plan.Automute.IsNull() {
		enableAutomute = plan.Automute.ValueBool()
	}

	var enableCSPM bool
	if !plan.EnableCspm.IsNull() {
		enableCSPM = plan.EnableCspm.ValueBool()
	}

	saInfo := datadogV2.GCPServiceAccountCreateRequestData{
		Data: &datadogV2.GCPServiceAccountData{
			Attributes: &datadogV2.GCPServiceAccountAttributes{
				ClientEmail:   stringToPointer(plan.ServiceAccountEmail.ValueString()),
				Automute:      boolToPointer(enableAutomute),
				IsCspmEnabled: boolToPointer(enableCSPM),
				HostFilters:   hostFilters,
			},
			Type: datadogV2.GCPSERVICEACCOUNTTYPE_GCP_SERVICE_ACCOUNT.Ptr(),
		},
	}

	createResponse, _, err := r.GcpApi.CreateGCPSTSAccount(r.Auth, saInfo)
	if err != nil {
		resp.Diagnostics.AddError("Error creating STS service account",
			"Error creating an entry within Datadog for your STS enabled service account: "+err.Error())
		return
	}
	createdServiceAccountInfo := createResponse.GetData()

	// Set the "computed" values.
	plan.ID = types.StringValue(createdServiceAccountInfo.GetId())
	plan.DelegateEmail = types.StringValue(delegateAttributes.GetDelegateAccountEmail())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read re-sets the Terraform state using the latest "pulled" data.
func (r *datadogIntegrationGCPSTSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datadogIntegrationGCPSTSResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	delegateResponse, _, err := r.GcpApi.GetGCPSTSDelegate(r.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving STS delegate"))
		return
	}

	delegateEmail, err := extractDelegateAccountEmail(delegateResponse)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting delegate account email"))
		return
	}
	state.DelegateEmail = delegateEmail

	stsEnabledAccounts, _, err := r.GcpApi.ListGCPSTSAccounts(r.Auth)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving STS service accounts",
			"Error listing GCP STS Accounts: "+err.Error())
		return
	}

	foundAccount, err := findServiceAccountByUniqueID(stsEnabledAccounts, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error finding your service account", err.Error())
		return
	}

	outputListValue, hostFilterCount := getHostFilters(foundAccount)

	// The section below handles optional fields in Terraform.
	// If an optional field is not specified, then Terraform state stores a nil value.
	// However, the Datadog GCP API always responds with a value for these optional fields (an empty list, a false boolean, etc).
	// If these optional fields aren't specified within Terraform Resources, then these fields should always remain nil.
	if state.HostFilters.IsNull() {
		if hostFilterCount > 0 {
			state.HostFilters = outputListValue
		}
	} else {
		state.HostFilters = outputListValue
	}

	accountAttributes := foundAccount.GetAttributes()
	if state.Automute.IsNull() {
		if accountAttributes.GetAutomute() {
			state.Automute = types.BoolValue(accountAttributes.GetAutomute())
		}
	} else {
		state.Automute = types.BoolValue(accountAttributes.GetAutomute())
	}

	if state.EnableCspm.IsNull() {
		if accountAttributes.GetIsCspmEnabled() {
			state.EnableCspm = types.BoolValue(accountAttributes.GetIsCspmEnabled())
		}
	} else {
		state.EnableCspm = types.BoolValue(accountAttributes.GetIsCspmEnabled())
	}

	state.ServiceAccountEmail = types.StringValue(accountAttributes.GetClientEmail())

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the Terraform state locally and on the Datadog "backend".
func (r *datadogIntegrationGCPSTSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan datadogIntegrationGCPSTSResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var currentState datadogIntegrationGCPSTSResourceModel
	currentDiagnostics := req.State.Get(ctx, &currentState)
	if currentDiagnostics.HasError() {
		return
	}

	var listOfHostFilters []string
	if plan.HostFilters.IsNull() {
		listOfHostFilters = make([]string, 0)
	} else {
		hostFilterPlanElements := plan.HostFilters.Elements()
		hostFilters, err := attributeListToStringList(ctx, hostFilterPlanElements)
		if err != nil {
			resp.Diagnostics.AddError("Error converting attribute list to strings",
				"Error converting attribute list to strings: "+err.Error())
			return
		}
		listOfHostFilters = hostFilters
	}

	var toEnableCSPM bool
	if !plan.EnableCspm.IsNull() {
		toEnableCSPM = plan.EnableCspm.ValueBool()
	}

	var toEnableAutomute bool
	if !plan.Automute.IsNull() {
		toEnableAutomute = plan.Automute.ValueBool()
	}

	updatedSAInfo := datadogV2.GCPServiceAccountUpdateRequest{
		Data: &datadogV2.GCPServiceAccountUpdateRequestData{
			Type: datadogV2.GCPSERVICEACCOUNTTYPE_GCP_SERVICE_ACCOUNT.Ptr(),
			Attributes: &datadogV2.GCPServiceAccountAttributes{
				IsCspmEnabled: boolToPointer(toEnableCSPM),
				Automute:      boolToPointer(toEnableAutomute),
				HostFilters:   listOfHostFilters,
			},
		},
	}

	uniqueAccountID := currentState.ID.ValueString()

	updateResponse, _, err := r.GcpApi.UpdateGCPSTSAccount(r.Auth, uniqueAccountID, updatedSAInfo)
	if err != nil {
		resp.Diagnostics.AddError("Error updating your service account",
			"Error: "+err.Error())
		return
	}

	dataBlock := updateResponse.GetData()
	blockAttributes := dataBlock.GetAttributes()

	plan.ID = basetypes.NewStringValue(dataBlock.GetId())
	plan.DelegateEmail = currentState.DelegateEmail
	plan.ServiceAccountEmail = basetypes.NewStringValue(blockAttributes.GetClientEmail())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete removes the resource from Terraform state.
func (r *datadogIntegrationGCPSTSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datadogIntegrationGCPSTSResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.GcpApi.DeleteGCPSTSAccount(r.Auth, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting your service account",
			"Error encountered when attempting to delete your service account from Datadog: "+err.Error())
		return
	}
}

func (r *datadogIntegrationGCPSTSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func stringToPointer(s string) *string {
	return &s
}

func boolToPointer(b bool) *bool {
	return &b
}

func getHostFilters(account *datadogV2.GCPSTSAccount) (basetypes.ListValue, int) {
	accountAttributes := account.GetAttributes()

	currentHostFilters := accountAttributes.GetHostFilters()

	var requiredAttributes []attr.Value
	for _, hostFilter := range currentHostFilters {
		requiredAttributes = append(requiredAttributes, types.StringValue(hostFilter))
	}

	outputListValue, _ := types.ListValue(types.StringType, requiredAttributes)

	return outputListValue, len(requiredAttributes)
}

func extractDelegateAccountEmail(delegateResponse datadogV2.GCPSTSDelegateResponse) (basetypes.StringValue, error) {
	delegateResponseData := delegateResponse.GetData()

	delegateAttributes := delegateResponseData.GetAttributes()

	delegateAccountEmail := delegateAttributes.GetDelegateAccountEmail()
	if delegateAccountEmail == "" {
		return basetypes.StringValue{}, errors.New("error, delegate account email is empty \"\"")
	}

	return types.StringValue(delegateAttributes.GetDelegateAccountEmail()), nil
}

func findServiceAccountByUniqueID(accounts datadogV2.GCPSTSEnabledAccountData, accountToFindID string) (*datadogV2.GCPSTSAccount, error) {
	if accountToFindID == "" {
		idEmptyError := errors.New("Error your service account's unique account ID is empty \"\"")
		return nil, idEmptyError
	}

	var foundAccount *datadogV2.GCPSTSAccount

	for _, accountObject := range accounts.GetData() {
		accountID := accountObject.GetId()

		if accountID == accountToFindID {
			foundAccount = &accountObject
			return foundAccount, nil
		}
	}

	return nil, fmt.Errorf("Error your service account with ID '%s' was not found within Datadog's backend", accountToFindID)
}

func attributeListToStringList(ctx context.Context, listOfAttributes []attr.Value) ([]string, error) {
	var listOfHostFilters []string

	// Convert each element into a Go Type, rather than a TF type
	for _, element := range listOfAttributes {
		stringElement, err := element.ToTerraformValue(ctx)
		if err != nil {
			return nil, err
		}

		var valuePointer string
		stringElement.Copy().As(&valuePointer)

		listOfHostFilters = append(listOfHostFilters, valuePointer)
	}

	return listOfHostFilters, nil
}
