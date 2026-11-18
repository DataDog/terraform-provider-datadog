package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &aiGatewayModelGroupResource{}
	_ resource.ResourceWithImportState = &aiGatewayModelGroupResource{}
)

// aiGatewayDefaultModelGroupID is the ID of the organization-wide model group. It always exists, cannot be deleted,
// and needs expected_revision on update, so this resource does not manage it.
const aiGatewayDefaultModelGroupID = "default"

type aiGatewayModelGroupResource struct {
	Api  *datadogV2.AIGatewayApi
	Auth context.Context
}

type aiGatewayModelGroupModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ModelIDs          types.Set    `tfsdk:"model_ids"`
	ProviderIDs       types.Set    `tfsdk:"provider_ids"`
	RoutingRuleSlugs  types.Set    `tfsdk:"routing_rule_slugs"`
	AllowFastMode     types.Bool   `tfsdk:"allow_fast_mode"`
	MaxReasoningLevel types.String `tfsdk:"max_reasoning_level"`
}

func NewAiGatewayModelGroupResource() resource.Resource {
	return &aiGatewayModelGroupResource{}
}

func (r *aiGatewayModelGroupResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewAIGatewayApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *aiGatewayModelGroupResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ai_gateway_model_group"
}

func (r *aiGatewayModelGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AI Gateway model group resource. A model group selects the models, providers, and routing rules that its members can use, " +
			"and the request options that apply to them. This resource manages named model groups only, not the organization-wide default model group. " +
			"This resource is in Preview.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of the model group.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"model_ids": schema.SetAttribute{
				Description: "IDs of individually selected model catalog entries. Individual selections do not include models added to a provider later.",
				Optional:    true,
				ElementType: types.Int64Type,
				Validators: []validator.Set{
					setvalidator.ValueInt64sAre(int64validator.AtLeast(1)),
				},
			},
			"provider_ids": schema.SetAttribute{
				Description: "IDs of selected providers. A provider selection covers the provider's current and future models.",
				Optional:    true,
				ElementType: types.Int64Type,
				Validators: []validator.Set{
					setvalidator.ValueInt64sAre(int64validator.AtLeast(1)),
				},
			},
			"routing_rule_slugs": schema.SetAttribute{
				Description: "Slugs of selected routing rules, such as `coding-route`.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 128)),
				},
			},
			"allow_fast_mode": schema.BoolAttribute{
				Description: "Whether fast mode is allowed for the models and routing rules in the group.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"max_reasoning_level": schema.StringAttribute{
				Description: "Maximum reasoning level permitted for requests that use the group. No group-specific limit applies when unset.",
				Optional:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayMaxReasoningLevelFromValue),
				},
			},
		},
	}
}

func (r *aiGatewayModelGroupResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if request.ID == aiGatewayDefaultModelGroupID {
		response.Diagnostics.AddError("unsupported model group", "the organization-wide default model group cannot be managed by datadog_ai_gateway_model_group")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *aiGatewayModelGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state aiGatewayModelGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetAiGatewayModelGroup(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AI Gateway model group"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayModelGroupState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayModelGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state aiGatewayModelGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes, diags := buildAiGatewayModelGroupCreateAttributes(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewAiGatewayModelGroupCreateRequest(*datadogV2.NewAiGatewayModelGroupCreateData(*attributes, datadogV2.AIGATEWAYMODELGROUPTYPE_MODEL_GROUPS))

	resp, _, err := r.Api.CreateAiGatewayModelGroup(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating AI Gateway model group"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayModelGroupState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayModelGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state aiGatewayModelGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	attributes, diags := buildAiGatewayModelGroupUpdateAttributes(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewAiGatewayModelGroupUpdateRequest(*datadogV2.NewAiGatewayModelGroupUpdateData(*attributes, id, datadogV2.AIGATEWAYMODELGROUPTYPE_MODEL_GROUPS))

	resp, _, err := r.Api.UpdateAiGatewayModelGroup(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating AI Gateway model group"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayModelGroupState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayModelGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state aiGatewayModelGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteAiGatewayModelGroup(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting AI Gateway model group"))
	}
}

// updateAiGatewayModelGroupState sets state from the API response. The response lists every saved selection, including
// ones the caller cannot view, so the selections are read from model_ids, provider_ids and routing_rules rather than
// from the embedded metadata.
func updateAiGatewayModelGroupState(ctx context.Context, state *aiGatewayModelGroupModel, data datadogV2.AiGatewayModelGroupResource) diag.Diagnostics {
	var diags diag.Diagnostics
	attributes := data.GetAttributes()

	if attributes.Revision != nil || data.GetId() == aiGatewayDefaultModelGroupID {
		diags.AddError("unsupported model group", "the organization-wide default model group cannot be managed by datadog_ai_gateway_model_group")
		return diags
	}

	slugs := make([]string, 0, len(attributes.GetRoutingRules()))
	for _, rule := range attributes.GetRoutingRules() {
		slugs = append(slugs, rule.GetSlug())
	}

	state.ID = types.StringValue(data.GetId())
	state.Name = types.StringValue(attributes.GetName())
	state.ModelIDs = aiGatewayModelGroupSetValue(ctx, &diags, state.ModelIDs, types.Int64Type, attributes.GetModelIds())
	state.ProviderIDs = aiGatewayModelGroupSetValue(ctx, &diags, state.ProviderIDs, types.Int64Type, attributes.GetProviderIds())
	state.RoutingRuleSlugs = aiGatewayModelGroupSetValue(ctx, &diags, state.RoutingRuleSlugs, types.StringType, slugs)
	state.AllowFastMode = types.BoolValue(attributes.GetAllowFastMode())
	if level := attributes.MaxReasoningLevel.Get(); level != nil {
		state.MaxReasoningLevel = types.StringValue(string(*level))
	} else {
		state.MaxReasoningLevel = types.StringNull()
	}

	return diags
}

// aiGatewayModelGroupSetValue maps an empty selection to null, so an omitted attribute does not diff. An attribute
// configured as an empty set stays empty, so the applied state matches the plan.
func aiGatewayModelGroupSetValue[T any](ctx context.Context, diags *diag.Diagnostics, prior types.Set, elementType attr.Type, values []T) types.Set {
	if len(values) == 0 {
		if !prior.IsNull() && !prior.IsUnknown() {
			return types.SetValueMust(elementType, []attr.Value{})
		}
		return types.SetNull(elementType)
	}
	value, d := types.SetValueFrom(ctx, elementType, values)
	diags.Append(d...)
	return value
}

func buildAiGatewayModelGroupCreateAttributes(ctx context.Context, state *aiGatewayModelGroupModel) (*datadogV2.AiGatewayModelGroupCreateAttributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewAiGatewayModelGroupCreateAttributes(state.Name.ValueString())

	attributes.SetAllowFastMode(state.AllowFastMode.ValueBool())
	if !state.ModelIDs.IsNull() {
		attributes.SetModelIds(aiGatewayModelGroupSetElements[int64](ctx, &diags, state.ModelIDs))
	}
	if !state.ProviderIDs.IsNull() {
		attributes.SetProviderIds(aiGatewayModelGroupSetElements[int64](ctx, &diags, state.ProviderIDs))
	}
	if !state.RoutingRuleSlugs.IsNull() {
		attributes.SetRoutingRuleSlugs(aiGatewayModelGroupSetElements[string](ctx, &diags, state.RoutingRuleSlugs))
	}
	if !state.MaxReasoningLevel.IsNull() {
		attributes.SetMaxReasoningLevel(datadogV2.AiGatewayMaxReasoningLevel(state.MaxReasoningLevel.ValueString()))
	}

	return attributes, diags
}

// buildAiGatewayModelGroupUpdateAttributes sends every field. The API keeps a selection that is omitted, so a
// selection removed from the configuration is sent as [], and an unset max_reasoning_level is sent as null.
// expected_revision is never sent: it is only for the default model group.
func buildAiGatewayModelGroupUpdateAttributes(ctx context.Context, state *aiGatewayModelGroupModel) (*datadogV2.AiGatewayModelGroupUpdateAttributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewAiGatewayModelGroupUpdateAttributes()

	attributes.SetName(state.Name.ValueString())
	attributes.SetAllowFastMode(state.AllowFastMode.ValueBool())
	attributes.SetModelIds(aiGatewayModelGroupSetElements[int64](ctx, &diags, state.ModelIDs))
	attributes.SetProviderIds(aiGatewayModelGroupSetElements[int64](ctx, &diags, state.ProviderIDs))
	attributes.SetRoutingRuleSlugs(aiGatewayModelGroupSetElements[string](ctx, &diags, state.RoutingRuleSlugs))
	if state.MaxReasoningLevel.IsNull() {
		attributes.SetMaxReasoningLevelNil()
	} else {
		attributes.SetMaxReasoningLevel(datadogV2.AiGatewayNullableMaxReasoningLevel(state.MaxReasoningLevel.ValueString()))
	}

	return attributes, diags
}

// aiGatewayModelGroupSetElements returns the set's elements, and a non-nil empty slice for a null or empty set, which
// the client serializes as [] instead of omitting it.
func aiGatewayModelGroupSetElements[T any](ctx context.Context, diags *diag.Diagnostics, set types.Set) []T {
	var values []T
	if !set.IsNull() && !set.IsUnknown() {
		diags.Append(set.ElementsAs(ctx, &values, false)...)
	}
	if values == nil {
		values = []T{}
	}
	return values
}
