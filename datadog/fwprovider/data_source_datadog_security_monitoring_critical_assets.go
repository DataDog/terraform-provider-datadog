package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &securityMonitoringCriticalAssetsDataSource{}
)

type securityMonitoringCriticalAssetsDataSourceModel struct {
	ID             types.String                                      `tfsdk:"id"`
	CriticalAssets []*securityMonitoringCriticalAssetDataSourceModel `tfsdk:"critical_assets"`
}

type securityMonitoringCriticalAssetsDataSource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

func NewSecurityMonitoringCriticalAssetsDataSource() datasource.DataSource {
	return &securityMonitoringCriticalAssetsDataSource{}
}

func (d *securityMonitoringCriticalAssetsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	d.Auth = providerData.Auth
}

func (d *securityMonitoringCriticalAssetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "security_monitoring_critical_assets"
}

func (d *securityMonitoringCriticalAssetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all critical assets for the current org.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"critical_assets": schema.ListAttribute{
				Computed:    true,
				Description: "List of critical assets",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":         types.StringType,
						"enabled":    types.BoolType,
						"query":      types.StringType,
						"rule_query": types.StringType,
						"severity":   types.StringType,
						"tags": types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *securityMonitoringCriticalAssetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state securityMonitoringCriticalAssetsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddResp, httpResp, err := d.Api.ListSecurityMonitoringCriticalAssets(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error getting critical assets"), ""))
		return
	}
	if err := utils.CheckForUnparsed(ddResp); err != nil {
		resp.Diagnostics.AddError("Failed to parse response", err.Error())
		return
	}

	d.updateState(ctx, &state, &ddResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *securityMonitoringCriticalAssetsDataSource) updateState(ctx context.Context, state *securityMonitoringCriticalAssetsDataSourceModel, assetsData *datadogV2.SecurityMonitoringCriticalAssetsResponse) {
	var assets []*securityMonitoringCriticalAssetDataSourceModel

	for _, asset := range assetsData.GetData() {
		attrs := asset.GetAttributes()

		tags, _ := types.ListValueFrom(ctx, types.StringType, attrs.GetTags())

		a := &securityMonitoringCriticalAssetDataSourceModel{
			Id:        types.StringValue(asset.GetId()),
			Enabled:   types.BoolValue(attrs.GetEnabled()),
			Query:     types.StringValue(attrs.GetQuery()),
			RuleQuery: types.StringValue(attrs.GetRuleQuery()),
			Severity:  types.StringValue(string(attrs.GetSeverity())),
			Tags:      tags,
		}

		assets = append(assets, a)
	}

	state.ID = types.StringValue("critical_assets")
	state.CriticalAssets = assets
}
