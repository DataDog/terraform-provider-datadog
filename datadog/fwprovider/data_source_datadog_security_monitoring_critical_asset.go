package fwprovider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &securityMonitoringCriticalAssetDataSource{}

type securityMonitoringCriticalAssetDataSource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

type securityMonitoringCriticalAssetDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Query     types.String `tfsdk:"query"`
	RuleQuery types.String `tfsdk:"rule_query"`
	Severity  types.String `tfsdk:"severity"`
	Tags      types.List   `tfsdk:"tags"`
}

func NewSecurityMonitoringCriticalAssetDataSource() datasource.DataSource {
	return &securityMonitoringCriticalAssetDataSource{}
}

func (d *securityMonitoringCriticalAssetDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "security_monitoring_critical_asset"
}

func (d *securityMonitoringCriticalAssetDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	d.auth = providerData.Auth
}

func (d *securityMonitoringCriticalAssetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing security monitoring critical asset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the critical asset.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the critical asset is enabled.",
			},
			"query": schema.StringAttribute{
				Computed:    true,
				Description: "The query used to match a critical asset and the associated signals.",
			},
			"rule_query": schema.StringAttribute{
				Computed:    true,
				Description: "The rule query to filter which detection rules this critical asset applies to.",
			},
			"severity": schema.StringAttribute{
				Computed:    true,
				Description: "The severity change applied to signals matching this critical asset.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A list of tags associated with the critical asset.",
			},
		},
	}
}

func (d *securityMonitoringCriticalAssetDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state securityMonitoringCriticalAssetDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	criticalAssetId := state.Id.ValueString()

	res, httpResponse, err := d.api.GetSecurityMonitoringCriticalAsset(d.auth, criticalAssetId)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.Diagnostics.AddError("Not Found", fmt.Sprintf("Critical asset with ID %s not found", criticalAssetId))
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching security monitoring critical asset"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	d.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *securityMonitoringCriticalAssetDataSource) updateStateFromResponse(ctx context.Context, state *securityMonitoringCriticalAssetDataSourceModel, res *datadogV2.SecurityMonitoringCriticalAssetResponse) {
	state.Id = types.StringValue(res.Data.GetId())

	attributes := res.Data.Attributes

	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Query = types.StringValue(attributes.GetQuery())
	state.RuleQuery = types.StringValue(attributes.GetRuleQuery())
	state.Severity = types.StringValue(string(attributes.GetSeverity()))

	if len(attributes.GetTags()) == 0 {
		state.Tags = types.ListNull(types.StringType)
	} else {
		state.Tags, _ = types.ListValueFrom(ctx, types.StringType, attributes.GetTags())
	}
}
