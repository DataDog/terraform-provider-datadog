package fwprovider

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &appKeyRegistrationResource{}
	_ resource.ResourceWithImportState = &appKeyRegistrationResource{}
)

type appKeyRegistrationResource struct {
	Api  *datadogV2.ActionConnectionApi
	Auth context.Context
}

type appKeyRegistrationResourceModel struct {
	ID types.String `tfsdk:"id"`
}

func NewAppKeyRegistrationResource() resource.Resource {
	return &appKeyRegistrationResource{}
}

func (r *appKeyRegistrationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetActionConnectionApiV2()
	r.Auth = providerData.Auth
}

func (r *appKeyRegistrationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "app_key_registration"
}

func (r *appKeyRegistrationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Registers App Keys to be used for Action Connection, App Builder, and Workflow Automation. This registration is required to enable API and Terraform use in these products.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The Application Key ID to register.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // This replace means if the ID is changed the previous app key registration will be deleted and a new one is created.
				},
			},
		},
	}
}

func (r *appKeyRegistrationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *appKeyRegistrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan appKeyRegistrationResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("error parsing ID as UUID", err.Error())
		return
	}

	_, httpResp, err := r.Api.RegisterAppKey(r.Auth, id.String())
	if err != nil {
		if httpResp != nil {
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("Error reading error response when registering app key", err.Error())
				return
			}
			response.Diagnostics.AddError("Error registering app key", string(body))
		} else {
			response.Diagnostics.AddError("Error registering app key", err.Error())
		}
		return
	}

	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *appKeyRegistrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state appKeyRegistrationResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	appKeyRegistration, httpResp, err := r.Api.GetAppKeyRegistration(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil {
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("Error reading error response when getting app key registration", err.Error())
				return
			}

			// If the app key registration is not found, we log a warning and remove the resource from state. This may be due to changes in the UI or attempting to import an app key that is not registered.
			if httpResp.StatusCode == http.StatusNotFound {
				response.Diagnostics.AddWarning("The application key with ID '"+state.ID.ValueString()+"' is not registered. It may have been unregistered outside of Terraform.", string(body))
				response.State.RemoveResource(ctx)
				return
			}

			response.Diagnostics.AddError("Error getting app key registration", string(body))
		} else {
			response.Diagnostics.AddError("Error getting app key registration", err.Error())
		}
		return
	}

	appKeyRegistrationModel := &appKeyRegistrationResourceModel{
		ID: types.StringValue(appKeyRegistration.Data.Id.String()),
	}

	diags = response.State.Set(ctx, appKeyRegistrationModel)
	response.Diagnostics.Append(diags...)
}

func (r *appKeyRegistrationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update should not be called", "Updating this resource should replace it.")
}

func (r *appKeyRegistrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state appKeyRegistrationResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.UnregisterAppKey(r.Auth, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error deleting app key registration", err.Error())
		return
	}

	if httpResp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			response.Diagnostics.AddError("Error reading error response when deleting app key registration", err.Error())
		} else {
			response.Diagnostics.AddError("Error deleting app key registration", string(body))
		}
	}
}
