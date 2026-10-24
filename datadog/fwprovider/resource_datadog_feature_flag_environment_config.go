package fwprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &featureFlagEnvironmentConfigResource{}
	_ resource.ResourceWithImportState = &featureFlagEnvironmentConfigResource{}
)

func NewFeatureFlagEnvironmentConfigResource() resource.Resource {
	return &featureFlagEnvironmentConfigResource{}
}

type featureFlagEnvironmentConfigResource struct {
	Api  *datadogV2.FeatureFlagsApi
	Auth context.Context
}

type featureFlagEnvironmentConfigResourceModel struct {
	ID               types.String `tfsdk:"id"`
	FeatureFlagID    types.String `tfsdk:"feature_flag_id"`
	EnvironmentID    types.String `tfsdk:"environment_id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Status           types.String `tfsdk:"status"`
	DefaultVariantID types.String `tfsdk:"default_variant_id"`
}

func (r *featureFlagEnvironmentConfigResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "feature_flag_environment_config"
}

func (r *featureFlagEnvironmentConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Controls a feature flag's enabled state in one environment. Deleting this resource disables the flag; it does not delete the flag or environment.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"feature_flag_id": schema.StringAttribute{
				Description: "UUID of the feature flag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "UUID of the feature flag environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the flag is enabled in this environment.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status returned by Datadog: ENABLED or DISABLED.",
				Computed:    true,
			},
			"default_variant_id": schema.StringAttribute{
				Description: "UUID of the environment's default variant. The public API exposes this value on reads but has no operation to change it.",
				Computed:    true,
			},
		},
	}
}

func (r *featureFlagEnvironmentConfigResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetFeatureFlagsApiV2()
	r.Auth = providerData.Auth
}

func (r *featureFlagEnvironmentConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan featureFlagEnvironmentConfigResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	flagID, environmentID, err := r.setEnabledFromModel(plan)
	if err != nil {
		response.Diagnostics.AddError("Feature flag environment configuration failed", err.Error())
		return
	}
	if err := r.refresh(flagID, environmentID, &plan); err != nil {
		response.Diagnostics.AddError("Feature flag environment read failed", err.Error())
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *featureFlagEnvironmentConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state featureFlagEnvironmentConfigResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	flagID, environmentID, err := parseFeatureFlagCompositeID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag environment config ID", err.Error())
		return
	}
	state.FeatureFlagID = types.StringValue(flagID.String())
	state.EnvironmentID = types.StringValue(environmentID.String())
	if err := r.refresh(flagID, environmentID, &state); err != nil {
		if errors.Is(err, errFeatureFlagEnvironmentNotFound) || isFeatureFlagHTTPNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Feature flag environment read failed", err.Error())
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *featureFlagEnvironmentConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan featureFlagEnvironmentConfigResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	flagID, environmentID, err := r.setEnabledFromModel(plan)
	if err != nil {
		response.Diagnostics.AddError("Feature flag environment configuration failed", err.Error())
		return
	}
	if err := r.refresh(flagID, environmentID, &plan); err != nil {
		response.Diagnostics.AddError("Feature flag environment read failed", err.Error())
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *featureFlagEnvironmentConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state featureFlagEnvironmentConfigResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	flagID, environmentID, err := parseFeatureFlagCompositeID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag environment config ID", err.Error())
		return
	}
	httpResponse, err := r.Api.DisableFeatureFlagEnvironment(r.Auth, flagID, environmentID)
	if err != nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound) {
		response.Diagnostics.AddError("Feature flag environment disable failed", utils.TranslateClientError(err, httpResponse, "error disabling feature flag environment").Error())
	}
}

func (r *featureFlagEnvironmentConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importFeatureFlagCompositeState(ctx, request, response)
}

func (r *featureFlagEnvironmentConfigResource) setEnabledFromModel(state featureFlagEnvironmentConfigResourceModel) (uuid.UUID, uuid.UUID, error) {
	flagID, err := parseFeatureFlagUUID(state.FeatureFlagID.ValueString(), "feature_flag_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	environmentID, err := parseFeatureFlagUUID(state.EnvironmentID.ValueString(), "environment_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	var httpResponse *http.Response
	if state.Enabled.ValueBool() {
		httpResponse, err = r.Api.EnableFeatureFlagEnvironment(r.Auth, flagID, environmentID)
	} else {
		httpResponse, err = r.Api.DisableFeatureFlagEnvironment(r.Auth, flagID, environmentID)
	}
	if err != nil {
		return uuid.Nil, uuid.Nil, utils.TranslateClientError(err, httpResponse, "error setting feature flag environment state")
	}
	return flagID, environmentID, nil
}

func (r *featureFlagEnvironmentConfigResource) refresh(flagID, environmentID uuid.UUID, state *featureFlagEnvironmentConfigResourceModel) error {
	flag, httpResponse, err := r.Api.GetFeatureFlag(r.Auth, flagID)
	if err != nil {
		return featureFlagResponseError(err, httpResponse, "error reading feature flag")
	}
	if err := checkFeatureFlagResponse(&flag); err != nil {
		return err
	}
	raw, err := findFeatureFlagEnvironment(&flag, environmentID)
	if err != nil {
		return err
	}
	status, _ := raw["status"].(string)
	if status != string(datadogV2.FEATUREFLAGSTATUS_ENABLED) && status != string(datadogV2.FEATUREFLAGSTATUS_DISABLED) {
		return fmt.Errorf("feature flag environment %s has unexpected status %q", environmentID, status)
	}
	state.ID = types.StringValue(featureFlagCompositeID(flagID, environmentID))
	state.FeatureFlagID = types.StringValue(flagID.String())
	state.EnvironmentID = types.StringValue(environmentID.String())
	state.Status = types.StringValue(status)
	state.Enabled = types.BoolValue(status == string(datadogV2.FEATUREFLAGSTATUS_ENABLED))
	if defaultVariantID, _ := raw["default_variant_id"].(string); defaultVariantID != "" {
		state.DefaultVariantID = types.StringValue(defaultVariantID)
	} else {
		state.DefaultVariantID = types.StringNull()
	}
	return nil
}
