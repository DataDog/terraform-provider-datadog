package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	AwsRegionsIncludeAll  *awsRegionsIncludeAllModel  `tfsdk:"aws_regions_include_all"`
	AwsRegionsIncludeOnly *awsRegionsIncludeOnlyModel `tfsdk:"aws_regions_include_only"`
}
type awsRegionsIncludeAllModel struct {
	IncludeAll types.Bool `tfsdk:"include_all"`
}
type awsRegionsIncludeOnlyModel struct {
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
	AwsNamespaceFiltersExcludeAll  *awsNamespaceFiltersExcludeAllModel  `tfsdk:"aws_namespace_filters_exclude_all"`
	AwsNamespaceFiltersExcludeOnly *awsNamespaceFiltersExcludeOnlyModel `tfsdk:"aws_namespace_filters_exclude_only"`
	AwsNamespaceFiltersIncludeAll  *awsNamespaceFiltersIncludeAllModel  `tfsdk:"aws_namespace_filters_include_all"`
	AwsNamespaceFiltersIncludeOnly *awsNamespaceFiltersIncludeOnlyModel `tfsdk:"aws_namespace_filters_include_only"`
}
type awsNamespaceFiltersExcludeAllModel struct {
	ExcludeAll types.Bool `tfsdk:"exclude_all"`
}
type awsNamespaceFiltersExcludeOnlyModel struct {
	ExcludeOnly types.List `tfsdk:"exclude_only"`
}
type awsNamespaceFiltersIncludeAllModel struct {
	IncludeAll types.Bool `tfsdk:"include_all"`
}
type awsNamespaceFiltersIncludeOnlyModel struct {
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
	XRayServicesIncludeAll  *xRayServicesIncludeAllModel  `tfsdk:"x_ray_services_include_all"`
	XRayServicesIncludeOnly *xRayServicesIncludeOnlyModel `tfsdk:"x_ray_services_include_only"`
}
type xRayServicesIncludeAllModel struct {
	IncludeAll types.Bool `tfsdk:"include_all"`
}
type xRayServicesIncludeOnlyModel struct {
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
								Description: "AWS Secret Access Key",
							},
						},
					},
					"aws_auth_config_role": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"external_id": schema.StringAttribute{
								Optional:    true,
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
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"aws_regions_include_all": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"include_all": schema.BoolAttribute{
								Optional:    true,
								Description: "Include all regions",
							},
						},
					},
					"aws_regions_include_only": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"include_only": schema.ListAttribute{
								Optional:    true,
								Description: "Include only these regions",
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			"logs_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"lambda_forwarder": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"lambdas": schema.ListAttribute{
								Required:    true,
								Description: "List of Datadog Lambda Log Forwarder ARNs",
								ElementType: types.StringType,
							},
							"sources": schema.ListAttribute{
								Required:    true,
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
						Optional:    true,
						Description: "Enable EC2 automute for AWS metrics",
					},
					"collect_cloudwatch_alarms": schema.BoolAttribute{
						Optional:    true,
						Description: "Enable CloudWatch alarms collection",
					},
					"collect_custom_metrics": schema.BoolAttribute{
						Optional:    true,
						Description: "Enable custom metrics collection",
					},
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Description: "Enable AWS metrics collection",
					},
				},
				Blocks: map[string]schema.Block{
					"tag_filters": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"namespace": schema.StringAttribute{
									Optional:    true,
									Description: "The AWS Namespace to apply the tag filters against",
								},
								"tags": schema.ListAttribute{
									Optional:    true,
									Description: "The tags to filter based on",
									ElementType: types.StringType,
								},
							},
						},
					},
					"namespace_filters": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{},
						Blocks: map[string]schema.Block{
							"aws_namespace_filters_exclude_all": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"exclude_all": schema.BoolAttribute{
										Optional:    true,
										Description: "Exclude all namespaces",
									},
								},
							},
							"aws_namespace_filters_exclude_only": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"exclude_only": schema.ListAttribute{
										Optional:    true,
										Description: "Exclude only these namespaces",
										ElementType: types.StringType,
									},
								},
							},
							"aws_namespace_filters_include_all": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"include_all": schema.BoolAttribute{
										Optional:    true,
										Description: "Include all namespaces",
									},
								},
							},
							"aws_namespace_filters_include_only": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"include_only": schema.ListAttribute{
										Optional:    true,
										Description: "Include only these namespaces",
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			"resources_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"cloud_security_posture_management_collection": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether Datadog collects cloud security posture management resources from your AWS account.",
					},
					"extended_collection": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether Datadog collects additional attributes and configuration information about the resources in your AWS account. Required for `cspm_resource_collection`.",
					},
				},
			},
			"traces_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"xray_services": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{},
						Blocks: map[string]schema.Block{
							"x_ray_services_include_all": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"include_all": schema.BoolAttribute{
										Optional:    true,
										Description: "Include all services",
									},
								},
							},
							"x_ray_services_include_only": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"include_only": schema.ListAttribute{
										Optional:    true,
										Description: "Include only these services",
										ElementType: types.StringType,
									},
								},
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

	aws_account_id := state.AwsAccountId.ValueString()
	resp, httpResp, err := r.Api.GetAWSAccount(r.Auth, aws_account_id)
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

	aws_account_id := state.AwsAccountId.ValueString()

	body, diags := r.buildAwsAccountV2UpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateAWSAccount(r.Auth, aws_account_id, *body)
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

	aws_account_id := state.AwsAccountId.ValueString()

	httpResp, err := r.Api.DeleteAWSAccount(r.Auth, aws_account_id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting aws_account_v2"))
		return
	}
}

func (r *awsAccountV2Resource) updateState(ctx context.Context, state *awsAccountV2Model, resp *datadogV2.AWSAccountResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.AwsAccountId = types.StringValue(attributes.GetAwsAccountId())

	state.AwsPartition = types.StringValue(string(attributes.GetAwsPartition()))

	if accountTags, ok := attributes.GetAccountTagsOk(); ok {
		if accountTags != nil && len(*accountTags) > 0 {
			state.AccountTags, _ = types.ListValueFrom(ctx, types.StringType, *accountTags)
		}
	}

	if logsConfig, ok := attributes.GetLogsConfigOk(); ok {

		logsConfigTf := logsConfigModel{}
		lambdaForwarderTf := lambdaForwarderModel{}
		if lambdaForwarder, ok := logsConfig.GetLambdaForwarderOk(); ok {
			if lambdas, ok := lambdaForwarder.GetLambdasOk(); ok && len(*lambdas) > 0 {
				lambdaForwarderTf.Lambdas, _ = types.ListValueFrom(ctx, types.StringType, *lambdas)
			} else {
				lambdaForwarderTf.Lambdas, _ = types.ListValueFrom(ctx, types.StringType, &[]string{})
			}

			if sources, ok := lambdaForwarder.GetSourcesOk(); ok && len(*sources) > 0 {
				lambdaForwarderTf.Sources, _ = types.ListValueFrom(ctx, types.StringType, *sources)
			} else {
				lambdaForwarderTf.Sources, _ = types.ListValueFrom(ctx, types.StringType, &[]string{})
			}
		}
		logsConfigTf.LambdaForwarder = &lambdaForwarderTf
		state.LogsConfig = &logsConfigTf
	}

	if metricsConfig, ok := attributes.GetMetricsConfigOk(); ok {

		metricsConfigTf := metricsConfigModel{}
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
			metricsConfigTf.TagFilters = []*tagFiltersModel{}
			for _, tagFiltersDd := range *tagFilters {
				tagFiltersTfItem := tagFiltersModel{}
				if namespace, ok := tagFiltersDd.GetNamespaceOk(); ok {
					tagFiltersTfItem.Namespace = types.StringValue(*namespace)
				}
				if tags, ok := tagFiltersDd.GetTagsOk(); ok && len(*tags) > 0 {
					tagFiltersTfItem.Tags, _ = types.ListValueFrom(ctx, types.StringType, *tags)
				}
				metricsConfigTf.TagFilters = append(metricsConfigTf.TagFilters, &tagFiltersTfItem)
			}
		}

		state.MetricsConfig = &metricsConfigTf
	}

	if resourcesConfig, ok := attributes.GetResourcesConfigOk(); ok {

		resourcesConfigTf := resourcesConfigModel{}
		if cloudSecurityPostureManagementCollection, ok := resourcesConfig.GetCloudSecurityPostureManagementCollectionOk(); ok {
			resourcesConfigTf.CloudSecurityPostureManagementCollection = types.BoolValue(*cloudSecurityPostureManagementCollection)
		}
		if extendedCollection, ok := resourcesConfig.GetExtendedCollectionOk(); ok {
			resourcesConfigTf.ExtendedCollection = types.BoolValue(*extendedCollection)
		}

		state.ResourcesConfig = &resourcesConfigTf
	}

	if tracesConfig, ok := attributes.GetTracesConfigOk(); ok {

		tracesConfigTf := tracesConfigModel{}

		tracesConfigTf.XrayServices = &xrayServicesModel{}

		if xrayServices, ok := tracesConfig.GetXrayServicesOk(); ok {
			if xrayServices.XRayServicesIncludeAll != nil {
				xrayServicesIncludeAllTf := xRayServicesIncludeAllModel{}
				if includeAll, ok := xrayServices.XRayServicesIncludeAll.GetIncludeAllOk(); ok {
					xrayServicesIncludeAllTf.IncludeAll = types.BoolValue(*includeAll)
				}

				tracesConfigTf.XrayServices.XRayServicesIncludeAll = &xrayServicesIncludeAllTf
			}

			if xrayServices.XRayServicesIncludeOnly != nil {
				xrayServicesIncludeOnlyTf := xRayServicesIncludeOnlyModel{}
				if includeOnly, ok := xrayServices.XRayServicesIncludeOnly.GetIncludeOnlyOk(); ok && len(*includeOnly) > 0 {
					xrayServicesIncludeOnlyTf.IncludeOnly, _ = types.ListValueFrom(ctx, types.StringType, *includeOnly)
				}

				tracesConfigTf.XrayServices.XRayServicesIncludeOnly = &xrayServicesIncludeOnlyTf
			}
		}

		state.TracesConfig = &tracesConfigTf
	}
}

func (r *awsAccountV2Resource) buildAwsAccountV2RequestBody(ctx context.Context, state *awsAccountV2Model) (*datadogV2.AWSAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAWSAccountCreateRequestAttributesWithDefaults()

	attributes.SetAwsAccountId(state.AwsAccountId.ValueString())
	attributes.SetAwsPartition(datadogV2.AWSAccountPartition(state.AwsPartition.ValueString()))

	if state.AuthConfig != nil {
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

		attributes.SetAuthConfig(authConfig)
	}

	if !state.AccountTags.IsNull() {
		var accountTags []string
		diags.Append(state.AccountTags.ElementsAs(ctx, &accountTags, false)...)
		attributes.SetAccountTags(accountTags)
	}

	if state.LogsConfig != nil {
		var logsConfig datadogV2.AWSLogsConfig

		if state.LogsConfig.LambdaForwarder != nil {
			var lambdaForwarder datadogV2.AWSLambdaForwarderConfig

			if !state.LogsConfig.LambdaForwarder.Lambdas.IsNull() {
				var lambdas []string
				diags.Append(state.LogsConfig.LambdaForwarder.Lambdas.ElementsAs(ctx, &lambdas, false)...)
				lambdaForwarder.SetLambdas(lambdas)
			}

			if !state.LogsConfig.LambdaForwarder.Sources.IsNull() {
				var sources []string
				diags.Append(state.LogsConfig.LambdaForwarder.Sources.ElementsAs(ctx, &sources, false)...)
				lambdaForwarder.SetSources(sources)
			}

			logsConfig.LambdaForwarder = &lambdaForwarder
		}

		attributes.LogsConfig = &logsConfig
	}

	if state.MetricsConfig != nil {
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

		if state.MetricsConfig.TagFilters != nil {
			var tagFilters []datadogV2.AWSNamespaceTagFilter
			for _, tagFiltersTFItem := range state.MetricsConfig.TagFilters {
				tagFiltersDDItem := datadogV2.NewAWSNamespaceTagFilter()

				if !tagFiltersTFItem.Namespace.IsNull() {
					tagFiltersDDItem.SetNamespace(tagFiltersTFItem.Namespace.ValueString())
				}

				if !tagFiltersTFItem.Tags.IsNull() {
					var tags []string
					diags.Append(tagFiltersTFItem.Tags.ElementsAs(ctx, &tags, false)...)
					tagFiltersDDItem.SetTags(tags)
				}
			}
			metricsConfig.SetTagFilters(tagFilters)
		}

		attributes.MetricsConfig = &metricsConfig
	}

	if state.ResourcesConfig != nil {
		var resourcesConfig datadogV2.AWSResourcesConfig

		if !state.ResourcesConfig.CloudSecurityPostureManagementCollection.IsNull() {
			resourcesConfig.SetCloudSecurityPostureManagementCollection(state.ResourcesConfig.CloudSecurityPostureManagementCollection.ValueBool())
		}
		if !state.ResourcesConfig.ExtendedCollection.IsNull() {
			resourcesConfig.SetExtendedCollection(state.ResourcesConfig.ExtendedCollection.ValueBool())
		}

		attributes.ResourcesConfig = &resourcesConfig
	}

	if state.TracesConfig != nil {
		var tracesConfig datadogV2.AWSTracesConfig

		attributes.TracesConfig = &tracesConfig
	}

	req := datadogV2.NewAWSAccountCreateRequestWithDefaults()
	req.Data = datadogV2.NewAWSAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *awsAccountV2Resource) buildAwsAccountV2UpdateRequestBody(ctx context.Context, state *awsAccountV2Model) (*datadogV2.AWSAccountPatchRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAWSAccountPatchRequestAttributesWithDefaults()

	attributes.SetAwsAccountId(state.AwsAccountId.ValueString())
	attributes.SetAwsPartition(datadogV2.AWSAccountPartition(state.AwsPartition.ValueString()))

	if state.AuthConfig != nil {
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

		attributes.SetAuthConfig(authConfig)
	}

	if !state.AccountTags.IsNull() {
		var accountTags []string
		diags.Append(state.AccountTags.ElementsAs(ctx, &accountTags, false)...)
		attributes.SetAccountTags(accountTags)
	}

	if state.LogsConfig != nil {
		var logsConfig datadogV2.AWSLogsConfig

		if state.LogsConfig.LambdaForwarder != nil {
			var lambdaForwarder datadogV2.AWSLambdaForwarderConfig

			if !state.LogsConfig.LambdaForwarder.Lambdas.IsNull() {
				var lambdas []string
				diags.Append(state.LogsConfig.LambdaForwarder.Lambdas.ElementsAs(ctx, &lambdas, false)...)
				lambdaForwarder.SetLambdas(lambdas)
			}

			if !state.LogsConfig.LambdaForwarder.Sources.IsNull() {
				var sources []string
				diags.Append(state.LogsConfig.LambdaForwarder.Sources.ElementsAs(ctx, &sources, false)...)
				lambdaForwarder.SetSources(sources)
			}

			logsConfig.LambdaForwarder = &lambdaForwarder
		}

		attributes.LogsConfig = &logsConfig
	}

	if state.MetricsConfig != nil {
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

		if state.MetricsConfig.TagFilters != nil {
			var tagFilters []datadogV2.AWSNamespaceTagFilter
			for _, tagFiltersTFItem := range state.MetricsConfig.TagFilters {
				tagFiltersDDItem := datadogV2.NewAWSNamespaceTagFilter()

				if !tagFiltersTFItem.Namespace.IsNull() {
					tagFiltersDDItem.SetNamespace(tagFiltersTFItem.Namespace.ValueString())
				}

				if !tagFiltersTFItem.Tags.IsNull() {
					var tags []string
					diags.Append(tagFiltersTFItem.Tags.ElementsAs(ctx, &tags, false)...)
					tagFiltersDDItem.SetTags(tags)
				}
			}
			metricsConfig.SetTagFilters(tagFilters)
		}

		attributes.MetricsConfig = &metricsConfig
	}

	if state.ResourcesConfig != nil {
		var resourcesConfig datadogV2.AWSResourcesConfig

		if !state.ResourcesConfig.CloudSecurityPostureManagementCollection.IsNull() {
			resourcesConfig.SetCloudSecurityPostureManagementCollection(state.ResourcesConfig.CloudSecurityPostureManagementCollection.ValueBool())
		}
		if !state.ResourcesConfig.ExtendedCollection.IsNull() {
			resourcesConfig.SetExtendedCollection(state.ResourcesConfig.ExtendedCollection.ValueBool())
		}

		attributes.ResourcesConfig = &resourcesConfig
	}

	if state.TracesConfig != nil {
		var tracesConfig datadogV2.AWSTracesConfig

		attributes.TracesConfig = &tracesConfig
	}

	req := datadogV2.NewAWSAccountPatchRequestWithDefaults()
	req.Data = datadogV2.NewAWSAccountPatchRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
