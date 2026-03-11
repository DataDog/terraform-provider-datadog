package fwprovider

import (
	"context"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &scorecardRuleDataSource{}

type scorecardRuleDataSource struct {
	Api  *datadogV2.ServiceScorecardsApi
	Auth context.Context
}

type scorecardRuleDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ScorecardName types.String `tfsdk:"scorecard_name"`
	Description   types.String `tfsdk:"description"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Level         types.String `tfsdk:"level"`
	Owner         types.String `tfsdk:"owner"`
	Custom        types.Bool   `tfsdk:"custom"`
}

func NewScorecardRuleDataSource() datasource.DataSource {
	return &scorecardRuleDataSource{}
}

func (d *scorecardRuleDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetServiceScorecardsApiV2()
	d.Auth = providerData.Auth
}

func (d *scorecardRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "scorecard_rule"
}

func (d *scorecardRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog scorecard rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the scorecard rule.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the rule.",
				Computed:    true,
			},
			"scorecard_name": schema.StringAttribute{
				Description: "Name of the scorecard this rule belongs to.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the rule.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is enabled.",
				Computed:    true,
			},
			"level": schema.StringAttribute{
				Description: "The maturity level of the rule.",
				Computed:    true,
			},
			"owner": schema.StringAttribute{
				Description: "Owner of the rule.",
				Computed:    true,
			},
			"custom": schema.BoolAttribute{
				Description: "Whether the rule is a custom rule.",
				Computed:    true,
			},
		},
	}
}

func (d *scorecardRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state scorecardRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	optParams := datadogV2.NewListScorecardRulesOptionalParameters()
	optParams.WithFilterRuleId(id)
	optParams.WithPageSize(1)

	ddResp, _, err := d.Api.ListScorecardRules(d.Auth, *optParams)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading scorecard rule"))
		return
	}
	if err := utils.CheckForUnparsed(ddResp); err != nil {
		resp.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	data := ddResp.GetData()
	if len(data) == 0 {
		resp.Diagnostics.AddError("scorecard rule not found", "no scorecard rule found with the given ID")
		return
	}

	item := data[0]
	attrs := item.GetAttributes()

	state.ID = types.StringValue(item.GetId())
	if v, ok := attrs.GetNameOk(); ok && v != nil {
		state.Name = types.StringValue(*v)
	}
	if v, ok := attrs.GetScorecardNameOk(); ok && v != nil {
		state.ScorecardName = types.StringValue(*v)
	}
	if v, ok := attrs.GetDescriptionOk(); ok && v != nil {
		state.Description = types.StringValue(*v)
	} else {
		state.Description = types.StringNull()
	}
	if v, ok := attrs.GetEnabledOk(); ok && v != nil {
		state.Enabled = types.BoolValue(*v)
	}
	if v, ok := attrs.GetLevelOk(); ok && v != nil {
		state.Level = types.StringValue(strconv.FormatInt(int64(*v), 10))
	}
	if v, ok := attrs.GetOwnerOk(); ok && v != nil {
		state.Owner = types.StringValue(*v)
	} else {
		state.Owner = types.StringNull()
	}
	if v, ok := attrs.GetCustomOk(); ok && v != nil {
		state.Custom = types.BoolValue(*v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
