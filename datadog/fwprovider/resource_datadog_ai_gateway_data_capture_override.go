package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure      = &aiGatewayDataCaptureOverrideResource{}
	_ resource.ResourceWithImportState    = &aiGatewayDataCaptureOverrideResource{}
	_ resource.ResourceWithValidateConfig = &aiGatewayDataCaptureOverrideResource{}
)

// aiGatewayDataCaptureTagKeyPattern matches the form the API stores a tag key in: lowercase and trimmed. The API
// rewrites any other key into this form, which would show as a change on every plan.
var aiGatewayDataCaptureTagKeyPattern = regexp.MustCompile(`^[^A-Z\s]([^A-Z]*[^A-Z\s])?$`)

// aiGatewayDataCaptureWriteAttempts bounds how often a write is retried after another change to the organization's
// data capture configuration makes its configuration revision stale.
const aiGatewayDataCaptureWriteAttempts = 3

// NewAiGatewayDataCaptureOverrideID generates the client-side UUID of a new override. Acceptance tests replace it so
// that recorded requests replay with the same UUID.
var NewAiGatewayDataCaptureOverrideID = uuid.New

type aiGatewayDataCaptureOverrideResource struct {
	Api  *datadogV2.AIGatewayApi
	Auth context.Context
}

type aiGatewayDataCaptureOverrideModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Mode               types.String `tfsdk:"mode"`
	MetadataCapture    types.Set    `tfsdk:"metadata_capture"`
	TeamID             types.String `tfsdk:"team_id"`
	UserID             types.String `tfsdk:"user_id"`
	Tags               types.Map    `tfsdk:"tags"`
	ConfirmFullCapture types.Bool   `tfsdk:"confirm_full_capture"`
	Revision           types.Int64  `tfsdk:"revision"`
}

func NewAiGatewayDataCaptureOverrideResource() resource.Resource {
	return &aiGatewayDataCaptureOverrideResource{}
}

func (r *aiGatewayDataCaptureOverrideResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewAIGatewayApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *aiGatewayDataCaptureOverrideResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ai_gateway_data_capture_override"
}

func (r *aiGatewayDataCaptureOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	scopeAttributes := []path.Expression{path.MatchRoot("team_id"), path.MatchRoot("user_id"), path.MatchRoot("tags")}

	response.Schema = schema.Schema{
		Description: "Provides a Datadog AI Gateway data capture override resource. An override sets the data capture mode for the requests that match its scope, instead of the organization default set by `datadog_ai_gateway_governance`. " +
			"When several overrides match a request, `metadata_only` takes precedence over `full_capture`. This resource is in Preview.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the override.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name of the override. Names need not be unique.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"mode": schema.StringAttribute{
				Description: "Data capture mode for matching requests. `metadata_only` records request metadata without prompt or completion content; `full_capture` also records content.",
				Required:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAiGatewayCaptureOverrideModeFromValue),
				},
			},
			"metadata_capture": schema.SetAttribute{
				Description: "Items that Metadata-only capture adds to matching spans. Some items require others: `tool_output_size`, `tool_names`, and `conversation_shape` require `tool_usage`, and `input_size_by_role` requires `tool_output_size`. " +
					"When unset, Terraform leaves the current items unchanged. To record no items, use an empty set.",
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
			"team_id": schema.StringAttribute{
				Description: "UUID of a Datadog Team. When set, the override matches requests that select this team with the `x-ai-gateway-team` header. At least one of `team_id`, `user_id`, and `tags` is required.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.AtLeastOneOf(scopeAttributes...)},
			},
			"user_id": schema.StringAttribute{
				Description: "UUID of a Datadog user. When set, the override matches requests from this user.",
				Optional:    true,
			},
			"tags": schema.MapAttribute{
				Description: "Request tags the override matches, such as `provider`, `model`, or `ml_app`. A request matches when it has all of these tags. Keys must be lowercase.",
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.Map{mapvalidator.SizeAtLeast(1)},
			},
			"confirm_full_capture": schema.BoolAttribute{
				Description: "Set to `true` to confirm a change that can turn on full capture for more requests, such as creating a `full_capture` override or deleting a `metadata_only` one. The API rejects such a change without it.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"revision": schema.Int64Attribute{
				Description: "Revision of the override. It increases with every change.",
				Computed:    true,
			},
		},
	}
}

// ValidateConfig rejects tag keys the API would rewrite, since the rewritten key would differ from the configuration.
func (r *aiGatewayDataCaptureOverrideResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var tags types.Map
	response.Diagnostics.Append(request.Config.GetAttribute(ctx, path.Root("tags"), &tags)...)
	if response.Diagnostics.HasError() || tags.IsNull() || tags.IsUnknown() {
		return
	}
	for key := range tags.Elements() {
		if !aiGatewayDataCaptureTagKeyPattern.MatchString(key) {
			response.Diagnostics.AddAttributeError(path.Root("tags"), "invalid tag key",
				fmt.Sprintf("tag key %q must be lowercase, with no surrounding whitespace", key))
		}
	}
}

func (r *aiGatewayDataCaptureOverrideResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *aiGatewayDataCaptureOverrideResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state aiGatewayDataCaptureOverrideModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway data capture override id must be a UUID")
		return
	}
	_, override, diags := r.readOverride(id)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if override == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(updateAiGatewayDataCaptureOverrideState(ctx, &state, aiGatewayDataCaptureOverrideFields{
		ID:              override.GetId(),
		Name:            override.GetName(),
		Mode:            string(override.GetMode()),
		MetadataCapture: override.GetMetadataCapture(),
		Scope:           override.GetScope(),
		ScopeRestricted: override.GetScopeRestricted(),
		Revision:        override.GetRevision(),
	})...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayDataCaptureOverrideResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var configured, state aiGatewayDataCaptureOverrideModel
	response.Diagnostics.Append(request.Config.Get(ctx, &configured)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	scope, diags := buildAiGatewayDataCaptureScope(ctx, &state)
	response.Diagnostics.Append(diags...)
	metadata, diags := buildAiGatewayDataCaptureMetadata(ctx, configured.MetadataCapture)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// The API takes a client-generated UUID, so a retried create can't make a second override.
	id := NewAiGatewayDataCaptureOverrideID()
	var resp datadogV2.AiGatewayCaptureOverrideResponse
	response.Diagnostics.Append(r.writeWithCurrentRevision(id, "creating", func(configurationRevision string, _ *datadogV2.AiGatewayCaptureOverride) (*http.Response, error) {
		attributes := datadogV2.NewAiGatewayCaptureCreateRequestDataAttributes(configurationRevision,
			datadogV2.AiGatewayCaptureCreateRequestDataAttributesMode(state.Mode.ValueString()), state.Name.ValueString(), *scope)
		attributes.SetConfirmFullCapture(state.ConfirmFullCapture.ValueBool())
		if metadata != nil {
			attributes.SetMetadataCapture(metadata)
		}
		body := datadogV2.NewAiGatewayCaptureCreateRequest(*datadogV2.NewAiGatewayCaptureCreateRequestData(*attributes, id, datadogV2.AIGATEWAYCAPTURECREATEREQUESTDATATYPE_DATA_CAPTURE_OVERRIDES))
		var httpResp *http.Response
		var err error
		resp, httpResp, err = r.Api.CreateAiGatewayDataCaptureOverride(r.Auth, *body)
		return httpResp, err
	})...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayDataCaptureOverrideState(ctx, &state, aiGatewayDataCaptureOverrideResponseFields(resp.GetData()))...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayDataCaptureOverrideResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var configured, state, prior aiGatewayDataCaptureOverrideModel
	response.Diagnostics.Append(request.Config.Get(ctx, &configured)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	response.Diagnostics.Append(request.State.Get(ctx, &prior)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(prior.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway data capture override id must be a UUID")
		return
	}
	metadata, diags := buildAiGatewayDataCaptureMetadata(ctx, configured.MetadataCapture)
	response.Diagnostics.Append(diags...)
	// The API keeps an omitted scope, and the scope it returns can withhold tags the caller can't view, so the scope
	// is sent only when the configuration changes it.
	var scope *datadogV2.AiGatewayCaptureScope
	if !aiGatewayDataCaptureScopeEqual(&state, &prior) {
		scope, diags = buildAiGatewayDataCaptureScope(ctx, &state)
		response.Diagnostics.Append(diags...)
	}
	if response.Diagnostics.HasError() {
		return
	}

	var resp datadogV2.AiGatewayCaptureOverrideResponse
	response.Diagnostics.Append(r.writeWithCurrentRevision(id, "updating", func(configurationRevision string, current *datadogV2.AiGatewayCaptureOverride) (*http.Response, error) {
		if err := checkAiGatewayDataCaptureOverrideRevision(current, prior.Revision.ValueInt64()); err != nil {
			return nil, err
		}
		attributes := datadogV2.NewAiGatewayCaptureUpdateRequestDataAttributes(configurationRevision,
			datadogV2.AiGatewayCaptureUpdateRequestDataAttributesMode(state.Mode.ValueString()), state.Name.ValueString(), prior.Revision.ValueInt64())
		attributes.SetConfirmFullCapture(state.ConfirmFullCapture.ValueBool())
		if metadata != nil {
			attributes.SetMetadataCapture(metadata)
		}
		if scope != nil {
			attributes.SetScope(*scope)
		}
		body := datadogV2.NewAiGatewayCaptureUpdateRequest(*datadogV2.NewAiGatewayCaptureUpdateRequestData(*attributes, id, datadogV2.AIGATEWAYCAPTUREUPDATEREQUESTDATATYPE_DATA_CAPTURE_OVERRIDES))
		var httpResp *http.Response
		var err error
		resp, httpResp, err = r.Api.UpdateAiGatewayDataCaptureOverride(r.Auth, id, *body)
		return httpResp, err
	})...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	response.Diagnostics.Append(updateAiGatewayDataCaptureOverrideState(ctx, &state, aiGatewayDataCaptureOverrideResponseFields(resp.GetData()))...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *aiGatewayDataCaptureOverrideResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state aiGatewayDataCaptureOverrideModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid id", "the AI Gateway data capture override id must be a UUID")
		return
	}
	response.Diagnostics.Append(r.writeWithCurrentRevision(id, "deleting", func(configurationRevision string, current *datadogV2.AiGatewayCaptureOverride) (*http.Response, error) {
		if current == nil {
			return nil, nil
		}
		if err := checkAiGatewayDataCaptureOverrideRevision(current, state.Revision.ValueInt64()); err != nil {
			return nil, err
		}
		attributes := datadogV2.NewAiGatewayCaptureDeleteRequestDataAttributes(configurationRevision, state.Revision.ValueInt64())
		attributes.SetConfirmFullCapture(state.ConfirmFullCapture.ValueBool())
		body := datadogV2.NewAiGatewayCaptureDeleteRequest(*datadogV2.NewAiGatewayCaptureDeleteRequestData(*attributes, id, datadogV2.AIGATEWAYCAPTUREDELETEREQUESTDATATYPE_DATA_CAPTURE_OVERRIDES))
		httpResp, err := r.Api.DeleteAiGatewayDataCaptureOverride(r.Auth, id, *body)
		if err != nil && httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return httpResp, err
	})...)
}

// readOverride returns the organization's configuration revision and the override with this ID, or nil when the
// configuration doesn't list it. The configuration lists only the overrides the caller can view.
func (r *aiGatewayDataCaptureOverrideResource) readOverride(id uuid.UUID) (string, *datadogV2.AiGatewayCaptureOverride, diag.Diagnostics) {
	var diags diag.Diagnostics
	resp, _, err := r.Api.GetAiGatewayDataCaptureConfiguration(r.Auth)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error retrieving AI Gateway data capture configuration"))
		return "", nil, diags
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("response contains unparsed object", err.Error())
		return "", nil, diags
	}
	data := resp.GetData()
	attributes := data.GetAttributes()
	for _, override := range attributes.GetOverrides() {
		if override.GetId() == id {
			return attributes.GetConfigurationRevision(), &override, diags
		}
	}
	return attributes.GetConfigurationRevision(), nil, diags
}

// writeWithCurrentRevision reads the current configuration revision and runs write with it. Any change to the
// organization's data capture configuration, including another override's, makes the revision stale and the write
// fail with HTTP 409, so the write is retried with a fresh revision a few times. A change to this override itself
// isn't retried: write checks the override's own revision against the state first.
func (r *aiGatewayDataCaptureOverrideResource) writeWithCurrentRevision(id uuid.UUID, action string, write func(configurationRevision string, current *datadogV2.AiGatewayCaptureOverride) (*http.Response, error)) diag.Diagnostics {
	var lastErr error
	for attempt := 0; attempt < aiGatewayDataCaptureWriteAttempts; attempt++ {
		configurationRevision, current, diags := r.readOverride(id)
		if diags.HasError() {
			return diags
		}
		httpResp, err := write(configurationRevision, current)
		if err == nil {
			return nil
		}
		lastErr = err
		if httpResp == nil || httpResp.StatusCode != http.StatusConflict {
			break
		}
	}
	return diag.Diagnostics{utils.FrameworkErrorDiag(lastErr, fmt.Sprintf("error %s AI Gateway data capture override", action))}
}

type aiGatewayDataCaptureOverrideChangedError struct {
	current, prior int64
}

func (e aiGatewayDataCaptureOverrideChangedError) Error() string {
	return fmt.Sprintf("the override changed outside Terraform (revision %d, Terraform last read revision %d); run terraform plan again to see the change", e.current, e.prior)
}

func checkAiGatewayDataCaptureOverrideRevision(current *datadogV2.AiGatewayCaptureOverride, prior int64) error {
	if current == nil {
		return fmt.Errorf("the override no longer exists or isn't visible to this API key")
	}
	if current.GetRevision() != prior {
		return aiGatewayDataCaptureOverrideChangedError{current: current.GetRevision(), prior: prior}
	}
	return nil
}

func buildAiGatewayDataCaptureScope(ctx context.Context, state *aiGatewayDataCaptureOverrideModel) (*datadogV2.AiGatewayCaptureScope, diag.Diagnostics) {
	var diags diag.Diagnostics
	scope := datadogV2.NewAiGatewayCaptureScope()

	for _, facet := range []struct {
		name  string
		value types.String
		set   func(uuid.UUID)
	}{
		{"team_id", state.TeamID, scope.SetTeamId},
		{"user_id", state.UserID, scope.SetUserId},
	} {
		if facet.value.IsNull() || facet.value.IsUnknown() {
			continue
		}
		id, err := uuid.Parse(facet.value.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root(facet.name), "invalid "+facet.name, facet.name+" must be a UUID")
			continue
		}
		facet.set(id)
	}
	if !state.Tags.IsNull() && !state.Tags.IsUnknown() {
		tags := map[string]string{}
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		scope.SetTags(tags)
	}

	return scope, diags
}

// buildAiGatewayDataCaptureMetadata returns nil when the configuration doesn't set metadata_capture, since the API
// keeps omitted items, and a non-nil empty slice, which is sent as [] to clear them, when it sets an empty set.
func buildAiGatewayDataCaptureMetadata(ctx context.Context, configured types.Set) ([]datadogV2.AiGatewayMetadataCaptureItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	if configured.IsNull() || configured.IsUnknown() {
		return nil, diags
	}
	var items []string
	diags.Append(configured.ElementsAs(ctx, &items, false)...)
	metadata := make([]datadogV2.AiGatewayMetadataCaptureItem, 0, len(items))
	for _, item := range items {
		metadata = append(metadata, datadogV2.AiGatewayMetadataCaptureItem(item))
	}
	return metadata, diags
}

func aiGatewayDataCaptureScopeEqual(a, b *aiGatewayDataCaptureOverrideModel) bool {
	return strings.EqualFold(a.TeamID.ValueString(), b.TeamID.ValueString()) && a.TeamID.IsNull() == b.TeamID.IsNull() &&
		strings.EqualFold(a.UserID.ValueString(), b.UserID.ValueString()) && a.UserID.IsNull() == b.UserID.IsNull() &&
		a.Tags.Equal(b.Tags)
}

type aiGatewayDataCaptureOverrideFields struct {
	ID              uuid.UUID
	Name            string
	Mode            string
	MetadataCapture []datadogV2.AiGatewayMetadataCaptureItem
	Scope           datadogV2.AiGatewayCaptureResponseScope
	ScopeRestricted bool
	Revision        int64
}

func aiGatewayDataCaptureOverrideResponseFields(data datadogV2.AiGatewayCaptureOverrideResponseData) aiGatewayDataCaptureOverrideFields {
	attributes := data.GetAttributes()
	return aiGatewayDataCaptureOverrideFields{
		ID:              data.GetId(),
		Name:            attributes.GetName(),
		Mode:            string(attributes.GetMode()),
		MetadataCapture: attributes.GetMetadataCapture(),
		Scope:           attributes.GetScope(),
		ScopeRestricted: attributes.GetScopeRestricted(),
		Revision:        attributes.GetRevision(),
	}
}

func updateAiGatewayDataCaptureOverrideState(ctx context.Context, state *aiGatewayDataCaptureOverrideModel, fields aiGatewayDataCaptureOverrideFields) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(fields.ID.String())
	state.Name = types.StringValue(fields.Name)
	state.Mode = types.StringValue(fields.Mode)
	state.Revision = types.Int64Value(fields.Revision)

	items := make([]string, 0, len(fields.MetadataCapture))
	for _, item := range fields.MetadataCapture {
		items = append(items, string(item))
	}
	value, d := types.SetValueFrom(ctx, types.StringType, items)
	diags.Append(d...)
	state.MetadataCapture = value

	// A restricted scope withholds tags the caller can't view, so keep the scope Terraform last saw.
	if fields.ScopeRestricted {
		return diags
	}
	state.TeamID = aiGatewayDataCaptureUUIDValue(state.TeamID, fields.Scope.TeamId)
	state.UserID = aiGatewayDataCaptureUUIDValue(state.UserID, fields.Scope.UserId)
	if tags := fields.Scope.GetTags(); len(tags) > 0 {
		value, d := types.MapValueFrom(ctx, types.StringType, tags)
		diags.Append(d...)
		state.Tags = value
	} else {
		state.Tags = types.MapNull(types.StringType)
	}

	return diags
}

// aiGatewayDataCaptureUUIDValue keeps the prior spelling of the same UUID, since the API returns UUIDs in lowercase
// and a configured uppercase UUID would otherwise differ from the plan.
func aiGatewayDataCaptureUUIDValue(prior types.String, value *uuid.UUID) types.String {
	if value == nil {
		return types.StringNull()
	}
	if !prior.IsNull() && !prior.IsUnknown() && strings.EqualFold(prior.ValueString(), value.String()) {
		return prior
	}
	return types.StringValue(value.String())
}
