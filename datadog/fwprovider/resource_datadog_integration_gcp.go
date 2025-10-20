package fwprovider

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const (
	defaultType                    = "service_account"
	defaultAuthURI                 = "https://accounts.google.com/o/oauth2/auth"
	defaultTokenURI                = "https://oauth2.googleapis.com/token"
	defaultAuthProviderX509CertURL = "https://www.googleapis.com/oauth2/v1/certs"
	defaultClientX509CertURLPrefix = "https://www.googleapis.com/robot/v1/metadata/x509/"
)

var (
	integrationGcpMutex sync.Mutex
	_                   resource.ResourceWithConfigure   = (*integrationGcpResource)(nil)
	_                   resource.ResourceWithImportState = (*integrationGcpResource)(nil)
)

type integrationGcpResource struct {
	api  *datadogV1.GCPIntegrationApi
	auth context.Context
}

type integrationGcpModel struct {
	ID                                types.String `tfsdk:"id"`
	ProjectID                         types.String `tfsdk:"project_id"`
	PrivateKeyId                      types.String `tfsdk:"private_key_id"`
	PrivateKey                        types.String `tfsdk:"private_key"`
	ClientEmail                       types.String `tfsdk:"client_email"`
	ClientId                          types.String `tfsdk:"client_id"`
	Automute                          types.Bool   `tfsdk:"automute"`
	HostFilters                       types.String `tfsdk:"host_filters"`               // DEPRECATED: use MonitoredResourceConfigs["gce_instance"]
	CloudRunRevisionFilters           types.Set    `tfsdk:"cloud_run_revision_filters"` // DEPRECATED: use MonitoredResourceConfigs["cloud_run_revision"]
	ResourceCollectionEnabled         types.Bool   `tfsdk:"resource_collection_enabled"`
	CspmResourceCollectionEnabled     types.Bool   `tfsdk:"cspm_resource_collection_enabled"`
	IsSecurityCommandCenterEnabled    types.Bool   `tfsdk:"is_security_command_center_enabled"`
	IsResourceChangeCollectionEnabled types.Bool   `tfsdk:"is_resource_change_collection_enabled"`
	MonitoredResourceConfigs          types.Set    `tfsdk:"monitored_resource_configs"`
}

func NewIntegrationGcpResource() resource.Resource {
	return &integrationGcpResource{}
}

func (r *integrationGcpResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetGCPIntegrationApiV1()
	r.auth = providerData.Auth
}

func (r *integrationGcpResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_gcp"
}

func (r *integrationGcpResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		// Avoid using default values for bool settings to prevent breaking changes for existing customers.
		// Customers who have previously modified these settings via the UI should not be impacted
		// https://github.com/DataDog/terraform-provider-datadog/pull/2424#issuecomment-2150871460
		Description: "This resource is deprecated—use the `datadog_integration_gcp_sts` resource instead. Provides a Datadog - Google Cloud Platform integration resource. This can be used to create and manage Datadog - Google Cloud Platform integration.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"project_id": schema.StringAttribute{
				Description: "Your Google Cloud project ID found in your JSON service account key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key_id": schema.StringAttribute{
				Description: "Your private key ID found in your JSON service account key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				Description: "Your private key name found in your JSON service account key.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_email": schema.StringAttribute{
				Description: "Your email found in your JSON service account key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_id": schema.StringAttribute{
				Description: "Your ID found in your JSON service account key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_filters": schema.StringAttribute{
				Optional:           true,
				Computed:           true,
				Description:        "List of filters to limit the VM instances that are pulled into Datadog by using tags. Only VM instance resources that apply to specified filters are imported into Datadog.",
				Default:            stringdefault.StaticString(""),
				DeprecationMessage: "**Note:** This field is deprecated. Instead, use `monitored_resource_configs` with `type=gce_instance`",
			},
			"cloud_run_revision_filters": schema.SetAttribute{
				Optional:           true,
				Computed:           true,
				Description:        "List of filters to limit the Cloud Run revisions that are pulled into Datadog by using tags. Only Cloud Run revision resources that apply to specified filters are imported into Datadog.",
				ElementType:        types.StringType,
				DeprecationMessage: "**Note:** This field is deprecated. Instead, use `monitored_resource_configs` with `type=cloud_run_revision`",
			},
			"automute": schema.BoolAttribute{
				Description: "Silence monitors for expected GCE instance shutdowns.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"resource_collection_enabled": schema.BoolAttribute{
				Description: "When enabled, Datadog scans for all resources in your GCP environment.",
				Optional:    true,
				Computed:    true,
			},
			"cspm_resource_collection_enabled": schema.BoolAttribute{
				Description: "Whether Datadog collects cloud security posture management resources from your GCP project. If enabled, requires `resource_collection_enabled` to also be enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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
			"monitored_resource_configs": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Configurations for GCP monitored resources. Only monitored resources that apply to specified filters are imported into Datadog.",
				ElementType: MonitoredResourceConfigSpec,
			},
		},
	}
}

func (r *integrationGcpResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationGcpResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationGcpModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	emptyFilters := extractEmptyMRCs(ctx, response.Diagnostics, tfCollectionToSlice[*MonitoredResourceConfigModel](ctx, response.Diagnostics, state.MonitoredResourceConfigs))

	if response.Diagnostics.HasError() {
		return
	}

	integration, err := r.getGCPIntegration(state)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing GCP integration"))
		return
	}

	if integration == nil {
		response.State.RemoveResource(ctx)
		return
	}

	// Save data into Terraform state
	r.updateState(ctx, &state, integration, emptyFilters)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationGcpModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var userCfg integrationGcpModel
	response.Diagnostics.Append(request.Config.Get(ctx, &userCfg)...)
	if response.Diagnostics.HasError() {
		return
	}

	integrationGcpMutex.Lock()
	defer integrationGcpMutex.Unlock()

	body := r.buildIntegrationGcpRequestBodyBase(state)
	r.addDefaultsToBody(body, state)
	r.addRequiredFieldsToBody(body, state)
	diags := r.addOptionalFieldsToBody(ctx, body, state)

	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, _, err := r.api.CreateGCPIntegration(r.auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating GCP integration"))
		return
	}
	integration, err := r.getGCPIntegration(state)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing GCP integration"))
		return
	}
	if integration == nil {
		response.Diagnostics.AddError("error retrieving GCP integration", "")
		return
	}

	emptyFilters := extractEmptyMRCs(ctx, response.Diagnostics, tfCollectionToSlice[*MonitoredResourceConfigModel](ctx, response.Diagnostics, userCfg.MonitoredResourceConfigs))
	// Save data into Terraform state
	r.updateState(ctx, &state, integration, emptyFilters)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationGcpModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var userCfg integrationGcpModel
	response.Diagnostics.Append(request.Config.Get(ctx, &userCfg)...)

	if response.Diagnostics.HasError() {
		return
	}

	integrationGcpMutex.Lock()
	defer integrationGcpMutex.Unlock()

	body := r.buildIntegrationGcpRequestBodyBase(state)
	diags := r.addOptionalFieldsToBody(ctx, body, state)

	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, _, err := r.api.UpdateGCPIntegration(r.auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating GCP integration"))
		return
	}
	integration, err := r.getGCPIntegration(state)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing GCP integration"))
		return
	}
	if integration == nil {
		response.Diagnostics.AddError("error retrieving GCP integration", "")
		return
	}

	// Save data into Terraform state
	emptyFilters := extractEmptyMRCs(ctx, response.Diagnostics, tfCollectionToSlice[*MonitoredResourceConfigModel](ctx, response.Diagnostics, userCfg.MonitoredResourceConfigs))
	r.updateState(ctx, &state, integration, emptyFilters)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationGcpModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	integrationGcpMutex.Lock()
	defer integrationGcpMutex.Unlock()

	diags := diag.Diagnostics{}
	body := r.buildIntegrationGcpRequestBodyBase(state)

	response.Diagnostics.Append(diags...)

	_, httpResp, err := r.api.DeleteGCPIntegration(r.auth, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting GCP integration"))
		return
	}
}

func (r *integrationGcpResource) updateState(ctx context.Context, state *integrationGcpModel, resp *datadogV1.GCPAccount, emptyFilters []*MonitoredResourceConfigModel) {
	projectId := types.StringValue(resp.GetProjectId())
	// ProjectID and ClientEmail are the only parameters required in all mutating API requests
	state.ID = projectId
	state.ProjectID = projectId
	state.ClientEmail = types.StringValue(resp.GetClientEmail())

	// Computed Values
	state.Automute = types.BoolValue(resp.GetAutomute())
	state.CspmResourceCollectionEnabled = types.BoolValue(resp.GetIsCspmEnabled())
	state.ResourceCollectionEnabled = types.BoolValue(resp.GetResourceCollectionEnabled())
	state.IsSecurityCommandCenterEnabled = types.BoolValue(resp.GetIsSecurityCommandCenterEnabled())
	state.IsResourceChangeCollectionEnabled = types.BoolValue(resp.GetIsResourceChangeCollectionEnabled())

	// Non-computed values
	if clientId, ok := resp.GetClientIdOk(); ok {
		state.ClientId = types.StringValue(*clientId)
	}
	if privateKey, ok := resp.GetPrivateKeyOk(); ok {
		state.PrivateKey = types.StringValue(*privateKey)
	}
	if privateKeyId, ok := resp.GetPrivateKeyIdOk(); ok {
		state.PrivateKeyId = types.StringValue(*privateKeyId)
	}

	state.HostFilters = types.StringValue(resp.GetHostFilters())
	state.CloudRunRevisionFilters, _ = types.SetValueFrom(ctx, types.StringType, resp.GetCloudRunRevisionFilters())
	mrcs := make([]*MonitoredResourceConfigModel, 0)
	for _, mrc := range resp.GetMonitoredResourceConfigs() {
		var mdl MonitoredResourceConfigModel
		mdl.Type = types.StringValue(string(mrc.GetType()))
		mdl.Filters, _ = types.SetValueFrom(ctx, types.StringType, mrc.GetFilters())
		mrcs = append(mrcs, &mdl)
	}
	// https://datadoghq.atlassian.net/browse/GCP-2978 - Currently the API explicitly filters out empty filters, so plan != state
	// Ideally we want to error for empty filters, but until next version update we resolve this by gracefully handling the issue:
	// we will re-inject empty filters from the user input
	mrcs = append(mrcs, emptyFilters...)
	state.MonitoredResourceConfigs, _ = types.SetValueFrom(ctx, MonitoredResourceConfigSpec, mrcs)
}

func (r *integrationGcpResource) getGCPIntegration(state integrationGcpModel) (*datadogV1.GCPAccount, error) {
	resp, _, err := r.api.ListGCPIntegration(r.auth)
	if err != nil {
		return nil, err
	}

	for _, integration := range resp {
		if integration.GetProjectId() == state.ProjectID.ValueString() && integration.GetClientEmail() == state.ClientEmail.ValueString() {
			if err := utils.CheckForUnparsed(integration); err != nil {
				return nil, err
			}
			return &integration, nil
		}
	}

	return nil, nil // Leave handling of how to deal with nil account to the caller
}

func (r *integrationGcpResource) buildIntegrationGcpRequestBodyBase(state integrationGcpModel) *datadogV1.GCPAccount {
	return &datadogV1.GCPAccount{
		ProjectId:   state.ProjectID.ValueStringPointer(),
		ClientEmail: state.ClientEmail.ValueStringPointer(),
	}
}

func (r *integrationGcpResource) addDefaultsToBody(body *datadogV1.GCPAccount, state integrationGcpModel) {
	body.SetType(defaultType)
	body.SetAuthUri(defaultAuthURI)
	body.SetAuthProviderX509CertUrl(defaultAuthProviderX509CertURL)
	body.SetClientX509CertUrl(defaultClientX509CertURLPrefix + state.ClientEmail.ValueString())
	body.SetTokenUri(defaultTokenURI)
}

func (r *integrationGcpResource) addRequiredFieldsToBody(body *datadogV1.GCPAccount, state integrationGcpModel) {
	body.SetClientId(state.ClientId.ValueString())
	body.SetPrivateKey(state.PrivateKey.ValueString())
	body.SetPrivateKeyId(state.PrivateKeyId.ValueString())
}

func (r *integrationGcpResource) addOptionalFieldsToBody(ctx context.Context, body *datadogV1.GCPAccount, state integrationGcpModel) diag.Diagnostics {
	diags := diag.Diagnostics{}
	body.SetAutomute(state.Automute.ValueBool())
	body.SetIsCspmEnabled(state.CspmResourceCollectionEnabled.ValueBool())
	body.SetIsSecurityCommandCenterEnabled(state.IsSecurityCommandCenterEnabled.ValueBool())

	if !state.ResourceCollectionEnabled.IsUnknown() {
		body.SetResourceCollectionEnabled(state.ResourceCollectionEnabled.ValueBool())
	}

	if !state.IsResourceChangeCollectionEnabled.IsUnknown() {
		body.SetIsResourceChangeCollectionEnabled(state.IsResourceChangeCollectionEnabled.ValueBool())
	}

	body.SetHostFilters(state.HostFilters.ValueString())
	body.SetCloudRunRevisionFilters(tfCollectionToSlice[string](ctx, diags, state.CloudRunRevisionFilters))
	mrcs := make([]datadogV1.GCPMonitoredResourceConfig, 0)
	for _, mrc := range tfCollectionToSlice[*MonitoredResourceConfigModel](ctx, diags, state.MonitoredResourceConfigs) {
		mrcs = append(mrcs, datadogV1.GCPMonitoredResourceConfig{
			Type:    ptrTo(datadogV1.GCPMonitoredResourceConfigType(mrc.Type.ValueString())),
			Filters: tfCollectionToSlice[string](ctx, diags, mrc.Filters),
		})
	}
	body.SetMonitoredResourceConfigs(mrcs)

	return diags
}
