package fwprovider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure      = &aiGatewayBudgetResource{}
	_ resource.ResourceWithImportState    = &aiGatewayBudgetResource{}
	_ resource.ResourceWithValidateConfig = &aiGatewayBudgetResource{}
)

// aiGatewayBudgetTagKeyPattern matches the form the API stores a tag key in: lowercase, trimmed, and without the
// x-dd-tag- header prefix. The API rewrites any other key into this form, and because tags and selector replace the
// budget, a key in another form would plan a replacement on every run.
var aiGatewayBudgetTagKeyPattern = regexp.MustCompile(`^[^A-Z\s]([^A-Z]*[^A-Z\s])?$`)

const aiGatewayBudgetTagHeaderPrefix = "x-dd-tag-"

type aiGatewayBudgetResource struct {
	Api  *datadogV2.AIGatewayApi
	Auth context.Context
}

type aiGatewayBudgetModel struct {
	ID                   types.String  `tfsdk:"id"`
	Name                 types.String  `tfsdk:"name"`
	CapacityMicrocents   types.Int64   `tfsdk:"capacity_microcents"`
	WindowPolicy         types.String  `tfsdk:"window_policy"`
	Enforcement          types.String  `tfsdk:"enforcement"`
	WarningThreshold     types.Float64 `tfsdk:"warning_threshold"`
	Enabled              types.Bool    `tfsdk:"enabled"`
	ApplyToCurrentPeriod types.Bool    `tfsdk:"apply_to_current_period"`
	TeamID               types.String  `tfsdk:"team_id"`
	UserID               types.String  `tfsdk:"user_id"`
	Tags                 types.Map     `tfsdk:"tags"`
	Selector             types.Set     `tfsdk:"selector"`
	MonitorID            types.Int64   `tfsdk:"monitor_id"`
}

type aiGatewayBudgetSelectorModel struct {
	Dim      types.String `tfsdk:"dim"`
	PerValue types.Bool   `tfsdk:"per_value"`
	Values   types.Set    `tfsdk:"values"`
}

var aiGatewayBudgetSelectorAttrTypes = map[string]attr.Type{
	"dim":       types.StringType,
	"per_value": types.BoolType,
	"values":    types.SetType{ElemType: types.StringType},
}

func NewAiGatewayBudgetResource() resource.Resource {
	return &aiGatewayBudgetResource{}
}

func (r *aiGatewayBudgetResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewAIGatewayApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *aiGatewayBudgetResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ai_gateway_budget"
}

func (r *aiGatewayBudgetResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AI Gateway budget resource. A budget caps spend over a daily, weekly, or monthly UTC period for the requests that match its scope. " +
			"This resource is in Preview.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of the budget.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(`^\S(.*\S)?$`), "must not have leading or trailing whitespace"),
				},
			},
			"capacity_microcents": schema.Int64Attribute{
				Description: "Spend allowed per period, in microcents: one US dollar is 100,000,000 microcents. A change applies from the next period unless `apply_to_current_period` is `true`.",
				Required:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"window_policy": schema.StringAttribute{
				Description: "UTC accounting period of the budget. Changing it replaces the budget.",
				Required:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayBudgetWindowPolicyFromValue),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enforcement": schema.StringAttribute{
				Description: "Whether the budget rejects requests over its capacity (`BLOCK`) or only notifies (`NOTIFY`).",
				Required:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayBudgetEnforcementFromValue),
				},
			},
			"warning_threshold": schema.Float64Attribute{
				Description: "Fraction of capacity that triggers a warning from the budget's monitor. Must be greater than 0 and at most 1.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(0.9),
				Validators: []validator.Float64{
					float64validator.AtMost(1),
					aiGatewayPositiveFloat64Validator{},
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the budget is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"apply_to_current_period": schema.BoolAttribute{
				Description: "Whether a change to `capacity_microcents` also applies to the current period. When `false`, the current period keeps its capacity. Not supported for budgets with a `selector`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"team_id": schema.StringAttribute{
				Description: "UUID of a Datadog Team. When set, only requests attributed to this team count toward the budget. Changing it replaces the budget.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "UUID of a Datadog user. When set, only this user's requests count toward the budget. Changing it replaces the budget.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.MapAttribute{
				Description: "Scope tags, such as `provider`, `model`, or `ml_app`. A request counts toward the budget when it has all of these tags. Keys must be lowercase. Changing them replaces the budget.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Map{
					// The API returns no tags as {}, which maps to null, so an empty map would never match state.
					mapvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"selector": schema.SetNestedAttribute{
				Description: "Dimension rules that make this budget a template budget. Each combination of values for the dimensions with `per_value = true` gets its own budget instance, and each request counts toward the instance that matches its values. Changing it replaces the budget.",
				Optional:    true,
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dim": schema.StringAttribute{
							Description: "Dimension to match: `user_id`, `team_id`, `model`, `provider`, or `tag:<key>` with a lowercase key.",
							Required:    true,
						},
						"per_value": schema.BoolAttribute{
							Description: "Whether each matched value gets its own budget instance. When `false`, spend is pooled across all matched values.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"values": schema.SetAttribute{
							Description: "Values the dimension must have. Omit to match any value, including requests that don't have the dimension.",
							Optional:    true,
							ElementType: types.StringType,
							Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
						},
					},
				},
			},
			"monitor_id": schema.Int64Attribute{
				Description: "ID of the Datadog monitor created for the budget. Unset when the monitor wasn't created.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// ValidateConfig rejects tag keys and selector entries the API would rewrite on create. Both replace the budget, so a
// rewritten value would plan a replacement on every run.
func (r *aiGatewayBudgetResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config aiGatewayBudgetModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}
	if !config.Tags.IsNull() && !config.Tags.IsUnknown() {
		for key := range config.Tags.Elements() {
			if err := validateAiGatewayBudgetTagKey(key); err != nil {
				response.Diagnostics.AddAttributeError(path.Root("tags"), "invalid tag key", err.Error())
			}
		}
	}
	if config.Selector.IsNull() || config.Selector.IsUnknown() {
		return
	}
	var entries []aiGatewayBudgetSelectorModel
	response.Diagnostics.Append(config.Selector.ElementsAs(ctx, &entries, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.Dim.IsUnknown() || entry.PerValue.IsUnknown() || entry.Values.IsUnknown() {
			continue
		}
		dim := entry.Dim.ValueString()
		if err := validateAiGatewayBudgetSelectorDim(dim); err != nil {
			response.Diagnostics.AddAttributeError(path.Root("selector"), "invalid selector dim", err.Error())
			continue
		}
		if seen[dim] {
			response.Diagnostics.AddAttributeError(path.Root("selector"), "invalid selector", fmt.Sprintf("dim %q appears more than once", dim))
		}
		seen[dim] = true

		perValue := entry.PerValue.ValueBool()
		values := len(entry.Values.Elements())
		switch {
		case entry.Values.IsNull() && !perValue:
			response.Diagnostics.AddAttributeError(path.Root("selector"), "invalid selector",
				fmt.Sprintf("dim %q has no values and per_value = false, so it matches everything; remove it", dim))
		case values == 1 && perValue:
			response.Diagnostics.AddAttributeError(path.Root("selector"), "invalid selector",
				fmt.Sprintf("dim %q has a single value, which can't be split; set per_value = false", dim))
		}
	}
}

func validateAiGatewayBudgetSelectorDim(dim string) error {
	switch dim {
	case "user_id", "team_id", "model", "provider":
		return nil
	}
	key, isTag := strings.CutPrefix(dim, "tag:")
	if !isTag || key == "" {
		return fmt.Errorf("dim %q must be user_id, team_id, model, provider, or tag:<key>", dim)
	}
	if err := validateAiGatewayBudgetTagKey(key); err != nil {
		return fmt.Errorf("dim %q: %w", dim, err)
	}
	return nil
}

func validateAiGatewayBudgetTagKey(key string) error {
	if !aiGatewayBudgetTagKeyPattern.MatchString(key) || strings.HasPrefix(key, aiGatewayBudgetTagHeaderPrefix) {
		return fmt.Errorf("tag key %q must be lowercase, with no surrounding whitespace and no %s prefix", key, aiGatewayBudgetTagHeaderPrefix)
	}
	return nil
}

func (r *aiGatewayBudgetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *aiGatewayBudgetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state aiGatewayBudgetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway budget id must be a UUID")
		return
	}
	resp, httpResp, err := r.Api.GetAiGatewayBudget(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AI Gateway budget"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayBudgetState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayBudgetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state aiGatewayBudgetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes, diags := buildAiGatewayBudgetCreateAttributes(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewAiGatewayBudgetCreateRequest(*datadogV2.NewAiGatewayBudgetCreateData(*attributes, datadogV2.AIGATEWAYBUDGETTYPE_BUDGETS))

	resp, _, err := r.Api.CreateAiGatewayBudget(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating AI Gateway budget"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}
	data := resp.GetData()

	// Create has no enabled field, so a budget configured as disabled is created and then disabled.
	if !state.Enabled.ValueBool() && data.Attributes.GetEnabled() {
		update := datadogV2.NewAiGatewayBudgetUpdateAttributes()
		update.SetEnabled(false)
		updateBody := datadogV2.NewAiGatewayBudgetUpdateRequest(*datadogV2.NewAiGatewayBudgetUpdateData(*update, data.GetId(), datadogV2.AIGATEWAYBUDGETTYPE_BUDGETS))
		updated, _, err := r.Api.UpdateAiGatewayBudget(r.Auth, data.GetId(), *updateBody)
		if err != nil {
			// Keep the created budget in state so the next apply retries the update instead of creating a duplicate.
			response.Diagnostics.Append(updateAiGatewayBudgetState(ctx, &state, data)...)
			response.Diagnostics.Append(response.State.Set(ctx, &state)...)
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error disabling AI Gateway budget"))
			return
		}
		data = updated.GetData()
	}

	response.Diagnostics.Append(updateAiGatewayBudgetState(ctx, &state, data)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayBudgetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, prior aiGatewayBudgetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	response.Diagnostics.Append(request.State.Get(ctx, &prior)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway budget id must be a UUID")
		return
	}
	attributes := buildAiGatewayBudgetUpdateAttributes(&state, &prior)
	body := datadogV2.NewAiGatewayBudgetUpdateRequest(*datadogV2.NewAiGatewayBudgetUpdateData(*attributes, id, datadogV2.AIGATEWAYBUDGETTYPE_BUDGETS))

	resp, _, err := r.Api.UpdateAiGatewayBudget(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating AI Gateway budget"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayBudgetState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayBudgetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state aiGatewayBudgetModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway budget id must be a UUID")
		return
	}
	httpResp, err := r.Api.DeleteAiGatewayBudget(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting AI Gateway budget"))
	}
}

func updateAiGatewayBudgetState(ctx context.Context, state *aiGatewayBudgetModel, data datadogV2.AiGatewayBudgetResource) diag.Diagnostics {
	var diags diag.Diagnostics
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId().String())
	state.Name = types.StringValue(attributes.GetName())
	state.CapacityMicrocents = types.Int64Value(attributes.GetCapacityMicrocents())
	state.WindowPolicy = types.StringValue(string(attributes.GetWindowPolicy()))
	state.Enforcement = types.StringValue(string(attributes.GetEnforcement()))
	state.WarningThreshold = types.Float64Value(attributes.GetWarningThreshold())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	// apply_to_current_period only shapes the next update, so it is never read back. Import leaves it null.
	if state.ApplyToCurrentPeriod.IsNull() || state.ApplyToCurrentPeriod.IsUnknown() {
		state.ApplyToCurrentPeriod = types.BoolValue(false)
	}
	state.TeamID = aiGatewayBudgetUUIDValue(state.TeamID, attributes.TeamId)
	state.UserID = aiGatewayBudgetUUIDValue(state.UserID, attributes.UserId)
	if monitorID := attributes.MonitorId.Get(); monitorID != nil {
		state.MonitorID = types.Int64Value(*monitorID)
	} else {
		state.MonitorID = types.Int64Null()
	}

	// The API returns no tags as an empty object.
	if tags := attributes.GetTags(); len(tags) > 0 {
		value, d := types.MapValueFrom(ctx, types.StringType, tags)
		diags.Append(d...)
		state.Tags = value
	} else {
		state.Tags = types.MapNull(types.StringType)
	}

	selectorType := types.ObjectType{AttrTypes: aiGatewayBudgetSelectorAttrTypes}
	if selector := attributes.GetSelector(); len(selector) > 0 {
		entries := make([]aiGatewayBudgetSelectorModel, 0, len(selector))
		for _, entry := range selector {
			model := aiGatewayBudgetSelectorModel{
				Dim:      types.StringValue(entry.GetDim()),
				PerValue: types.BoolValue(entry.GetPerValue()),
				Values:   types.SetNull(types.StringType),
			}
			if values := entry.GetValues(); len(values) > 0 {
				value, d := types.SetValueFrom(ctx, types.StringType, values)
				diags.Append(d...)
				model.Values = value
			}
			entries = append(entries, model)
		}
		value, d := types.SetValueFrom(ctx, selectorType, entries)
		diags.Append(d...)
		state.Selector = value
	} else {
		state.Selector = types.SetNull(selectorType)
	}

	return diags
}

func buildAiGatewayBudgetCreateAttributes(ctx context.Context, state *aiGatewayBudgetModel) (*datadogV2.AiGatewayBudgetCreateAttributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewAiGatewayBudgetCreateAttributes(
		state.CapacityMicrocents.ValueInt64(),
		datadogV2.AiGatewayBudgetEnforcement(state.Enforcement.ValueString()),
		state.Name.ValueString(),
		datadogV2.AiGatewayBudgetWindowPolicy(state.WindowPolicy.ValueString()),
	)
	attributes.SetWarningThreshold(state.WarningThreshold.ValueFloat64())

	for _, scope := range []struct {
		name  string
		value types.String
		set   func(uuid.UUID)
	}{
		{"team_id", state.TeamID, attributes.SetTeamId},
		{"user_id", state.UserID, attributes.SetUserId},
	} {
		if scope.value.IsNull() {
			continue
		}
		id, err := uuid.Parse(scope.value.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root(scope.name), "invalid "+scope.name, scope.name+" must be a UUID")
			continue
		}
		scope.set(id)
	}
	if !state.Tags.IsNull() {
		tags := map[string]string{}
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}
	if !state.Selector.IsNull() {
		var entries []aiGatewayBudgetSelectorModel
		diags.Append(state.Selector.ElementsAs(ctx, &entries, false)...)
		selector := make([]datadogV2.AiGatewayBudgetSelectorEntry, 0, len(entries))
		for _, entry := range entries {
			item := datadogV2.NewAiGatewayBudgetSelectorEntry(entry.Dim.ValueString(), entry.PerValue.ValueBool())
			if !entry.Values.IsNull() {
				var values []string
				diags.Append(entry.Values.ElementsAs(ctx, &values, false)...)
				item.SetValues(values)
			}
			selector = append(selector, *item)
		}
		attributes.SetSelector(selector)
	}

	return attributes, diags
}

// buildAiGatewayBudgetUpdateAttributes sends only the fields that changed. A name or warning_threshold change makes
// the API update the budget's monitor in the same transaction, so sending them unchanged would tie every update to
// the monitor.
func buildAiGatewayBudgetUpdateAttributes(state, prior *aiGatewayBudgetModel) *datadogV2.AiGatewayBudgetUpdateAttributes {
	attributes := datadogV2.NewAiGatewayBudgetUpdateAttributes()

	if !state.Name.Equal(prior.Name) {
		attributes.SetName(state.Name.ValueString())
	}
	if !state.CapacityMicrocents.Equal(prior.CapacityMicrocents) {
		attributes.SetCapacityMicrocents(state.CapacityMicrocents.ValueInt64())
		if state.ApplyToCurrentPeriod.ValueBool() {
			attributes.SetApplyToCurrentPeriod(true)
		}
	}
	if !state.Enforcement.Equal(prior.Enforcement) {
		attributes.SetEnforcement(datadogV2.AiGatewayBudgetEnforcement(state.Enforcement.ValueString()))
	}
	if !state.WarningThreshold.Equal(prior.WarningThreshold) {
		attributes.SetWarningThreshold(state.WarningThreshold.ValueFloat64())
	}
	if !state.Enabled.Equal(prior.Enabled) {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}

	return attributes
}

// aiGatewayBudgetUUIDValue keeps the prior spelling of the same UUID, since the API returns UUIDs in lowercase and a
// configured uppercase UUID would otherwise differ from the plan.
func aiGatewayBudgetUUIDValue(prior types.String, value *uuid.UUID) types.String {
	if value == nil {
		return types.StringNull()
	}
	if !prior.IsNull() && !prior.IsUnknown() && strings.EqualFold(prior.ValueString(), value.String()) {
		return prior
	}
	return types.StringValue(value.String())
}
