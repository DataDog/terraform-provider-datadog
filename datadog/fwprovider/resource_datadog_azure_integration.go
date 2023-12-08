package fwprovider

import (
	"context"
	"fmt"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &integrationAzureResource{}
	_ resource.ResourceWithImportState = &integrationAzureResource{}
)

var integrationAzureMutex = sync.Mutex{}

type integrationAzureResource struct {
	Api  *datadogV1.AzureIntegrationApi
	Auth context.Context
}

type integrationAzureModel struct {
	ID                    types.String `tfsdk:"id"`
	AppServicePlanFilters types.String `tfsdk:"app_service_plan_filters"`
	Automute              types.Bool   `tfsdk:"automute"`
	ClientId              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	ContainerAppFilters   types.String `tfsdk:"container_app_filters"`
	CspmEnabled           types.Bool   `tfsdk:"cspm_enabled"`
	CustomMetricsEnabled  types.Bool   `tfsdk:"custom_metrics_enabled"`
	HostFilters           types.String `tfsdk:"host_filters"`
	TenantName            types.String `tfsdk:"tenant_name"`
}

func NewIntegrationAzureResource() resource.Resource {
	return &integrationAzureResource{}
}

func (r *integrationAzureResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAzureIntegrationApiV1()
	r.Auth = providerData.Auth
}

func (r *integrationAzureResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_azure"
}

func (r *integrationAzureResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Your Azure web application ID.",
			},
			"client_secret": schema.StringAttribute{
				Required:    true,
				Description: "(Required for Initial Creation) Your Azure web application secret key.",
				Sensitive:   true,
			},
			"tenant_name": schema.StringAttribute{
				Required:    true,
				Description: "Your Azure Active Directory ID.",
			},
			"automute": schema.BoolAttribute{
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Optional:    true,
				Description: "Silence monitors for expected Azure VM shutdowns.",
			},
			"cspm_enabled": schema.BoolAttribute{
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Optional:    true,
				Description: "Enable Cloud Security Management Misconfigurations for your organization.",
			},
			"custom_metrics_enabled": schema.BoolAttribute{
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Optional:    true,
				Description: "Enable custom metrics for your organization.",
			},
			"host_filters": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "String of host tag(s) (in the form `key:value,key:value`) defines a filter that Datadog will use when collecting metrics from Azure. Limit the Azure instances that are pulled into Datadog by using tags. Only hosts that match one of the defined tags are imported into Datadog. e.x. `env:production,deploymentgroup:red`",
				Default:     stringdefault.StaticString(""),
			},
			"container_app_filters": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "This comma-separated list of tags (in the form `key:value,key:value`) defines a filter that Datadog uses when collecting metrics from Azure Container Apps. Only Container Apps that match one of the defined tags are imported into Datadog.",
				Default:     stringdefault.StaticString(""),
			},
			"app_service_plan_filters": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "This comma-separated list of tags (in the form `key:value,key:value`) defines a filter that Datadog uses when collecting metrics from Azure App Service Plans. Only App Service Plans that match one of the defined tags are imported into Datadog. The rest, including the apps and functions running on them, are ignored. This also filters the metrics for any App or Function running on the App Service Plan(s).",
				Default:     stringdefault.StaticString(""),
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationAzureResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationAzureResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationAzureModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	account, diags := r.getAzureAccount(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if account == nil {
		response.State.RemoveResource(ctx)
		return
	}
	r.updateState(ctx, &state, account)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAzureResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationAzureModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	integrationAzureMutex.Lock()
	defer integrationAzureMutex.Unlock()

	body := r.buildIntegrationAzureRequestBody(ctx, &state, state.TenantName.ValueString(), state.ClientId.ValueString(), false)

	_, _, err := r.Api.CreateAzureIntegration(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating an Azure integration"))
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s:%s", state.TenantName.ValueString(), state.ClientId.ValueString()))

	account, diags := r.getAzureAccount(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if account == nil {
		response.Diagnostics.AddError("error retrieving Azure integration", "")
		return
	}

	r.updateState(ctx, &state, account)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAzureResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationAzureModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	integrationAzureMutex.Lock()
	defer integrationAzureMutex.Unlock()

	tenantName, clientId, err := utils.TenantAndClientFromID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, ""))
		return
	}
	body := r.buildIntegrationAzureRequestBody(ctx, &state, tenantName, clientId, true)

	_, _, err = r.Api.UpdateAzureIntegration(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Azure integration"))
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s:%s", state.TenantName.ValueString(), state.ClientId.ValueString()))

	account, diags := r.getAzureAccount(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if account == nil {
		response.Diagnostics.AddError("error retrieving Azure integration", "")
		return
	}

	r.updateState(ctx, &state, account)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAzureResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationAzureModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	integrationAzureMutex.Lock()
	defer integrationAzureMutex.Unlock()

	tenantName, clientId, err := utils.TenantAndClientFromID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, ""))
		return
	}
	body := r.buildIntegrationAzureRequestBody(ctx, &state, tenantName, clientId, false)

	_, httpResp, err := r.Api.DeleteAzureIntegration(r.Auth, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting azure_integration"))
		return
	}
}

func (r *integrationAzureResource) updateState(ctx context.Context, state *integrationAzureModel, account *datadogV1.AzureAccount) {
	state.TenantName = types.StringValue(account.GetTenantName())
	state.ClientId = types.StringValue(account.GetClientId())
	state.Automute = types.BoolValue(account.GetAutomute())
	state.CspmEnabled = types.BoolValue(account.GetCspmEnabled())
	state.CustomMetricsEnabled = types.BoolValue(account.GetCustomMetricsEnabled())

	hostFilters, exists := account.GetHostFiltersOk()
	if exists {
		state.HostFilters = types.StringValue(*hostFilters)
	}
	appServicePlanFilters, exists := account.GetAppServicePlanFiltersOk()
	if exists {
		state.AppServicePlanFilters = types.StringValue(*appServicePlanFilters)
	}
	containerAppFilters, exists := account.GetContainerAppFiltersOk()
	if exists {
		state.ContainerAppFilters = types.StringValue(*containerAppFilters)
	}

	state.ID = types.StringValue(fmt.Sprintf("%s:%s", account.GetTenantName(), account.GetClientId()))
}

func (r *integrationAzureResource) getAzureAccount(ctx context.Context, state *integrationAzureModel) (*datadogV1.AzureAccount, diag.Diagnostics) {
	var diags diag.Diagnostics

	tenantName, clientId, err := utils.TenantAndClientFromID(state.ID.ValueString())
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, ""))
		return nil, diags
	}

	resp, _, err := r.Api.ListAzureIntegration(r.Auth)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error listing azure integration"))
		return nil, diags
	}

	var account *datadogV1.AzureAccount
	for _, integration := range resp {
		if integration.GetTenantName() == tenantName && integration.GetClientId() == clientId {
			if err := utils.CheckForUnparsed(integration); err != nil {
				diags.AddError("response contains unparsedObject", err.Error())
				return nil, diags
			}

			account = &integration
			break
		}
	}

	return account, diags
}

func (r *integrationAzureResource) buildIntegrationAzureRequestBody(ctx context.Context, state *integrationAzureModel, tenantName string, clientID string, update bool) *datadogV1.AzureAccount {
	datadogDefinition := datadogV1.NewAzureAccount()
	// Required params
	datadogDefinition.SetTenantName(tenantName)
	datadogDefinition.SetClientId(clientID)
	// Optional params
	datadogDefinition.SetHostFilters(state.HostFilters.ValueString())
	datadogDefinition.SetAppServicePlanFilters(state.AppServicePlanFilters.ValueString())
	datadogDefinition.SetContainerAppFilters(state.ContainerAppFilters.ValueString())
	datadogDefinition.SetAutomute(state.Automute.ValueBool())
	datadogDefinition.SetCspmEnabled(state.CspmEnabled.ValueBool())
	datadogDefinition.SetCustomMetricsEnabled(state.CustomMetricsEnabled.ValueBool())

	if !state.ClientSecret.IsNull() {
		datadogDefinition.SetClientSecret(state.ClientSecret.ValueString())
	}
	// Only do the following if building for the Update
	if update {
		if !state.TenantName.IsNull() {
			datadogDefinition.SetNewTenantName(state.TenantName.ValueString())
		}
		if !state.ClientId.IsNull() {
			datadogDefinition.SetNewClientId(state.ClientId.ValueString())
		}
	}
	return datadogDefinition
}
