package fwprovider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure      = &aiGatewayRateLimitResource{}
	_ resource.ResourceWithImportState    = &aiGatewayRateLimitResource{}
	_ resource.ResourceWithValidateConfig = &aiGatewayRateLimitResource{}
)

// aiGatewayRateLimitTagKeyPattern matches the form the API stores a tag key in: lowercase and trimmed. The API
// rewrites any other key into this form, which would show as a change on every plan.
var aiGatewayRateLimitTagKeyPattern = regexp.MustCompile(`^[^A-Z\s]([^A-Z]*[^A-Z\s])?$`)

const aiGatewayRateLimitTagHeaderPrefix = "x-dd-tag-"

type aiGatewayRateLimitResource struct {
	Api  *datadogV2.AIGatewayApi
	Auth context.Context
}

type aiGatewayRateLimitModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Mode                  types.String `tfsdk:"mode"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	RequestsPerMinute     types.Int64  `tfsdk:"requests_per_minute"`
	InputTokensPerMinute  types.Int64  `tfsdk:"input_tokens_per_minute"`
	OutputTokensPerMinute types.Int64  `tfsdk:"output_tokens_per_minute"`
	TeamID                types.String `tfsdk:"team_id"`
	UserID                types.String `tfsdk:"user_id"`
	Tags                  types.Map    `tfsdk:"tags"`
}

func NewAiGatewayRateLimitResource() resource.Resource {
	return &aiGatewayRateLimitResource{}
}

func (r *aiGatewayRateLimitResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewAIGatewayApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *aiGatewayRateLimitResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ai_gateway_rate_limit"
}

func (r *aiGatewayRateLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	limitValidators := func(others ...string) []validator.Int64 {
		expressions := make([]path.Expression, 0, len(others))
		for _, other := range others {
			expressions = append(expressions, path.MatchRoot(other))
		}
		return []validator.Int64{
			int64validator.AtLeast(1),
			int64validator.AtLeastOneOf(expressions...),
		}
	}

	response.Schema = schema.Schema{
		Description: "Provides a Datadog AI Gateway rate limit resource. A rate limit caps requests or tokens per rolling minute for the requests that match its scope. " +
			"This resource is in Preview.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of the rate limit.",
				Required:    true,
			},
			"mode": schema.StringAttribute{
				Description: "Whether the rate limit rejects requests over the limit (`ENFORCE`) or only alerts (`ALERT`). Defaults to the server's default when unset.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayRateLimitModeFromValue),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rate limit is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"requests_per_minute": schema.Int64Attribute{
				Description: "Maximum requests per rolling minute. At least one of `requests_per_minute`, `input_tokens_per_minute`, and `output_tokens_per_minute` is required.",
				Optional:    true,
				Validators:  limitValidators("input_tokens_per_minute", "output_tokens_per_minute"),
			},
			"input_tokens_per_minute": schema.Int64Attribute{
				Description: "Maximum input tokens per rolling minute.",
				Optional:    true,
				Validators:  limitValidators("requests_per_minute", "output_tokens_per_minute"),
			},
			"output_tokens_per_minute": schema.Int64Attribute{
				Description: "Maximum output tokens per rolling minute.",
				Optional:    true,
				Validators:  limitValidators("requests_per_minute", "input_tokens_per_minute"),
			},
			"team_id": schema.StringAttribute{
				Description: "UUID of a Datadog Team. When set, only requests attributed to this team count toward the limit.",
				Optional:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "UUID of a Datadog user. When set, only requests attributed to this user count toward the limit.",
				Optional:    true,
			},
			"tags": schema.MapAttribute{
				Description: "Scope tags, such as `provider`, `model` or `ml_app`. A request counts toward the limit when it has all of these tags. Keys must be lowercase.",
				Optional:    true,
				ElementType: types.StringType,
				// The API returns cleared tags as {}, which maps to null, so an empty map would never match state.
				Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			},
		},
	}
}

// ValidateConfig rejects tag keys the API would rewrite, since the rewritten key would differ from the configuration.
func (r *aiGatewayRateLimitResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var tags types.Map
	response.Diagnostics.Append(request.Config.GetAttribute(ctx, path.Root("tags"), &tags)...)
	if response.Diagnostics.HasError() || tags.IsNull() || tags.IsUnknown() {
		return
	}
	for key := range tags.Elements() {
		if err := validateAiGatewayRateLimitTagKey(key); err != nil {
			response.Diagnostics.AddAttributeError(path.Root("tags"), "invalid tag key", err.Error())
		}
	}
}

func validateAiGatewayRateLimitTagKey(key string) error {
	if !aiGatewayRateLimitTagKeyPattern.MatchString(key) || strings.HasPrefix(key, aiGatewayRateLimitTagHeaderPrefix) {
		return fmt.Errorf("tag key %q must be lowercase, with no surrounding whitespace and no %s prefix", key, aiGatewayRateLimitTagHeaderPrefix)
	}
	return nil
}

func (r *aiGatewayRateLimitResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *aiGatewayRateLimitResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state aiGatewayRateLimitModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway rate limit id must be a UUID")
		return
	}
	resp, httpResp, err := r.Api.GetAiGatewayRateLimit(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AI Gateway rate limit"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayRateLimitState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayRateLimitResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state aiGatewayRateLimitModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes, diags := buildAiGatewayRateLimitCreateAttributes(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewAiGatewayRateLimitCreateRequest(*datadogV2.NewAiGatewayRateLimitCreateData(*attributes, datadogV2.AIGATEWAYRATELIMITTYPE_RATE_LIMITS))

	resp, _, err := r.Api.CreateAiGatewayRateLimit(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating AI Gateway rate limit"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}
	data := resp.GetData()

	// Create has no enabled field, so a rate limit configured as disabled is created and then disabled.
	if !state.Enabled.ValueBool() && data.Attributes.GetEnabled() {
		update := datadogV2.NewAiGatewayRateLimitUpdateAttributes()
		update.SetEnabled(false)
		updateBody := datadogV2.NewAiGatewayRateLimitUpdateRequest(*datadogV2.NewAiGatewayRateLimitUpdateData(*update, data.GetId(), datadogV2.AIGATEWAYRATELIMITTYPE_RATE_LIMITS))
		updated, _, err := r.Api.UpdateAiGatewayRateLimit(r.Auth, data.GetId(), *updateBody)
		if err != nil {
			// Keep the created rate limit in state so the next apply retries the update instead of creating a duplicate.
			response.Diagnostics.Append(updateAiGatewayRateLimitState(ctx, &state, data)...)
			response.Diagnostics.Append(response.State.Set(ctx, &state)...)
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error disabling AI Gateway rate limit"))
			return
		}
		data = updated.GetData()
	}

	response.Diagnostics.Append(updateAiGatewayRateLimitState(ctx, &state, data)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayRateLimitResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state aiGatewayRateLimitModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway rate limit id must be a UUID")
		return
	}
	attributes, diags := buildAiGatewayRateLimitUpdateAttributes(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewAiGatewayRateLimitUpdateRequest(*datadogV2.NewAiGatewayRateLimitUpdateData(*attributes, id, datadogV2.AIGATEWAYRATELIMITTYPE_RATE_LIMITS))

	resp, _, err := r.Api.UpdateAiGatewayRateLimit(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating AI Gateway rate limit"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayRateLimitState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayRateLimitResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state aiGatewayRateLimitModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway rate limit id must be a UUID")
		return
	}
	httpResp, err := r.Api.DeleteAiGatewayRateLimit(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting AI Gateway rate limit"))
	}
}

func updateAiGatewayRateLimitState(ctx context.Context, state *aiGatewayRateLimitModel, data datadogV2.AiGatewayRateLimitResource) diag.Diagnostics {
	var diags diag.Diagnostics
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId().String())
	state.Name = types.StringValue(attributes.GetName())
	state.Mode = types.StringValue(string(attributes.GetMode()))
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.RequestsPerMinute = positiveInt64Value(attributes.RequestsPerMinute)
	state.InputTokensPerMinute = positiveInt64Value(attributes.InputTokensPerMinute)
	state.OutputTokensPerMinute = positiveInt64Value(attributes.OutputTokensPerMinute)
	state.TeamID = nullableUUIDValue(state.TeamID, attributes.TeamId)
	state.UserID = nullableUUIDValue(state.UserID, attributes.UserId)

	// The API returns cleared tags as an empty object.
	if tags := attributes.GetTags(); len(tags) > 0 {
		value, d := types.MapValueFrom(ctx, types.StringType, tags)
		diags.Append(d...)
		state.Tags = value
	} else {
		state.Tags = types.MapNull(types.StringType)
	}

	return diags
}

func buildAiGatewayRateLimitCreateAttributes(ctx context.Context, state *aiGatewayRateLimitModel) (*datadogV2.AiGatewayRateLimitCreateAttributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewAiGatewayRateLimitCreateAttributes(state.Name.ValueString())

	if !state.Mode.IsNull() && !state.Mode.IsUnknown() {
		attributes.SetMode(datadogV2.AiGatewayRateLimitMode(state.Mode.ValueString()))
	}
	if !state.RequestsPerMinute.IsNull() {
		attributes.SetRequestsPerMinute(state.RequestsPerMinute.ValueInt64())
	}
	if !state.InputTokensPerMinute.IsNull() {
		attributes.SetInputTokensPerMinute(state.InputTokensPerMinute.ValueInt64())
	}
	if !state.OutputTokensPerMinute.IsNull() {
		attributes.SetOutputTokensPerMinute(state.OutputTokensPerMinute.ValueInt64())
	}
	if !state.TeamID.IsNull() {
		teamID, err := uuid.Parse(state.TeamID.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("team_id"), "invalid team_id", "team_id must be a UUID")
		} else {
			attributes.SetTeamId(teamID)
		}
	}
	if !state.UserID.IsNull() {
		userID, err := uuid.Parse(state.UserID.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("user_id"), "invalid user_id", "user_id must be a UUID")
		} else {
			attributes.SetUserId(userID)
		}
	}
	if !state.Tags.IsNull() {
		tags := map[string]string{}
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	return attributes, diags
}

// buildAiGatewayRateLimitUpdateAttributes sends every field. The API keeps a field that is omitted or null,
// so a field removed from the configuration is sent as its clear value: 0 for a limit, "" for a scope ID, and
// {} for tags.
func buildAiGatewayRateLimitUpdateAttributes(ctx context.Context, state *aiGatewayRateLimitModel) (*datadogV2.AiGatewayRateLimitUpdateAttributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewAiGatewayRateLimitUpdateAttributes()

	attributes.SetName(state.Name.ValueString())
	attributes.SetEnabled(state.Enabled.ValueBool())
	if !state.Mode.IsNull() && !state.Mode.IsUnknown() {
		attributes.SetMode(datadogV2.AiGatewayRateLimitMode(state.Mode.ValueString()))
	}
	attributes.SetRequestsPerMinute(state.RequestsPerMinute.ValueInt64())
	attributes.SetInputTokensPerMinute(state.InputTokensPerMinute.ValueInt64())
	attributes.SetOutputTokensPerMinute(state.OutputTokensPerMinute.ValueInt64())

	for _, scope := range []struct {
		name  string
		value types.String
		set   func(string)
	}{
		{"team_id", state.TeamID, attributes.SetTeamId},
		{"user_id", state.UserID, attributes.SetUserId},
	} {
		if scope.value.IsNull() {
			scope.set("")
			continue
		}
		if _, err := uuid.Parse(scope.value.ValueString()); err != nil {
			diags.AddAttributeError(path.Root(scope.name), "invalid "+scope.name, scope.name+" must be a UUID")
			continue
		}
		scope.set(scope.value.ValueString())
	}

	tags := map[string]string{}
	if !state.Tags.IsNull() {
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	}
	attributes.SetTags(tags)

	return attributes, diags
}

// positiveInt64Value maps a missing limit, or one cleared to 0, to null.
func positiveInt64Value(value datadog.NullableInt64) types.Int64 {
	if v := value.Get(); v != nil && *v > 0 {
		return types.Int64Value(*v)
	}
	return types.Int64Null()
}

// nullableUUIDValue keeps the prior spelling of the same UUID, since the API returns UUIDs in lowercase and a
// configured uppercase UUID would otherwise differ from the plan.
func nullableUUIDValue(prior types.String, value datadog.NullableUUID) types.String {
	v := value.Get()
	if v == nil {
		return types.StringNull()
	}
	if !prior.IsNull() && !prior.IsUnknown() && strings.EqualFold(prior.ValueString(), v.String()) {
		return prior
	}
	return types.StringValue(v.String())
}
