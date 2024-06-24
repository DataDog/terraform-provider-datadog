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
	_ resource.ResourceWithConfigure   = &awsAccountResource{}
	_ resource.ResourceWithImportState = &awsAccountResource{}
)

type awsAccountResource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

type awsAccountModel struct {
	ID              types.String          `tfsdk:"id"`
	AwsAccountId    types.String          `tfsdk:"aws_account_id"`
	AccountTags     types.List            `tfsdk:"account_tags"`
	AuthConfig      *authConfigModel      `tfsdk:"auth_config"`
	AwsRegions      *awsRegionsModel      `tfsdk:"aws_regions"`
	LogsConfig      *logsConfigModel      `tfsdk:"logs_config"`
	MetricsConfig   *metricsConfigModel   `tfsdk:"metrics_config"`
	ResourcesConfig *resourcesConfigModel `tfsdk:"resources_config"`
	TracesConfig    *tracesConfigModel    `tfsdk:"traces_config"`
}

type authConfigModel struct {
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	ExternalId      types.String `tfsdk:"external_id"`
	RoleName        types.String `tfsdk:"role_name"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
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
	ExcludeAll  types.Bool `tfsdk:"exclude_all"`
	IncludeAll  types.Bool `tfsdk:"include_all"`
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

func NewAwsAccountResource() resource.Resource {
	return &awsAccountResource{}
}

func (r *awsAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *awsAccountResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_account"
}

func (r *awsAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AwsAccount resource. This can be used to create and manage Datadog aws_account.",
		Attributes: map[string]schema.Attribute{
			"aws_account_id": schema.StringAttribute{
				Optional:    true,
				Description: "AWS Account ID",
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
				Attributes: map[string]schema.Attribute{
					"access_key_id": schema.StringAttribute{
						Optional:    true,
						Description: "AWS Access Key ID",
					},
					"external_id": schema.StringAttribute{
						Optional:    true,
						Description: "AWS IAM External ID for associated role",
					},
					"role_name": schema.StringAttribute{
						Optional:    true,
						Description: "AWS IAM Role name",
					},
					"secret_access_key": schema.StringAttribute{
						Optional:    true,
						Description: "AWS Secret Access Key",
					},
				},
			},
			"aws_regions": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"include_all": schema.BoolAttribute{
						Optional:    true,
						Description: "Include all regions",
					},
					"include_only": schema.ListAttribute{
						Optional:    true,
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
								Description: "List of Datadog Lambda Log Forwarder ARNs",
								ElementType: types.StringType,
							},
							"sources": schema.ListAttribute{
								Optional:    true,
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
						Attributes: map[string]schema.Attribute{
							"exclude_all": schema.BoolAttribute{
								Optional:    true,
								Description: "Exclude all namespaces",
							},
							"include_all": schema.BoolAttribute{
								Optional:    true,
								Description: "Include all namespaces",
							},
							"exclude_only": schema.ListAttribute{
								Optional:    true,
								Description: "Exclude only these namespaces",
								ElementType: types.StringType,
							},
							"include_only": schema.ListAttribute{
								Optional:    true,
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
						Attributes: map[string]schema.Attribute{
							"include_all": schema.BoolAttribute{
								Optional:    true,
								Description: "Include all services",
							},
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
	}
}

func (r *awsAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *awsAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state awsAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetAWSAccountv2(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AwsAccount"))
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

func (r *awsAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state awsAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildAwsAccountRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateAWSAccountv2(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AwsAccount"))
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

func (r *awsAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state awsAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildAwsAccountUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.PatchAWSAccountv2(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AwsAccount"))
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

func (r *awsAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state awsAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteAWSAccountv2(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting aws_account"))
		return
	}
}

func (r *awsAccountResource) updateState(ctx context.Context, state *awsAccountModel, resp *datadogV2.AWSAccountResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if awsAccountId, ok := attributes.GetAwsAccountIdOk(); ok {
		state.AwsAccountId = types.StringValue(*awsAccountId)
	}

	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(*createdAt)
	}

	if modifiedAt, ok := attributes.GetModifiedAtOk(); ok {
		state.ModifiedAt = types.StringValue(*modifiedAt)
	}

	if accountTags, ok := attributes.GetAccountTagsOk(); ok && len(*accountTags) > 0 {
		state.AccountTags, _ = types.ListValueFrom(ctx, types.StringType, *accountTags)
	}

	if authConfig, ok := attributes.GetAuthConfigOk(); ok {

		authConfigTf := authConfigModel{}
		if accessKeyId, ok := authConfig.GetAccessKeyIdOk(); ok {
			authConfigTf.AccessKeyId = types.StringValue(*accessKeyId)
		}
		if externalId, ok := authConfig.GetExternalIdOk(); ok {
			authConfigTf.ExternalId = types.StringValue(*externalId)
		}
		if roleName, ok := authConfig.GetRoleNameOk(); ok {
			authConfigTf.RoleName = types.StringValue(*roleName)
		}
		if secretAccessKey, ok := authConfig.GetSecretAccessKeyOk(); ok {
			authConfigTf.SecretAccessKey = types.StringValue(*secretAccessKey)
		}

		state.AuthConfig = &authConfigTf
	}

	if awsRegions, ok := attributes.GetAwsRegionsOk(); ok {

		awsRegionsTf := awsRegionsModel{}
		if includeAll, ok := awsRegions.GetIncludeAllOk(); ok {
			awsRegionsTf.IncludeAll = types.BoolValue(*includeAll)
		}
		if includeOnly, ok := awsRegions.GetIncludeOnlyOk(); ok && len(*includeOnly) > 0 {
			awsRegionsTf.IncludeOnly, _ = types.ListValueFrom(ctx, types.StringType, *includeOnly)
		}

		state.AwsRegions = &awsRegionsTf
	}

	if logsConfig, ok := attributes.GetLogsConfigOk(); ok {

		logsConfigTf := logsConfigModel{}
		if lambdaForwarder, ok := logsConfig.GetLambdaForwarderOk(); ok {

			lambdaForwarderTf := lambdaForwarderModel{}
			if lambdas, ok := lambdaForwarder.GetLambdasOk(); ok && len(*lambdas) > 0 {
				lambdaForwarderTf.Lambdas, _ = types.ListValueFrom(ctx, types.StringType, *lambdas)
			}
			if sources, ok := lambdaForwarder.GetSourcesOk(); ok && len(*sources) > 0 {
				lambdaForwarderTf.Sources, _ = types.ListValueFrom(ctx, types.StringType, *sources)
			}

			logsConfigTf.LambdaForwarder = &lambdaForwarderTf
		}

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
		if namespaceFilters, ok := metricsConfig.GetNamespaceFiltersOk(); ok {

			namespaceFiltersTf := namespaceFiltersModel{}
			if excludeAll, ok := namespaceFilters.GetExcludeAllOk(); ok {
				namespaceFiltersTf.ExcludeAll = types.BoolValue(*excludeAll)
			}
			if excludeOnly, ok := namespaceFilters.GetExcludeOnlyOk(); ok && len(*excludeOnly) > 0 {
				namespaceFiltersTf.ExcludeOnly, _ = types.ListValueFrom(ctx, types.StringType, *excludeOnly)
			}
			if includeAll, ok := namespaceFilters.GetIncludeAllOk(); ok {
				namespaceFiltersTf.IncludeAll = types.BoolValue(*includeAll)
			}
			if includeOnly, ok := namespaceFilters.GetIncludeOnlyOk(); ok && len(*includeOnly) > 0 {
				namespaceFiltersTf.IncludeOnly, _ = types.ListValueFrom(ctx, types.StringType, *includeOnly)
			}

			metricsConfigTf.NamespaceFilters = &namespaceFiltersTf
		}
		if tagFilters, ok := metricsConfig.GetTagFiltersOk(); ok && len(*tagFilters) > 0 {
			metricsConfigTf.TagFilters = []*tagFiltersModel{}
			for _, tagFiltersDd := range *tagFilters {
				tagFiltersTfItem := tagFiltersModel{}

				if tagFilters, ok := tagFiltersDd.GetTagFiltersOk(); ok {

					tagFiltersTf := tagFiltersModel{}
					if namespace, ok := tagFilters.GetNamespaceOk(); ok {
						tagFiltersTf.Namespace = types.StringValue(*namespace)
					}
					if tags, ok := tagFilters.GetTagsOk(); ok && len(*tags) > 0 {
						tagFiltersTf.Tags, _ = types.ListValueFrom(ctx, types.StringType, *tags)
					}

					tagFiltersTfItem.TagFilters = &tagFiltersTf
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
		if xrayServices, ok := tracesConfig.GetXrayServicesOk(); ok {

			xrayServicesTf := xrayServicesModel{}
			if includeAll, ok := xrayServices.GetIncludeAllOk(); ok {
				xrayServicesTf.IncludeAll = types.BoolValue(*includeAll)
			}
			if includeOnly, ok := xrayServices.GetIncludeOnlyOk(); ok && len(*includeOnly) > 0 {
				xrayServicesTf.IncludeOnly, _ = types.ListValueFrom(ctx, types.StringType, *includeOnly)
			}

			tracesConfigTf.XrayServices = &xrayServicesTf
		}

		state.TracesConfig = &tracesConfigTf
	}
}

func (r *awsAccountResource) buildAwsAccountRequestBody(ctx context.Context, state *awsAccountModel) (*datadogV2.AWSAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAWSAccountCreateAttributesWithDefaults()

	attributes.SetAwsAccountId(state.AwsAccountId.ValueString())

	if !state.AccountTags.IsNull() {
		var accountTags []string
		diags.Append(state.AccountTags.ElementsAs(ctx, &accountTags, false)...)
		attributes.SetAccountTags(accountTags)
	}

	if state.AuthConfig != nil {
		var authConfig datadogV2.AWSAuthConfig

		if !state.AuthConfig.AccessKeyId.IsNull() {
			authConfig.SetAccessKeyId(state.AuthConfig.AccessKeyId.ValueString())
		}
		if !state.AuthConfig.ExternalId.IsNull() {
			authConfig.SetExternalId(state.AuthConfig.ExternalId.ValueString())
		}
		if !state.AuthConfig.RoleName.IsNull() {
			authConfig.SetRoleName(state.AuthConfig.RoleName.ValueString())
		}
		if !state.AuthConfig.SecretAccessKey.IsNull() {
			authConfig.SetSecretAccessKey(state.AuthConfig.SecretAccessKey.ValueString())
		}

		attributes.AuthConfig = &authConfig
	}

	if state.AwsRegions != nil {
		var awsRegions datadogV2.AWSRegionsList

		if !state.AwsRegions.IncludeAll.IsNull() {
			awsRegions.SetIncludeAll(state.AwsRegions.IncludeAll.ValueBool())
		}

		if !state.AwsRegions.IncludeOnly.IsNull() {
			var includeOnly []string
			diags.Append(state.AwsRegions.IncludeOnly.ElementsAs(ctx, &includeOnly, false)...)
			awsRegions.SetIncludeOnly(includeOnly)
		}

		attributes.AwsRegions = &awsRegions
	}

	if state.LogsConfig != nil {
		var logsConfig datadogV2.AWSLogs

		if state.LogsConfig.LambdaForwarder != nil {
			var lambdaForwarder datadogV2.AWSLambdaForwarder

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
		var metricsConfig datadogV2.AWSMetrics

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

		if state.MetricsConfig.NamespaceFilters != nil {
			var namespaceFilters datadogV2.AWSNamespacesList

			if !state.MetricsConfig.NamespaceFilters.ExcludeAll.IsNull() {
				namespaceFilters.SetExcludeAll(state.MetricsConfig.NamespaceFilters.ExcludeAll.ValueBool())
			}
			if !state.MetricsConfig.NamespaceFilters.IncludeAll.IsNull() {
				namespaceFilters.SetIncludeAll(state.MetricsConfig.NamespaceFilters.IncludeAll.ValueBool())
			}

			if !state.MetricsConfig.NamespaceFilters.ExcludeOnly.IsNull() {
				var excludeOnly []string
				diags.Append(state.MetricsConfig.NamespaceFilters.ExcludeOnly.ElementsAs(ctx, &excludeOnly, false)...)
				namespaceFilters.SetExcludeOnly(excludeOnly)
			}

			if !state.MetricsConfig.NamespaceFilters.IncludeOnly.IsNull() {
				var includeOnly []string
				diags.Append(state.MetricsConfig.NamespaceFilters.IncludeOnly.ElementsAs(ctx, &includeOnly, false)...)
				namespaceFilters.SetIncludeOnly(includeOnly)
			}

			metricsConfig.NamespaceFilters = &namespaceFilters
		}

		attributes.MetricsConfig = &metricsConfig
	}

	if state.ResourcesConfig != nil {
		var resourcesConfig datadogV2.AWSResources

		if !state.ResourcesConfig.CloudSecurityPostureManagementCollection.IsNull() {
			resourcesConfig.SetCloudSecurityPostureManagementCollection(state.ResourcesConfig.CloudSecurityPostureManagementCollection.ValueBool())
		}
		if !state.ResourcesConfig.ExtendedCollection.IsNull() {
			resourcesConfig.SetExtendedCollection(state.ResourcesConfig.ExtendedCollection.ValueBool())
		}

		attributes.ResourcesConfig = &resourcesConfig
	}

	if state.TracesConfig != nil {
		var tracesConfig datadogV2.AWSTraces

		if state.TracesConfig.XrayServices != nil {
			var xrayServices datadogV2.AWSXRayServicesList

			if !state.TracesConfig.XrayServices.IncludeAll.IsNull() {
				xrayServices.SetIncludeAll(state.TracesConfig.XrayServices.IncludeAll.ValueBool())
			}

			if !state.TracesConfig.XrayServices.IncludeOnly.IsNull() {
				var includeOnly []string
				diags.Append(state.TracesConfig.XrayServices.IncludeOnly.ElementsAs(ctx, &includeOnly, false)...)
				xrayServices.SetIncludeOnly(includeOnly)
			}

			tracesConfig.XrayServices = &xrayServices
		}

		attributes.TracesConfig = &tracesConfig
	}

	req := datadogV2.NewAWSAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewAWSAccountCreateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *awsAccountResource) buildAwsAccountUpdateRequestBody(ctx context.Context, state *awsAccountModel) (*datadogV2.AWSAccountPatchRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAWSAccountPatchAttributesWithDefaults()

	if !state.AwsAccountId.IsNull() {
		attributes.SetAwsAccountId(state.AwsAccountId.ValueString())
	}
	if !state.AwsAccountName.IsNull() {
		attributes.SetAwsAccountName(state.AwsAccountName.ValueString())
	}
	if !state.CreatedAt.IsNull() {
		attributes.SetCreatedAt(state.CreatedAt.ValueString())
	}
	if !state.ModifiedAt.IsNull() {
		attributes.SetModifiedAt(state.ModifiedAt.ValueString())
	}

	if !state.AccountTags.IsNull() {
		var accountTags []string
		diags.Append(state.AccountTags.ElementsAs(ctx, &accountTags, false)...)
		attributes.SetAccountTags(accountTags)
	}

	if state.AuthConfig != nil {
		var authConfig datadogV2.AWSAuthConfig

		if !state.AuthConfig.AccessKeyId.IsNull() {
			authConfig.SetAccessKeyId(state.AuthConfig.AccessKeyId.ValueString())
		}
		if !state.AuthConfig.ExternalId.IsNull() {
			authConfig.SetExternalId(state.AuthConfig.ExternalId.ValueString())
		}
		if !state.AuthConfig.RoleName.IsNull() {
			authConfig.SetRoleName(state.AuthConfig.RoleName.ValueString())
		}
		if !state.AuthConfig.SecretAccessKey.IsNull() {
			authConfig.SetSecretAccessKey(state.AuthConfig.SecretAccessKey.ValueString())
		}

		attributes.AuthConfig = &authConfig
	}

	if state.AwsRegions != nil {
		var awsRegions datadogV2.AWSRegionsList

		if !state.AwsRegions.IncludeAll.IsNull() {
			awsRegions.SetIncludeAll(state.AwsRegions.IncludeAll.ValueBool())
		}

		if !state.AwsRegions.IncludeOnly.IsNull() {
			var includeOnly []string
			diags.Append(state.AwsRegions.IncludeOnly.ElementsAs(ctx, &includeOnly, false)...)
			awsRegions.SetIncludeOnly(includeOnly)
		}

		attributes.AwsRegions = &awsRegions
	}

	if state.LogsConfig != nil {
		var logsConfig datadogV2.AWSLogs

		if state.LogsConfig.LambdaForwarder != nil {
			var lambdaForwarder datadogV2.AWSLambdaForwarder

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
		var metricsConfig datadogV2.AWSMetrics

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

		if state.MetricsConfig.NamespaceFilters != nil {
			var namespaceFilters datadogV2.AWSNamespacesList

			if !state.MetricsConfig.NamespaceFilters.ExcludeAll.IsNull() {
				namespaceFilters.SetExcludeAll(state.MetricsConfig.NamespaceFilters.ExcludeAll.ValueBool())
			}
			if !state.MetricsConfig.NamespaceFilters.IncludeAll.IsNull() {
				namespaceFilters.SetIncludeAll(state.MetricsConfig.NamespaceFilters.IncludeAll.ValueBool())
			}

			if !state.MetricsConfig.NamespaceFilters.ExcludeOnly.IsNull() {
				var excludeOnly []string
				diags.Append(state.MetricsConfig.NamespaceFilters.ExcludeOnly.ElementsAs(ctx, &excludeOnly, false)...)
				namespaceFilters.SetExcludeOnly(excludeOnly)
			}

			if !state.MetricsConfig.NamespaceFilters.IncludeOnly.IsNull() {
				var includeOnly []string
				diags.Append(state.MetricsConfig.NamespaceFilters.IncludeOnly.ElementsAs(ctx, &includeOnly, false)...)
				namespaceFilters.SetIncludeOnly(includeOnly)
			}

			metricsConfig.NamespaceFilters = &namespaceFilters
		}

		attributes.MetricsConfig = &metricsConfig
	}

	if state.ResourcesConfig != nil {
		var resourcesConfig datadogV2.AWSResources

		if !state.ResourcesConfig.CloudSecurityPostureManagementCollection.IsNull() {
			resourcesConfig.SetCloudSecurityPostureManagementCollection(state.ResourcesConfig.CloudSecurityPostureManagementCollection.ValueBool())
		}
		if !state.ResourcesConfig.ExtendedCollection.IsNull() {
			resourcesConfig.SetExtendedCollection(state.ResourcesConfig.ExtendedCollection.ValueBool())
		}

		attributes.ResourcesConfig = &resourcesConfig
	}

	if state.TracesConfig != nil {
		var tracesConfig datadogV2.AWSTraces

		if state.TracesConfig.XrayServices != nil {
			var xrayServices datadogV2.AWSXRayServicesList

			if !state.TracesConfig.XrayServices.IncludeAll.IsNull() {
				xrayServices.SetIncludeAll(state.TracesConfig.XrayServices.IncludeAll.ValueBool())
			}

			if !state.TracesConfig.XrayServices.IncludeOnly.IsNull() {
				var includeOnly []string
				diags.Append(state.TracesConfig.XrayServices.IncludeOnly.ElementsAs(ctx, &includeOnly, false)...)
				xrayServices.SetIncludeOnly(includeOnly)
			}

			tracesConfig.XrayServices = &xrayServices
		}

		attributes.TracesConfig = &tracesConfig
	}

	req := datadogV2.NewAWSAccountPatchRequestWithDefaults()
	req.Data = datadogV2.NewAWSAccountPatchWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
