package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &aiGatewayGovernanceResource{}
	_ resource.ResourceWithImportState = &aiGatewayGovernanceResource{}
)

// aiGatewayGovernanceID is the only ID the API accepts: an organization has exactly one set of governance controls.
const aiGatewayGovernanceID = "organization"

type aiGatewayGovernanceResource struct {
	Api  *datadogV2.AIGatewayApi
	Auth context.Context
}

type aiGatewayGovernanceModel struct {
	ID              types.String `tfsdk:"id"`
	DataCaptureMode types.String `tfsdk:"data_capture_mode"`
	MetadataCapture types.Set    `tfsdk:"metadata_capture"`
	ZdrMode         types.String `tfsdk:"zdr_mode"`
}

func NewAiGatewayGovernanceResource() resource.Resource {
	return &aiGatewayGovernanceResource{}
}

func (r *aiGatewayGovernanceResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewAIGatewayApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *aiGatewayGovernanceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ai_gateway_governance"
}

func (r *aiGatewayGovernanceResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages the organization-wide AI Gateway governance controls: the default data capture mode, the metadata that Metadata-only capture records, and the zero data retention (ZDR) requirement. " +
			"This resource is in Preview. An organization has one set of these controls, so declare this resource at most once per organization. " +
			"Deleting the resource removes it from the Terraform state and leaves the controls unchanged.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Always `organization`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_capture_mode": schema.StringAttribute{
				Description: "Default data capture mode, used for requests that don't match a data capture override. `metadata_only` records request metadata without prompt or completion content. `full_capture` also records prompt and completion content.",
				Required:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayDataCaptureModeFromValue),
				},
			},
			"metadata_capture": schema.SetAttribute{
				Description: "Items that Metadata-only capture adds to spans. Some items require others: `tool_output_size`, `tool_names`, and `conversation_shape` require `tool_usage`, and `input_size_by_role` requires `tool_output_size`. " +
					"When unset, Terraform leaves the current items unchanged. To record no items, set this attribute to an empty set (`[]`).",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayMetadataCaptureItemFromValue)),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"zdr_mode": schema.StringAttribute{
				Description: "Whether every request must reach a provider covered by a zero data retention agreement (`required`) or not (`off`). When unset, Terraform leaves the current requirement unchanged.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayZDRModeFromValue),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *aiGatewayGovernanceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if request.ID != aiGatewayGovernanceID {
		response.Diagnostics.AddError("invalid import ID", `the AI Gateway governance import ID must be "organization"`)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *aiGatewayGovernanceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state aiGatewayGovernanceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.GetAiGatewayGovernance(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AI Gateway governance"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayGovernanceState(ctx, &state, resp.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Create replaces the controls the configuration sets, because an organization's controls always exist.
func (r *aiGatewayGovernanceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	r.replace(ctx, request.Config, request.Plan, &response.State, &response.Diagnostics)
}

func (r *aiGatewayGovernanceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	r.replace(ctx, request.Config, request.Plan, &response.State, &response.Diagnostics)
}

// Delete leaves the controls as they are. Resetting them would change data capture or zero data retention for the
// whole organization, and the API has no defaults to reset to.
func (r *aiGatewayGovernanceResource) Delete(_ context.Context, _ resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddWarning("AI Gateway governance left unchanged",
		"The resource was removed from the Terraform state. The organization's AI Gateway governance controls keep their current values.")
}

type aiGatewayGovernancePlan interface {
	Get(ctx context.Context, target interface{}) diag.Diagnostics
}

type aiGatewayGovernanceState interface {
	Set(ctx context.Context, value interface{}) diag.Diagnostics
}

// replace builds the request from the configuration rather than the plan: the plan carries the last-read value of
// an unset Optional+Computed field, and sending it would write that field even though the configuration leaves it alone.
func (r *aiGatewayGovernanceResource) replace(ctx context.Context, config, plan aiGatewayGovernancePlan, target aiGatewayGovernanceState, diags *diag.Diagnostics) {
	var configured, state aiGatewayGovernanceModel
	diags.Append(config.Get(ctx, &configured)...)
	diags.Append(plan.Get(ctx, &state)...)
	if diags.HasError() {
		return
	}

	attributes, d := buildAiGatewayGovernanceAttributes(ctx, &configured)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	data := datadogV2.NewAiGatewayGovernanceReplaceData(*attributes, datadogV2.AIGATEWAYGOVERNANCETYPE_GOVERNANCE)
	data.SetId(aiGatewayGovernanceID)

	resp, _, err := r.Api.UpdateAiGatewayGovernance(r.Auth, *datadogV2.NewAiGatewayGovernanceReplaceRequest(*data))
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error updating AI Gateway governance"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("response contains unparsed object", err.Error())
		return
	}

	diags.Append(updateAiGatewayGovernanceState(ctx, &state, resp.GetData())...)
	diags.Append(target.Set(ctx, &state)...)
}

// buildAiGatewayGovernanceAttributes takes the configuration and omits metadata_capture and zdr_mode when it doesn't set them,
// since the API keeps an omitted field. Sending zdr_mode only when configured means Terraform never turns off an
// organization's zero data retention requirement by accident.
func buildAiGatewayGovernanceAttributes(ctx context.Context, state *aiGatewayGovernanceModel) (*datadogV2.AiGatewayGovernanceReplaceAttributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := datadogV2.NewAiGatewayGovernanceReplaceAttributes(datadogV2.AiGatewayDataCaptureMode(state.DataCaptureMode.ValueString()))

	if !state.MetadataCapture.IsNull() && !state.MetadataCapture.IsUnknown() {
		var items []string
		diags.Append(state.MetadataCapture.ElementsAs(ctx, &items, false)...)
		// A non-nil empty slice is sent as [], which clears the items.
		metadata := make([]datadogV2.AiGatewayMetadataCaptureItem, 0, len(items))
		for _, item := range items {
			metadata = append(metadata, datadogV2.AiGatewayMetadataCaptureItem(item))
		}
		attributes.SetMetadataCapture(metadata)
	}
	if !state.ZdrMode.IsNull() && !state.ZdrMode.IsUnknown() {
		attributes.SetZdrMode(datadogV2.AiGatewayZDRMode(state.ZdrMode.ValueString()))
	}

	return attributes, diags
}

func updateAiGatewayGovernanceState(ctx context.Context, state *aiGatewayGovernanceModel, data datadogV2.AiGatewayGovernanceResource) diag.Diagnostics {
	var diags diag.Diagnostics
	attributes := data.GetAttributes()

	state.ID = types.StringValue(aiGatewayGovernanceID)
	state.DataCaptureMode = types.StringValue(string(attributes.GetDataCaptureMode()))
	state.ZdrMode = types.StringValue(string(attributes.GetZdrMode()))

	items := make([]string, 0, len(attributes.GetMetadataCapture()))
	for _, item := range attributes.GetMetadataCapture() {
		items = append(items, string(item))
	}
	value, d := types.SetValueFrom(ctx, types.StringType, items)
	diags.Append(d...)
	state.MetadataCapture = value

	return diags
}
