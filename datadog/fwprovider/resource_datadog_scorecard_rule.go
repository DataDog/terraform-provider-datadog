package fwprovider

import (
	"context"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &scorecardRuleResource{}
	_ resource.ResourceWithImportState = &scorecardRuleResource{}
)

type scorecardRuleResource struct {
	Api  *datadogV2.ServiceScorecardsApi
	Auth context.Context
}

type scorecardRuleModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ScorecardName types.String `tfsdk:"scorecard_name"`
	Description   types.String `tfsdk:"description"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Level         types.String `tfsdk:"level"`
	Owner         types.String `tfsdk:"owner"`
	Custom        types.Bool   `tfsdk:"custom"`
}

func NewScorecardRuleResource() resource.Resource {
	return &scorecardRuleResource{}
}

func (r *scorecardRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetServiceScorecardsApiV2()
	r.Auth = providerData.Auth
}

func (r *scorecardRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "scorecard_rule"
}

func (r *scorecardRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Scorecard Rule resource. This can be used to create and manage scorecard rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of the rule.",
				Required:    true,
			},
			"scorecard_name": schema.StringAttribute{
				Description: "Name of the scorecard this rule belongs to. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the rule.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"level": schema.StringAttribute{
				Description: "The maturity level of the rule.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("1", "2", "3"),
				},
			},
			"owner": schema.StringAttribute{
				Description: "Owner of the rule.",
				Optional:    true,
			},
			"custom": schema.BoolAttribute{
				Description: "Whether the rule is a custom rule.",
				Computed:    true,
			},
		},
	}
}

func (r *scorecardRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *scorecardRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state scorecardRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attrs := r.buildRuleAttributes(&state)
	ruleType := datadogV2.RULETYPE_RULE

	body := datadogV2.CreateRuleRequest{
		Data: &datadogV2.CreateRuleRequestData{
			Attributes: &attrs,
			Type:       &ruleType,
		},
	}

	resp, _, err := r.Api.CreateScorecardRule(r.Auth, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating scorecard rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	data := resp.GetData()
	r.updateState(&state, data.GetId(), data.GetAttributes())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *scorecardRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state scorecardRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	optParams := datadogV2.NewListScorecardRulesOptionalParameters()
	optParams.WithFilterRuleId(id)
	optParams.WithPageSize(1)

	resp, _, err := r.Api.ListScorecardRules(r.Auth, *optParams)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading scorecard rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	data := resp.GetData()
	if len(data) == 0 {
		response.State.RemoveResource(ctx)
		return
	}

	item := data[0]
	r.updateState(&state, item.GetId(), item.GetAttributes())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *scorecardRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state scorecardRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the ID from the current state since it's computed
	var currentState scorecardRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &currentState)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := currentState.ID.ValueString()

	attrs := r.buildRuleAttributes(&state)
	ruleType := datadogV2.RULETYPE_RULE

	body := datadogV2.UpdateRuleRequest{
		Data: &datadogV2.UpdateRuleRequestData{
			Attributes: &attrs,
			Type:       &ruleType,
		},
	}

	resp, _, err := r.Api.UpdateScorecardRule(r.Auth, id, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating scorecard rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	data := resp.GetData()
	r.updateState(&state, data.GetId(), data.GetAttributes())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *scorecardRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state scorecardRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	httpResp, err := r.Api.DeleteScorecardRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting scorecard rule"))
		return
	}
}

func (r *scorecardRuleResource) buildRuleAttributes(state *scorecardRuleModel) datadogV2.RuleAttributes {
	attrs := datadogV2.RuleAttributes{}

	attrs.SetName(state.Name.ValueString())
	attrs.SetScorecardName(state.ScorecardName.ValueString())
	attrs.SetEnabled(state.Enabled.ValueBool())

	if !state.Description.IsNull() {
		attrs.SetDescription(state.Description.ValueString())
	}
	if !state.Owner.IsNull() {
		attrs.SetOwner(state.Owner.ValueString())
	}

	// Convert string level to int32 — validated by schema OneOf("1","2","3")
	if !state.Level.IsNull() {
		v, _ := strconv.ParseInt(state.Level.ValueString(), 10, 32)
		attrs.SetLevel(int32(v))
	}

	return attrs
}

func (r *scorecardRuleResource) updateState(state *scorecardRuleModel, id string, attrs datadogV2.RuleAttributes) {
	state.ID = types.StringValue(id)

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
}
