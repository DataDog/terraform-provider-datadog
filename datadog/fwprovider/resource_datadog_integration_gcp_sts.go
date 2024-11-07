package fwprovider

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	integrationGcpStsMutex sync.Mutex
	_                      resource.ResourceWithConfigure   = &integrationGcpStsResource{}
	_                      resource.ResourceWithImportState = &integrationGcpStsResource{}
)

type integrationGcpStsResource struct {
	Api  *datadogV2.GCPIntegrationApi
	Auth context.Context
}

type MetricNamespaceConfigModel struct {
	ID       types.String `tfsdk:"id"`
	Disabled types.Bool   `tfsdk:"disabled"`
}

type integrationGcpStsModel struct {
	ID                                types.String                  `tfsdk:"id"`
	AccountTags                       types.Set                     `tfsdk:"account_tags"`
	Automute                          types.Bool                    `tfsdk:"automute"`
	ClientEmail                       types.String                  `tfsdk:"client_email"`
	DelegateAccountEmail              types.String                  `tfsdk:"delegate_account_email"`
	HostFilters                       types.Set                     `tfsdk:"host_filters"`
	CloudRunRevisionFilters           types.Set                     `tfsdk:"cloud_run_revision_filters"`
	MetricNamespaceConfigs            []*MetricNamespaceConfigModel `tfsdk:"metric_namespace_configs"`
	IsCspmEnabled                     types.Bool                    `tfsdk:"is_cspm_enabled"`
	IsSecurityCommandCenterEnabled    types.Bool                    `tfsdk:"is_security_command_center_enabled"`
	IsResourceChangeCollectionEnabled types.Bool                    `tfsdk:"is_resource_change_collection_enabled"`
	ResourceCollectionEnabled         types.Bool                    `tfsdk:"resource_collection_enabled"`
}

func NewIntegrationGcpStsResource() resource.Resource {
	return &integrationGcpStsResource{}
}

func (r *integrationGcpStsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetGCPIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationGcpStsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_gcp_sts"
}

func (r *integrationGcpStsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		// Avoid using default values for bool settings to prevent breaking changes for existing customers.
		// Customers who have previously modified these settings via the UI should not be impacted
		// https://github.com/DataDog/terraform-provider-datadog/pull/2424#issuecomment-2150871460
		Description: "Provides a Datadog Integration GCP Sts resource. This can be used to create and manage Datadog - Google Cloud Platform integration.",
		Attributes: map[string]schema.Attribute{
			"account_tags": schema.SetAttribute{
				Optional:    true,
				Description: "Tags to be associated with GCP metrics and service checks from your account.",
				ElementType: types.StringType,
			},
			"automute": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Silence monitors for expected GCE instance shutdowns.",
			},
			"client_email": schema.StringAttribute{
				Required:    true,
				Description: "Your service account email address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delegate_account_email": schema.StringAttribute{
				Computed:    true,
				Description: "Datadog's STS Delegate Email.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host_filters": schema.SetAttribute{
				Optional:    true,
				Description: "Your Host Filters.",
				ElementType: types.StringType,
			},
			"cloud_run_revision_filters": schema.SetAttribute{
				Optional:    true,
				Description: "Tags to filter which Cloud Run revisions are imported into Datadog. Only revisions that meet specified criteria are monitored.",
				ElementType: types.StringType,
			},
			"metric_namespace_configs": schema.SetAttribute{
				Optional:    true,
				Description: "Configuration for a GCP metric namespace.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":       types.StringType,
						"disabled": types.BoolType,
					},
				},
			},
			"is_cspm_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether Datadog collects cloud security posture management resources from your GCP project. If enabled, requires `resource_collection_enabled` to also be enabled.",
			},
			"is_security_command_center_enabled": schema.BoolAttribute{
				Description: "When enabled, Datadog will attempt to collect Security Command Center Findings. Note: This requires additional permissions on the service account.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_resource_change_collection_enabled": schema.BoolAttribute{
				Description: "When enabled, Datadog scans for all resource change data in your Google Cloud environment.",
				Optional:    true,
				Computed:    true,
			},
			"resource_collection_enabled": schema.BoolAttribute{
				Description: "When enabled, Datadog scans for all resources in your GCP environment.",
				Optional:    true,
				Computed:    true,
			}, "id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationGcpStsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationGcpStsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	resp, httpResp, err := r.Api.ListGCPSTSAccounts(r.Auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Integration Gcp Sts"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	found := false
	for _, account := range resp.GetData() {
		if account.GetId() == state.ID.ValueString() {
			found = true
			r.updateState(ctx, &state, &account)
			break
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpStsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	integrationGcpStsMutex.Lock()
	defer integrationGcpStsMutex.Unlock()

	// This resource is special and uses datadog delegate account.
	// The datadog delegate account cannot mutate after creation hence it is safe
	// to call MakeGCPSTSDelegate multiple times. And to ensure it is created, we call it once before creating
	// gcp sts resource.
	delegateResponse, _, err := r.Api.MakeGCPSTSDelegate(r.Auth, *datadogV2.NewMakeGCPSTSDelegateOptionalParameters())
	if err != nil {
		response.Diagnostics.AddError("Error creating GCP Delegate within Datadog",
			"Could not create Delegate Service Account, unexpected error: "+err.Error())
		return
	}
	delegateEmail := delegateResponse.Data.Attributes.GetDelegateAccountEmail()
	state.DelegateAccountEmail = types.StringValue(delegateEmail)

	attributes, diags := r.buildIntegrationGcpStsRequestBody(ctx, &state)
	if !state.ClientEmail.IsNull() {
		attributes.SetClientEmail(state.ClientEmail.ValueString())
	}

	body := datadogV2.NewGCPSTSServiceAccountCreateRequestWithDefaults()
	body.Data = datadogV2.NewGCPSTSServiceAccountDataWithDefaults()
	body.Data.SetAttributes(attributes)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateGCPSTSAccount(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Integration Gcp Sts"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, resp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpStsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	integrationGcpStsMutex.Lock()
	defer integrationGcpStsMutex.Unlock()

	id := state.ID.ValueString()

	attributes, diags := r.buildIntegrationGcpStsRequestBody(ctx, &state)
	body := datadogV2.NewGCPSTSServiceAccountUpdateRequestWithDefaults()
	body.Data = datadogV2.NewGCPSTSServiceAccountUpdateRequestDataWithDefaults()
	body.Data.SetAttributes(attributes)

	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateGCPSTSAccount(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Integration Gcp Sts"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, resp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpStsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	integrationGcpStsMutex.Lock()
	defer integrationGcpStsMutex.Unlock()

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteGCPSTSAccount(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_gcp_sts"))
		return
	}
}

func (r *integrationGcpStsResource) updateState(ctx context.Context, state *integrationGcpStsModel, resp *datadogV2.GCPSTSServiceAccount) {
	state.ID = types.StringValue(resp.GetId())

	attributes := resp.GetAttributes()
	if accountTags, ok := attributes.GetAccountTagsOk(); ok && len(*accountTags) > 0 {
		state.AccountTags, _ = types.SetValueFrom(ctx, types.StringType, *accountTags)
	}
	if automute, ok := attributes.GetAutomuteOk(); ok {
		state.Automute = types.BoolValue(*automute)
	}
	if clientEmail, ok := attributes.GetClientEmailOk(); ok {
		state.ClientEmail = types.StringValue(*clientEmail)
	}
	if hostFilters, ok := attributes.GetHostFiltersOk(); ok && len(*hostFilters) > 0 {
		state.HostFilters, _ = types.SetValueFrom(ctx, types.StringType, *hostFilters)
	}
	if runFilters, ok := attributes.GetCloudRunRevisionFiltersOk(); ok && len(*runFilters) > 0 {
		state.CloudRunRevisionFilters, _ = types.SetValueFrom(ctx, types.StringType, *runFilters)
	}
	if namespaceConfigs, ok := attributes.GetMetricNamespaceConfigsOk(); ok && len(*namespaceConfigs) > 0 {
		state.MetricNamespaceConfigs = make([]*MetricNamespaceConfigModel, len(*namespaceConfigs))
		for i, namespaceConfig := range *namespaceConfigs {
			state.MetricNamespaceConfigs[i] = &MetricNamespaceConfigModel{
				ID:       types.StringValue(namespaceConfig.GetId()),
				Disabled: types.BoolValue(namespaceConfig.GetDisabled()),
			}
		}
	}
	if isCspmEnabled, ok := attributes.GetIsCspmEnabledOk(); ok {
		state.IsCspmEnabled = types.BoolValue(*isCspmEnabled)
	}
	if isSecurityCommandCenterEnabled, ok := attributes.GetIsSecurityCommandCenterEnabledOk(); ok {
		state.IsSecurityCommandCenterEnabled = types.BoolValue(*isSecurityCommandCenterEnabled)
	}
	if isResourceChangeCollectionEnabled, ok := attributes.GetIsResourceChangeCollectionEnabledOk(); ok {
		state.IsResourceChangeCollectionEnabled = types.BoolValue(*isResourceChangeCollectionEnabled)
	}
	if resourceCollectionEnabled, ok := attributes.GetResourceCollectionEnabledOk(); ok {
		state.ResourceCollectionEnabled = types.BoolValue(*resourceCollectionEnabled)
	}
}

func (r *integrationGcpStsResource) buildIntegrationGcpStsRequestBody(ctx context.Context, state *integrationGcpStsModel) (datadogV2.GCPSTSServiceAccountAttributes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.GCPSTSServiceAccountAttributes{}

	accountTags := make([]string, 0)
	if !state.AccountTags.IsNull() {
		diags.Append(state.AccountTags.ElementsAs(ctx, &accountTags, false)...)
	}
	attributes.SetAccountTags(accountTags)

	if !state.Automute.IsNull() {
		attributes.SetAutomute(state.Automute.ValueBool())
	}
	if !state.IsCspmEnabled.IsNull() {
		attributes.SetIsCspmEnabled(state.IsCspmEnabled.ValueBool())
	}

	hostFilters := make([]string, 0)
	if !state.HostFilters.IsNull() {
		diags.Append(state.HostFilters.ElementsAs(ctx, &hostFilters, false)...)
	}
	attributes.SetHostFilters(hostFilters)

	runFilters := make([]string, 0)
	if !state.CloudRunRevisionFilters.IsNull() {
		diags.Append(state.CloudRunRevisionFilters.ElementsAs(ctx, &runFilters, false)...)
	}
	attributes.SetCloudRunRevisionFilters(runFilters)

	namespaceConfigs := make([]datadogV2.GCPMetricNamespaceConfig, 0)
	if len(state.MetricNamespaceConfigs) > 0 {
		for _, namespaceConfig := range state.MetricNamespaceConfigs {
			namespaceConfigs = append(namespaceConfigs, datadogV2.GCPMetricNamespaceConfig{
				Id:       namespaceConfig.ID.ValueStringPointer(),
				Disabled: namespaceConfig.Disabled.ValueBoolPointer(),
			})
		}
	}
	attributes.SetMetricNamespaceConfigs(namespaceConfigs)

	if !state.IsSecurityCommandCenterEnabled.IsUnknown() {
		attributes.SetIsSecurityCommandCenterEnabled(state.IsSecurityCommandCenterEnabled.ValueBool())
	}
	if !state.IsResourceChangeCollectionEnabled.IsUnknown() {
		attributes.SetIsResourceChangeCollectionEnabled(state.IsResourceChangeCollectionEnabled.ValueBool())
	}
	if !state.ResourceCollectionEnabled.IsUnknown() {
		attributes.SetResourceCollectionEnabled(state.ResourceCollectionEnabled.ValueBool())
	}

	return attributes, diags
}
