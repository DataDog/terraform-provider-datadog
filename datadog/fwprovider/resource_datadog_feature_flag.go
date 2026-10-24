package fwprovider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure      = &featureFlagResource{}
	_ resource.ResourceWithImportState    = &featureFlagResource{}
	_ resource.ResourceWithValidateConfig = &featureFlagResource{}
)

func NewFeatureFlagResource() resource.Resource {
	return &featureFlagResource{}
}

type featureFlagResource struct {
	Api  *datadogV2.FeatureFlagsApi
	Auth context.Context
}

type featureFlagResourceModel struct {
	ID                types.String              `tfsdk:"id"`
	Key               types.String              `tfsdk:"key"`
	Name              types.String              `tfsdk:"name"`
	Description       types.String              `tfsdk:"description"`
	ValueType         types.String              `tfsdk:"value_type"`
	DefaultVariantKey types.String              `tfsdk:"default_variant_key"`
	JSONSchema        jsontypes.Normalized      `tfsdk:"json_schema"`
	RequireApproval   types.Bool                `tfsdk:"require_approval"`
	UnarchiveOnCreate types.Bool                `tfsdk:"unarchive_on_create"`
	Variants          []featureFlagVariantModel `tfsdk:"variants"`
	CreatedAt         types.String              `tfsdk:"created_at"`
	UpdatedAt         types.String              `tfsdk:"updated_at"`
	ArchivedAt        types.String              `tfsdk:"archived_at"`
}

type featureFlagVariantModel struct {
	Key   types.String `tfsdk:"key"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (r *featureFlagResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "feature_flag"
}

func (r *featureFlagResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages a Datadog Feature Flags flag. Deleting this resource archives the flag because the API does not expose a physical delete.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"key": schema.StringAttribute{
				Description: "Stable, organization-unique key referenced by SDKs. Archived flags retain their keys.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name shown in Datadog.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description shown in Datadog.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"value_type": schema.StringAttribute{
				Description: "Type of every variant value. Valid values are BOOLEAN, INTEGER, NUMERIC, STRING, and JSON.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("BOOLEAN", "INTEGER", "NUMERIC", "STRING", "JSON"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"default_variant_key": schema.StringAttribute{
				Description: "Key of the variant returned when no allocation matches. The API does not return this as a flag-level field, so an explicitly configured value is preserved in state.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"json_schema": schema.StringAttribute{
				Description: "JSON Schema used to validate variant values when value_type is JSON. JSON formatting differences are normalized.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"require_approval": schema.BoolAttribute{
				Description: "Whether flag changes require approval. Setting this requires the Approvals Override permission and changes force replacement because the update API does not support it.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"unarchive_on_create": schema.BoolAttribute{
				Description: "When creation collides with an archived flag of the same key, adopt and unarchive it, then reconcile its variants and mutable fields. The value type, approval setting, and any readable default variant must match.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"variants": schema.SetNestedAttribute{
				Description: "Variants the flag can return. Changes force replacement; use unarchive_on_create when retaining the same key.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Stable variant key, unique within this flag.",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Human-readable variant name.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Variant value encoded as a string, such as true, 42, or a JSON literal.",
							Required:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Time the flag was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Time the flag was last updated.",
				Computed:    true,
			},
			"archived_at": schema.StringAttribute{
				Description: "Time the flag was archived, or null when it is active.",
				Computed:    true,
			},
		},
	}
}

func (r *featureFlagResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetFeatureFlagsApiV2()
	r.Auth = providerData.Auth
}

func (r *featureFlagResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config featureFlagResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}
	keys := make(map[string]struct{}, len(config.Variants))
	for _, variant := range config.Variants {
		if variant.Key.IsNull() || variant.Key.IsUnknown() {
			continue
		}
		key := variant.Key.ValueString()
		if _, exists := keys[key]; exists {
			response.Diagnostics.AddError("Duplicate feature flag variant key", fmt.Sprintf("Variant key %q is declared more than once.", key))
		}
		keys[key] = struct{}{}
	}
	if !config.DefaultVariantKey.IsNull() && !config.DefaultVariantKey.IsUnknown() {
		if _, exists := keys[config.DefaultVariantKey.ValueString()]; !exists {
			response.Diagnostics.AddError("Unknown default variant", fmt.Sprintf("default_variant_key %q must match one of the declared variant keys.", config.DefaultVariantKey.ValueString()))
		}
	}
}

func (r *featureFlagResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan featureFlagResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, httpResponse, err := r.Api.CreateFeatureFlag(r.Auth, buildCreateFeatureFlagRequest(plan))
	if err != nil {
		if plan.UnarchiveOnCreate.ValueBool() && httpResponse != nil && (httpResponse.StatusCode == http.StatusConflict || httpResponse.StatusCode == http.StatusBadRequest) {
			r.createFromArchivedFlag(ctx, &plan, response)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error creating feature flag"), "Feature flag creation failed"))
		return
	}
	if err := checkFeatureFlagResponse(&created); err != nil {
		response.Diagnostics.AddError("Invalid feature flag response", err.Error())
		return
	}
	updateFeatureFlagState(&plan, &created)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *featureFlagResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state featureFlagResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, err := parseFeatureFlagUUID(state.ID.ValueString(), "id")
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag ID", err.Error())
		return
	}
	flag, httpResponse, err := r.Api.GetFeatureFlag(r.Auth, id)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error reading feature flag"), "Feature flag read failed"))
		return
	}
	if err := checkFeatureFlagResponse(&flag); err != nil {
		response.Diagnostics.AddError("Invalid feature flag response", err.Error())
		return
	}
	updateFeatureFlagState(&state, &flag)
	if state.UnarchiveOnCreate.IsNull() || state.UnarchiveOnCreate.IsUnknown() {
		state.UnarchiveOnCreate = types.BoolValue(false)
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *featureFlagResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan featureFlagResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, err := parseFeatureFlagUUID(plan.ID.ValueString(), "id")
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag ID", err.Error())
		return
	}
	updated, httpResponse, err := r.updateFeatureFlag(id, plan)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error updating feature flag"), "Feature flag update failed"))
		return
	}
	if err := checkFeatureFlagResponse(updated); err != nil {
		response.Diagnostics.AddError("Invalid feature flag response", err.Error())
		return
	}
	updateFeatureFlagState(&plan, updated)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *featureFlagResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state featureFlagResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, err := parseFeatureFlagUUID(state.ID.ValueString(), "id")
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag ID", err.Error())
		return
	}
	if !state.ArchivedAt.IsNull() && state.ArchivedAt.ValueString() != "" {
		return
	}
	_, httpResponse, err := r.Api.ArchiveFeatureFlag(r.Auth, id)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error archiving feature flag"), "Feature flag delete failed"))
	}
}

func (r *featureFlagResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func buildCreateFeatureFlagRequest(plan featureFlagResourceModel) datadogV2.CreateFeatureFlagRequest {
	variants := make([]datadogV2.CreateVariant, 0, len(plan.Variants))
	for _, variant := range plan.Variants {
		variants = append(variants, *datadogV2.NewCreateVariant(variant.Key.ValueString(), variant.Name.ValueString(), variant.Value.ValueString()))
	}
	attributes := datadogV2.NewCreateFeatureFlagAttributes(
		plan.Key.ValueString(),
		plan.Name.ValueString(),
		datadogV2.ValueType(plan.ValueType.ValueString()),
		variants,
	)
	description := plan.Description.ValueString()
	attributes.Description = &description
	if !plan.RequireApproval.IsNull() && !plan.RequireApproval.IsUnknown() {
		requireApproval := plan.RequireApproval.ValueBool()
		attributes.RequireApproval = &requireApproval
	}
	if !plan.DefaultVariantKey.IsNull() && !plan.DefaultVariantKey.IsUnknown() {
		value := plan.DefaultVariantKey.ValueString()
		attributes.DefaultVariantKey = *datadog.NewNullableString(&value)
	}
	if !plan.JSONSchema.IsNull() && !plan.JSONSchema.IsUnknown() {
		value := plan.JSONSchema.ValueString()
		attributes.JsonSchema = *datadog.NewNullableString(&value)
	}
	data := datadogV2.NewCreateFeatureFlagData(*attributes, datadogV2.CREATEFEATUREFLAGDATATYPE_FEATURE_FLAGS)
	return *datadogV2.NewCreateFeatureFlagRequest(*data)
}

func buildUpdateFeatureFlagRequest(plan featureFlagResourceModel) datadogV2.UpdateFeatureFlagRequest {
	attributes := datadogV2.NewUpdateFeatureFlagAttributes()
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()
	attributes.Name = &name
	attributes.Description = &description
	attributes.JsonSchema.Set(nil)
	if !plan.JSONSchema.IsNull() && !plan.JSONSchema.IsUnknown() {
		value := plan.JSONSchema.ValueString()
		attributes.JsonSchema.Set(&value)
	}
	data := datadogV2.NewUpdateFeatureFlagData(*attributes, datadogV2.UPDATEFEATUREFLAGDATATYPE_FEATURE_FLAGS)
	return *datadogV2.NewUpdateFeatureFlagRequest(*data)
}

func (r *featureFlagResource) updateFeatureFlag(id uuid.UUID, plan featureFlagResourceModel) (*datadogV2.FeatureFlagResponse, *http.Response, error) {
	updated, httpResponse, err := r.Api.UpdateFeatureFlag(r.Auth, id, buildUpdateFeatureFlagRequest(plan))
	if err != nil {
		return nil, httpResponse, err
	}
	return &updated, httpResponse, nil
}

func updateFeatureFlagState(state *featureFlagResourceModel, response *datadogV2.FeatureFlagResponse) {
	attributes := response.Data.Attributes
	state.ID = types.StringValue(response.Data.Id.String())
	state.Key = types.StringValue(attributes.Key)
	state.Name = types.StringValue(attributes.Name)
	state.Description = types.StringValue(attributes.Description)
	state.ValueType = types.StringValue(string(attributes.ValueType))
	state.RequireApproval = types.BoolValue(attributes.GetRequireApproval())
	state.CreatedAt = featureFlagTimeValue(attributes.CreatedAt)
	state.UpdatedAt = featureFlagTimeValue(attributes.UpdatedAt)
	if attributes.ArchivedAt.IsSet() {
		state.ArchivedAt = featureFlagTimeValue(attributes.ArchivedAt.Get())
	} else {
		state.ArchivedAt = types.StringNull()
	}
	if attributes.JsonSchema.IsSet() && attributes.JsonSchema.Get() != nil {
		state.JSONSchema = jsontypes.NewNormalizedValue(*attributes.JsonSchema.Get())
	} else {
		state.JSONSchema = jsontypes.NewNormalizedNull()
	}
	state.Variants = make([]featureFlagVariantModel, 0, len(attributes.Variants))
	for _, variant := range attributes.Variants {
		state.Variants = append(state.Variants, featureFlagVariantModel{
			Key:   types.StringValue(variant.Key),
			Name:  types.StringValue(variant.Name),
			Value: types.StringValue(variant.Value),
		})
	}
	if state.DefaultVariantKey.IsNull() || state.DefaultVariantKey.IsUnknown() {
		state.DefaultVariantKey = defaultFeatureFlagVariantKey(response)
	}
}

func (r *featureFlagResource) createFromArchivedFlag(ctx context.Context, plan *featureFlagResourceModel, response *resource.CreateResponse) {
	parameters := datadogV2.NewListFeatureFlagsOptionalParameters().WithKey(plan.Key.ValueString()).WithIsArchived(true).WithPageLimit(100)
	list, httpResponse, err := r.Api.ListFeatureFlags(r.Auth, *parameters)
	if err != nil {
		response.Diagnostics.AddError("Archived feature flag lookup failed", utils.TranslateClientError(err, httpResponse, "error finding archived feature flag").Error())
		return
	}
	var archived *datadogV2.FeatureFlagListItem
	for i := range list.Data {
		if list.Data[i].Attributes.Key == plan.Key.ValueString() {
			archived = &list.Data[i]
			break
		}
	}
	if archived == nil {
		response.Diagnostics.AddError("Feature flag key collision", fmt.Sprintf("No archived flag with key %q was found. Set unarchive_on_create only when adopting an archived flag, or choose another key.", plan.Key.ValueString()))
		return
	}
	if err := validateArchivedFeatureFlagAdoption(*archived, *plan); err != nil {
		response.Diagnostics.AddError("Archived feature flag cannot be adopted", err.Error())
		return
	}

	flagID := archived.Id
	unarchived, httpResponse, err := r.Api.UnarchiveFeatureFlag(r.Auth, flagID)
	if err != nil {
		response.Diagnostics.AddError("Feature flag unarchive failed", utils.TranslateClientError(err, httpResponse, "error unarchiving feature flag").Error())
		return
	}
	if err := r.syncFeatureFlagVariants(flagID, unarchived.Data.Attributes.Variants, plan.Variants); err != nil {
		response.Diagnostics.AddError("Archived feature flag was unarchived but not fully reconciled", fmt.Sprintf("Flag %s is active. Reconcile it manually or import it after fixing the error: %v", flagID, err))
		return
	}
	if _, httpResponse, err := r.updateFeatureFlag(flagID, *plan); err != nil {
		response.Diagnostics.AddError("Archived feature flag was unarchived but not fully reconciled", fmt.Sprintf("Flag %s is active. Reconcile it manually or import it after fixing the error: %v", flagID, utils.TranslateClientError(err, httpResponse, "error updating adopted feature flag")))
		return
	}
	flag, httpResponse, err := r.Api.GetFeatureFlag(r.Auth, flagID)
	if err != nil {
		response.Diagnostics.AddError("Adopted feature flag read failed", utils.TranslateClientError(err, httpResponse, "error reading adopted feature flag").Error())
		return
	}
	if err := checkFeatureFlagResponse(&flag); err != nil {
		response.Diagnostics.AddError("Invalid feature flag response", err.Error())
		return
	}
	updateFeatureFlagState(plan, &flag)
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func validateArchivedFeatureFlagAdoption(archived datadogV2.FeatureFlagListItem, plan featureFlagResourceModel) error {
	attributes := archived.Attributes
	if string(attributes.ValueType) != plan.ValueType.ValueString() {
		return fmt.Errorf("value_type is immutable: archived flag has %s but configuration has %s", attributes.ValueType, plan.ValueType.ValueString())
	}
	if !plan.RequireApproval.IsNull() && !plan.RequireApproval.IsUnknown() && attributes.GetRequireApproval() != plan.RequireApproval.ValueBool() {
		return fmt.Errorf("require_approval is immutable and differs on the archived flag")
	}
	if plan.DefaultVariantKey.IsNull() || plan.DefaultVariantKey.IsUnknown() {
		return nil
	}
	variantKeys := make(map[uuid.UUID]string, len(attributes.Variants))
	for _, variant := range attributes.Variants {
		variantKeys[variant.Id] = variant.Key
	}
	desired := plan.DefaultVariantKey.ValueString()
	for _, environment := range attributes.FeatureFlagEnvironments {
		if !environment.DefaultVariantId.IsSet() || environment.DefaultVariantId.Get() == nil {
			continue
		}
		id, err := uuid.Parse(*environment.DefaultVariantId.Get())
		if err != nil {
			continue
		}
		if key := variantKeys[id]; key != "" && key != desired {
			return fmt.Errorf("default_variant_key is immutable: an environment uses variant %q but configuration requests %q", key, desired)
		}
	}
	return nil
}

func (r *featureFlagResource) syncFeatureFlagVariants(flagID uuid.UUID, current []datadogV2.Variant, desired []featureFlagVariantModel) error {
	currentByKey := make(map[string]datadogV2.Variant, len(current))
	for _, variant := range current {
		currentByKey[variant.Key] = variant
	}
	desiredKeys := make(map[string]struct{}, len(desired))
	for _, variant := range desired {
		key := variant.Key.ValueString()
		desiredKeys[key] = struct{}{}
		if existing, ok := currentByKey[key]; ok {
			if existing.Name == variant.Name.ValueString() && existing.Value == variant.Value.ValueString() {
				continue
			}
			body := *datadogV2.NewUpdateVariantRequest()
			name := variant.Name.ValueString()
			value := variant.Value.ValueString()
			body.Name = &name
			body.Value = &value
			_, httpResponse, err := r.Api.UpdateVariantForFeatureFlag(r.Auth, flagID, existing.Id, body)
			if err != nil {
				return utils.TranslateClientError(err, httpResponse, fmt.Sprintf("error updating variant %q", key))
			}
			continue
		}
		body := *datadogV2.NewCreateVariant(key, variant.Name.ValueString(), variant.Value.ValueString())
		_, httpResponse, err := r.Api.CreateVariantForFeatureFlag(r.Auth, flagID, body)
		if err != nil {
			return utils.TranslateClientError(err, httpResponse, fmt.Sprintf("error creating variant %q", key))
		}
	}
	for key, variant := range currentByKey {
		if _, keep := desiredKeys[key]; keep {
			continue
		}
		httpResponse, err := r.Api.DeleteVariantFromFeatureFlag(r.Auth, flagID, variant.Id)
		if err != nil {
			return utils.TranslateClientError(err, httpResponse, fmt.Sprintf("error deleting variant %q", key))
		}
	}
	return nil
}
