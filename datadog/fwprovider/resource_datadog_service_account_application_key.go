package fwprovider

import (
	"context"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &serviceAccountApplicationKeyResource{}
	_ resource.ResourceWithImportState = &serviceAccountApplicationKeyResource{}
)

type serviceAccountApplicationKeyResource struct {
	Api       *datadogV2.ServiceAccountsApi
	ActionApi *datadogV2.ActionConnectionApi
	Auth      context.Context
}

type serviceAccountApplicationKeyModel struct {
	ID                     types.String `tfsdk:"id"`
	ServiceAccountId       types.String `tfsdk:"service_account_id"`
	Name                   types.String `tfsdk:"name"`
	Key                    types.String `tfsdk:"key"`
	CreatedAt              types.String `tfsdk:"created_at"`
	Last4                  types.String `tfsdk:"last4"`
	Scopes                 types.Set    `tfsdk:"scopes"`
	EnableActionsApiAccess types.Bool   `tfsdk:"enable_actions_api_access"`
}

func NewServiceAccountApplicationKeyResource() resource.Resource {
	return &serviceAccountApplicationKeyResource{}
}

func (r *serviceAccountApplicationKeyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetServiceAccountsApiV2()
	r.ActionApi = providerData.DatadogApiInstances.GetActionConnectionApiV2()
	r.Auth = providerData.Auth
}

func (r *serviceAccountApplicationKeyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "service_account_application_key"
}

func (r *serviceAccountApplicationKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog `service_account_application_key` resource. This can be used to create and manage Datadog service account application keys.",
		Attributes: map[string]schema.Attribute{
			"service_account_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the service account that owns this key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the application key.",
			},
			"scopes": schema.SetAttribute{
				Description: "Authorization scopes for the Application Key. Application Keys configured with no scopes have full access.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"enable_actions_api_access": schema.BoolAttribute{
				Description: "(Preview) Enable Actions API access for this service account application key. When true, the key will be automatically registered for use with Action Connection, App Builder, and Workflow Automation. Defaults to `false`.",
				Optional:    true,
			},
			"key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The value of the service account application key. This value cannot be imported.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation date of the application key.",
			},
			"last4": schema.StringAttribute{
				Computed:    true,
				Description: "The last four characters of the application key.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *serviceAccountApplicationKeyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("error retrieving service_account_id or application_key_id from given ID", "")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("service_account_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("key"), "")...)
	response.Diagnostics.AddWarning("Importing a service account application key will not import the key value.", "")

}

func (r *serviceAccountApplicationKeyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state serviceAccountApplicationKeyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	serviceAccountId := state.ServiceAccountId.ValueString()

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetServiceAccountApplicationKey(r.Auth, serviceAccountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ServiceAccountApplicationKey"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateStatePartialKey(ctx, &state, &resp)

	// Check Actions API registration status
	isRegistered := r.isRegisteredForActionsApi(state.ID.ValueString())
	state.EnableActionsApiAccess = types.BoolValue(isRegistered)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *serviceAccountApplicationKeyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state serviceAccountApplicationKeyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serviceAccountId := state.ServiceAccountId.ValueString()

	body, diags := r.buildServiceAccountApplicationKeyRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateServiceAccountApplicationKey(r.Auth, serviceAccountId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating ServiceAccountApplicationKey"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateStateFullKey(ctx, &state, &resp)

	// Handle Actions API access registration if enabled
	if !state.EnableActionsApiAccess.IsNull() && state.EnableActionsApiAccess.ValueBool() {
		if err := r.registerForActionsApi(state.ID.ValueString()); err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error registering service account application key for Actions API access"))
			return
		}
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *serviceAccountApplicationKeyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state serviceAccountApplicationKeyModel
	var plan serviceAccountApplicationKeyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serviceAccountId := plan.ServiceAccountId.ValueString()

	id := plan.ID.ValueString()

	body, diags := r.buildServiceAccountApplicationKeyUpdateRequestBody(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateServiceAccountApplicationKey(r.Auth, serviceAccountId, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating ServiceAccountApplicationKey"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateStatePartialKey(ctx, &plan, &resp)

	// Handle Actions API access changes
	oldEnabled := !state.EnableActionsApiAccess.IsNull() && state.EnableActionsApiAccess.ValueBool()
	newEnabled := !plan.EnableActionsApiAccess.IsNull() && plan.EnableActionsApiAccess.ValueBool()

	if oldEnabled != newEnabled {
		if newEnabled {
			if err := r.registerForActionsApi(plan.ID.ValueString()); err != nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error registering service account application key for Actions API access"))
				return
			}
		} else {
			if err := r.unregisterFromActionsApi(plan.ID.ValueString()); err != nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error unregistering service account application key from Actions API access"))
				return
			}
		}
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *serviceAccountApplicationKeyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state serviceAccountApplicationKeyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	serviceAccountId := state.ServiceAccountId.ValueString()

	id := state.ID.ValueString()

	// Unregister from Actions API if registered
	if !state.EnableActionsApiAccess.IsNull() && state.EnableActionsApiAccess.ValueBool() {
		if err := r.unregisterFromActionsApi(state.ID.ValueString()); err != nil {
			// Log warning but don't fail deletion
			response.Diagnostics.AddWarning("Failed to unregister from Actions API", err.Error())
		}
	}

	httpResp, err := r.Api.DeleteServiceAccountApplicationKey(r.Auth, serviceAccountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting service_account_application_key"))
		return
	}
}

func (r *serviceAccountApplicationKeyResource) updateStatePartialKey(ctx context.Context, state *serviceAccountApplicationKeyModel, resp *datadogV2.PartialApplicationKeyResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(*createdAt)
	}

	if last4, ok := attributes.GetLast4Ok(); ok {
		state.Last4 = types.StringValue(*last4)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if attributes.HasScopes() {
		state.Scopes, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetScopes())
	}
}

func (r *serviceAccountApplicationKeyResource) updateStateFullKey(ctx context.Context, state *serviceAccountApplicationKeyModel, resp *datadogV2.ApplicationKeyResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if key, ok := attributes.GetKeyOk(); ok {
		state.Key = types.StringValue(*key)
	}

	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		timeStr := createdAt.Format(time.RFC3339)
		state.CreatedAt = types.StringValue(timeStr)

	}

	if last4, ok := attributes.GetLast4Ok(); ok {
		state.Last4 = types.StringValue(*last4)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if attributes.HasScopes() {
		state.Scopes, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetScopes())
	}
}

func (r *serviceAccountApplicationKeyResource) buildServiceAccountApplicationKeyRequestBody(ctx context.Context, state *serviceAccountApplicationKeyModel) (*datadogV2.ApplicationKeyCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewApplicationKeyCreateAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	attributes.SetScopes(getScopesFromStateAttribute(state.Scopes))

	req := datadogV2.NewApplicationKeyCreateRequestWithDefaults()
	req.Data = *datadogV2.NewApplicationKeyCreateDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *serviceAccountApplicationKeyResource) buildServiceAccountApplicationKeyUpdateRequestBody(ctx context.Context, state *serviceAccountApplicationKeyModel) (*datadogV2.ApplicationKeyUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewApplicationKeyUpdateAttributesWithDefaults()

	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}

	attributes.SetScopes(getScopesFromStateAttribute(state.Scopes))
	req := datadogV2.NewApplicationKeyUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewApplicationKeyUpdateDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	if !state.ID.IsNull() {
		req.Data.SetId(state.ID.ValueString())
	}

	return req, diags
}

// registerForActionsApi registers the service account application key for Actions API access
func (r *serviceAccountApplicationKeyResource) registerForActionsApi(keyId string) error {
	_, _, err := r.ActionApi.RegisterAppKey(r.Auth, keyId)
	return err
}

// unregisterFromActionsApi unregisters the service account application key from Actions API access
func (r *serviceAccountApplicationKeyResource) unregisterFromActionsApi(keyId string) error {
	_, err := r.ActionApi.UnregisterAppKey(r.Auth, keyId)
	return err
}

// isRegisteredForActionsApi checks if the service account application key is registered for Actions API access
func (r *serviceAccountApplicationKeyResource) isRegisteredForActionsApi(keyId string) bool {
	_, _, err := r.ActionApi.GetAppKeyRegistration(r.Auth, keyId)
	return err == nil
}
