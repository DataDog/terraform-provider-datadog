package fwprovider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &scorecardRulesDataSource{}

type scorecardRulesDataSource struct {
	Api  *datadogV2.ServiceScorecardsApi
	Auth context.Context
}

type scorecardRuleItemModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ScorecardName types.String `tfsdk:"scorecard_name"`
	Description   types.String `tfsdk:"description"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Level         types.String `tfsdk:"level"`
	Owner         types.String `tfsdk:"owner"`
	Custom        types.Bool   `tfsdk:"custom"`
}

type scorecardRulesDataSourceModel struct {
	// Query Parameters
	FilterName        types.String `tfsdk:"filter_name"`
	FilterEnabled     types.Bool   `tfsdk:"filter_enabled"`
	FilterCustom      types.Bool   `tfsdk:"filter_custom"`
	FilterDescription types.String `tfsdk:"filter_description"`

	// Results
	ID    types.String              `tfsdk:"id"`
	Rules []*scorecardRuleItemModel `tfsdk:"rules"`
}

func NewScorecardRulesDataSource() datasource.DataSource {
	return &scorecardRulesDataSource{}
}

func (d *scorecardRulesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetServiceScorecardsApiV2()
	d.Auth = providerData.Auth
}

func (d *scorecardRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "scorecard_rules"
}

func (d *scorecardRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing Datadog scorecard rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"filter_name": schema.StringAttribute{
				Description: "Filter rules by name.",
				Optional:    true,
			},
			"filter_enabled": schema.BoolAttribute{
				Description: "Filter for enabled rules only.",
				Optional:    true,
			},
			"filter_custom": schema.BoolAttribute{
				Description: "Filter for custom rules only.",
				Optional:    true,
			},
			"filter_description": schema.StringAttribute{
				Description: "Filter rules by description.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"rules": schema.ListNestedBlock{
				Description: "List of scorecard rules.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the scorecard rule.",
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *scorecardRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state scorecardRulesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	optParams := datadogV2.NewListScorecardRulesOptionalParameters()
	optParams.WithPageSize(100)
	if !state.FilterName.IsNull() {
		optParams.WithFilterRuleName(state.FilterName.ValueString())
	}
	if !state.FilterEnabled.IsNull() {
		optParams.WithFilterRuleEnabled(state.FilterEnabled.ValueBool())
	}
	if !state.FilterCustom.IsNull() {
		optParams.WithFilterRuleCustom(state.FilterCustom.ValueBool())
	}
	if !state.FilterDescription.IsNull() {
		optParams.WithFilterRuleDescription(state.FilterDescription.ValueString())
	}

	var allRules []datadogV2.ListRulesResponseDataItem
	offset := int64(0)
	for {
		optParams.WithPageOffset(offset)
		ddResp, _, err := d.Api.ListScorecardRules(d.Auth, *optParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing scorecard rules"))
			return
		}
		if err := utils.CheckForUnparsed(ddResp); err != nil {
			resp.Diagnostics.AddError("response contains unparsedObject", err.Error())
			return
		}

		data := ddResp.GetData()
		if len(data) == 0 {
			break
		}
		allRules = append(allRules, data...)
		offset += int64(len(data))
	}

	d.updateState(&state, allRules)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *scorecardRulesDataSource) updateState(state *scorecardRulesDataSourceModel, rules []datadogV2.ListRulesResponseDataItem) {
	ruleModels := make([]*scorecardRuleItemModel, 0, len(rules))
	for _, item := range rules {
		attrs := item.GetAttributes()
		m := &scorecardRuleItemModel{
			ID:   types.StringValue(item.GetId()),
			Name: types.StringValue(attrs.GetName()),
		}
		if v, ok := attrs.GetScorecardNameOk(); ok && v != nil {
			m.ScorecardName = types.StringValue(*v)
		}
		if v, ok := attrs.GetDescriptionOk(); ok && v != nil {
			m.Description = types.StringValue(*v)
		}
		if v, ok := attrs.GetEnabledOk(); ok && v != nil {
			m.Enabled = types.BoolValue(*v)
		}
		if v, ok := attrs.GetLevelOk(); ok && v != nil {
			m.Level = types.StringValue(strconv.FormatInt(int64(*v), 10))
		}
		if v, ok := attrs.GetOwnerOk(); ok && v != nil {
			m.Owner = types.StringValue(*v)
		}
		if v, ok := attrs.GetCustomOk(); ok && v != nil {
			m.Custom = types.BoolValue(*v)
		}
		ruleModels = append(ruleModels, m)
	}

	hashingData := fmt.Sprintf("%s:%t:%t:%s",
		state.FilterName.ValueString(),
		state.FilterEnabled.ValueBool(),
		state.FilterCustom.ValueBool(),
		state.FilterDescription.ValueString(),
	)
	state.ID = types.StringValue(utils.ConvertToSha256(hashingData))
	state.Rules = ruleModels
}
