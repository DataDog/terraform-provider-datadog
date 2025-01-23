package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		Description: "Provides a Datadog App resource. This can be used to create and manage a Datadog App",
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

func (r *appResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state appResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetApp(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving App"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	diags := r.updateStateForRead(ctx, &state, &resp)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state appResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildCreateAppRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateApp(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating App"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	state.ID = types.StringValue(resp.Data.GetId())

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state appResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildUpdateAppRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateApp(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating App"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	diags = r.updateStateForUpdate(ctx, &state, &resp)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state appResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	_, httpResp, err := r.Api.DeleteApp(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting App"))
		return
	}
}

func (r *appResource) updateStateForRead(ctx context.Context, state *appResourceModel, resp *datadogV2.GetAppResponse) diag.Diagnostics {
	diags := diag.Diagnostics{}
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	// use indent to format the json string and match the input
	bytes, err := json.MarshalIndent(attributes, "", "  ")
	if err != nil {
		diags.AddError("attributes", fmt.Sprintf("error marshaling attributes: %s", err))
		return diags
	}
	state.AppJson = types.StringValue(string(bytes) + "\n") // add newline to match TF input (b/c it uses json encoder)
	return nil
}

func (r *appResource) updateStateForUpdate(ctx context.Context, state *appResourceModel, resp *datadogV2.UpdateAppResponse) diag.Diagnostics {
	diags := diag.Diagnostics{}
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	// use indent to format the json string and match the input
	bytes, err := json.MarshalIndent(attributes, "", "  ")
	if err != nil {
		diags.AddError("attributes", fmt.Sprintf("error marshaling attributes: %s", err))
		return diags
	}
	state.AppJson = types.StringValue(string(bytes) + "\n") // add newline to match TF input (b/c it uses json encoder)
	return nil
}

func (r *appResource) buildCreateAppRequestBody(ctx context.Context, state *appResourceModel) (*datadogV2.CreateAppRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCreateAppRequestDataAttributesWithDefaults()

	tflog.Debug(ctx, "tflog app json", map[string]interface{}{"app json": state.AppJson.String()})

	// TODO: figure out why it's being encoded twice
	// decode encoded json into string and then decode that into the attributes struct
	var appJsonString string
	err := json.Unmarshal([]byte(state.AppJson.String()), &appJsonString)
	if err != nil {
		diags.AddError("app json", fmt.Sprintf("error unmarshalling app json to string: %s", err))
		return nil, diags
	}

	tflog.Debug(ctx, "tflog app json string", map[string]interface{}{"app json string": appJsonString})

	err = json.Unmarshal([]byte(appJsonString), attributes)
	if err != nil {
		diags.AddError("app json", fmt.Sprintf("error unmarshalling app json string to attributes struct: %s", err))
		return nil, diags
	}

	req := datadogV2.NewCreateAppRequestWithDefaults()
	req.Data = datadogV2.NewCreateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	// keep state consistent after Create as well -> marshal attributes back to appJson string (gets rid of newlines/whitespace)
	// bytes, err := json.Marshal(attributes)
	// if err != nil {
	// 	diags.AddError("app json", fmt.Sprintf("error marshalling attributes struct back to json: %s", err))
	// 	return nil, diags
	// }
	// state.AppJson = types.StringValue(string(bytes))

	return req, diags
}

func (r *appResource) buildUpdateAppRequestBody(ctx context.Context, state *appResourceModel) (*datadogV2.UpdateAppRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUpdateAppRequestDataAttributesWithDefaults()

	// decode encoded json into string and then decode that into the attributes struct
	var appJsonString string
	err := json.Unmarshal([]byte(state.AppJson.String()), &appJsonString)
	if err != nil {
		diags.AddError("app json", fmt.Sprintf("error unmarshalling app json to string: %s", err))
		return nil, diags
	}

	err = json.Unmarshal([]byte(appJsonString), attributes)
	if err != nil {
		diags.AddError("app json", fmt.Sprintf("error unmarshalling app json string to attributes struct: %s", err))
		return nil, diags
	}

	req := datadogV2.NewUpdateAppRequestWithDefaults()
	req.Data = datadogV2.NewUpdateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
