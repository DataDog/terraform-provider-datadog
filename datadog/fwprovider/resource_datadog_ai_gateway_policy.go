package fwprovider

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &aiGatewayPolicyResource{}
	_ resource.ResourceWithImportState = &aiGatewayPolicyResource{}
)

type aiGatewayPolicyResource struct {
	Api  *datadogV2.AIGatewayApi
	Auth context.Context
}

type aiGatewayPolicyModel struct {
	ID       types.String          `tfsdk:"id"`
	Inbound  types.String          `tfsdk:"inbound"`
	UserID   types.String          `tfsdk:"user_id"`
	Priority types.Int64           `tfsdk:"priority"`
	Tiers    []aiGatewayPolicyTier `tfsdk:"tiers"`
}

type aiGatewayPolicyTier struct {
	Targets []aiGatewayPolicyTarget `tfsdk:"targets"`
}

type aiGatewayPolicyTarget struct {
	Provider types.String `tfsdk:"provider"`
	Model    types.String `tfsdk:"model"`
}

func NewAiGatewayPolicyResource() resource.Resource {
	return &aiGatewayPolicyResource{}
}

func (r *aiGatewayPolicyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewAIGatewayApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *aiGatewayPolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ai_gateway_policy"
}

func (r *aiGatewayPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AI Gateway routing policy resource. A policy routes requests that match an inbound pattern to ordered tiers of " +
			"provider and model targets: the gateway tries the first tier, then falls back to the next. Each inbound pattern has one policy " +
			"per scope, organization-wide or for one user. This resource is in Preview.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"inbound": schema.StringAttribute{
				Description: "Inbound request pattern the policy applies to, such as `coding-agent` or `gpt-5*`. It must not contain whitespace, and `*` is allowed " +
					"only as a trailing wildcard. Changing it replaces the policy.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "UUID of a Datadog user. When set, the policy applies only to that user's requests; otherwise it applies to the whole organization. " +
					"Changing it replaces the policy.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"priority": schema.Int64Attribute{
				Description: "Routing priority of the policy.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"tiers": schema.ListNestedAttribute{
				Description: "Ordered fallback tiers. The gateway routes to a target in the first tier and falls back to later tiers.",
				Required:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"targets": schema.SetNestedAttribute{
							Description: "Targets in the tier.",
							Required:    true,
							Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"provider": schema.StringAttribute{
										Description: "Slug of the provider to route to, or `*` to fan out to every provider. When unset, the gateway resolves the provider from the model.",
										Optional:    true,
										Validators:  []validator.String{stringvalidator.LengthBetween(1, 256)},
									},
									"model": schema.StringAttribute{
										Description: "Model name, or a pattern with a trailing `*`.",
										Required:    true,
										Validators:  []validator.String{stringvalidator.LengthBetween(1, 256)},
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

// The import ID is the inbound pattern for an organization-wide policy, or `<user_id>:<inbound>` for a user's policy.
func (r *aiGatewayPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	userID, inbound := parseAiGatewayPolicyID(request.ID)
	if inbound == "" {
		response.Diagnostics.AddError("invalid import ID", "expected `<inbound>` or `<user_id>:<inbound>`")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("inbound"), inbound)...)
	if userID != nil {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_id"), userID.String())...)
	}
}

func (r *aiGatewayPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state aiGatewayPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	userID, diags := aiGatewayPolicyUserID(state.UserID)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	policy, err := r.findPolicy(userID, state.Inbound.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AI Gateway policy"))
		return
	}
	if policy == nil {
		response.State.RemoveResource(ctx)
		return
	}

	updateAiGatewayPolicyState(&state, *policy)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state aiGatewayPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	userID, diags := aiGatewayPolicyUserID(state.UserID)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	// The API upserts by inbound pattern, so creating over an existing policy would silently take it over.
	existing, err := r.findPolicy(userID, state.Inbound.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error checking for an existing AI Gateway policy"))
		return
	}
	if existing != nil {
		response.Diagnostics.AddAttributeError(path.Root("inbound"), "AI Gateway policy already exists",
			fmt.Sprintf("a policy for inbound %q already exists in this scope; import it with `terraform import` instead of creating it", state.Inbound.ValueString()))
		return
	}

	r.upsert(&state, userID, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state aiGatewayPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	userID, diags := aiGatewayPolicyUserID(state.UserID)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	r.upsert(&state, userID, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state aiGatewayPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	userID, diags := aiGatewayPolicyUserID(state.UserID)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	params := datadogV2.NewDeleteAiGatewayPolicyOptionalParameters()
	if userID != nil {
		params = params.WithUserId(*userID)
	}
	httpResp, err := r.Api.DeleteAiGatewayPolicy(r.Auth, state.Inbound.ValueString(), *params)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting AI Gateway policy"))
	}
}

func (r *aiGatewayPolicyResource) upsert(state *aiGatewayPolicyModel, userID *uuid.UUID, diags *diag.Diagnostics) {
	params := datadogV2.NewUpsertAiGatewayPolicyOptionalParameters()
	if userID != nil {
		params = params.WithUserId(*userID)
	}
	body := datadogV2.NewAiGatewayPolicyRequest(*datadogV2.NewAiGatewayPolicyRequestData(*buildAiGatewayPolicyRequestAttributes(state), datadogV2.AIGATEWAYPOLICYTYPE_ROUTING_DOCUMENTS))

	resp, _, err := r.Api.UpsertAiGatewayPolicy(r.Auth, *body, *params)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error saving AI Gateway policy"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("response contains unparsed object", err.Error())
		return
	}
	updateAiGatewayPolicyState(state, resp.GetData())
}

// findPolicy returns the policy for inbound in the scope, or nil when there is none. It lists the scope's policies
// rather than calling GetAiGatewayPolicy, which does not return 404 for a missing policy.
func (r *aiGatewayPolicyResource) findPolicy(userID *uuid.UUID, inbound string) (*datadogV2.AiGatewayPolicyResource, error) {
	params := datadogV2.NewListAiGatewayPoliciesOptionalParameters()
	if userID != nil {
		params = params.WithUserId(*userID)
	}
	resp, _, err := r.Api.ListAiGatewayPolicies(r.Auth, *params)
	if err != nil {
		return nil, err
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return nil, err
	}
	for _, policy := range resp.GetData() {
		if policy.GetId() == inbound {
			return &policy, nil
		}
	}
	return nil, nil
}

// buildAiGatewayPolicyRequestAttributes numbers the tiers by position, from 1: the API merges tiers that share a
// number, so sending them unnumbered would collapse every tier into one, and the database rejects tier 0.
func buildAiGatewayPolicyRequestAttributes(state *aiGatewayPolicyModel) *datadogV2.AiGatewayPolicyRequestAttributes {
	tiers := make([]datadogV2.AiGatewayPolicyTierRequest, 0, len(state.Tiers))
	for i, tier := range state.Tiers {
		targets := make([]datadogV2.AiGatewayPolicyTargetRequest, 0, len(tier.Targets))
		for _, target := range tier.Targets {
			request := datadogV2.NewAiGatewayPolicyTargetRequest()
			request.SetModel(target.Model.ValueString())
			if !target.Provider.IsNull() {
				request.SetProvider(target.Provider.ValueString())
			}
			targets = append(targets, *request)
		}
		tierRequest := datadogV2.NewAiGatewayPolicyTierRequest(targets)
		tierRequest.SetPriority(int64(i + 1))
		tiers = append(tiers, *tierRequest)
	}

	attributes := datadogV2.NewAiGatewayPolicyRequestAttributes(state.Inbound.ValueString(), tiers)
	attributes.SetPriority(state.Priority.ValueInt64())
	return attributes
}

// updateAiGatewayPolicyState sets state from the API response. The response doesn't include tier numbers, so tiers are
// ordered by their oldest saved target: every write re-inserts a policy's targets tier by tier, so target IDs grow with
// the tier number. A tier reordered outside Terraform therefore shows as a diff instead of being hidden.
func updateAiGatewayPolicyState(state *aiGatewayPolicyModel, data datadogV2.AiGatewayPolicyResource) {
	attributes := data.GetAttributes()

	apiTiers := append([]datadogV2.AiGatewayPolicyTier(nil), attributes.GetTiers()...)
	sort.SliceStable(apiTiers, func(i, j int) bool {
		return minAiGatewayPolicyTargetID(apiTiers[i]) < minAiGatewayPolicyTargetID(apiTiers[j])
	})

	tiers := make([]aiGatewayPolicyTier, 0, len(apiTiers))
	for _, tier := range apiTiers {
		tierModel := aiGatewayPolicyTier{Targets: make([]aiGatewayPolicyTarget, 0, len(tier.GetTargets()))}
		for _, target := range tier.GetTargets() {
			targetModel := aiGatewayPolicyTarget{Provider: types.StringNull(), Model: types.StringNull()}
			if provider := target.Provider.Get(); provider != nil {
				targetModel.Provider = types.StringValue(*provider)
			}
			if model, ok := target.GetModelOk(); ok && model != nil && *model != "" {
				targetModel.Model = types.StringValue(*model)
			}
			tierModel.Targets = append(tierModel.Targets, targetModel)
		}
		tiers = append(tiers, tierModel)
	}
	state.Tiers = tiers

	state.Inbound = types.StringValue(data.GetId())
	state.Priority = types.Int64Value(attributes.GetPriority())
	state.ID = types.StringValue(formatAiGatewayPolicyID(state.UserID, data.GetId()))
}

func minAiGatewayPolicyTargetID(tier datadogV2.AiGatewayPolicyTier) int64 {
	min := int64(math.MaxInt64)
	for _, target := range tier.GetTargets() {
		if target.GetId() < min {
			min = target.GetId()
		}
	}
	return min
}

func aiGatewayPolicyUserID(value types.String) (*uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	userID, err := uuid.Parse(value.ValueString())
	if err != nil {
		diags.AddAttributeError(path.Root("user_id"), "invalid user_id", "user_id must be a UUID")
		return nil, diags
	}
	return &userID, diags
}

func formatAiGatewayPolicyID(userID types.String, inbound string) string {
	if userID.IsNull() || userID.IsUnknown() {
		return inbound
	}
	return userID.ValueString() + ":" + inbound
}

// parseAiGatewayPolicyID splits an ID from formatAiGatewayPolicyID. An ID is user-scoped only when it starts with a UUID
// followed by a colon, so an organization-wide inbound pattern that contains a colon still parses.
func parseAiGatewayPolicyID(id string) (*uuid.UUID, string) {
	if prefix, inbound, ok := strings.Cut(id, ":"); ok {
		if userID, err := uuid.Parse(prefix); err == nil && len(prefix) == 36 {
			return &userID, inbound
		}
	}
	return nil, id
}
