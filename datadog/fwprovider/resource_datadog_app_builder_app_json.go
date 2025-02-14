package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes" // v0.1.0, else breaking
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &appBuilderAppJSONResource{}
	_ resource.ResourceWithImportState = &appBuilderAppJSONResource{}
)

type appBuilderAppJSONResource struct {
	Api  *datadogV2.AppBuilderApi
	Auth context.Context
}

// try single property JSON input -> validation will be handled on the API side
type appBuilderAppJSONResourceModel struct {
	ID      types.String         `tfsdk:"id"`
	AppJson jsontypes.Normalized `tfsdk:"app_json"`
}

func NewAppBuilderAppJSONResource() resource.Resource {
	return &appBuilderAppJSONResource{}
}

func (r *appBuilderAppJSONResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAppBuilderApiV2()
	r.Auth = providerData.Auth
}

func (r *appBuilderAppJSONResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "app_builder_app_json"
}

func (r *appBuilderAppJSONResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog App JSON resource for creating and managing Datadog Apps from App Builder using the JSON definition.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"app_json": schema.StringAttribute{
				Required:    true,
				Description: "The JSON representation of the App.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				CustomType: jsontypes.NormalizedType{},
			},
		},
	}
}

func (r *appBuilderAppJSONResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *appBuilderAppJSONResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan appBuilderAppJSONResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	createRequest, err := appBuilderAppJSONModelToCreateApiRequest(plan)
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
	plan.ID = types.StringValue(resp.Data.GetId().String())

	// Save data into Terraform state
	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
}

func (r *appBuilderAppJSONResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state appBuilderAppJSONResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error parsing id as uuid", err.Error())
		return
	}

	appBuilderAppJSONModel, err := readAppBuilderAppJSON(r.Auth, r.Api, id)
	if err != nil {
		response.Diagnostics.AddError("Error reading app", err.Error())
		return
	}

	// Save data into Terraform state
	diags = response.State.Set(ctx, appBuilderAppJSONModel)
	response.Diagnostics.Append(diags...)
}

func (r *appBuilderAppJSONResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan appBuilderAppJSONResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error parsing id as uuid", err.Error())
		return
	}

	updateRequest, err := appBuilderAppJSONModelToUpdateApiRequest(plan)
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

func (r *appBuilderAppJSONResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state appBuilderAppJSONResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error parsing id as uuid", err.Error())
		return
	}
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

func apiResponseToAppBuilderAppJSONModel(resp datadogV2.GetAppResponse) (*appBuilderAppJSONResourceModel, error) {
	appBuilderAppJSONModel := &appBuilderAppJSONResourceModel{
		ID: types.StringValue(resp.Data.GetId().String()),
	}

	data := resp.GetData()
	attributes := data.GetAttributes()

	marshalledBytes, err := json.Marshal(attributes)
	if err != nil {
		err = fmt.Errorf("error marshaling attributes: %s", err)
		return nil, err
	}
	appBuilderAppJSONModel.AppJson = jsontypes.NewNormalizedValue(string(marshalledBytes))

	return appBuilderAppJSONModel, nil
}

func appBuilderAppJSONModelToCreateApiRequest(plan appBuilderAppJSONResourceModel) (*datadogV2.CreateAppRequest, error) {
	attributes := datadogV2.NewCreateAppRequestDataAttributesWithDefaults()

	// decode encoded json into the attributes struct
	err := json.Unmarshal([]byte(plan.AppJson.ValueString()), attributes)
	if err != nil {
		err = fmt.Errorf("error unmarshalling app json string to attributes struct: %s", err)
		return nil, err
	}

	req := datadogV2.NewCreateAppRequestWithDefaults()
	req.Data = datadogV2.NewCreateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, nil
}

func appBuilderAppJSONModelToUpdateApiRequest(plan appBuilderAppJSONResourceModel) (*datadogV2.UpdateAppRequest, error) {
	attributes := datadogV2.NewUpdateAppRequestDataAttributesWithDefaults()

	// decode encoded json into the attributes struct
	err := json.Unmarshal([]byte(plan.AppJson.ValueString()), attributes)
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
func readAppBuilderAppJSON(ctx context.Context, api *datadogV2.AppBuilderApi, id uuid.UUID) (*appBuilderAppJSONResourceModel, error) {
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

	appModel, err := apiResponseToAppBuilderAppJSONModel(resp)
	if err != nil {
		return nil, err
	}

	return appModel, nil
}
