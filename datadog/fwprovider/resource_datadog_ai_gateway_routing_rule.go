package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &aiGatewayRoutingRuleResource{}
	_ resource.ResourceWithImportState = &aiGatewayRoutingRuleResource{}
)

// The API's defaults for values a configuration may leave unset. State keeps them unset when the API returns the default.
const (
	aiGatewayRoutingRuleDefaultWeight              = int64(1)
	aiGatewayRoutingRuleDefaultConfidenceThreshold = 0.5
)

type aiGatewayRoutingRuleResource struct {
	Api  *datadogV2.AIGatewayApi
	Auth context.Context
}

type aiGatewayRoutingRuleModel struct {
	ID      types.String               `tfsdk:"id"`
	Slug    types.String               `tfsdk:"slug"`
	RouteID types.Int64                `tfsdk:"route_id"`
	Tiers   []aiGatewayRoutingRuleTier `tfsdk:"tiers"`
}

type aiGatewayRoutingRuleTier struct {
	Ordering   types.String                    `tfsdk:"ordering"`
	Classifier *aiGatewayRoutingRuleClassifier `tfsdk:"classifier"`
	Targets    []aiGatewayRoutingRuleTarget    `tfsdk:"targets"`
}

type aiGatewayRoutingRuleClassifier struct {
	Provider            types.String  `tfsdk:"provider"`
	Model               types.String  `tfsdk:"model"`
	Context             types.String  `tfsdk:"context"`
	ConfidenceThreshold types.Float64 `tfsdk:"confidence_threshold"`
}

type aiGatewayRoutingRuleTarget struct {
	Provider       types.String `tfsdk:"provider"`
	Model          types.String `tfsdk:"model"`
	RoutingRule    types.String `tfsdk:"routing_rule"`
	Weight         types.Int64  `tfsdk:"weight"`
	ReasoningLevel types.String `tfsdk:"reasoning_level"`
	Rubric         types.String `tfsdk:"rubric"`
	Default        types.Bool   `tfsdk:"default"`
}

func NewAiGatewayRoutingRuleResource() resource.Resource {
	return &aiGatewayRoutingRuleResource{}
}

func (r *aiGatewayRoutingRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewAIGatewayApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *aiGatewayRoutingRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ai_gateway_routing_rule"
}

func (r *aiGatewayRoutingRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AI Gateway routing rule resource. A routing rule is a named route that sends requests to ordered tiers of " +
			"destinations: the gateway tries the first tier, then falls back to the next. Requests select a rule by its slug. This resource is in Preview.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"slug": schema.StringAttribute{
				Description: "Organization-unique slug of the routing rule, in lowercase kebab case. Changing it replaces the routing rule.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"route_id": schema.Int64Attribute{
				Description: "Numeric ID of the routing rule. Restriction policies on the routing rule use it as `ai-gateway-route:<route_id>`.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"tiers": schema.ListNestedAttribute{
				Description: "Ordered fallback tiers. The gateway routes to a destination in the first tier and falls back to later tiers.",
				Required:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ordering": schema.StringAttribute{
							Description: "How the tier orders its targets. `configured` tries them in the order listed. `weighted_affinity` spreads requests by " +
								"each target's `weight`. `classified`, which lets a decision model pick one target by its `rubric`, is allowed only on the first tier.",
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(string(datadogV2.AIGATEWAYROUTINGRULETIERORDERING_CONFIGURED)),
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(datadogV2.AIGATEWAYROUTINGRULETIERORDERING_CONFIGURED),
									string(datadogV2.AIGATEWAYROUTINGRULETIERORDERING_WEIGHTED_AFFINITY),
									string(datadogV2.AIGATEWAYROUTINGRULETIERORDERING_CLASSIFIED),
								),
							},
						},
						"classifier": schema.SingleNestedAttribute{
							Description: "Decision model of a `classified` tier. Required on a `classified` tier and not allowed on others.",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"provider": schema.StringAttribute{
									Description: "Slug of the provider that serves the decision model. It must be an OpenAI, Baseten, or TypeSafe provider.",
									Required:    true,
								},
								"model": schema.StringAttribute{
									Description: "Canonical name of the decision model on that provider.",
									Required:    true,
								},
								"context": schema.StringAttribute{
									Description: "Prompt that tells the decision model about the rule's traffic, up to 2000 characters.",
									Optional:    true,
									Validators:  []validator.String{stringvalidator.LengthBetween(1, 2000)},
								},
								"confidence_threshold": schema.Float64Attribute{
									Description: "Minimum confidence that the picked target must reach. If the confidence is lower, the request goes to the default target.",
									Optional:    true,
									Computed:    true,
									Default:     float64default.StaticFloat64(aiGatewayRoutingRuleDefaultConfidenceThreshold),
									Validators:  []validator.Float64{float64validator.Between(0, 1)},
								},
							},
						},
						"targets": schema.ListNestedAttribute{
							Description: "Destinations in the tier, in order. A target is either a model, set with `provider` and `model`, or another routing rule, " +
								"set with `routing_rule`. A routing rule target must be the only target in its tier.",
							Required:   true,
							Validators: []validator.List{listvalidator.SizeAtLeast(1)},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"provider": schema.StringAttribute{
										Description: "Slug of the provider to route to, or `*` to route to any provider that serves the model. Required for a model target.",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("routing_rule")),
										},
									},
									"model": schema.StringAttribute{
										Description: "Model to route to. Required for a model target.",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
											stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("provider")),
										},
									},
									"routing_rule": schema.StringAttribute{
										Description: "Slug of another routing rule to route to.",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
											stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("model")),
										},
									},
									"weight": schema.Int64Attribute{
										Description: "Relative weight of the target in a `weighted_affinity` tier. The API uses 1 when it's unset. " +
											"Not allowed in other tiers.",
										Optional:   true,
										Validators: []validator.Int64{int64validator.Between(1, 10000)},
									},
									"reasoning_level": schema.StringAttribute{
										Description: "Reasoning effort override for a model target. When unset, the caller's and provider's defaults apply.",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("none", "minimal", "low", "medium", "high", "xhigh", "max"),
										},
									},
									"rubric": schema.StringAttribute{
										Description: "Description of when the decision model should pick this target, up to 300 characters. Required on every target of a `classified` tier " +
											"except the default one, and not allowed in other tiers.",
										Optional:   true,
										Validators: []validator.String{stringvalidator.LengthBetween(1, 300)},
									},
									"default": schema.BoolAttribute{
										Description: "Marks the default target of a `classified` tier, which receives requests that don't match any rubric. Each `classified` tier must have exactly one default target.",
										Optional:    true,
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

func (r *aiGatewayRoutingRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("slug"), request.ID)...)
}

func (r *aiGatewayRoutingRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state aiGatewayRoutingRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetAiGatewayRoutingRule(r.Auth, state.Slug.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AI Gateway routing rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	updateAiGatewayRoutingRuleState(&state, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayRoutingRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state aiGatewayRoutingRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	tiers, diags := buildAiGatewayRoutingRuleTierRequests(state.Tiers)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	attributes := datadogV2.NewAiGatewayRoutingRuleCreateAttributes(state.Slug.ValueString(), tiers)
	body := datadogV2.NewAiGatewayRoutingRuleCreateRequest(*datadogV2.NewAiGatewayRoutingRuleCreateData(*attributes, datadogV2.AIGATEWAYROUTINGRULETYPE_ROUTING_RULES))

	resp, _, err := r.Api.CreateAiGatewayRoutingRule(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating AI Gateway routing rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	updateAiGatewayRoutingRuleState(&state, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayRoutingRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state aiGatewayRoutingRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := buildAiGatewayRoutingRuleReplaceRequest(&state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateAiGatewayRoutingRule(r.Auth, state.Slug.ValueString(), *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating AI Gateway routing rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	updateAiGatewayRoutingRuleState(&state, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Delete doesn't send `force`, so deleting a routing rule that other rules or policies still use fails with 409 instead of rewriting them.
func (r *aiGatewayRoutingRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state aiGatewayRoutingRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteAiGatewayRoutingRule(r.Auth, state.Slug.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting AI Gateway routing rule"))
	}
}

// buildAiGatewayRoutingRuleReplaceRequest replaces every tier. Targets are sent without their saved IDs, so the API saves
// exactly the configured targets and drops the ones removed from the configuration; an ID would instead keep that target's
// saved settings for any field left unset.
func buildAiGatewayRoutingRuleReplaceRequest(state *aiGatewayRoutingRuleModel) (*datadogV2.AiGatewayRoutingRuleReplaceRequest, diag.Diagnostics) {
	tiers, diags := buildAiGatewayRoutingRuleTierRequests(state.Tiers)
	if diags.HasError() {
		return nil, diags
	}
	attributes := datadogV2.NewAiGatewayRoutingRuleReplaceAttributes()
	attributes.SetTiers(tiers)
	data := datadogV2.NewAiGatewayRoutingRuleReplaceData(*attributes, state.Slug.ValueString(), datadogV2.AIGATEWAYROUTINGRULETYPE_ROUTING_RULES)
	return datadogV2.NewAiGatewayRoutingRuleReplaceRequest(*data), diags
}

// buildAiGatewayRoutingRuleTierRequests sends ordering and classifier on every tier, with null for no classifier: on
// update, the API keeps a tier's saved ordering and classifier when they're omitted.
func buildAiGatewayRoutingRuleTierRequests(tiers []aiGatewayRoutingRuleTier) ([]datadogV2.AiGatewayRoutingRuleTierRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	requests := make([]datadogV2.AiGatewayRoutingRuleTierRequest, 0, len(tiers))
	for _, tier := range tiers {
		targets := make([]datadogV2.AiGatewayRoutingRuleTargetRequest, 0, len(tier.Targets))
		for _, target := range tier.Targets {
			targets = append(targets, *buildAiGatewayRoutingRuleTargetRequest(target))
		}
		request := datadogV2.NewAiGatewayRoutingRuleTierRequest(targets)

		ordering, err := datadogV2.NewAiGatewayNullableRoutingRuleTierOrderingFromValue(tier.Ordering.ValueString())
		if err != nil {
			diags.AddError("invalid ordering", err.Error())
			return nil, diags
		}
		request.Ordering = *datadogV2.NewNullableAiGatewayNullableRoutingRuleTierOrdering(ordering)

		if tier.Classifier == nil {
			request.Classifier = *datadogV2.NewNullableAiGatewayRoutingClassifierRequest(nil)
		} else {
			classifier := datadogV2.NewAiGatewayRoutingClassifierRequest()
			classifier.SetProvider(tier.Classifier.Provider.ValueString())
			classifier.SetModel(tier.Classifier.Model.ValueString())
			if !tier.Classifier.Context.IsNull() {
				classifier.SetContext(tier.Classifier.Context.ValueString())
			}
			classifier.ConfidenceThreshold = *datadog.NewNullableFloat64(tier.Classifier.ConfidenceThreshold.ValueFloat64Pointer())
			request.Classifier = *datadogV2.NewNullableAiGatewayRoutingClassifierRequest(classifier)
		}
		requests = append(requests, *request)
	}
	return requests, diags
}

func buildAiGatewayRoutingRuleTargetRequest(target aiGatewayRoutingRuleTarget) *datadogV2.AiGatewayRoutingRuleTargetRequest {
	if !target.RoutingRule.IsNull() {
		request := datadogV2.NewAiGatewayRoutingRuleTargetRequest(datadogV2.AIGATEWAYROUTINGRULETARGETTYPE_ROUTING_RULE)
		request.SetSlug(target.RoutingRule.ValueString())
		return request
	}

	request := datadogV2.NewAiGatewayRoutingRuleTargetRequest(datadogV2.AIGATEWAYROUTINGRULETARGETTYPE_MODEL)
	request.SetProvider(target.Provider.ValueString())
	request.SetModel(target.Model.ValueString())
	if !target.Weight.IsNull() {
		request.Weight = *datadog.NewNullableInt64(target.Weight.ValueInt64Pointer())
	}
	if !target.ReasoningLevel.IsNull() {
		level := datadogV2.AiGatewayRoutingReasoningLevel(target.ReasoningLevel.ValueString())
		request.ReasoningLevel = *datadogV2.NewNullableAiGatewayRoutingReasoningLevel(&level)
	}
	if !target.Rubric.IsNull() {
		request.Rubric = *datadog.NewNullableString(target.Rubric.ValueStringPointer())
	}
	if !target.Default.IsNull() {
		request.Default = *datadog.NewNullableBool(target.Default.ValueBoolPointer())
	}
	return request
}

// updateAiGatewayRoutingRuleState sets state from the API response. The API returns weight 1 for a target that left it
// unset and omits `default = false`. Unless the prior state at the same position set them, both read back as unset, so a
// configuration or import that leaves them unset doesn't show a diff.
func updateAiGatewayRoutingRuleState(state *aiGatewayRoutingRuleModel, data datadogV2.AiGatewayRoutingRuleResource) {
	attributes := data.GetAttributes()
	prior := state.Tiers

	tiers := make([]aiGatewayRoutingRuleTier, 0, len(attributes.GetTiers()))
	for tierIndex, tier := range attributes.GetTiers() {
		var priorTier *aiGatewayRoutingRuleTier
		if tierIndex < len(prior) {
			priorTier = &prior[tierIndex]
		}

		tierModel := aiGatewayRoutingRuleTier{
			Ordering: types.StringValue(string(tier.GetOrdering())),
			Targets:  make([]aiGatewayRoutingRuleTarget, 0, len(tier.GetTargets())),
		}
		if classifier, ok := tier.GetClassifierOk(); ok && classifier != nil {
			tierModel.Classifier = &aiGatewayRoutingRuleClassifier{
				Provider:            types.StringValue(classifier.GetProvider()),
				Model:               types.StringValue(classifier.GetModel()),
				Context:             types.StringNull(),
				ConfidenceThreshold: types.Float64Value(classifier.GetConfidenceThreshold()),
			}
			if context := classifier.GetContext(); context != "" {
				tierModel.Classifier.Context = types.StringValue(context)
			}
		}

		for targetIndex, target := range tier.GetTargets() {
			var priorTarget *aiGatewayRoutingRuleTarget
			if priorTier != nil && targetIndex < len(priorTier.Targets) {
				priorTarget = &priorTier.Targets[targetIndex]
			}
			tierModel.Targets = append(tierModel.Targets, aiGatewayRoutingRuleTargetState(target, priorTarget))
		}
		tiers = append(tiers, tierModel)
	}
	state.Tiers = tiers

	state.Slug = types.StringValue(data.GetId())
	state.ID = types.StringValue(data.GetId())
	state.RouteID = types.Int64Value(attributes.GetRouteId())
}

func aiGatewayRoutingRuleTargetState(target datadogV2.AiGatewayRoutingRuleTarget, prior *aiGatewayRoutingRuleTarget) aiGatewayRoutingRuleTarget {
	model := aiGatewayRoutingRuleTarget{
		Provider:       types.StringNull(),
		Model:          types.StringNull(),
		RoutingRule:    types.StringNull(),
		Weight:         types.Int64Null(),
		ReasoningLevel: types.StringNull(),
		Rubric:         types.StringNull(),
		Default:        types.BoolNull(),
	}
	if target.GetType() == datadogV2.AIGATEWAYROUTINGRULETARGETTYPE_ROUTING_RULE {
		if slug := target.GetSlug(); slug != "" {
			model.RoutingRule = types.StringValue(slug)
		}
		return model
	}

	if provider := target.GetProvider(); provider != "" {
		model.Provider = types.StringValue(provider)
	}
	if name := target.GetModel(); name != "" {
		model.Model = types.StringValue(name)
	}
	if weight, ok := target.GetWeightOk(); ok && weight != nil {
		if *weight != aiGatewayRoutingRuleDefaultWeight || prior != nil && !prior.Weight.IsNull() {
			model.Weight = types.Int64Value(*weight)
		}
	}
	if level, ok := target.GetReasoningLevelOk(); ok && level != nil {
		model.ReasoningLevel = types.StringValue(string(*level))
	}
	if rubric := target.GetRubric(); rubric != "" {
		model.Rubric = types.StringValue(rubric)
	}
	if target.GetDefault() {
		model.Default = types.BoolValue(true)
	} else if prior != nil && !prior.Default.IsNull() && !prior.Default.ValueBool() {
		model.Default = types.BoolValue(false)
	}
	return model
}
