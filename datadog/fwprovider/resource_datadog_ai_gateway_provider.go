package fwprovider

import (
	"context"
	"net/http"
	"regexp"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &aiGatewayProviderResource{}
	_ resource.ResourceWithImportState = &aiGatewayProviderResource{}
)

type aiGatewayProviderResource struct {
	Api  *datadogV2.AIGatewayApi
	Auth context.Context
}

type aiGatewayProviderModel struct {
	ID                   types.String  `tfsdk:"id"`
	Name                 types.String  `tfsdk:"name"`
	Slug                 types.String  `tfsdk:"slug"`
	KnownProvider        types.String  `tfsdk:"known_provider"`
	IntegrationAccountID types.String  `tfsdk:"integration_account_id"`
	IntegrationConfig    types.Object  `tfsdk:"integration_config"`
	CostMultiplier       types.Float64 `tfsdk:"cost_multiplier"`
}

type aiGatewayProviderIntegrationConfigModel struct {
	ProjectID           types.String `tfsdk:"project_id"`
	Location            types.String `tfsdk:"location"`
	BedrockRegion       types.String `tfsdk:"bedrock_region"`
	FoundryResourceID   types.String `tfsdk:"foundry_resource_id"`
	FoundryProject      types.String `tfsdk:"foundry_project"`
	SupportedOperations types.List   `tfsdk:"supported_operations"`
}

var aiGatewayProviderIntegrationConfigAttrTypes = map[string]attr.Type{
	"project_id":           types.StringType,
	"location":             types.StringType,
	"bedrock_region":       types.StringType,
	"foundry_resource_id":  types.StringType,
	"foundry_project":      types.StringType,
	"supported_operations": types.ListType{ElemType: types.StringType},
}

func NewAiGatewayProviderResource() resource.Resource {
	return &aiGatewayProviderResource{}
}

func (r *aiGatewayProviderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewAIGatewayApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *aiGatewayProviderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ai_gateway_provider"
}

func (r *aiGatewayProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	// The API can only update name and cost_multiplier, so every other field forces a new provider.
	// integration_config is replaced field by field rather than as a whole object. Object plan modifiers run
	// before the nested ones, so an object-level RequiresReplace would compare the object before
	// supported_operations, which is computed, had its UseStateForUnknown applied.
	immutableString := func(description string, maxLength int) schema.StringAttribute {
		return schema.StringAttribute{
			Description: description,
			Optional:    true,
			Validators:  []validator.String{stringvalidator.LengthAtMost(maxLength)},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		}
	}

	response.Schema = schema.Schema{
		Description: "Provides a Datadog AI Gateway provider resource. A provider connects the AI Gateway to a model vendor through an integration account, and its models become routable under its slug. " +
			"This resource is in Preview.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Human-readable name of the provider.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 256)},
			},
			"slug": schema.StringAttribute{
				Description: "Routing identifier of the provider. Changing it forces a new provider.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9_-]+$`), "must contain only letters, digits, hyphens, and underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"known_provider": schema.StringAttribute{
				Description: "Provider implementation. Use `custom` for an OpenAI-compatible endpoint. Changing it forces a new provider.",
				Required:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayKnownProviderFromValue),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"integration_account_id": schema.StringAttribute{
				Description: "ID of the integration account that holds the vendor credentials. Changing it forces a new provider.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 256)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"integration_config": schema.SingleNestedAttribute{
				Description: "Provider-specific integration settings. The API returns this block for every `custom` provider, so it is computed when unset. Changing any field forces a new provider.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"project_id":          immutableString("GCP project ID, for `gemini` providers.", 256),
					"location":            immutableString("GCP location, for `gemini` providers.", 256),
					"bedrock_region":      immutableString("AWS region, for `bedrock` providers.", 64),
					"foundry_resource_id": immutableString("ARM resource ID of the Azure AI Foundry resource, for `azure_foundry` providers.", 512),
					"foundry_project":     immutableString("Azure AI Foundry project name, for `azure_foundry` providers.", 256),
					"supported_operations": schema.ListAttribute{
						Description: "Operations that a `custom` provider discovered and manually added models support. The API defaults to `OPENAI_CHAT_COMPLETIONS`.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ValueStringsAre(validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayCustomProviderSupportedOperationFromValue)),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
							listplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"cost_multiplier": schema.Float64Attribute{
				Description: "Multiplier applied to provider costs. Must be greater than 0 and at most 1000000. The API defaults to 1.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Float64{
					float64validator.AtMost(1000000),
					aiGatewayPositiveFloat64Validator{},
				},
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *aiGatewayProviderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *aiGatewayProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state aiGatewayProviderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, diags := parseAiGatewayProviderID(state.ID)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	resp, httpResp, err := r.Api.GetAiGatewayProvider(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AI Gateway provider"))
		return
	}
	// The models in resp.Included can carry operations newer than the client; the resource doesn't read them.
	if err := utils.CheckForUnparsed(resp.Data); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayProviderState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state aiGatewayProviderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes, diags := buildAiGatewayProviderCreateAttributes(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewAiGatewayProviderCreateRequest(*datadogV2.NewAiGatewayProviderCreateData(*attributes, datadogV2.AIGATEWAYPROVIDERTYPE_PROVIDERS))

	resp, _, err := r.Api.CreateAiGatewayProvider(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating AI Gateway provider"))
		return
	}
	if err := utils.CheckForUnparsed(resp.Data); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayProviderState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state aiGatewayProviderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, diags := parseAiGatewayProviderID(state.ID)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body := datadogV2.NewAiGatewayProviderUpdateRequest(*datadogV2.NewAiGatewayProviderUpdateData(*buildAiGatewayProviderUpdateAttributes(&state), state.ID.ValueString(), datadogV2.AIGATEWAYPROVIDERTYPE_PROVIDERS))

	resp, _, err := r.Api.UpdateAiGatewayProvider(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating AI Gateway provider"))
		return
	}
	if err := utils.CheckForUnparsed(resp.Data); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayProviderState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state aiGatewayProviderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, diags := parseAiGatewayProviderID(state.ID)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	// force is never sent: deleting a provider that routing rules still use must not silently rewrite or delete those rules.
	httpResp, err := r.Api.DeleteAiGatewayProvider(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting AI Gateway provider: it is still used by routing rules. Remove it from those rules, or delete them, before deleting the provider"))
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting AI Gateway provider"))
	}
}

// parseAiGatewayProviderID converts the string resource ID to the numeric ID that the provider endpoints take.
func parseAiGatewayProviderID(value types.String) (int64, diag.Diagnostics) {
	var diags diag.Diagnostics
	id, err := strconv.ParseInt(value.ValueString(), 10, 64)
	if err != nil || id < 1 {
		diags.AddError("invalid id", "the AI Gateway provider id must be a positive integer, got "+strconv.Quote(value.ValueString()))
	}
	return id, diags
}

func updateAiGatewayProviderState(ctx context.Context, state *aiGatewayProviderModel, data datadogV2.AiGatewayProviderResource) diag.Diagnostics {
	var diags diag.Diagnostics
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())
	state.Name = types.StringValue(attributes.GetName())
	state.Slug = types.StringValue(attributes.GetSlug())
	state.KnownProvider = types.StringValue(string(attributes.GetKnownProvider()))
	state.CostMultiplier = types.Float64Value(attributes.GetCostMultiplier())

	// The API returns an empty string for a provider created without an integration account.
	if accountID := attributes.GetIntegrationAccountId(); accountID != "" {
		state.IntegrationAccountID = types.StringValue(accountID)
	} else {
		state.IntegrationAccountID = types.StringNull()
	}

	config, ok := attributes.GetIntegrationConfigOk()
	if !ok || config == nil {
		state.IntegrationConfig = types.ObjectNull(aiGatewayProviderIntegrationConfigAttrTypes)
		return diags
	}
	model := aiGatewayProviderIntegrationConfigModel{
		ProjectID:           types.StringPointerValue(config.ProjectId),
		Location:            types.StringPointerValue(config.Location),
		BedrockRegion:       types.StringPointerValue(config.BedrockRegion),
		FoundryResourceID:   types.StringPointerValue(config.FoundryResourceId),
		FoundryProject:      types.StringPointerValue(config.FoundryProject),
		SupportedOperations: types.ListNull(types.StringType),
	}
	if operations := config.GetSupportedOperations(); len(operations) > 0 {
		value, d := types.ListValueFrom(ctx, types.StringType, operations)
		diags.Append(d...)
		model.SupportedOperations = value
	}
	value, d := types.ObjectValueFrom(ctx, aiGatewayProviderIntegrationConfigAttrTypes, model)
	diags.Append(d...)
	state.IntegrationConfig = value

	return diags
}

// buildAiGatewayProviderCreateAttributes never sets initial_bindings, which the API rejects with HTTP 503 even when
// empty, or models, which this resource does not manage.
func buildAiGatewayProviderCreateAttributes(ctx context.Context, state *aiGatewayProviderModel) (*datadogV2.AiGatewayProviderCreateAttributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewAiGatewayProviderCreateAttributes(state.Name.ValueString(), state.Slug.ValueString())
	// The constructor fills in the spec default of 1. Leave an unset cost_multiplier to the server instead.
	attributes.CostMultiplier = nil

	attributes.SetKnownProvider(datadogV2.AiGatewayKnownProvider(state.KnownProvider.ValueString()))
	if !state.IntegrationAccountID.IsNull() && !state.IntegrationAccountID.IsUnknown() {
		attributes.SetIntegrationAccountId(state.IntegrationAccountID.ValueString())
	}
	if !state.CostMultiplier.IsNull() && !state.CostMultiplier.IsUnknown() {
		attributes.SetCostMultiplier(state.CostMultiplier.ValueFloat64())
	}

	if !state.IntegrationConfig.IsNull() && !state.IntegrationConfig.IsUnknown() {
		var model aiGatewayProviderIntegrationConfigModel
		diags.Append(state.IntegrationConfig.As(ctx, &model, basetypes.ObjectAsOptions{})...)
		config := datadogV2.NewAiGatewayProviderIntegrationConfig()
		for _, field := range []struct {
			value types.String
			set   func(string)
		}{
			{model.ProjectID, config.SetProjectId},
			{model.Location, config.SetLocation},
			{model.BedrockRegion, config.SetBedrockRegion},
			{model.FoundryResourceID, config.SetFoundryResourceId},
			{model.FoundryProject, config.SetFoundryProject},
		} {
			if !field.value.IsNull() && !field.value.IsUnknown() {
				field.set(field.value.ValueString())
			}
		}
		if !model.SupportedOperations.IsNull() && !model.SupportedOperations.IsUnknown() {
			var operations []string
			diags.Append(model.SupportedOperations.ElementsAs(ctx, &operations, false)...)
			supported := make([]datadogV2.AiGatewayCustomProviderSupportedOperation, 0, len(operations))
			for _, operation := range operations {
				supported = append(supported, datadogV2.AiGatewayCustomProviderSupportedOperation(operation))
			}
			config.SetSupportedOperations(supported)
		}
		attributes.SetIntegrationConfig(*config)
	}

	return attributes, diags
}

// buildAiGatewayProviderUpdateAttributes sends only name and cost_multiplier, the two fields the API can update.
// Every other field forces a replacement.
func buildAiGatewayProviderUpdateAttributes(state *aiGatewayProviderModel) *datadogV2.AiGatewayProviderUpdateAttributes {
	attributes := datadogV2.NewAiGatewayProviderUpdateAttributes()
	attributes.SetName(state.Name.ValueString())
	if !state.CostMultiplier.IsNull() && !state.CostMultiplier.IsUnknown() {
		attributes.SetCostMultiplier(state.CostMultiplier.ValueFloat64())
	}
	return attributes
}
