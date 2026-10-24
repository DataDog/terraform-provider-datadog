package fwprovider

import (
	"context"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &featureFlagEnvironmentResource{}
	_ resource.ResourceWithImportState = &featureFlagEnvironmentResource{}
)

func NewFeatureFlagEnvironmentResource() resource.Resource {
	return &featureFlagEnvironmentResource{}
}

type featureFlagEnvironmentResource struct {
	Api  *datadogV2.FeatureFlagsApi
	Auth context.Context
}

type featureFlagEnvironmentResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Key                        types.String `tfsdk:"key"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	Queries                    []string     `tfsdk:"queries"`
	IsProduction               types.Bool   `tfsdk:"is_production"`
	RequireFeatureFlagApproval types.Bool   `tfsdk:"require_feature_flag_approval"`
	CreatedAt                  types.String `tfsdk:"created_at"`
	UpdatedAt                  types.String `tfsdk:"updated_at"`
}

func (r *featureFlagEnvironmentResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "feature_flag_environment"
}

func (r *featureFlagEnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages a Datadog Feature Flags environment definition.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"key": schema.StringAttribute{
				Description: "Server-generated environment key.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Human-readable environment name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Environment description returned by Datadog. The environment write API does not support changing it.",
				Computed:    true,
			},
			"queries": schema.ListAttribute{
				Description: "Datadog tag queries that define the environment scope.",
				Required:    true,
				ElementType: types.StringType,
			},
			"is_production": schema.BoolAttribute{
				Description: "Whether this is a production environment.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"require_feature_flag_approval": schema.BoolAttribute{
				Description: "Whether feature flag changes in this environment require approval.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"created_at": schema.StringAttribute{
				Description: "Time the environment was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Time the environment was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *featureFlagEnvironmentResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetFeatureFlagsApiV2()
	r.Auth = providerData.Auth
}

func (r *featureFlagEnvironmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan featureFlagEnvironmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	attributes := *datadogV2.NewCreateEnvironmentAttributes(plan.Name.ValueString(), plan.Queries)
	attributes.SetIsProduction(plan.IsProduction.ValueBool())
	attributes.SetRequireFeatureFlagApproval(plan.RequireFeatureFlagApproval.ValueBool())
	data := datadogV2.NewCreateEnvironmentData(attributes, datadogV2.CREATEENVIRONMENTDATATYPE_ENVIRONMENTS)
	body := datadogV2.NewCreateEnvironmentRequest(*data)
	created, httpResponse, err := r.Api.CreateFeatureFlagsEnvironment(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error creating feature flag environment"), "Environment creation failed"))
		return
	}

	// The create response can omit server-derived fields such as key.
	environment, httpResponse, err := r.Api.GetFeatureFlagsEnvironment(r.Auth, created.Data.Id)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error reading feature flag environment after creation"), "Environment read failed"))
		return
	}
	if err := utils.CheckForUnparsed(environment); err != nil {
		response.Diagnostics.AddError("Invalid feature flag environment response", err.Error())
		return
	}
	updateFeatureFlagEnvironmentState(&plan, &environment)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *featureFlagEnvironmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state featureFlagEnvironmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, err := parseFeatureFlagUUID(state.ID.ValueString(), "id")
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag environment ID", err.Error())
		return
	}
	environment, httpResponse, err := r.Api.GetFeatureFlagsEnvironment(r.Auth, id)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error reading feature flag environment"), "Environment read failed"))
		return
	}
	if err := utils.CheckForUnparsed(environment); err != nil {
		response.Diagnostics.AddError("Invalid feature flag environment response", err.Error())
		return
	}
	updateFeatureFlagEnvironmentState(&state, &environment)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *featureFlagEnvironmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan featureFlagEnvironmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, err := parseFeatureFlagUUID(plan.ID.ValueString(), "id")
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag environment ID", err.Error())
		return
	}
	name := plan.Name.ValueString()
	attributes := datadogV2.NewUpdateEnvironmentAttributes()
	attributes.Name = &name
	attributes.Queries = plan.Queries
	attributes.SetIsProduction(plan.IsProduction.ValueBool())
	attributes.SetRequireFeatureFlagApproval(plan.RequireFeatureFlagApproval.ValueBool())
	data := datadogV2.NewUpdateEnvironmentData(*attributes, datadogV2.UPDATEENVIRONMENTDATATYPE_ENVIRONMENTS)
	body := datadogV2.NewUpdateEnvironmentRequest(*data)
	environment, httpResponse, err := r.Api.UpdateFeatureFlagsEnvironment(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error updating feature flag environment"), "Environment update failed"))
		return
	}
	if err := utils.CheckForUnparsed(environment); err != nil {
		response.Diagnostics.AddError("Invalid feature flag environment response", err.Error())
		return
	}
	updateFeatureFlagEnvironmentState(&plan, &environment)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *featureFlagEnvironmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state featureFlagEnvironmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, err := parseFeatureFlagUUID(state.ID.ValueString(), "id")
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag environment ID", err.Error())
		return
	}
	httpResponse, err := r.Api.DeleteFeatureFlagsEnvironment(r.Auth, id)
	if err != nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound) {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error deleting feature flag environment"), "Environment delete failed"))
	}
}

func (r *featureFlagEnvironmentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func updateFeatureFlagEnvironmentState(state *featureFlagEnvironmentResourceModel, response *datadogV2.EnvironmentResponse) {
	attributes := response.Data.Attributes
	state.ID = types.StringValue(response.Data.Id.String())
	state.Key = types.StringValue(attributes.GetKey())
	state.Name = types.StringValue(attributes.Name)
	state.Queries = append([]string{}, attributes.Queries...)
	state.IsProduction = types.BoolValue(attributes.GetIsProduction())
	state.RequireFeatureFlagApproval = types.BoolValue(attributes.GetRequireFeatureFlagApproval())
	state.CreatedAt = featureFlagTimeValue(attributes.CreatedAt)
	state.UpdatedAt = featureFlagTimeValue(attributes.UpdatedAt)
	if attributes.Description.IsSet() && attributes.Description.Get() != nil {
		state.Description = types.StringValue(*attributes.Description.Get())
	} else {
		state.Description = types.StringNull()
	}
}
