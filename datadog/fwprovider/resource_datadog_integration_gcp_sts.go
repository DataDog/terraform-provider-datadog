package fwprovider

import (
	"context"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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

var MonitoredResourceConfigSpec = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"filters": types.SetType{
			ElemType: types.StringType,
		},
	},
}

type MonitoredResourceConfigModel struct {
	Type    types.String `tfsdk:"type"`
	Filters types.Set    `tfsdk:"filters"`
}

type integrationGcpStsModel struct {
	ID                                types.String                  `tfsdk:"id"`
	AccountTags                       types.Set                     `tfsdk:"account_tags"`
	Automute                          types.Bool                    `tfsdk:"automute"`
	ClientEmail                       types.String                  `tfsdk:"client_email"`
	DelegateAccountEmail              types.String                  `tfsdk:"delegate_account_email"`
	HostFilters                       types.Set                     `tfsdk:"host_filters"`               // DEPRECATED: use MonitoredResourceConfigs["gce_instance"]
	CloudRunRevisionFilters           types.Set                     `tfsdk:"cloud_run_revision_filters"` // DEPRECATED: use MonitoredResourceConfigs["cloud_run_revision"]
	IsCspmEnabled                     types.Bool                    `tfsdk:"is_cspm_enabled"`
	IsSecurityCommandCenterEnabled    types.Bool                    `tfsdk:"is_security_command_center_enabled"`
	IsResourceChangeCollectionEnabled types.Bool                    `tfsdk:"is_resource_change_collection_enabled"`
	IsPerProjectQuotaEnabled          types.Bool                    `tfsdk:"is_per_project_quota_enabled"`
	ResourceCollectionEnabled         types.Bool                    `tfsdk:"resource_collection_enabled"`
	MetricNamespaceConfigs            []*MetricNamespaceConfigModel `tfsdk:"metric_namespace_configs"`
	MonitoredResourceConfigs          types.Set                     `tfsdk:"monitored_resource_configs"`
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
			"id": utils.ResourceIDAttribute(),
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
				Optional:           true,
				Computed:           true,
				Description:        "List of filters to limit the VM instances that are pulled into Datadog by using tags. Only VM instance resources that apply to specified filters are imported into Datadog.",
				ElementType:        types.StringType,
				DeprecationMessage: "**Note:** This field is deprecated. Instead, use `monitored_resource_configs` with `type=gce_instance`",
			},
			"cloud_run_revision_filters": schema.SetAttribute{
				Optional:           true,
				Computed:           true,
				Description:        "List of filters to limit the Cloud Run revisions that are pulled into Datadog by using tags. Only Cloud Run revision resources that apply to specified filters are imported into Datadog.",
				ElementType:        types.StringType,
				DeprecationMessage: "**Note:** This field is deprecated. Instead, use `monitored_resource_configs` with `type=cloud_run_revision`",
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
			"is_per_project_quota_enabled": schema.BoolAttribute{
				Description: "When enabled, Datadog includes the `X-Goog-User-Project` header to attribute Google Cloud billing and quota usage to the monitored project instead of the default service account project.",
				Optional:    true,
				Computed:    true,
			},
			"resource_collection_enabled": schema.BoolAttribute{
				Description: "When enabled, Datadog scans for all resources in your GCP environment.",
				Optional:    true,
				Computed:    true,
			},
			"metric_namespace_configs": schema.SetAttribute{
				Optional:    true,
				Description: "Configurations for GCP metric namespaces.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":       types.StringType,
						"disabled": types.BoolType,
					},
				},
			},
			"monitored_resource_configs": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Configurations for GCP monitored resources. Only monitored resources that apply to specified filters are imported into Datadog.",
				ElementType: MonitoredResourceConfigSpec,
			},
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

	empties, d := extractEmptyMRCs(ctx, state.MonitoredResourceConfigs)
	response.Diagnostics.Append(d...)
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
			// Re-inject empty filters into state to preserve user config
			merged, d2 := mergeMRCs(ctx, state.MonitoredResourceConfigs, empties)
			response.Diagnostics.Append(d2...)
			if response.Diagnostics.HasError() {
				return
			}
			state.MonitoredResourceConfigs = merged

			break
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpStsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var userCfg integrationGcpStsModel
	response.Diagnostics.Append(request.Config.Get(ctx, &userCfg)...)
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

	// The API explicitly filters out empty filters, so plan != state
	// To resolve this we will re-inject empty filters from the user input
	empties, d1 := extractEmptyMRCs(ctx, userCfg.MonitoredResourceConfigs)
	response.Diagnostics.Append(d1...)
	merged, d2 := mergeMRCs(ctx, state.MonitoredResourceConfigs, empties)
	response.Diagnostics.Append(d2...)
	state.MonitoredResourceConfigs = merged

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)

}

func (r *integrationGcpStsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var userCfg integrationGcpStsModel
	response.Diagnostics.Append(request.Config.Get(ctx, &userCfg)...)
	if response.Diagnostics.HasError() {
		return
	}

	empties, d := extractEmptyMRCs(ctx, userCfg.MonitoredResourceConfigs)
	response.Diagnostics.Append(d...)
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

	// Re-inject empty filters from the user config back into state so state matches config
	merged, d2 := mergeMRCs(ctx, state.MonitoredResourceConfigs, empties)
	response.Diagnostics.Append(d2...)
	if response.Diagnostics.HasError() {
		return
	}
	state.MonitoredResourceConfigs = merged

	// 6) Save state
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
	if automute, ok := attributes.GetAutomuteOk(); ok {
		state.Automute = types.BoolValue(*automute)
	}
	if clientEmail, ok := attributes.GetClientEmailOk(); ok {
		state.ClientEmail = types.StringValue(*clientEmail)
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
	if isPerProjectQuotaEnabled, ok := attributes.GetIsPerProjectQuotaEnabledOk(); ok {
		state.IsPerProjectQuotaEnabled = types.BoolValue(*isPerProjectQuotaEnabled)
	}
	if resourceCollectionEnabled, ok := attributes.GetResourceCollectionEnabledOk(); ok {
		state.ResourceCollectionEnabled = types.BoolValue(*resourceCollectionEnabled)
	}

	if accountTags := attributes.GetAccountTags(); len(accountTags) > 0 {
		state.AccountTags, _ = types.SetValueFrom(ctx, types.StringType, accountTags)
	}

	if mncs := attributes.GetMetricNamespaceConfigs(); len(mncs) > 0 {
		state.MetricNamespaceConfigs = make([]*MetricNamespaceConfigModel, 0, len(mncs))
		for _, mnc := range mncs {
			state.MetricNamespaceConfigs = append(state.MetricNamespaceConfigs, &MetricNamespaceConfigModel{
				ID:       types.StringValue(mnc.GetId()),
				Disabled: types.BoolValue(mnc.GetDisabled()),
			})
		}
	}

	state.HostFilters, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetHostFilters())
	state.CloudRunRevisionFilters, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetCloudRunRevisionFilters())
	mrcs := make([]*MonitoredResourceConfigModel, 0)
	for _, mrc := range attributes.GetMonitoredResourceConfigs() {
		var mdl MonitoredResourceConfigModel
		mdl.Type = types.StringValue(string(mrc.GetType()))
		mdl.Filters, _ = types.SetValueFrom(ctx, types.StringType, mrc.GetFilters())
		mrcs = append(mrcs, &mdl)
	}
	state.MonitoredResourceConfigs, _ = types.SetValueFrom(ctx, MonitoredResourceConfigSpec, mrcs)
}

func (r *integrationGcpStsResource) buildIntegrationGcpStsRequestBody(ctx context.Context, state *integrationGcpStsModel) (datadogV2.GCPSTSServiceAccountAttributes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.GCPSTSServiceAccountAttributes{}

	if !state.Automute.IsNull() {
		attributes.SetAutomute(state.Automute.ValueBool())
	}
	if !state.IsCspmEnabled.IsNull() {
		attributes.SetIsCspmEnabled(state.IsCspmEnabled.ValueBool())
	}
	if !state.IsSecurityCommandCenterEnabled.IsUnknown() {
		attributes.SetIsSecurityCommandCenterEnabled(state.IsSecurityCommandCenterEnabled.ValueBool())
	}
	if !state.IsResourceChangeCollectionEnabled.IsUnknown() {
		attributes.SetIsResourceChangeCollectionEnabled(state.IsResourceChangeCollectionEnabled.ValueBool())
	}
	if !state.ResourceCollectionEnabled.IsUnknown() {
		attributes.SetResourceCollectionEnabled(state.ResourceCollectionEnabled.ValueBool())
	}
	if !state.IsPerProjectQuotaEnabled.IsUnknown() {
		attributes.SetIsPerProjectQuotaEnabled(state.IsPerProjectQuotaEnabled.ValueBool())
	}

	attributes.SetAccountTags(tfCollectionToSlice[string](ctx, diags, state.AccountTags))

	mncs := make([]datadogV2.GCPMetricNamespaceConfig, 0)
	for _, mnc := range state.MetricNamespaceConfigs {
		mncs = append(mncs, datadogV2.GCPMetricNamespaceConfig{
			Id:       mnc.ID.ValueStringPointer(),
			Disabled: mnc.Disabled.ValueBoolPointer(),
		})
	}
	attributes.SetMetricNamespaceConfigs(mncs)

	attributes.SetHostFilters(tfCollectionToSlice[string](ctx, diags, state.HostFilters))
	attributes.SetCloudRunRevisionFilters(tfCollectionToSlice[string](ctx, diags, state.CloudRunRevisionFilters))
	mrcs := make([]datadogV2.GCPMonitoredResourceConfig, 0)
	for _, mrc := range tfCollectionToSlice[*MonitoredResourceConfigModel](ctx, diags, state.MonitoredResourceConfigs) {
		mrcs = append(mrcs, datadogV2.GCPMonitoredResourceConfig{
			Type:    ptrTo(datadogV2.GCPMonitoredResourceConfigType(mrc.Type.ValueString())),
			Filters: tfCollectionToSlice[string](ctx, diags, mrc.Filters),
		})
	}
	attributes.SetMonitoredResourceConfigs(mrcs)

	return attributes, diags
}

func extractEmptyMRCs(ctx context.Context, set types.Set) ([]MonitoredResourceConfigModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	out := []MonitoredResourceConfigModel{}
	if set.IsNull() || set.IsUnknown() {
		return out, diags
	}

	var items []MonitoredResourceConfigModel
	diags.Append(set.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return out, diags
	}

	for _, item := range items {
		if item.Type.IsNull() || item.Type.IsUnknown() || item.Filters.IsNull() || item.Filters.IsUnknown() {
			continue
		}
		var fs []string
		diags.Append(item.Filters.ElementsAs(ctx, &fs, false)...)
		if diags.HasError() {
			return out, diags
		}
		if len(fs) == 0 {
			emptyFilters, _ := types.SetValueFrom(ctx, types.StringType, []string{})
			out = append(out, MonitoredResourceConfigModel{
				Type:    types.StringValue(item.Type.ValueString()),
				Filters: emptyFilters,
			})
		}
	}
	return out, diags
}

func mergeMRCs(ctx context.Context, base types.Set, toAdd []MonitoredResourceConfigModel) (types.Set, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var baseItems []MonitoredResourceConfigModel
	if !base.IsNull() && !base.IsUnknown() {
		diags.Append(base.ElementsAs(ctx, &baseItems, false)...)
		if diags.HasError() {
			return base, diags
		}
	}

	baseItems = append(baseItems, toAdd...)

	newSet, d := types.SetValueFrom(ctx, MonitoredResourceConfigSpec, baseItems)
	diags.Append(d...)
	if diags.HasError() {
		return base, diags
	}
	return newSet, diags
}

func tfCollectionToSlice[T any](ctx context.Context, diags diag.Diagnostics, col tfCollection) []T {
	slice := make([]T, 0)
	if !col.IsNull() {
		diags.Append(col.ElementsAs(ctx, &slice, false)...)
	}
	return slice
}

func ptrTo[T any](item T) *T {
	return &item
}

type tfCollection interface {
	IsNull() bool
	ElementsAs(ctx context.Context, target interface{}, allowUnhandled bool) diag.Diagnostics
}
