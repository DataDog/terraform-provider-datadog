package fwprovider

import (
	"context"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSourceWithConfigure = &datadogTeamDataSource{}
)

type securityMonitoringSuppressionsDataSourceModel struct {
	Id             types.String                         `tfsdk:"id"`
	SuppressionIds types.List                           `tfsdk:"suppression_ids"`
	Suppressions   []securityMonitoringSuppressionModel `tfsdk:"suppressions"`
}

type securityMonitoringSuppressionDataSource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

func NewSecurityMonitoringSuppressionDataSource() datasource.DataSource {
	return &securityMonitoringSuppressionDataSource{}
}

func (r *securityMonitoringSuppressionDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (*securityMonitoringSuppressionDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "security_monitoring_suppressions"
}

func (r *securityMonitoringSuppressionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state securityMonitoringSuppressionsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, _, err := r.api.ListSecurityMonitoringSuppressions(r.auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching suppressions"))
		return
	}

	data := res.GetData()

	suppressionIds := make([]string, len(data))
	suppressions := make([]securityMonitoringSuppressionModel, len(data))

	for idx, suppression := range res.GetData() {
		var suppressionModel securityMonitoringSuppressionModel
		suppressionModel.Id = types.StringValue(suppression.GetId())
		attributes := suppression.Attributes

		suppressionModel.Name = types.StringValue(attributes.GetName())
		suppressionModel.Description = types.StringValue(attributes.GetDescription())
		suppressionModel.Enabled = types.BoolValue(attributes.GetEnabled())
		suppressionModel.RuleQuery = types.StringValue(attributes.GetRuleQuery())
		suppressionModel.SuppressionQuery = types.StringValue(attributes.GetSuppressionQuery())

		if attributes.StartDate == nil {
			suppressionModel.StartDate = types.StringNull()
		} else {
			startDate := time.UnixMilli(*attributes.StartDate).Format(time.RFC3339)
			suppressionModel.StartDate = types.StringValue(startDate)
		}
		if attributes.ExpirationDate == nil {
			suppressionModel.ExpirationDate = types.StringNull()
		} else {
			expirationDate := time.UnixMilli(*attributes.ExpirationDate).Format(time.RFC3339)
			suppressionModel.ExpirationDate = types.StringValue(expirationDate)
		}

		suppressionIds[idx] = suppression.GetId()
		suppressions[idx] = suppressionModel
	}

	// Build the resource ID based on the suppression IDs
	state.Id = types.StringValue(strings.Join(suppressionIds, "--"))
	tfSuppressionIds, diags := types.ListValueFrom(ctx, types.StringType, suppressionIds)
	response.Diagnostics.Append(diags...)
	state.SuppressionIds = tfSuppressionIds
	state.Suppressions = suppressions

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (*securityMonitoringSuppressionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing suppression rules, and use them in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"suppression_ids": schema.ListAttribute{
				Computed:    true,
				Description: "List of IDs of suppressions",
				ElementType: types.StringType,
			},
			"suppressions": schema.ListAttribute{
				Computed:    true,
				Description: "List of suppressions",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                   types.StringType,
						"name":                 types.StringType,
						"description":          types.StringType,
						"enabled":              types.BoolType,
						"start_date":           types.StringType,
						"expiration_date":      types.StringType,
						"rule_query":           types.StringType,
						"suppression_query":    types.StringType,
						"data_exclusion_query": types.StringType,
					},
				},
			},
		},
	}
}
