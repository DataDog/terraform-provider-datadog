package fwprovider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
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
	CreatedAt     types.String `tfsdk:"created_at"`
	ModifiedAt    types.String `tfsdk:"modified_at"`
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Custom        types.Bool   `tfsdk:"custom"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Owner         types.String `tfsdk:"owner"`
	ScorecardName types.String `tfsdk:"scorecard_name"`
	Level         types.Int32  `tfsdk:"level"`
	ScopeQuery    types.String `tfsdk:"scope_query"`
}

func NewScorecardRuleResource() resource.Resource {
	return &scorecardRuleResource{}
}

func (r *scorecardRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetServiceScorecardsApiV2()
	r.Auth = providerData.Auth
}

func (r *scorecardRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "service_scorecard_rule"
}

func (r *scorecardRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Service Scorecard Rule resource. This can be used to create and manage Datadog Service Scorecard Rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of the rule.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 140),
				},
			},
			"description": schema.StringAttribute{
				Description: "Explanation of the rule.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "If enabled, the rule is calculated as part of the score",
				Optional:    true,
				Default:     booldefault.StaticBool(true),
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"custom": schema.BoolAttribute{
				Description: "Defines if the rule is a custom rule",
				Computed:    true,
			},
			"owner": schema.StringAttribute{
				Description: "Owner of the rule.",
				Optional:    true,
			},
			"scorecard_name": schema.StringAttribute{
				Description: "The scorecard name to which this rule must belong",
				Required:    true,
			},
			"level": schema.Int32Attribute{
				Description: "The criticality level of the rule",
				Optional:    true,
				Default:     int32default.StaticInt32(3),
				Computed:    true,
				Validators: []validator.Int32{
					int32validator.OneOf(int32(1), int32(2), int32(3)),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"scope_query": schema.StringAttribute{
				Description: "The scope query to apply to the rule",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^$|^([a-zA-Z0-9_-]+:\S+)(\s[a-zA-Z0-9_-]+:\S+)*$`),
						"Scope query must be a valid Datadog query (e.g. tier:critical team:team-a)",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Creation time of the rule",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Last modification time of the rule",
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
	attrs, diags := r.buildServiceScorecardRule(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	ruleReq := datadogV2.NewCreateRuleRequest()
	ruleReq.SetData(*datadogV2.NewCreateRuleRequestData())
	ruleReq.Data.SetAttributes(*attrs)

	res, _, err := r.Api.CreateScorecardRule(r.Auth, *ruleReq)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Scorecard Rule"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	state.ID = types.StringValue(res.Data.GetId())
	response.Diagnostics.Append(r.updateState(ctx, &state, res.Data.Attributes)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *scorecardRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state scorecardRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	optParams := datadogV2.NewListScorecardRulesOptionalParameters().WithFilterRuleId(state.ID.ValueString())
	res, _, err := r.Api.ListScorecardRules(r.Auth, *optParams)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Scorecard Rule"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	if len(res.Data) < 1 {
		response.Diagnostics.AddError(
			fmt.Sprintf("Scorecard Rule with ID %s not found", state.ID.ValueString()),
			"No rule was returned by the API.",
		)
		return
	}
	response.Diagnostics.Append(r.updateState(ctx, &state, res.Data[0].Attributes)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *scorecardRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state scorecardRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attrs, diags := r.buildServiceScorecardRule(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	ruleReq := datadogV2.NewUpdateRuleRequest()
	ruleReq.SetData(*datadogV2.NewUpdateRuleRequestData())
	ruleReq.Data.SetAttributes(*attrs)

	res, _, err := r.Api.UpdateScorecardRule(r.Auth, state.ID.ValueString(), *ruleReq)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Scorecard Rule"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, res.Data.Attributes)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *scorecardRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state scorecardRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteScorecardRule(r.Auth, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting Scorecard Rule"))
		return
	}
	if httpResp.StatusCode == 404 {
		return // already deleted
	}
}

func (r *scorecardRuleResource) buildServiceScorecardRule(ctx context.Context, state *scorecardRuleModel) (*datadogV2.RuleAttributes, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attrs := datadogV2.RuleAttributes{}
	if attrs.AdditionalProperties == nil {
		attrs.AdditionalProperties = make(map[string]interface{})
	}

	attrs.SetName(state.Name.ValueString())
	attrs.SetEnabled(state.Enabled.ValueBool())
	attrs.SetOwner(state.Owner.ValueString())
	attrs.SetScorecardName(state.ScorecardName.ValueString())
	attrs.AdditionalProperties["level"] = state.Level.ValueInt32()

	if !state.Description.IsNull() {
		attrs.SetDescription(state.Description.ValueString())
	}

	if !state.ScopeQuery.IsNull() {
		attrs.AdditionalProperties["scope_query"] = state.ScopeQuery.ValueString()
	}

	return &attrs, diags
}

func (r *scorecardRuleResource) updateState(ctx context.Context, state *scorecardRuleModel, attrs *datadogV2.RuleAttributes) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if createdAt, ok := attrs.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(createdAt.Format(time.RFC3339Nano))
	}

	if modifiedAt, ok := attrs.GetModifiedAtOk(); ok {
		state.ModifiedAt = types.StringValue(modifiedAt.Format(time.RFC3339Nano))
	}

	state.Name = types.StringPointerValue(attrs.Name)
	state.ScorecardName = types.StringPointerValue(attrs.ScorecardName)
	state.Enabled = types.BoolPointerValue(attrs.Enabled)

	if owner, ok := attrs.GetOwnerOk(); ok {
		state.Owner = types.StringPointerValue(owner)
	}

	if desc, ok := attrs.GetDescriptionOk(); ok {
		state.Description = types.StringPointerValue(desc)
	}

	if custom, ok := attrs.GetCustomOk(); ok {
		state.Custom = types.BoolPointerValue(custom)
	}

	if sq, ok := attrs.AdditionalProperties["scope_query"].(string); ok {
		state.ScopeQuery = types.StringValue(sq)
	}

	if level, ok := attrs.AdditionalProperties["level"].(int32); ok {
		state.Level = types.Int32Value(level)
	}

	return diags
}
