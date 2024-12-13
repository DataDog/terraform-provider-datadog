package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// TODO: figure out rest of model - typing?
type appResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Tags             types.List   `tfsdk:"tags"`
	RootInstanceName types.String `tfsdk:"root_instance_name"`
	Components       types.List   `tfsdk:"components"`
	EmbeddedQueries  types.List   `tfsdk:"embedded_queries"`
	Scripts          types.List   `tfsdk:"scripts"`
	InputSchema      types.String `tfsdk:"input_schema"`
}

// type GetAppResponseDataAttributes struct {
// 	Components       []ComponentGrid `json:"components,omitempty"`
// 	Description      *string         `json:"description,omitempty"`
// 	EmbeddedQueries  []Query         `json:"embeddedQueries,omitempty"`
// 	Favorite         *bool           `json:"favorite,omitempty"`
// 	InputSchema      *InputSchema    `json:"inputSchema,omitempty"`
// 	Name             *string         `json:"name,omitempty"`
// 	RootInstanceName *string         `json:"rootInstanceName,omitempty"`
// 	Scripts          []Script        `json:"scripts,omitempty"`
// 	Tags             []string        `json:"tags,omitempty"`
// }

// type App struct {
// 	Id               uuid.UUID                        `json:"id" jsonapi:"primary,appDefinitions"`
// 	Name             string                           `json:"name" jsonapi:"attribute"`
// 	Description      string                           `json:"description" jsonapi:"attribute"`
// 	Favorite         bool                             `json:"favorite" jsonapi:"attribute"`
// 	Tags             []string                         `json:"tags" jsonapi:"attribute"`
// 	RootInstanceName string                           `json:"rootInstanceName,omitempty" jsonapi:"attribute"`
// 	Components       []any                            `json:"components,omitempty" jsonapi:"attribute"`
// 	Queries          []*Query                         `json:"embeddedQueries,omitempty" jsonapi:"attribute"`
// 	Scripts          []*Script                        `json:"scripts,omitempty" jsonapi:"attribute"`
// 	InputSchema      *InputSchema                     `json:"inputSchema,omitempty" jsonapi:"attribute"`
// 	Deployment       *Deployment                      `json:"deployment,omitempty" jsonapi:"relationship"`
// 	Connections      []*actionserver.CustomConnection `json:"connections,omitempty" jsonapi:"relationship"`
// 	Meta             AppMeta                          `json:"-" jsonapi:"meta"`
// 	SelfService      bool                             `json:"selfService" jsonapi:"attribute"`
// }

// type AppJSON struct {
// 	{
// 		"queries": [],
// 		"id": "2b779a42-41d6-4ce9-b496-02a4ecacaff2",
// 		"components": [
// 		  {
// 			"events": [],
// 			"name": "grid0",
// 			"properties": {
// 			  "children": []
// 			},
// 			"type": "grid"
// 		  }
// 		],
// 		"description": "",
// 		"favorite": false,
// 		"name": "Tiffany's App Mon, Dec 16, 4:56:31 pm",
// 		"rootInstanceName": "grid0",
// 		"selfService": false,
// 		"tags": null
// 	  }
// }

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

// TODO: figure out rest of Schema
func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog App resource. This can be used to create and manage a Datadog App",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				// Required:    true,
				Description: "The name of the App.",
			},
			"description": schema.StringAttribute{
				Description: "The description of the App.",
			},
			"tags": schema.ListAttribute{
				Description: "The tags of the App.",
				ElementType: types.StringType,
			},
			"root_instance_name": schema.StringAttribute{
				Description: "The root instance name of the App.",
			},
			"components": schema.ListAttribute{
				Description: "The components of the App.",
				ElementType: types.StringType,
			},
			"embedded_queries": schema.ListAttribute{
				Description: "The embedded queries of the App.",
				ElementType: types.StringType,
			},
			"scripts": schema.ListAttribute{
				Description: "The scripts of the App.",
				ElementType: types.StringType,
			},
			"input_schema": schema.StringAttribute{
				Description: "The input schema of the App.",
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

	r.updateStateForRead(ctx, &state, &resp)

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
	r.updateStateForRead(ctx, &state, &resp)

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
	r.updateStateForRead(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state teamModel
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

// TODO
func (r *appResource) updateStateForRead(ctx context.Context, state *appResourceModel, resp *datadogV2.GetAppResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if name, ok := attributes.GetNameOk(); ok && name != nil {
		state.Name = types.StringValue(*name)
	}

	if description, ok := attributes.GetDescriptionOk(); ok && description != nil {
		state.Description = types.StringValue(*description)
	}

	if tags, ok := attributes.GetTagsOk(); ok && tags != nil {
		state.Tags, _ = types.ListValueFrom(ctx, types.StringType, tags)
	}

	if rootInstanceName, ok := attributes.GetRootInstanceNameOk(); ok && rootInstanceName != nil {
		state.RootInstanceName = types.StringValue(*rootInstanceName)
	}

	if components, ok := attributes.GetComponentsOk(); ok && components != nil {
		state.Components, _ = types.ListValueFrom(ctx, types.StringType, components)
	}

	if embeddedQueries, ok := attributes.GetEmbeddedQueriesOk(); ok && embeddedQueries != nil {
		state.EmbeddedQueries, _ = types.ListValueFrom(ctx, types.StringType, embeddedQueries)
	}

	if scripts, ok := attributes.GetScriptsOk(); ok && scripts != nil {
		state.Scripts, _ = types.ListValueFrom(ctx, types.StringType, scripts)
	}

	if inputSchema, ok := attributes.GetInputSchemaOk(); ok && inputSchema != nil {
		state.InputSchema = types.StringValue(*inputSchema)
	}
}

// TODO
func (r *appResource) buildCreateAppRequestBody(ctx context.Context, state *appResourceModel) (*datadogV2.CreateAppRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCreateAppRequestDataAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	// attributes.SetHandle(state.Handle.ValueString())

	attributes.SetName(state.Name.ValueString())

	req := datadogV2.NewCreateAppRequestWithDefaults()
	req.Data = datadogV2.NewCreateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

// TODO
func (r *appResource) buildUpdateAppRequestBody(ctx context.Context, state *appResourceModel) (*datadogV2.UpdateAppRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUpdateAppRequestDataAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	// attributes.SetHandle(state.Handle.ValueString())

	attributes.SetName(state.Name.ValueString())

	req := datadogV2.NewUpdateAppRequestWithDefaults()
	req.Data = datadogV2.NewUpdateAppRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
