package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	GcpApi *datadogV2.GCPIntegrationSTSApi
}

// NewDatadogIntegrationGCPSTSResource is a helper function to simplify the provider implementation.
func NewDatadogIntegrationGCPSTSResource() resource.Resource {
	return &datadogIntegrationGCPSTSResource{}
}

// datadogIntegrationGCPSTSResourceModel
type datadogIntegrationGCPSTSResourceModel struct {
	ServiceAccountEmail types.String `tfsdk:"service_account_email"`
	GeneratedSaId       types.String `tfsdk:"generated_sa_id"`
	DelegateEmail       types.String `tfsdk:"delegate_email"`
	Automute            types.Bool   `tfsdk:"automute"`
	EnableCspm          types.Bool   `tfsdk:"enable_cspm"`
	HostFilters         types.List   `tfsdk:"host_filters"`
}

const (
	defaultType = "gcp_service_account"
)

// Metadata returns the resource type name. Resource Name within Terraform is "datadog_integration_gcp_sts".
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
	r.GcpApi = providerData.DatadogApiInstances.GetGCPStsIntegrationApiV2()
}

// Schema defines the configuration used as input within your Terraform Resource.
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
			"generated_sa_id": schema.StringAttribute{
				Description: "Datadog's Unique ID generated for your STS-enabled GCP service account.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host_filters": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Your Datadog Host Filters.",
				Optional:    true,
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

// Create creates the resource and sets the initial Terraform state.
func (r *datadogIntegrationGCPSTSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Get current TF state.
	var plan datadogIntegrationGCPSTSResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a Delegate Service Account within Datadog.
	delegateResponse, _, err := r.GcpApi.CreateGCPSTSDelegate(r.Auth, *datadogV2.NewCreateGCPSTSDelegateOptionalParameters())
	if err != nil {
		resp.Diagnostics.AddError("Error creating GCP Delegate within Datadog",
			"Could not create Delegate Service Account, unexpected error:"+err.Error())
		return
	}

	delegateInfoResponse := delegateResponse.GetData()
	delegateAttributes := delegateInfoResponse.GetAttributes()
	hostFilterPlanElements := plan.HostFilters.Elements()

	listOfHostFilters, err := attributeListToStringList(ctx, hostFilterPlanElements)
	if err != nil {
		resp.Diagnostics.AddError("Error converting attribute list to strings",
			"Error converting attribute list to strings:"+err.Error())
		return
	}

	// Create an entry wthin Datadog for your STS enabled service account.
	saInfo := datadogV2.ServiceAccountToBeCreatedData{
		Data: &datadogV2.ServiceAccountMetadata{
			Attributes: &datadogV2.AttributeMetadata{
				ClientEmail:   stringToPointer(plan.ServiceAccountEmail.ValueString()),
				Automute:      boolToPointer(plan.Automute.ValueBool()),
				IsCspmEnabled: boolToPointer(plan.EnableCspm.ValueBool()),
				HostFilters:   listOfHostFilters,
			},
			Type: stringToPointer(defaultType),
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
	plan.GeneratedSaId = types.StringValue(createdServiceAccountInfo.GetId())
	plan.DelegateEmail = types.StringValue(delegateAttributes.GetDelegateAccountEmail())

	// Write state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resets the Terraform state using the latest "pulled" data. Read() is called when running Terraform Plans.
func (r *datadogIntegrationGCPSTSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get Current State.
	var state datadogIntegrationGCPSTSResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the Delegate email.
	delegateResponse, _, err := r.GcpApi.GetGCPSTSDelegate(r.Auth, *datadogV2.NewGetGCPSTSDelegateOptionalParameters())
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving STS delegate"))
		return
	}

	delegateResponseData := delegateResponse.GetData()
	delegateAttributes := delegateResponseData.GetAttributes()
	state.DelegateEmail = types.StringValue(delegateAttributes.GetDelegateAccountEmail())

	stsEnabledAccounts, _, err := r.GcpApi.ListGCPSTSAccounts(r.Auth)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving STS service accounts",
			"Error listing GCP STS Accounts:"+err.Error())
		return
	}

	// Find Service Account by ID.
	var foundAccount *datadogV2.GCPSTSAccounts
	for _, accountObject := range stsEnabledAccounts.GetData() {
		accountUniqueID := accountObject.GetId()

		if accountUniqueID == state.GeneratedSaId.ValueString() {
			foundAccount = &accountObject
			break
		}
	}
	if foundAccount == nil {
		resp.Diagnostics.AddError("Error finding your service account",
			"Error couldn't find your service account with ID:"+state.GeneratedSaId.ValueString())
		return
	}

	// Retrieve Host Filters.
	accountAttributes := foundAccount.GetAttributes()
	currentHostFilters := accountAttributes.GetHostFilters()

	var requiredAttributes []attr.Value
	for _, hostFilter := range currentHostFilters {
		requiredAttributes = append(requiredAttributes, types.StringValue(hostFilter))
	}
	outputListValue, _ := types.ListValue(types.StringType, requiredAttributes)

	state.HostFilters = outputListValue
	state.Automute = types.BoolValue(accountAttributes.GetAutomute())
	state.EnableCspm = types.BoolValue(accountAttributes.GetIsCspmEnabled())
	state.ServiceAccountEmail = types.StringValue(accountAttributes.GetClientEmail())

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update re-sets the Terraform state locally and on the Datadog "backend".
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

	hostFilterPlanElements := plan.HostFilters.Elements()
	listOfHostFilters, err := attributeListToStringList(ctx, hostFilterPlanElements)
	if err != nil {
		resp.Diagnostics.AddError("Error converting attribute list to strings",
			"Error converting attribute list to strings:"+err.Error())
		return
	}

	updatedSAInfo := datadogV2.DataObjectPatch{
		Data: &datadogV2.ServiceAccountInfoPatch{
			Type: stringToPointer(defaultType),
			Attributes: &datadogV2.ServiceAccountInfoPatchAttributes{
				IsCspmEnabled: boolToPointer(plan.EnableCspm.ValueBool()),
				Automute:      boolToPointer(plan.Automute.ValueBool()),
				HostFilters:   listOfHostFilters,
			},
		},
	}

	uniqueAccountID := currentState.GeneratedSaId.ValueString()
	updateResponse, _, err := r.GcpApi.UpdateGCPSTSAccount(r.Auth, uniqueAccountID, updatedSAInfo)
	if err != nil {
		resp.Diagnostics.AddError("Error updating your service account",
			"Error:"+err.Error())
		return
	}

	dataBlock := updateResponse.GetData()
	blockAttributes := dataBlock.GetAttributes()

	plan.GeneratedSaId = basetypes.NewStringValue(dataBlock.GetId())
	plan.DelegateEmail = currentState.DelegateEmail
	plan.ServiceAccountEmail = basetypes.NewStringValue(blockAttributes.GetClientEmail())

	if plan.Automute.IsNull() {
		plan.Automute = currentState.Automute
	}
	if plan.EnableCspm.IsNull() {
		plan.EnableCspm = currentState.EnableCspm
	}
	if plan.HostFilters.IsNull() {
		plan.HostFilters = currentState.HostFilters
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource, and removes the Terraform state on success.
func (r *datadogIntegrationGCPSTSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datadogIntegrationGCPSTSResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.GcpApi.DeleteGCPSTSAccount(r.Auth, state.GeneratedSaId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting your service account",
			"Error encountered when attempting to delete your service account from Datadog"+err.Error())
		return
	}
}

func stringToPointer(s string) *string {
	return &s
}

func boolToPointer(b bool) *bool {
	return &b
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