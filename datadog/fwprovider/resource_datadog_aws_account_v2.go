package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &awsAccountV2Resource{}
	_ resource.ResourceWithImportState = &awsAccountV2Resource{}
)

type awsAccountV2Resource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

type awsAccountV2Model struct {
	ID              types.String          `tfsdk:"id"`
	AwsAccountId    types.String          `tfsdk:"aws_account_id"`
	AwsPartition    types.String          `tfsdk:"aws_partition"`
	AccountTags     types.List            `tfsdk:"account_tags"`
	AuthConfig      *authConfigModel      `tfsdk:"auth_config"`
	AwsRegions      *awsRegionsModel      `tfsdk:"aws_regions"`
	LogsConfig      *logsConfigModel      `tfsdk:"logs_config"`
	MetricsConfig   *metricsConfigModel   `tfsdk:"metrics_config"`
	ResourcesConfig *resourcesConfigModel `tfsdk:"resources_config"`
	TracesConfig    *tracesConfigModel    `tfsdk:"traces_config"`
}

type authConfigModel struct {
	AwsAuthConfigKeys *awsAuthConfigKeysModel `tfsdk:"aws_auth_config_keys"`
	AwsAuthConfigRole *awsAuthConfigRoleModel `tfsdk:"aws_auth_config_role"`
}
type awsAuthConfigKeysModel struct {
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}
type awsAuthConfigRoleModel struct {
	ExternalId types.String `tfsdk:"external_id"`
	RoleName   types.String `tfsdk:"role_name"`
}

type awsRegionsModel struct {
	IncludeAll  types.Bool `tfsdk:"include_all"`
	IncludeOnly types.List `tfsdk:"include_only"`
}

type logsConfigModel struct {
	LambdaForwarder *lambdaForwarderModel `tfsdk:"lambda_forwarder"`
}
type lambdaForwarderModel struct {
	Lambdas types.List `tfsdk:"lambdas"`
	Sources types.List `tfsdk:"sources"`
}

type metricsConfigModel struct {
	AutomuteEnabled         types.Bool             `tfsdk:"automute_enabled"`
	CollectCloudwatchAlarms types.Bool             `tfsdk:"collect_cloudwatch_alarms"`
	CollectCustomMetrics    types.Bool             `tfsdk:"collect_custom_metrics"`
	Enabled                 types.Bool             `tfsdk:"enabled"`
	TagFilters              []*tagFiltersModel     `tfsdk:"tag_filters"`
	NamespaceFilters        *namespaceFiltersModel `tfsdk:"namespace_filters"`
}
type tagFiltersModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Tags      types.List   `tfsdk:"tags"`
}

type namespaceFiltersModel struct {
	ExcludeOnly types.List `tfsdk:"exclude_only"`
	IncludeOnly types.List `tfsdk:"include_only"`
}

type resourcesConfigModel struct {
	CloudSecurityPostureManagementCollection types.Bool `tfsdk:"cloud_security_posture_management_collection"`
	ExtendedCollection                       types.Bool `tfsdk:"extended_collection"`
}

type tracesConfigModel struct {
	XrayServices *xrayServicesModel `tfsdk:"xray_services"`
}
type xrayServicesModel struct {
	IncludeAll  types.Bool `tfsdk:"include_all"`
	IncludeOnly types.List `tfsdk:"include_only"`
}

func NewAwsAccountV2Resource() resource.Resource {
	return &awsAccountV2Resource{}
}

func (r *awsAccountV2Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *awsAccountV2Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_account_v2"
}

func (r *awsAccountV2Resource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("auth_config").AtName("aws_auth_config_keys"),
			path.MatchRoot("auth_config").AtName("aws_auth_config_role"),
		),
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("traces_config").AtName("xray_services").AtName("include_all"),
			path.MatchRoot("traces_config").AtName("xray_services").AtName("include_only"),
		),
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("metrics_config").AtName("namespace_filters").AtName("include_only"),
			path.MatchRoot("metrics_config").AtName("namespace_filters").AtName("exclude_only"),
		),
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("aws_regions").AtName("include_all"),
			path.MatchRoot("aws_regions").AtName("include_only"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("auth_config").AtName("aws_auth_config_keys").AtName("access_key_id"),
			path.MatchRoot("auth_config").AtName("aws_auth_config_keys").AtName("secret_access_key"),
		),
	}
}

func (r *awsAccountV2Resource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AwsAccountV2 resource. This can be used to create and manage Datadog aws_account_v2.",
		Attributes: map[string]schema.Attribute{
			"aws_account_id": schema.StringAttribute{
				Required:    true,
				Description: "AWS Account ID",
			},
			"aws_partition": schema.StringAttribute{
				Required:    true,
				Description: "AWS Account partition",
			},
			"account_tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Tags to apply to all metrics in the account",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"auth_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"aws_auth_config_keys": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"access_key_id": schema.StringAttribute{
								Optional:    true,
								Description: "AWS Access Key ID",
							},
							"secret_access_key": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "AWS Secret Access Key",
							},
						},
					},
					"aws_auth_config_role": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"external_id": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "AWS IAM External ID for associated role",
							},
							"role_name": schema.StringAttribute{
								Optional:    true,
								Description: "AWS IAM Role name",
							},
						},
					},
				},
			},
			"aws_regions": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"include_all": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Include all regions",
					},
					"include_only": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Include only these regions",
						ElementType: types.StringType,
					},
				},
			},
			"logs_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"lambda_forwarder": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"lambdas": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								Description: "List of Datadog Lambda Log Forwarder ARNs",
								ElementType: types.StringType,
							},
							"sources": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								Description: "List of AWS services that will send logs to the Datadog Lambda Log Forwarder",
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			"metrics_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"automute_enabled": schema.BoolAttribute{
						Required:    true,
						Description: "Enable EC2 automute for AWS metrics",
					},
					"collect_cloudwatch_alarms": schema.BoolAttribute{
						Required:    true,
						Description: "Enable CloudWatch alarms collection",
					},
					"collect_custom_metrics": schema.BoolAttribute{
						Required:    true,
						Description: "Enable custom metrics collection",
					},
					"enabled": schema.BoolAttribute{
						Required:    true,
						Description: "Enable AWS metrics collection",
					},
				},
				Blocks: map[string]schema.Block{
					"tag_filters": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"namespace": schema.StringAttribute{
									Required:    true,
									Description: "The AWS Namespace to apply the tag filters against",
								},
								"tags": schema.ListAttribute{
									Optional:    true,
									Computed:    true,
									Description: "The tags to filter based on",
									ElementType: types.StringType,
								},
							},
						},
					},
					"namespace_filters": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"exclude_only": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Exclude only these namespaces",
								ElementType: types.StringType,
							},
							"include_only": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Include only these namespaces",
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			"resources_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"cloud_security_posture_management_collection": schema.BoolAttribute{
						Required:    true,
						Description: "Whether Datadog collects cloud security posture management resources from your AWS account.",
					},
					"extended_collection": schema.BoolAttribute{
						Required:    true,
						Description: "Whether Datadog collects additional attributes and configuration information about the resources in your AWS account. Required for `cloud_security_posture_management_collection`.",
					},
				},
			},
			"traces_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"xray_services": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"include_all": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Include all services",
							},
							"include_only": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Include only these services",
								ElementType: types.StringType,
							},
						},
					},
				},
			},
		},
	}
}

func (r *awsAccountV2Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *awsAccountV2Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state awsAccountV2Model
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountConfigId := state.ID.String()
	resp, httpResp, err := r.Api.GetAWSAccount(r.Auth, awsAccountConfigId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AwsAccountV2"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *awsAccountV2Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state awsAccountV2Model
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildAwsAccountV2RequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateAWSAccount(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AwsAccountV2"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *awsAccountV2Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state awsAccountV2Model
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildAwsAccountV2UpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountConfigId := state.ID.String()
	resp, _, err := r.Api.UpdateAWSAccount(r.Auth, awsAccountConfigId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AwsAccountV2"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *awsAccountV2Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state awsAccountV2Model
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountConfigId := state.ID.String()
	httpResp, err := r.Api.DeleteAWSAccount(r.Auth, awsAccountConfigId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting aws_account_v2"))
		return
	}
}

func buildStateAuthConfig(attributes datadogV2.AWSAccountResponseAttributes) *authConfigModel {
	authConfigTf := authConfigModel{}
	if authConfig, ok := attributes.GetAuthConfigOk(); ok {
		if authConfig.AWSAuthConfigKeys != nil {
			authConfigTf.AwsAuthConfigKeys = &awsAuthConfigKeysModel{}
			authConfigTf.AwsAuthConfigKeys.AccessKeyId = types.StringValue(authConfig.AWSAuthConfigKeys.GetAccessKeyId())
			authConfigTf.AwsAuthConfigKeys.SecretAccessKey = types.StringValue(authConfig.AWSAuthConfigKeys.GetSecretAccessKey())
		} else if authConfig.AWSAuthConfigRole != nil {
			authConfigTf.AwsAuthConfigRole = &awsAuthConfigRoleModel{}
			authConfigTf.AwsAuthConfigRole.RoleName = types.StringValue(authConfig.AWSAuthConfigRole.GetRoleName())
			authConfigTf.AwsAuthConfigRole.ExternalId = types.StringValue(authConfig.AWSAuthConfigRole.GetExternalId())
		}
	}
	return &authConfigTf
}

func buildStateLogsConfig(ctx context.Context, attributes datadogV2.AWSAccountResponseAttributes, diags diag.Diagnostics) *logsConfigModel {
	logsConfig := attributes.GetLogsConfig()

	logsConfigTf := logsConfigModel{}
	lambdaForwarderTf := lambdaForwarderModel{}

	if lambdaForwarder, ok := logsConfig.GetLambdaForwarderOk(); ok {
		if lambdaForwarder != nil && (lambdaForwarder.HasLambdas() || lambdaForwarder.HasSources()) {
			lambdas := lambdaForwarder.GetLambdas()
			var d diag.Diagnostics
			lambdaForwarderTf.Lambdas, d = types.ListValueFrom(ctx, types.StringType, lambdas)
			diags.Append(d...)

			sources := lambdaForwarder.GetSources()
			lambdaForwarderTf.Sources, d = types.ListValueFrom(ctx, types.StringType, sources)
			diags.Append(d...)

			logsConfigTf.LambdaForwarder = &lambdaForwarderTf
		}
	}

	return &logsConfigTf
}

func buildStateAwsRegions(ctx context.Context, attributes datadogV2.AWSAccountResponseAttributes, diags diag.Diagnostics) *awsRegionsModel {
	awsRegions := attributes.GetAwsRegions()

	awsRegionsTf := awsRegionsModel{}
	if awsRegions.AWSRegionsIncludeAll != nil {
		awsRegionsTf.IncludeAll = types.BoolValue(awsRegions.AWSRegionsIncludeAll.GetIncludeAll())
	} else if awsRegions.AWSRegionsIncludeOnly != nil {
		includeOnly, d := types.ListValueFrom(ctx, types.StringType, awsRegions.AWSRegionsIncludeOnly.GetIncludeOnly())
		awsRegionsTf.IncludeOnly = includeOnly
		diags.Append(d...)
	}

	return &awsRegionsTf
}

func buildStateMetricsConfig(ctx context.Context, attributes datadogV2.AWSAccountResponseAttributes, diags diag.Diagnostics) *metricsConfigModel {
	metricsConfig := attributes.GetMetricsConfig()
	metricsConfigTf := metricsConfigModel{}
	metricsConfigTf.TagFilters = []*tagFiltersModel{}
	metricsConfigTf.NamespaceFilters = &namespaceFiltersModel{}
	if automuteEnabled, ok := metricsConfig.GetAutomuteEnabledOk(); ok {
		metricsConfigTf.AutomuteEnabled = types.BoolValue(*automuteEnabled)
	}
	if collectCloudwatchAlarms, ok := metricsConfig.GetCollectCloudwatchAlarmsOk(); ok {
		metricsConfigTf.CollectCloudwatchAlarms = types.BoolValue(*collectCloudwatchAlarms)
	}
	if collectCustomMetrics, ok := metricsConfig.GetCollectCustomMetricsOk(); ok {
		metricsConfigTf.CollectCustomMetrics = types.BoolValue(*collectCustomMetrics)
	}
	if enabled, ok := metricsConfig.GetEnabledOk(); ok {
		metricsConfigTf.Enabled = types.BoolValue(*enabled)
	}

	if tagFilters, ok := metricsConfig.GetTagFiltersOk(); ok && len(*tagFilters) > 0 {
		for _, tagFiltersDd := range *tagFilters {
			tagFiltersTfItem := tagFiltersModel{}
			if namespace, ok := tagFiltersDd.GetNamespaceOk(); ok {
				tagFiltersTfItem.Namespace = types.StringValue(*namespace)
			}
			if tags, ok := tagFiltersDd.GetTagsOk(); ok {
				tagsTf, d := types.ListValueFrom(ctx, types.StringType, *tags)
				tagFiltersTfItem.Tags = tagsTf
				diags.Append(d...)
			}
			metricsConfigTf.TagFilters = append(metricsConfigTf.TagFilters, &tagFiltersTfItem)
		}
	}

	if namespaceFilters, ok := metricsConfig.GetNamespaceFiltersOk(); ok {
		nsFiltersTf := namespaceFiltersModel{}
		if namespaceFilters.AWSNamespaceFiltersExcludeOnly != nil {
			excludeOnly, _ := types.ListValueFrom(ctx, types.StringType, namespaceFilters.AWSNamespaceFiltersExcludeOnly.GetExcludeOnly())
			nsFiltersTf.ExcludeOnly = excludeOnly
		}
		if namespaceFilters.AWSNamespaceFiltersIncludeOnly != nil {
			includeOnly, _ := types.ListValueFrom(ctx, types.StringType, namespaceFilters.AWSNamespaceFiltersIncludeOnly.GetIncludeOnly())
			nsFiltersTf.IncludeOnly = includeOnly
		}

		metricsConfigTf.NamespaceFilters = &nsFiltersTf
	}

	return &metricsConfigTf
}

func buildStateResourcesConfig(attributes datadogV2.AWSAccountResponseAttributes) *resourcesConfigModel {
	resourcesConfig := attributes.GetResourcesConfig()
	resourcesConfigTf := resourcesConfigModel{}
	resourcesConfigTf.CloudSecurityPostureManagementCollection = types.BoolValue(resourcesConfig.GetCloudSecurityPostureManagementCollection())
	resourcesConfigTf.ExtendedCollection = types.BoolValue(resourcesConfig.GetExtendedCollection())
	return &resourcesConfigTf
}

func buildStateTracesConfig(ctx context.Context, attributes datadogV2.AWSAccountResponseAttributes, diags diag.Diagnostics) *tracesConfigModel {
	tracesConfig := attributes.GetTracesConfig()
	tracesConfigTf := tracesConfigModel{}

	if xrayServices, ok := tracesConfig.GetXrayServicesOk(); ok {
		xrayServicesTf := xrayServicesModel{}
		if xrayServices.XRayServicesIncludeAll != nil {
			xrayServicesTf.IncludeAll = types.BoolValue(xrayServices.XRayServicesIncludeAll.GetIncludeAll())
		} else if xrayServices.XRayServicesIncludeOnly != nil {
			includeOnly, d := types.ListValueFrom(ctx, types.StringType, xrayServices.XRayServicesIncludeOnly.GetIncludeOnly())
			xrayServicesTf.IncludeOnly = includeOnly
			diags.Append(d...)
		}
		tracesConfigTf.XrayServices = &xrayServicesTf
	}

	return &tracesConfigTf
}

func buildStateAccountTags(ctx context.Context, attributes datadogV2.AWSAccountResponseAttributes, diags diag.Diagnostics) types.List {
	accountTagsDd := attributes.GetAccountTags()
	if accountTagsDd == nil {
		accountTagsDd = []string{}
	}
	accountTags, _ := types.ListValueFrom(ctx, types.StringType, accountTagsDd)
	return accountTags
}

func (r *awsAccountV2Resource) updateState(ctx context.Context, state *awsAccountV2Model, resp *datadogV2.AWSAccountResponse) {
	state.ID = types.StringValue(resp.Data.GetId())
	diags := diag.Diagnostics{}

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.AwsAccountId = types.StringValue(attributes.GetAwsAccountId())
	state.AwsPartition = types.StringValue(string(attributes.GetAwsPartition()))
	state.AwsRegions = buildStateAwsRegions(ctx, attributes, diags)
	state.AuthConfig = buildStateAuthConfig(attributes)
	state.AccountTags = buildStateAccountTags(ctx, attributes, diags)
	state.LogsConfig = buildStateLogsConfig(ctx, attributes, diags)
	state.MetricsConfig = buildStateMetricsConfig(ctx, attributes, diags)
	state.ResourcesConfig = buildStateResourcesConfig(attributes)
	state.TracesConfig = buildStateTracesConfig(ctx, attributes, diags)
}

func (r *awsAccountV2Resource) buildAwsAccountV2RequestBody(ctx context.Context, state *awsAccountV2Model) (*datadogV2.AWSAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAWSAccountCreateRequestAttributesWithDefaults()

	attributes.SetAwsAccountId(state.AwsAccountId.ValueString())
	attributes.SetAwsPartition(datadogV2.AWSAccountPartition(state.AwsPartition.ValueString()))
	attributes.SetAwsRegions(buildRequestAwsRegions(ctx, state, diags))
	attributes.SetAuthConfig(buildRequestAuthConfig(state))
	attributes.SetAccountTags(buildRequestAccountTags(ctx, state, diags))
	attributes.SetLogsConfig(buildRequestLogsConfig(ctx, state, diags))
	attributes.SetMetricsConfig(buildRequestMetricsConfig(ctx, state, diags))
	attributes.SetResourcesConfig(buildRequestResourcesConfig(state))
	attributes.SetTracesConfig(buildRequestTracesConfig(ctx, state, diags))

	req := datadogV2.NewAWSAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewAWSAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func buildRequestAwsRegions(ctx context.Context, state *awsAccountV2Model, diags diag.Diagnostics) datadogV2.AWSRegions {
	regions := datadogV2.AWSRegions{}
	if !state.AwsRegions.IncludeAll.IsUnknown() {
		regions.AWSRegionsIncludeAll = datadogV2.NewAWSRegionsIncludeAllWithDefaults()
		regions.AWSRegionsIncludeAll.IncludeAll = state.AwsRegions.IncludeAll.ValueBool()
	}
	if !state.AwsRegions.IncludeOnly.IsUnknown() {
		regions.AWSRegionsIncludeOnly = datadogV2.NewAWSRegionsIncludeOnlyWithDefaults()
		var includeOnly []string
		diags.Append(state.AwsRegions.IncludeOnly.ElementsAs(ctx, &includeOnly, false)...)
		regions.AWSRegionsIncludeOnly.IncludeOnly = includeOnly
	}
	return regions
}
func buildRequestAuthConfig(state *awsAccountV2Model) datadogV2.AWSAuthConfig {
	authConfig := datadogV2.AWSAuthConfig{}

	if state.AuthConfig.AwsAuthConfigKeys != nil {
		authConfig.AWSAuthConfigKeys = datadogV2.NewAWSAuthConfigKeysWithDefaults()
		if !state.AuthConfig.AwsAuthConfigKeys.AccessKeyId.IsNull() {
			authConfig.AWSAuthConfigKeys.SetAccessKeyId(state.AuthConfig.AwsAuthConfigKeys.AccessKeyId.ValueString())
		}
		if !state.AuthConfig.AwsAuthConfigKeys.SecretAccessKey.IsNull() {
			authConfig.AWSAuthConfigKeys.SetSecretAccessKey(state.AuthConfig.AwsAuthConfigKeys.SecretAccessKey.ValueString())
		}
	}

	if state.AuthConfig.AwsAuthConfigRole != nil {
		authConfig.AWSAuthConfigRole = datadogV2.NewAWSAuthConfigRoleWithDefaults()
		if !state.AuthConfig.AwsAuthConfigRole.ExternalId.IsNull() {
			authConfig.AWSAuthConfigRole.SetExternalId(state.AuthConfig.AwsAuthConfigRole.ExternalId.ValueString())
		}
		if !state.AuthConfig.AwsAuthConfigRole.RoleName.IsNull() {
			authConfig.AWSAuthConfigRole.SetRoleName(state.AuthConfig.AwsAuthConfigRole.RoleName.ValueString())
		}
	}

	return authConfig
}

func buildRequestAccountTags(ctx context.Context, state *awsAccountV2Model, diags diag.Diagnostics) []string {
	accountTags := []string{}
	if !state.AccountTags.IsNull() {
		diags.Append(state.AccountTags.ElementsAs(ctx, &accountTags, false)...)
	}

	return accountTags
}

func buildRequestLogsConfig(ctx context.Context, state *awsAccountV2Model, diags diag.Diagnostics) datadogV2.AWSLogsConfig {
	logsConfig := datadogV2.AWSLogsConfig{}
	lambdaForwarder := datadogV2.AWSLambdaForwarderConfig{}
	lambdas := []string{}
	sources := []string{}
	if state.LogsConfig != nil && state.LogsConfig.LambdaForwarder != nil {
		if !state.LogsConfig.LambdaForwarder.Lambdas.IsNull() {
			diags.Append(state.LogsConfig.LambdaForwarder.Lambdas.ElementsAs(ctx, &lambdas, false)...)
		}
		if !state.LogsConfig.LambdaForwarder.Sources.IsNull() {
			diags.Append(state.LogsConfig.LambdaForwarder.Sources.ElementsAs(ctx, &sources, false)...)
		}
	}

	lambdaForwarder.SetLambdas(lambdas)
	lambdaForwarder.SetSources(sources)
	logsConfig.LambdaForwarder = &lambdaForwarder
	return logsConfig
}

func buildRequestMetricsConfig(ctx context.Context, state *awsAccountV2Model, diags diag.Diagnostics) datadogV2.AWSMetricsConfig {
	var metricsConfig datadogV2.AWSMetricsConfig

	if !state.MetricsConfig.AutomuteEnabled.IsNull() {
		metricsConfig.SetAutomuteEnabled(state.MetricsConfig.AutomuteEnabled.ValueBool())
	}
	if !state.MetricsConfig.CollectCloudwatchAlarms.IsNull() {
		metricsConfig.SetCollectCloudwatchAlarms(state.MetricsConfig.CollectCloudwatchAlarms.ValueBool())
	}
	if !state.MetricsConfig.CollectCustomMetrics.IsNull() {
		metricsConfig.SetCollectCustomMetrics(state.MetricsConfig.CollectCustomMetrics.ValueBool())
	}
	if !state.MetricsConfig.Enabled.IsNull() {
		metricsConfig.SetEnabled(state.MetricsConfig.Enabled.ValueBool())
	}

	tagFilters := []datadogV2.AWSNamespaceTagFilter{}
	for _, tagFiltersTFItem := range state.MetricsConfig.TagFilters {
		tagFiltersDDItem := datadogV2.NewAWSNamespaceTagFilterWithDefaults()

		if !tagFiltersTFItem.Namespace.IsNull() {
			tagFiltersDDItem.SetNamespace(tagFiltersTFItem.Namespace.ValueString())
		}

		if !tagFiltersTFItem.Tags.IsNull() {
			tags := []string{}
			diags.Append(tagFiltersTFItem.Tags.ElementsAs(ctx, &tags, false)...)
			tagFiltersDDItem.SetTags(tags)
		}
		tagFilters = append(tagFilters, *tagFiltersDDItem)
	}
	metricsConfig.SetTagFilters(tagFilters)

	namespaceFiltersDD := datadogV2.AWSNamespaceFilters{}
	nsFiltersTf := state.MetricsConfig.NamespaceFilters
	if !nsFiltersTf.ExcludeOnly.IsUnknown() {
		var excludeOnly []string
		namespaceFiltersDD.AWSNamespaceFiltersExcludeOnly = datadogV2.NewAWSNamespaceFiltersExcludeOnlyWithDefaults()
		diags.Append(nsFiltersTf.ExcludeOnly.ElementsAs(ctx, &excludeOnly, false)...)
		namespaceFiltersDD.AWSNamespaceFiltersExcludeOnly.SetExcludeOnly(excludeOnly)
	} else if !nsFiltersTf.ExcludeOnly.IsUnknown() {
		var includeOnly []string
		namespaceFiltersDD.AWSNamespaceFiltersIncludeOnly = datadogV2.NewAWSNamespaceFiltersIncludeOnlyWithDefaults()
		diags.Append(nsFiltersTf.IncludeOnly.ElementsAs(ctx, &includeOnly, false)...)
		namespaceFiltersDD.AWSNamespaceFiltersIncludeOnly.SetIncludeOnly(includeOnly)
	}
	metricsConfig.SetNamespaceFilters(namespaceFiltersDD)

	return metricsConfig
}

func buildRequestResourcesConfig(state *awsAccountV2Model) datadogV2.AWSResourcesConfig {
	var resourcesConfig datadogV2.AWSResourcesConfig

	if !state.ResourcesConfig.CloudSecurityPostureManagementCollection.IsNull() {
		resourcesConfig.SetCloudSecurityPostureManagementCollection(state.ResourcesConfig.CloudSecurityPostureManagementCollection.ValueBool())
	}
	if !state.ResourcesConfig.ExtendedCollection.IsNull() {
		resourcesConfig.SetExtendedCollection(state.ResourcesConfig.ExtendedCollection.ValueBool())
	}

	return resourcesConfig
}

func buildRequestTracesConfig(ctx context.Context, state *awsAccountV2Model, diags diag.Diagnostics) datadogV2.AWSTracesConfig {
	tracesConfig := datadogV2.NewAWSTracesConfigWithDefaults()

	if state.TracesConfig != nil {
		if state.TracesConfig.XrayServices != nil {
			var ddXRayServiceList datadogV2.XRayServicesList

			if !state.TracesConfig.XrayServices.IncludeAll.IsUnknown() {
				includeAll := state.TracesConfig.XrayServices.IncludeAll.ValueBool()
				ddXRayServiceList = datadogV2.XRayServicesIncludeAllAsXRayServicesList(&datadogV2.XRayServicesIncludeAll{IncludeAll: includeAll})
			} else if !state.TracesConfig.XrayServices.IncludeOnly.IsUnknown() {
				includeOnlyTf := state.TracesConfig.XrayServices.IncludeOnly
				var ddIncludeOnly []string
				diags.Append(includeOnlyTf.ElementsAs(ctx, &ddIncludeOnly, false)...)
				ddXRayServiceList = datadogV2.XRayServicesIncludeOnlyAsXRayServicesList(&datadogV2.XRayServicesIncludeOnly{IncludeOnly: ddIncludeOnly})
			}
			tracesConfig.SetXrayServices(ddXRayServiceList)
		}
	}

	return *tracesConfig
}

func (r *awsAccountV2Resource) buildAwsAccountV2UpdateRequestBody(ctx context.Context, state *awsAccountV2Model) (*datadogV2.AWSAccountUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAWSAccountUpdateRequestAttributesWithDefaults()

	attributes.SetAwsAccountId(state.AwsAccountId.ValueString())
	attributes.SetAwsPartition(datadogV2.AWSAccountPartition(state.AwsPartition.ValueString()))
	attributes.SetAwsRegions(buildRequestAwsRegions(ctx, state, diags))
	attributes.SetAuthConfig(buildRequestAuthConfig(state))
	attributes.SetAccountTags(buildRequestAccountTags(ctx, state, diags))
	attributes.SetLogsConfig(buildRequestLogsConfig(ctx, state, diags))
	attributes.SetMetricsConfig(buildRequestMetricsConfig(ctx, state, diags))
	attributes.SetResourcesConfig(buildRequestResourcesConfig(state))
	attributes.SetTracesConfig(buildRequestTracesConfig(ctx, state, diags))

	req := datadogV2.NewAWSAccountUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewAWSAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
