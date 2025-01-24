package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &appResource{}
	_ resource.ResourceWithImportState = &appResource{}
)

type appResource struct {
	Api  *datadogV2.AppsApi
	Auth context.Context
}

// try single property JSON input -> validation will be handled on the API side
type appResourceModel struct {
	ID      types.String `tfsdk:"id"`
	AppJson types.String `tfsdk:"app_json"`
}

func NewAppResource() resource.Resource {
	return &appResource{}
}

func (r *appResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAppsApiV2()
	r.Auth = providerData.Auth
}

func (r *appResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "app"
}

func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog App resource. This can be used to create and manage a Datadog App from the App Builder product.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"app_json": schema.StringAttribute{
				Required:    true,
				Description: "The JSON representation of the App.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (r *appResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *appResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan appResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	createRequest, err := appModelToCreateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("Error building create app request", err.Error())
		return
	}

	resp, httpResp, err := r.Api.CreateApp(r.Auth, *createRequest)
	if err != nil {
		if httpResp != nil {
			// error body may have useful info for the user
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("Error reading error response", err.Error())
				return
			}
			response.Diagnostics.AddError("Error creating app", string(body))
		} else {
			response.Diagnostics.AddError("Error creating app", err.Error())
		}
		return
	}

	// set computed values
	plan.ID = types.StringValue(resp.Data.GetId())

	// Save data into Terraform state
	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *appResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state appResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	appModel, err := readApp(r.Auth, r.Api, id)
	if err != nil {
		response.Diagnostics.AddError("Error reading app", err.Error())
		return
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, appModel)
	response.Diagnostics.Append(diags...)
}

func (r *appResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan appResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	updateRequest, err := appModelToUpdateApiRequest(plan)
	if err != nil {
		response.Diagnostics.AddError("Error building update app request", err.Error())
		return
	}

	_, httpResp, err := r.Api.UpdateApp(r.Auth, id, *updateRequest)
	if err != nil {
		if httpResp != nil {
			// error body may have useful info for the user
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("Error reading error response", err.Error())
				return
			}
			response.Diagnostics.AddError("Error updating app", string(body))
		} else {
			response.Diagnostics.AddError("Error updating app", err.Error())
		}
		return
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *appResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state appResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	_, httpResp, err := r.Api.DeleteApp(r.Auth, id)
	if err != nil {
		if httpResp != nil {
			// error body may have useful info for the user
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				response.Diagnostics.AddError("Error reading error response", err.Error())
				return
			}
			response.Diagnostics.AddError("Error deleting app", string(body))
		} else {
			response.Diagnostics.AddError("Error deleting app", err.Error())
		}
		return
	}
}

func apiResponseToAppModel(resp datadogV2.GetAppResponse) (*appResourceModel, error) {
	appModel := &appResourceModel{
		ID: types.StringValue(resp.Data.GetId()),
	}

	data := resp.GetData()
	attributes := data.GetAttributes()

	bytes, err := json.Marshal(attributes)
	if err != nil {
		err = fmt.Errorf("error marshaling attributes: %s", err)
		return nil, err
	}
	appModel.AppJson = types.StringValue(string(bytes))

	return appModel, nil
}

func appModelToCreateApiRequest(appModel appResourceModel) (*datadogV2.CreateAppRequest, error) {
	attributes := datadogV2.NewCreateAppRequestDataAttributesWithDefaults()

	// decode encoded json into string and then decode that into the attributes struct
	var appJsonString string
	err := json.Unmarshal([]byte(appModel.AppJson.String()), &appJsonString)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app json to string: %s", err)
		return nil, err
	}

	err = json.Unmarshal([]byte(appJsonString), attributes)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app json string to attributes struct: %s", err)
		return nil, err
	}

	req := datadogV2.NewCreateAppRequestWithDefaults()
	req.Data = datadogV2.NewCreateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

func appModelToUpdateApiRequest(plan appResourceModel) (*datadogV2.UpdateAppRequest, error) {
	attributes := datadogV2.NewUpdateAppRequestDataAttributesWithDefaults()

	// decode encoded json into string and then decode that into the attributes struct
	var appJsonString string
	err := json.Unmarshal([]byte(plan.AppJson.String()), &appJsonString)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app json to string: %s", err)
		return nil, err
	}

	err = json.Unmarshal([]byte(appJsonString), attributes)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app json string to attributes struct: %s", err)
		return nil, err
	}

	req := datadogV2.NewUpdateAppRequestWithDefaults()
	req.Data = datadogV2.NewUpdateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

// Read logic is shared between data source and resource
func readApp(ctx context.Context, api *datadogV2.AppsApi, id string) (*appResourceModel, error) {
	resp, httpResp, err := api.GetApp(ctx, id)
	if err != nil {
		if httpResp != nil {
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, fmt.Errorf("could not read error response")
			}
			return nil, fmt.Errorf("%s", body)
		}
		return nil, err
	}

	appModel, err := apiResponseToAppModel(resp)
	if err != nil {
		return nil, err
	}

	return appModel, nil
}
