package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// integration_name path parameter for the Databricks AMS endpoint:
// /api/v2/web-integrations/databricks/accounts
const databricksIntegrationName = "databricks"

var (
	_ resource.ResourceWithConfigure        = &integrationDatabricksAccountResource{}
	_ resource.ResourceWithImportState      = &integrationDatabricksAccountResource{}
	_ resource.ResourceWithConfigValidators = &integrationDatabricksAccountResource{}
)

var (
	databricksAuthConfigPath = frameworkPath.MatchRoot("auth_config")
	databricksOauthPath      = databricksAuthConfigPath.AtName("oauth")
	databricksPatPath        = databricksAuthConfigPath.AtName("pat")
)

type integrationDatabricksAccountResource struct {
	Api  *datadogV2.WebIntegrationsApi
	Auth context.Context
}

type integrationDatabricksAccountModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	WorkspaceUrl types.String `tfsdk:"workspace_url"`

	DdApiKeyId                 types.String `tfsdk:"dd_api_key_id"`
	DdApiKeySecret             types.String `tfsdk:"dd_api_key_secret"`
	SystemTablesSqlWarehouseId types.String `tfsdk:"system_tables_sql_warehouse_id"`
	ModelServingEndpointName   types.String `tfsdk:"model_serving_endpoint_name"`
	UcVolumePath               types.String `tfsdk:"uc_volume_path"`
	DoCrawlersCron             types.String `tfsdk:"do_crawlers_cron"`

	DjmEnabled                 types.Bool `tfsdk:"djm_enabled"`
	DjmGlobalInitScriptEnabled types.Bool `tfsdk:"djm_global_init_script_enabled"`
	DjmClusterPolicyEnabled    types.Bool `tfsdk:"djm_cluster_policy_enabled"`
	CcmEnabled                 types.Bool `tfsdk:"ccm_enabled"`
	DoEnabled                  types.Bool `tfsdk:"do_enabled"`
	ModelServingMetricsEnabled types.Bool `tfsdk:"model_serving_metrics_enabled"`
	ScriptLogsEnabled          types.Bool `tfsdk:"script_logs_enabled"`
	ScriptGpumEnabled          types.Bool `tfsdk:"script_gpum_enabled"`
	TableLineageEnabled        types.Bool `tfsdk:"table_lineage_enabled"`
	ServerlessJobsEnabled      types.Bool `tfsdk:"serverless_jobs_enabled"`

	AuthConfig                       *databricksAuthConfigModel                       `tfsdk:"auth_config"`
	PrivateActionRunnerConfiguration *databricksPrivateActionRunnerConfigurationModel `tfsdk:"private_action_runner_configuration"`
}

type databricksAuthConfigModel struct {
	Oauth *databricksOauthModel `tfsdk:"oauth"`
	Pat   *databricksPatModel   `tfsdk:"pat"`
}

type databricksOauthModel struct {
	ClientId            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	DatabricksAccountId types.String `tfsdk:"databricks_account_id"`
	AzureTenantId       types.String `tfsdk:"azure_tenant_id"`
}

type databricksPatModel struct {
	Token types.String `tfsdk:"token"`
}

type databricksPrivateActionRunnerConfigurationModel struct {
	ConnectionId types.String `tfsdk:"connection_id"`
	UserUuid     types.String `tfsdk:"user_uuid"`
	SecretPath   types.String `tfsdk:"secret_path"`
}

func NewIntegrationDatabricksAccountResource() resource.Resource {
	return &integrationDatabricksAccountResource{}
}

func (r *integrationDatabricksAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetWebIntegrationsApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationDatabricksAccountResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_databricks_account"
}

func (r *integrationDatabricksAccountResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			databricksOauthPath,
			databricksPatPath,
		),
	}
}

func (r *integrationDatabricksAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	secretStateModifiers := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}

	response.Schema = schema.Schema{
		Description: "Provides a Datadog Databricks integration account resource. " +
			"Manages a Databricks workspace connection used for Data Jobs Monitoring, " +
			"Cloud Cost Management, Data Observability, and related Databricks-driven products.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the account.",
			},
			"workspace_url": schema.StringAttribute{
				Required:    true,
				Description: "The URL of your Databricks workspace (e.g., https://your-workspace.cloud.databricks.com).",
			},
			"dd_api_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "Datadog API Key ID used for the Data Jobs Monitoring init script when managed by Datadog.",
			},
			"dd_api_key_secret": schema.StringAttribute{
				Optional:      true,
				Sensitive:     true,
				Description:   "Datadog API Key value (not ID) used for the Data Jobs Monitoring init script when managed by Datadog. This value is write-only; changes made outside of Terraform will not be drift-detected.",
				PlanModifiers: secretStateModifiers,
			},
			"system_tables_sql_warehouse_id": schema.StringAttribute{
				Optional:    true,
				Description: "SQL Warehouse ID for querying Databricks System Tables. Required for Cloud Cost Management.",
			},
			"model_serving_endpoint_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the Databricks model serving endpoint to monitor.",
			},
			"uc_volume_path": schema.StringAttribute{
				Optional: true,
				Description: "Unity Catalog volume path in `catalog.schema.volume` format where the Datadog init script will be stored " +
					"(e.g. `main.default.datadog_volume`). Required when `djm_cluster_policy_enabled` is true.",
			},
			"do_crawlers_cron": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0 * * * *"),
				Description: "Cron schedule controlling how often Datadog crawls the Databricks warehouse for metadata. Defaults to hourly.",
			},
			"djm_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Enable Data Jobs Monitoring for this workspace. Defaults to true.",
			},
			"djm_global_init_script_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Description: "When enabled, Datadog installs and manages the Agent with a global init script in the workspace. " +
					"Installation can take up to 15 minutes. Requires Workspace Admin permissions.",
			},
			"djm_cluster_policy_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Description: "When enabled, Datadog installs and manages the Agent using a cluster policy and Unity Catalog Volume. " +
					"Requires a Unity Catalog-enabled workspace with DBR 13.3 LTS+ and `uc_volume_path`.",
			},
			"ccm_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enable Cloud Cost Management to collect cost data from Databricks System Tables. Requires `system_tables_sql_warehouse_id`.",
			},
			"do_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enable Data Observability to collect data for viewing in Datadog Data Observability.",
			},
			"model_serving_metrics_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Retrieve health and usage metrics from Databricks model serving endpoints.",
			},
			"script_logs_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Collect driver and worker logs from Databricks clusters when using a Datadog-managed init script.",
			},
			"script_gpum_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Collect GPU metrics from Databricks clusters when using a Datadog-managed init script.",
			},
			"table_lineage_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enable table lineage tracking for Databricks tables.",
			},
			"serverless_jobs_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Serverless opt-in for Data Jobs Monitoring. Defaults to true.",
			},
		},
		Blocks: map[string]schema.Block{
			"auth_config": schema.SingleNestedBlock{
				Description: "Configure how Datadog authenticates to your Databricks workspace. " +
					"Exactly one of `oauth` or `pat` must be provided.",
				Blocks: map[string]schema.Block{
					"oauth": schema.SingleNestedBlock{
						Description: "OAuth (service principal) authentication. Recommended for new deployments.",
						Attributes: map[string]schema.Attribute{
							"client_id": schema.StringAttribute{
								Required:    true,
								Description: "OAuth Client ID for the Databricks service principal.",
							},
							"client_secret": schema.StringAttribute{
								Required:      true,
								Sensitive:     true,
								Description:   "OAuth Client Secret for the Databricks service principal. This value is write-only; changes made outside of Terraform will not be drift-detected.",
								PlanModifiers: secretStateModifiers,
							},
							"databricks_account_id": schema.StringAttribute{
								Required:    true,
								Description: "Databricks Account ID (UUID format). Found in your Databricks profile in the upper-right corner.",
							},
							"azure_tenant_id": schema.StringAttribute{
								Optional:    true,
								Description: "Azure Tenant ID (UUID format) for authenticating via Microsoft Entra ID. Only set when using Azure Entra ID OAuth.",
							},
						},
					},
					"pat": schema.SingleNestedBlock{
						Description: "Personal Access Token authentication. Deprecated in favor of `oauth`; kept for backwards compatibility.",
						Attributes: map[string]schema.Attribute{
							"token": schema.StringAttribute{
								Required:      true,
								Sensitive:     true,
								Description:   "Databricks Personal Access Token (PAT). Generate from Settings > Developer > Access tokens. This value is write-only; changes made outside of Terraform will not be drift-detected.",
								PlanModifiers: secretStateModifiers,
							},
						},
					},
				},
			},
			"private_action_runner_configuration": schema.SingleNestedBlock{
				Description: "Run Datadog crawlers behind a Private Action Runner instead of from Datadog's network.",
				Attributes: map[string]schema.Attribute{
					"connection_id": schema.StringAttribute{
						Optional:    true,
						Description: "Private Action Runner connection ID.",
					},
					"user_uuid": schema.StringAttribute{
						Optional:    true,
						Description: "Service Account UUID used to execute Private Action Runner actions.",
					},
					"secret_path": schema.StringAttribute{
						Optional:    true,
						Description: "Path to the stored secret holding Databricks credentials inside the Private Action Runner.",
					},
				},
			},
		},
	}
}

func (r *integrationDatabricksAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationDatabricksAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationDatabricksAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetWebIntegrationAccount(r.Auth, databricksIntegrationName, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Databricks integration account"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationDatabricksAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan integrationDatabricksAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildCreateRequestBody(&plan)
	resp, _, err := r.Api.CreateWebIntegrationAccount(r.Auth, databricksIntegrationName, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Databricks integration account"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&plan, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *integrationDatabricksAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan integrationDatabricksAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	body := r.buildUpdateRequestBody(&plan)
	resp, _, err := r.Api.UpdateWebIntegrationAccount(r.Auth, databricksIntegrationName, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Databricks integration account"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&plan, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *integrationDatabricksAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationDatabricksAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	httpResp, err := r.Api.DeleteWebIntegrationAccount(r.Auth, databricksIntegrationName, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting Databricks integration account"))
	}
}

// updateState writes name + settings from the API response into TF state.
// Secrets are never returned by the API and must NOT be touched here — the
// framework preserves any prior state for fields we leave alone.
func (r *integrationDatabricksAccountResource) updateState(state *integrationDatabricksAccountModel, resp *datadogV2.WebIntegrationAccountResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	attrs := data.GetAttributes()
	state.Name = types.StringValue(attrs.GetName())

	settings := attrs.GetSettings()
	state.WorkspaceUrl = settingString(settings, "workspace_url")

	state.DdApiKeyId = settingStringOrNull(state.DdApiKeyId, settings, "dd_api_key_id")
	state.SystemTablesSqlWarehouseId = settingStringOrNull(state.SystemTablesSqlWarehouseId, settings, "system_tables_sql_warehouse_id")
	state.ModelServingEndpointName = settingStringOrNull(state.ModelServingEndpointName, settings, "model_serving_endpoint_name")
	state.UcVolumePath = settingStringOrNull(state.UcVolumePath, settings, "uc_volume_path")
	state.DoCrawlersCron = settingStringDefault(settings, "do_crawlers_cron", "0 * * * *")

	state.DjmEnabled = settingBoolDefault(settings, "djm_enabled", true)
	state.DjmGlobalInitScriptEnabled = settingBoolDefault(settings, "djm_global_init_script_enabled", false)
	state.DjmClusterPolicyEnabled = settingBoolDefault(settings, "djm_cluster_policy_enabled", false)
	state.CcmEnabled = settingBoolDefault(settings, "ccm_enabled", false)
	state.DoEnabled = settingBoolDefault(settings, "do_enabled", false)
	state.ModelServingMetricsEnabled = settingBoolDefault(settings, "model_serving_metrics_enabled", false)
	state.ScriptLogsEnabled = settingBoolDefault(settings, "script_logs_enabled", false)
	state.ScriptGpumEnabled = settingBoolDefault(settings, "script_gpum_enabled", false)
	state.TableLineageEnabled = settingBoolDefault(settings, "table_lineage_enabled", false)
	state.ServerlessJobsEnabled = settingBoolDefault(settings, "serverless_jobs_enabled", true)

	// OAuth settings live in the wire `settings` object but in the TF schema they
	// nest under `auth_config.oauth`. Reshuffle here on Read so a user importing
	// (or refreshing) an OAuth-configured account sees the right structure.
	if state.AuthConfig != nil && state.AuthConfig.Oauth != nil {
		state.AuthConfig.Oauth.ClientId = settingStringOrNull(state.AuthConfig.Oauth.ClientId, settings, "client_id")
		state.AuthConfig.Oauth.DatabricksAccountId = settingStringOrNull(state.AuthConfig.Oauth.DatabricksAccountId, settings, "databricks_account_id")
		state.AuthConfig.Oauth.AzureTenantId = settingStringOrNull(state.AuthConfig.Oauth.AzureTenantId, settings, "azure_tenant_id")
	}

	if parc, ok := settings["private_action_runner_configuration"].(map[string]interface{}); ok {
		if state.PrivateActionRunnerConfiguration == nil {
			state.PrivateActionRunnerConfiguration = &databricksPrivateActionRunnerConfigurationModel{}
		}
		state.PrivateActionRunnerConfiguration.ConnectionId = settingStringOrNull(state.PrivateActionRunnerConfiguration.ConnectionId, parc, "connection_id")
		state.PrivateActionRunnerConfiguration.UserUuid = settingStringOrNull(state.PrivateActionRunnerConfiguration.UserUuid, parc, "user_uuid")
		state.PrivateActionRunnerConfiguration.SecretPath = settingStringOrNull(state.PrivateActionRunnerConfiguration.SecretPath, parc, "secret_path")
	}
}

func (r *integrationDatabricksAccountResource) buildCreateRequestBody(plan *integrationDatabricksAccountModel) *datadogV2.WebIntegrationAccountCreateRequest {
	attrs := datadogV2.WebIntegrationAccountCreateRequestAttributes{
		Name:     plan.Name.ValueString(),
		Settings: r.buildSettings(plan),
		Secrets:  r.buildSecrets(plan),
	}

	data := datadogV2.NewWebIntegrationAccountCreateRequestDataWithDefaults()
	data.SetAttributes(attrs)

	req := datadogV2.NewWebIntegrationAccountCreateRequestWithDefaults()
	req.SetData(*data)
	return req
}

func (r *integrationDatabricksAccountResource) buildUpdateRequestBody(plan *integrationDatabricksAccountModel) *datadogV2.WebIntegrationAccountUpdateRequest {
	attrs := datadogV2.NewWebIntegrationAccountUpdateRequestAttributesWithDefaults()
	attrs.SetName(plan.Name.ValueString())
	attrs.SetSettings(r.buildSettings(plan))
	attrs.SetSecrets(r.buildSecrets(plan))

	data := datadogV2.NewWebIntegrationAccountUpdateRequestDataWithDefaults()
	data.SetAttributes(*attrs)

	req := datadogV2.NewWebIntegrationAccountUpdateRequestWithDefaults()
	req.SetData(*data)
	return req
}

// buildSettings flattens the TF schema (Option B: OAuth identifiers nested
// under auth_config) into the wire `settings` map expected by the AMS API.
// Server schema has `additionalProperties: false`, so only known keys are sent.
func (r *integrationDatabricksAccountResource) buildSettings(plan *integrationDatabricksAccountModel) map[string]interface{} {
	settings := map[string]interface{}{
		"workspace_url": plan.WorkspaceUrl.ValueString(),

		"djm_enabled":                    plan.DjmEnabled.ValueBool(),
		"djm_global_init_script_enabled": plan.DjmGlobalInitScriptEnabled.ValueBool(),
		"djm_cluster_policy_enabled":     plan.DjmClusterPolicyEnabled.ValueBool(),
		"ccm_enabled":                    plan.CcmEnabled.ValueBool(),
		"do_enabled":                     plan.DoEnabled.ValueBool(),
		"model_serving_metrics_enabled":  plan.ModelServingMetricsEnabled.ValueBool(),
		"script_logs_enabled":            plan.ScriptLogsEnabled.ValueBool(),
		"script_gpum_enabled":            plan.ScriptGpumEnabled.ValueBool(),
		"table_lineage_enabled":          plan.TableLineageEnabled.ValueBool(),
		"serverless_jobs_enabled":        plan.ServerlessJobsEnabled.ValueBool(),
		"do_crawlers_cron":               plan.DoCrawlersCron.ValueString(),
	}

	setIfKnown(settings, "dd_api_key_id", plan.DdApiKeyId)
	setIfKnown(settings, "system_tables_sql_warehouse_id", plan.SystemTablesSqlWarehouseId)
	setIfKnown(settings, "model_serving_endpoint_name", plan.ModelServingEndpointName)
	setIfKnown(settings, "uc_volume_path", plan.UcVolumePath)

	if plan.AuthConfig != nil && plan.AuthConfig.Oauth != nil {
		oauth := plan.AuthConfig.Oauth
		setIfKnown(settings, "client_id", oauth.ClientId)
		setIfKnown(settings, "databricks_account_id", oauth.DatabricksAccountId)
		setIfKnown(settings, "azure_tenant_id", oauth.AzureTenantId)
	}

	if parc := plan.PrivateActionRunnerConfiguration; parc != nil {
		nested := map[string]interface{}{}
		setIfKnown(nested, "connection_id", parc.ConnectionId)
		setIfKnown(nested, "user_uuid", parc.UserUuid)
		setIfKnown(nested, "secret_path", parc.SecretPath)
		if len(nested) > 0 {
			settings["private_action_runner_configuration"] = nested
		}
	}

	return settings
}

func (r *integrationDatabricksAccountResource) buildSecrets(plan *integrationDatabricksAccountModel) map[string]interface{} {
	secrets := map[string]interface{}{}

	if plan.AuthConfig != nil {
		if plan.AuthConfig.Oauth != nil {
			setIfKnown(secrets, "client_secret", plan.AuthConfig.Oauth.ClientSecret)
		}
		if plan.AuthConfig.Pat != nil {
			setIfKnown(secrets, "token", plan.AuthConfig.Pat.Token)
		}
	}
	setIfKnown(secrets, "dd_api_key_secret", plan.DdApiKeySecret)

	return secrets
}

// Helpers for safe-casting settings map values returned by the API.
func settingString(settings map[string]interface{}, key string) types.String {
	if v, ok := settings[key].(string); ok {
		return types.StringValue(v)
	}
	return types.StringNull()
}

func settingStringOrNull(prior types.String, settings map[string]interface{}, key string) types.String {
	if v, ok := settings[key].(string); ok && v != "" {
		return types.StringValue(v)
	}
	// API omitted the field: keep prior state if it had a value, otherwise null.
	if !prior.IsNull() && !prior.IsUnknown() && prior.ValueString() != "" {
		return prior
	}
	return types.StringNull()
}

func settingStringDefault(settings map[string]interface{}, key, def string) types.String {
	if v, ok := settings[key].(string); ok && v != "" {
		return types.StringValue(v)
	}
	return types.StringValue(def)
}

func settingBoolDefault(settings map[string]interface{}, key string, def bool) types.Bool {
	if v, ok := settings[key].(bool); ok {
		return types.BoolValue(v)
	}
	return types.BoolValue(def)
}

// setIfKnown writes the TF value into the wire map only when the user actually
// set it. Keeps the request payload clean — important because the AMS schema
// is `additionalProperties: false` AND the server deep-merges PATCH bodies,
// so sending an explicit "" would overwrite the existing value with empty.
func setIfKnown(dst map[string]interface{}, key string, v types.String) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	dst[key] = v.ValueString()
}
