package fwprovider

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	datadogCommunity "github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/fwutils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ provider.Provider = &FrameworkProvider{}

var Resources = []func() resource.Resource{
	NewAgentlessScanningAwsScanOptionsResource,
	NewAgentlessScanningGcpScanOptionsResource,
	NewOpenapiApiResource,
	NewAPIKeyResource,
	NewApplicationKeyResource,
	NewApmRetentionFilterResource,
	NewApmRetentionFiltersOrderResource,
	NewIntegrationAwsAccountResource,
	NewCatalogEntityResource,
	NewDashboardListResource,
	NewDatasetResource,
	NewDomainAllowlistResource,
	NewDowntimeScheduleResource,
	NewIntegrationAzureResource,
	NewIntegrationAwsEventBridgeResource,
	NewIntegrationAwsExternalIDResource,
	NewIntegrationCloudflareAccountResource,
	NewIntegrationConfluentAccountResource,
	NewIntegrationConfluentResourceResource,
	NewIntegrationFastlyAccountResource,
	NewIntegrationFastlyServiceResource,
	NewIntegrationGcpResource,
	NewIntegrationGcpStsResource,
	NewCloudInventorySyncConfigResource,
	NewIpAllowListResource,
	NewMonitorNotificationRuleResource,
	NewSecurityNotificationRuleResource,
	NewRestrictionPolicyResource,
	NewRumApplicationResource,
	NewRumMetricResource,
	NewRumRetentionFilterResource,
	NewRumRetentionFiltersOrderResource,
	NewSensitiveDataScannerGroupOrder,
	NewServiceAccountApplicationKeyResource,
	NewSpansMetricResource,
	NewSyntheticsConcurrencyCapResource,
	NewSyntheticsGlobalVariableResource,
	NewSyntheticsPrivateLocationResource,
	NewSyntheticsSuiteResource,
	NewTeamLinkResource,
	NewTeamMembershipResource,
	NewTeamNotificationRuleResource,
	NewTeamPermissionSettingResource,
	NewTeamResource,
	NewTeamHierarchyLinksResource,
	NewUserRoleResource,
	NewSecurityMonitoringSuppressionResource,
	NewSecurityMonitoringCriticalAssetResource,
	NewServiceAccountResource,
	NewWebhookResource,
	NewWebhookCustomVariableResource,
	NewLogsCustomDestinationResource,
	NewLogsRestrictionQueryResource,
	NewTenantBasedHandleResource,
	NewAppsecWafExclusionFilterResource,
	NewAppsecWafCustomRuleResource,
	NewWorkflowsWebhookHandleResource,
	NewActionConnectionResource,
	NewWorkflowAutomationResource,
	NewAppBuilderAppResource,
	NewObservabilitPipelineResource,
	NewOnCallEscalationPolicyResource,
	NewOnCallScheduleResource,
	NewOnCallTeamRoutingRulesResource,
	NewOnCallUserNotificationChannelResource,
	NewOnCallUserNotificationRuleResource,
	NewOrgConnectionResource,
	NewComplianceResourceEvaluationFilter,
	NewSecurityMonitoringRuleJSONResource,
	NewComplianceCustomFrameworkResource,
	NewCostBudgetResource,
	NewTagPipelineRulesetResource,
	NewTagPipelineRulesetsResource,
	NewCSMThreatsAgentRuleResource,
	NewCSMThreatsPolicyResource,
	NewAppKeyRegistrationResource,
	NewIncidentTypeResource,
	NewIncidentNotificationTemplateResource,
	NewIncidentNotificationRuleResource,
	NewAwsCurConfigResource,
	NewGcpUcConfigResource,
	NewDatadogCustomAllocationRuleResource,
	NewCustomAllocationRulesResource,
	NewAzureUcConfigResource,
	NewDeploymentGateResource,
	NewReferenceTableResource,
	NewDatastoreResource,
	NewDatastoreItemResource,
}

var Datasources = []func() datasource.DataSource{
	NewAPIKeyDataSource,
	NewApplicationKeyDataSource,
	NewAwsAvailableNamespacesDataSource,
	NewAwsIntegrationExternalIDDataSource,
	NewAwsIntegrationIAMPermissionsDataSource,
	NewAwsIntegrationIAMPermissionsStandardDataSource,
	NewAwsIntegrationIAMPermissionsResourceCollectionDataSource,
	NewAwsLogsServicesDataSource,
	NewDatadogApmRetentionFiltersOrderDataSource,
	NewDatadogDashboardListDataSource,
	NewDatadogIntegrationAWSNamespaceRulesDatasource,
	NewDatadogMetricActiveTagsAndAggregationsDataSource,
	NewDatadogMetricMetadataDataSource,
	NewDatadogMetricTagsDataSource,
	NewDatadogMetricsDataSource,
	NewDatadogPowerpackDataSource,
	NewDatadogServiceAccountDatasource,
	NewDatadogSoftwareCatalogDataSource,
	NewDatadogTeamDataSource,
	NewDatadogTeamHierarchyLinksDataSource,
	NewDatadogTeamMembershipsDataSource,
	NewDatadogTeamNotificationRuleDataSource,
	NewDatadogTeamNotificationRulesDataSource,
	NewHostsDataSource,
	NewIPRangesDataSource,
	NewRumApplicationDataSource,
	NewRumRetentionFiltersDataSource,
	NewSensitiveDataScannerGroupOrderDatasource,
	NewDatadogUsersDataSource,
	NewDatadogRoleUsersDataSource,
	NewSecurityMonitoringSuppressionDataSource,
	NewSecurityMonitoringCriticalAssetDataSource,
	NewSecurityMonitoringCriticalAssetsDataSource,
	NewLogsPipelinesOrderDataSource,
	NewDatadogTeamsDataSource,
	NewDatadogActionConnectionDataSource,
	NewDatadogSyntheticsGlobalVariableDataSource,
	NewDatadogSyntheticsLocationsDataSource,
	NewWorkflowAutomationDataSource,
	NewDatadogAppBuilderAppDataSource,
	NewCostBudgetDataSource,
	NewTagPipelineRulesetDataSource,
	NewCSMThreatsAgentRulesDataSource,
	NewCSMThreatsPoliciesDataSource,
	NewIncidentTypeDataSource,
	NewIncidentNotificationTemplateDataSource,
	NewIncidentNotificationRuleDataSource,
	NewDatadogAwsCurConfigDataSource,
	NewDatadogGcpUcConfigDataSource,
	NewDatadogCustomAllocationRuleDataSource,
	NewDatadogAzureUcConfigDataSource,
	NewDatadogReferenceTableDataSource,
	NewDatadogReferenceTableRowsDataSource,
	NewOrganizationSettingsDataSource,
	NewDatadogDatastoreDataSource,
	NewDatastoreItemDataSource,
}

// FrameworkProvider struct
type FrameworkProvider struct {
	CommunityClient     *datadogCommunity.Client
	DatadogApiInstances *utils.ApiInstances
	Auth                context.Context

	ConfigureCallbackFunc func(p *FrameworkProvider, request *provider.ConfigureRequest, config *ProviderSchema) diag.Diagnostics
	Now                   func() time.Time
	DefaultTags           map[string]string
}

// ProviderSchema struct
type ProviderSchema struct {
	ApiKey                           types.String `tfsdk:"api_key"`
	AppKey                           types.String `tfsdk:"app_key"`
	ApiUrl                           types.String `tfsdk:"api_url"`
	Validate                         types.String `tfsdk:"validate"`
	CloudProviderType                types.String `tfsdk:"cloud_provider_type"`
	CloudProviderRegion              types.String `tfsdk:"cloud_provider_region"`
	OrgUuid                          types.String `tfsdk:"org_uuid"`
	AWSAccessKeyId                   types.String `tfsdk:"aws_access_key_id"`
	AWSSecretAccessKey               types.String `tfsdk:"aws_secret_access_key"`
	AWSSessionToken                  types.String `tfsdk:"aws_session_token"`
	HttpClientRetryEnabled           types.String `tfsdk:"http_client_retry_enabled"`
	HttpClientRetryTimeout           types.Int64  `tfsdk:"http_client_retry_timeout"`
	HttpClientRetryBackoffMultiplier types.Int64  `tfsdk:"http_client_retry_backoff_multiplier"`
	HttpClientRetryBackoffBase       types.Int64  `tfsdk:"http_client_retry_backoff_base"`
	HttpClientRetryMaxRetries        types.Int64  `tfsdk:"http_client_retry_max_retries"`
	DefaultTags                      []DefaultTag `tfsdk:"default_tags"`
}

type DefaultTag struct {
	Tags types.Map `tfsdk:"tags"`
}

func New() provider.Provider {
	return &FrameworkProvider{
		ConfigureCallbackFunc: defaultConfigureFunc,
	}
}

func (p *FrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	var wrappedResources []func() resource.Resource
	for _, f := range Resources {
		r := f()
		wrappedResources = append(wrappedResources, func() resource.Resource { return NewFrameworkResourceWrapper(&r) })
	}

	if utils.UseMonitorFrameworkProvider() {
		monitorResource := NewMonitorResource()
		wrappedResources = append(wrappedResources, func() resource.Resource { return NewFrameworkResourceWrapper(&monitorResource) })
	}

	return wrappedResources
}

func (p *FrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	var wrappedDatasources []func() datasource.DataSource
	for _, f := range Datasources {
		r := f()
		wrappedDatasources = append(wrappedDatasources, func() datasource.DataSource { return NewFrameworkDatasourceWrapper(&r) })
	}

	return wrappedDatasources
}

func (p *FrameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "datadog_"
}

func (p *FrameworkProvider) MetaSchema(_ context.Context, _ provider.MetaSchemaRequest, _ *provider.MetaSchemaResponse) {
}

func (p *FrameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "(Required unless validate is false) Datadog API key. This can also be set via the DD_API_KEY environment variable.",
			},
			"app_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "(Required unless validate is false) Datadog APP key. This can also be set via the DD_APP_KEY environment variable.",
			},
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "The API URL. This can also be set via the DD_HOST environment variable, and defaults to `https://api.datadoghq.com`. Note that this URL must not end with the `/api/` path. For example, `https://api.datadoghq.com/` is a correct value, while `https://api.datadoghq.com/api/` is not. And if you're working with \"EU\" version of Datadog, use `https://api.datadoghq.eu/`. Other Datadog region examples: `https://api.us5.datadoghq.com/`, `https://api.us3.datadoghq.com/` and `https://api.ddog-gov.com/`. See https://docs.datadoghq.com/getting_started/site/ for all available regions.",
			},
			"validate": schema.StringAttribute{
				Optional:    true,
				Description: "Enables validation of the provided API key during provider initialization. Valid values are [`true`, `false`]. Default is true. When false, api_key won't be checked.",
			},
			"cloud_provider_type": schema.StringAttribute{
				Optional:    true,
				Description: "Specifies the cloud provider used for cloud-provider-based authentication, enabling keyless access without API or app keys. Only [`aws`] is supported. This feature is in Preview. If you'd like to enable it for your organization, contact [support](https://docs.datadoghq.com/help/).",
			},
			"cloud_provider_region": schema.StringAttribute{
				Optional:    true,
				Description: "The cloud provider region specifier; used for cloud-provider-based authentication. For example, `us-east-1` for AWS.",
			},
			"org_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "The organization UUID; used for cloud-provider-based authentication. See the [Datadog API documentation](https://docs.datadoghq.com/api/v1/organizations/) for more information.",
			},
			"aws_access_key_id": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS access key ID; used for cloud-provider-based authentication. This can also be set using the `AWS_ACCESS_KEY_ID` environment variable. Required when using `cloud_provider_type` set to `aws`.",
			},
			"aws_secret_access_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS secret access key; used for cloud-provider-based authentication. This can also be set using the `AWS_SECRET_ACCESS_KEY` environment variable. Required when using `cloud_provider_type` set to `aws`.",
			},
			"aws_session_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS session token; used for cloud-provider-based authentication. This can also be set using the `AWS_SESSION_TOKEN` environment variable. Required when using `cloud_provider_type` set to `aws` and using temporary credentials.",
			},
			"http_client_retry_enabled": schema.StringAttribute{
				Optional:    true,
				Description: "Enables request retries on HTTP status codes 429 and 5xx. Valid values are [`true`, `false`]. Defaults to `true`.",
			},
			"http_client_retry_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request retry timeout period. Defaults to 60 seconds.",
			},
			"http_client_retry_backoff_multiplier": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request retry back off multiplier. Defaults to 2.",
			},
			"http_client_retry_backoff_base": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request retry back off base. Defaults to 2.",
			},
			"http_client_retry_max_retries": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request maximum retry number. Defaults to 3.",
			},
		},
		Blocks: map[string]schema.Block{
			"default_tags": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "[Experimental - Logs Pipelines, Monitors, Security Monitoring Rules, and Service Level Objectives only] Configuration block containing settings to apply default resource tags across all resources.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tags": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "[Experimental - Logs Pipelines, Monitors, Security Monitoring Rules, and Service Level Objectives only] Resource tags to be applied by default across all resources.",
						},
					},
				},
			},
		},
	}
}

func (p *FrameworkProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config ProviderSchema
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	diags := p.ConfigureConfigDefaults(ctx, &config)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
	}

	response.Diagnostics.Append(p.ConfigureCallbackFunc(p, &request, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Make config available for data sources and resources
	response.DataSourceData = p
	response.ResourceData = p
}

func (p *FrameworkProvider) ConfigureConfigDefaults(ctx context.Context, config *ProviderSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	if config.ApiKey.IsNull() {
		apiKey, err := utils.GetMultiEnvVar(utils.APIKeyEnvVars[:]...)
		if err == nil {
			config.ApiKey = types.StringValue(apiKey)
		}
	}

	if config.AppKey.IsNull() {
		appKey, err := utils.GetMultiEnvVar(utils.APPKeyEnvVars[:]...)
		if err == nil {
			config.AppKey = types.StringValue(appKey)
		}
	}

	if config.ApiUrl.IsNull() {
		apiUrl, err := utils.GetMultiEnvVar(utils.APIUrlEnvVars[:]...)
		if err == nil {
			config.ApiUrl = types.StringValue(apiUrl)
		}
	}

	if config.OrgUuid.IsNull() {
		orgUUID, err := utils.GetMultiEnvVar(utils.OrgUUIDEnvVars[:]...)
		if err == nil {
			config.OrgUuid = types.StringValue(orgUUID)
		}
	}
	if config.AWSAccessKeyId.IsNull() {
		awsAccessKeyId, err := utils.GetMultiEnvVar(utils.AWSAccessKeyId)
		if err == nil {
			config.AWSAccessKeyId = types.StringValue(awsAccessKeyId)
		}
	}
	if config.AWSSecretAccessKey.IsNull() {
		awsSecretAccessKey, err := utils.GetMultiEnvVar(utils.AWSSecretAccessKey)
		if err == nil {
			config.AWSSecretAccessKey = types.StringValue(awsSecretAccessKey)
		}
	}
	if config.AWSSessionToken.IsNull() {
		awsSessionToken, err := utils.GetMultiEnvVar(utils.AWSSessionToken)
		if err == nil {
			config.AWSSessionToken = types.StringValue(awsSessionToken)
		}
	}

	if config.HttpClientRetryEnabled.IsNull() {
		retryEnabled, err := utils.GetMultiEnvVar(utils.DDHTTPRetryEnabled)
		if err == nil {
			config.HttpClientRetryEnabled = types.StringValue(retryEnabled)
		}
	}

	if config.HttpClientRetryTimeout.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar(utils.DDHTTPRetryTimeout)
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryTimeout = types.Int64Value(int64(v))
		}
	}

	if config.HttpClientRetryBackoffMultiplier.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar(utils.DDHTTPRetryBackoffMultiplier)
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryBackoffMultiplier = types.Int64Value(int64(v))
		}
	}

	if config.HttpClientRetryBackoffBase.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar(utils.DDHTTPRetryBackoffBase)
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryBackoffBase = types.Int64Value(int64(v))
		}
	}

	if config.HttpClientRetryMaxRetries.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar(utils.DDHTTPRetryMaxRetries)
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryMaxRetries = types.Int64Value(int64(v))
		}
	}

	// Configure defaults for booleans.
	// Remove this once fully migrated to framework
	if config.Validate.IsNull() {
		config.Validate = types.StringValue("true")
	}
	if config.HttpClientRetryEnabled.IsNull() {
		config.HttpClientRetryEnabled = types.StringValue("true")
	}

	// Run validations on the provider config after defaults and values from
	// env var has been set.
	diags.Append(p.ValidateConfigValues(ctx, config)...)

	return diags
}

func (p *FrameworkProvider) ValidateConfigValues(ctx context.Context, config *ProviderSchema) diag.Diagnostics {
	var diags diag.Diagnostics
	// Init validators we need for purposes of config validation only
	oneOfStringValidator := stringvalidator.OneOf("true", "false")
	int64AtLeastValidator := int64validator.AtLeast(1)
	int64BetweenValidator := int64validator.Between(1, 5)

	if !config.Validate.IsNull() {
		res := validator.StringResponse{}
		oneOfStringValidator.ValidateString(ctx, validator.StringRequest{ConfigValue: config.Validate}, &res)
		diags.Append(res.Diagnostics...)
	}

	if !config.HttpClientRetryEnabled.IsNull() {
		res := validator.StringResponse{}
		oneOfStringValidator.ValidateString(ctx, validator.StringRequest{ConfigValue: config.HttpClientRetryEnabled}, &res)
		diags.Append(res.Diagnostics...)
	}

	if !config.HttpClientRetryBackoffMultiplier.IsNull() {
		res := validator.Int64Response{}
		int64AtLeastValidator.ValidateInt64(ctx, validator.Int64Request{ConfigValue: config.HttpClientRetryBackoffMultiplier}, &res)
		diags.Append(res.Diagnostics...)
	}

	if !config.HttpClientRetryBackoffBase.IsNull() {
		res := validator.Int64Response{}
		int64AtLeastValidator.ValidateInt64(ctx, validator.Int64Request{ConfigValue: config.HttpClientRetryBackoffBase}, &res)
		diags.Append(res.Diagnostics...)
	}

	if !config.HttpClientRetryMaxRetries.IsNull() {
		res := validator.Int64Response{}
		int64BetweenValidator.ValidateInt64(ctx, validator.Int64Request{ConfigValue: config.HttpClientRetryMaxRetries}, &res)
		diags.Append(res.Diagnostics...)
	}

	return diags
}

// Helper method to configure the provider
func defaultConfigureFunc(p *FrameworkProvider, request *provider.ConfigureRequest, config *ProviderSchema) diag.Diagnostics {
	diags := diag.Diagnostics{}
	validate, _ := strconv.ParseBool(config.Validate.ValueString())
	httpClientRetryEnabled, _ := strconv.ParseBool(config.HttpClientRetryEnabled.ValueString())

	cloudProviderType := config.CloudProviderType.ValueString()
	cloudProviderRegion := config.CloudProviderRegion.ValueString()
	orgUUID := config.OrgUuid.ValueString()
	awsAccessKeyId := config.AWSAccessKeyId.ValueString()
	awsSecretAccessKey := config.AWSSecretAccessKey.ValueString()
	awsSessionToken := config.AWSSessionToken.ValueString()

	if validate {
		if cloudProviderType == "" && (config.ApiKey.ValueString() == "" || config.AppKey.ValueString() == "") {
			diags.AddError("api_key and app_key or orgUUID must be set unless validate = false", "")
			return diags
		} else if cloudProviderType != "" && orgUUID == "" {
			diags.AddError("orgUUID must be set when using cloud provider auth unless validate = false", "")
			return diags
		}
	}

	// Initialize the community client
	p.CommunityClient = datadogCommunity.NewClient(config.ApiKey.ValueString(), config.AppKey.ValueString())
	if !config.ApiUrl.IsNull() && config.ApiUrl.ValueString() != "" {
		p.CommunityClient.SetBaseUrl(config.ApiUrl.ValueString())
	}
	c := cleanhttp.DefaultClient()
	p.CommunityClient.ExtraHeader["User-Agent"] = utils.GetUserAgentFramework(fmt.Sprintf(
		"datadog-api-client-go/%s (go %s; os %s; arch %s)",
		"go-datadog-api",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	), request.TerraformVersion)
	p.CommunityClient.HttpClient = c

	// Initialize the official Datadog V1 API client
	auth := context.Background()
	// Check cloud_provider_type first - explicit config takes precedence over API keys
	if cloudProviderType != "" {
		// Allows for delegated token authentication
		auth = context.WithValue(
			auth,
			datadog.ContextDelegatedToken,
			&datadog.DelegatedTokenCredentials{},
		)
		switch cloudProviderType {
		case "aws":
			auth = context.WithValue(
				auth,
				datadog.ContextAWSVariables,
				map[string]string{
					datadog.AWSAccessKeyIdName:     awsAccessKeyId,
					datadog.AWSSecretAccessKeyName: awsSecretAccessKey,
					datadog.AWSSessionTokenName:    awsSessionToken,
				},
			)
		default:
			diags.AddError("cloud_provider_type must be set to a valid value unless validate = false", "")
			return diags
		}
	} else if config.ApiKey.ValueString() != "" || config.AppKey.ValueString() != "" {
		auth = context.WithValue(
			auth,
			datadog.ContextAPIKeys,
			map[string]datadog.APIKey{
				"apiKeyAuth": {
					Key: config.ApiKey.ValueString(),
				},
				"appKeyAuth": {
					Key: config.AppKey.ValueString(),
				},
			},
		)
	}
	ddClientConfig := datadog.NewConfiguration()
	ddClientConfig.UserAgent = utils.GetUserAgentFramework(ddClientConfig.UserAgent, request.TerraformVersion)
	ddClientConfig.Debug = logging.IsDebugOrHigher()

	ddClientConfig.SetUnstableOperationEnabled("v2.CreateOpenAPI", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateOpenAPI", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetOpenAPI", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteOpenAPI", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.ListAWSLogsServices", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetDataset", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateDataset", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateDataset", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteDataset", true)

	// Enable Logs Restriction Queries
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateRestrictionQuery", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetRestrictionQuery", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateRestrictionQuery", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteRestrictionQuery", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.AddRoleToRestrictionQuery", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.RemoveRoleFromRestrictionQuery", true)

	// Enable Observability Pipelines
	ddClientConfig.SetUnstableOperationEnabled("v2.CreatePipeline", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetPipeline", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdatePipeline", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeletePipeline", true)

	// Enable MonitorNotificationRule
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateMonitorNotificationRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetMonitorNotificationRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteMonitorNotificationRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateMonitorNotificationRule", true)

	// Enable IncidentType
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateIncidentType", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetIncidentType", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateIncidentType", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteIncidentType", true)

	// Enable IncidentNotificationTemplate
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateIncidentNotificationTemplate", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetIncidentNotificationTemplate", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateIncidentNotificationTemplate", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteIncidentNotificationTemplate", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.ListIncidentNotificationTemplates", true)

	// Enable IncidentNotificationRule
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateIncidentNotificationRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetIncidentNotificationRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateIncidentNotificationRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteIncidentNotificationRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.ListIncidentNotificationRules", true)

	// Enable AWS CUR Config
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateCostAWSCURConfig", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.ListCostAWSCURConfigs", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateCostAWSCURConfig", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteCostAWSCURConfig", true)

	// Enable Deployment Gates & Rules
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateDeploymentGate", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateDeploymentGate", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteDeploymentGate", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetDeploymentGate", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.CreateDeploymentRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.UpdateDeploymentRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.DeleteDeploymentRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetDeploymentRule", true)
	ddClientConfig.SetUnstableOperationEnabled("v2.GetDeploymentGateRules", true)

	if !config.ApiUrl.IsNull() && config.ApiUrl.ValueString() != "" {
		parsedAPIURL, parseErr := url.Parse(config.ApiUrl.ValueString())
		if parseErr != nil {
			diags.AddError("invalid API URL", parseErr.Error())
			return diags
		}
		if parsedAPIURL.Host == "" || parsedAPIURL.Scheme == "" {
			diags.AddError("invalid API URL", fmt.Sprintf("API URL '%s' missing protocol or host", config.ApiUrl.ValueString()))
			return diags
		}
		// If api url is passed, set and use the api name and protocol on ServerIndex{1}
		auth = context.WithValue(auth, datadog.ContextServerIndex, 1)
		auth = context.WithValue(auth, datadog.ContextServerVariables, map[string]string{
			"name":     parsedAPIURL.Host,
			"protocol": parsedAPIURL.Scheme,
		})

		// Configure URL's per operation
		// IPRangesApiService.GetIPRanges
		ipRangesDNSNameArr := strings.Split(parsedAPIURL.Hostname(), ".")
		// Parse out subdomain if it exists
		if len(ipRangesDNSNameArr) > 2 {
			ipRangesDNSNameArr = ipRangesDNSNameArr[1:]
		}
		ipRangesDNSNameArr = append([]string{utils.BaseIPRangesSubdomain}, ipRangesDNSNameArr...)

		auth = context.WithValue(auth, datadog.ContextOperationServerIndices, map[string]int{
			"v1.IPRangesApi.GetIPRanges": 1,
		})
		auth = context.WithValue(auth, datadog.ContextOperationServerVariables, map[string]map[string]string{
			"v1.IPRangesApi.GetIPRanges": {
				"name": strings.Join(ipRangesDNSNameArr, "."),
			},
		})
	}

	if httpClientRetryEnabled {
		ddClientConfig.RetryConfiguration.EnableRetry = httpClientRetryEnabled

		if !config.HttpClientRetryBackoffMultiplier.IsNull() {
			timeout := time.Duration(config.HttpClientRetryBackoffMultiplier.ValueInt64()) * time.Second
			ddClientConfig.RetryConfiguration.HTTPRetryTimeout = timeout
		}

		if !config.HttpClientRetryBackoffBase.IsNull() {
			ddClientConfig.RetryConfiguration.BackOffBase = float64(config.HttpClientRetryBackoffBase.ValueInt64())
		}

		if !config.HttpClientRetryMaxRetries.IsNull() {
			ddClientConfig.RetryConfiguration.MaxRetries = int(config.HttpClientRetryMaxRetries.ValueInt64())
		}
	}

	ddClientConfig.HTTPClient = utils.NewHTTPClient()
	// If cloud_provider_type is set, use cloud auth (takes precedence over API keys)
	if cloudProviderType != "" {
		switch cloudProviderType {
		case "aws":
			ddClientConfig.DelegatedTokenConfig = &datadog.DelegatedTokenConfig{
				OrgUUID: orgUUID,
				ProviderAuth: &datadog.AWSAuth{
					AwsRegion: cloudProviderRegion,
				},
				Provider: "aws",
			}
		}
	}
	datadogClient := datadog.NewAPIClient(ddClientConfig)

	p.DatadogApiInstances = &utils.ApiInstances{HttpClient: datadogClient}
	p.Auth = auth

	var defaultTags map[string]string
	if len(config.DefaultTags) > 0 && !config.DefaultTags[0].Tags.IsNull() {
		tagBlock := config.DefaultTags[0]
		diags.Append(tagBlock.Tags.ElementsAs(auth, &defaultTags, false)...)
	}
	p.DefaultTags = defaultTags
	/*  Commented out due to duplicate validation in SDK provider - remove after Framework migration is complete.
	if validate {
		log.Println("[INFO] Datadog client successfully initialized, now validating...")
		if cloudProviderType != "" { // Validate the cloud auth credentials
			delegatedConfig, err := datadogClient.GetDelegatedToken(auth)
			if err != nil {
				diags.AddError("[ERROR] Datadog Client validation error: %v", err.Error())
				return diags
			}
			if delegatedConfig.DelegatedToken == "" {
				msg := fmt.Sprintf(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your OrgUUID is correct and your cloud auth credentials for "%s" are valid and are for the correct region, see https://www.terraform.io/docs/providers/datadog/ for more information on providing credentials for the Datadog Provider`, cloudProviderType)
				err := errors.New(msg)
				diags.AddError("[ERROR] Datadog Client validation error: %v", err.Error())
				return diags
			}
		} else { // Validate the API and APP keys
			resp, _, err := p.DatadogApiInstances.GetAuthenticationApiV1().Validate(auth)
			if err != nil {
				diags.AddError("[ERROR] Datadog Client validation error", err.Error())
				return diags
			}
			valid, ok := resp.GetValidOk()
			if (ok && !*valid) || !ok {
				err := errors.New(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and are for the correct region, see https://www.terraform.io/docs/providers/datadog/ for more information on providing credentials for the Datadog Provider`)
				diags.AddError("[ERROR] Datadog Client validation error", err.Error())
				return diags
			}
		}
	} else {
		log.Println("[INFO] Skipping key validation (validate = false)")
	}
	log.Printf("[INFO] Datadog Client successfully validated.")
	*/
	return nil
}

var (
	_ resource.ResourceWithConfigure        = &FrameworkResourceWrapper{}
	_ resource.ResourceWithImportState      = &FrameworkResourceWrapper{}
	_ resource.ResourceWithConfigValidators = &FrameworkResourceWrapper{}
	_ resource.ResourceWithModifyPlan       = &FrameworkResourceWrapper{}
	_ resource.ResourceWithUpgradeState     = &FrameworkResourceWrapper{}
	_ resource.ResourceWithValidateConfig   = &FrameworkResourceWrapper{}
)

func NewFrameworkResourceWrapper(i *resource.Resource) resource.Resource {
	return &FrameworkResourceWrapper{
		innerResource: i,
	}
}

type FrameworkResourceWrapper struct {
	innerResource *resource.Resource
}

func (r *FrameworkResourceWrapper) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	rCasted, ok := (*r.innerResource).(resource.ResourceWithConfigure)
	if ok {
		if req.ProviderData == nil {
			return
		}
		_, ok := req.ProviderData.(*FrameworkProvider)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Resource Configure Type", "")
			return
		}

		rCasted.Configure(ctx, req, resp)
	}
}

func (r *FrameworkResourceWrapper) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	(*r.innerResource).Metadata(ctx, req, resp)
	resp.TypeName = req.ProviderTypeName + resp.TypeName
}

func (r *FrameworkResourceWrapper) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	(*r.innerResource).Schema(ctx, req, resp)
	fwutils.EnrichFrameworkResourceSchema(&resp.Schema)
}

func (r *FrameworkResourceWrapper) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	(*r.innerResource).Create(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	(*r.innerResource).Read(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	(*r.innerResource).Update(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	(*r.innerResource).Delete(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if rCasted, ok := (*r.innerResource).(resource.ResourceWithImportState); ok {
		if req.ID == "" {
			resp.Diagnostics.AddError("resource ID is required for import and cannot be empty", "")
			return
		}
		rCasted.ImportState(ctx, req, resp)
		return
	}

	resp.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (r *FrameworkResourceWrapper) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if rCasted, ok := (*r.innerResource).(resource.ResourceWithConfigValidators); ok {
		return rCasted.ConfigValidators(ctx)
	}
	return nil
}

func (r *FrameworkResourceWrapper) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if v, ok := (*r.innerResource).(resource.ResourceWithModifyPlan); ok {
		// If the plan is null, no need to modify the plan
		// Plan is null in case destroy planning
		if req.Plan.Raw.IsNull() {
			return
		}
		v.ModifyPlan(ctx, req, resp)
	}
}

func (r *FrameworkResourceWrapper) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := (*r.innerResource).(resource.ResourceWithUpgradeState); ok {
		return v.UpgradeState(ctx)
	}
	return nil
}

func (r *FrameworkResourceWrapper) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if v, ok := (*r.innerResource).(resource.ResourceWithValidateConfig); ok {
		v.ValidateConfig(ctx, req, resp)
	}
}

var (
	_ datasource.DataSourceWithConfigure        = &FrameworkDatasourceWrapper{}
	_ datasource.DataSourceWithConfigValidators = &FrameworkDatasourceWrapper{}
	_ datasource.DataSourceWithValidateConfig   = &FrameworkDatasourceWrapper{}
	_ datasource.DataSource                     = &FrameworkDatasourceWrapper{}
)

func NewFrameworkDatasourceWrapper(i *datasource.DataSource) datasource.DataSource {
	return &FrameworkDatasourceWrapper{
		innerDatasource: i,
	}
}

type FrameworkDatasourceWrapper struct {
	innerDatasource *datasource.DataSource
}

func (r *FrameworkDatasourceWrapper) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	rCasted, ok := (*r.innerDatasource).(datasource.DataSourceWithConfigure)
	if ok {
		if req.ProviderData == nil {
			return
		}
		_, ok := req.ProviderData.(*FrameworkProvider)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "")
			return
		}

		rCasted.Configure(ctx, req, resp)
	}
}

func (r *FrameworkDatasourceWrapper) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if rCasted, ok := (*r.innerDatasource).(datasource.DataSourceWithValidateConfig); ok {
		rCasted.ValidateConfig(ctx, req, resp)
	}
}

func (r *FrameworkDatasourceWrapper) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if rCasted, ok := (*r.innerDatasource).(datasource.DataSourceWithConfigValidators); ok {
		return rCasted.ConfigValidators(ctx)
	}
	return nil
}

func (r *FrameworkDatasourceWrapper) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	(*r.innerDatasource).Metadata(ctx, req, resp)
	resp.TypeName = req.ProviderTypeName + resp.TypeName
}

func (r *FrameworkDatasourceWrapper) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	(*r.innerDatasource).Schema(ctx, req, resp)
	fwutils.EnrichFrameworkDatasourceSchema(&resp.Schema)
}

func (r *FrameworkDatasourceWrapper) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	(*r.innerDatasource).Read(ctx, req, resp)
}
