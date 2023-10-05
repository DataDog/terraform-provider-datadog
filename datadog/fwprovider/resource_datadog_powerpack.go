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
	_ resource.ResourceWithConfigure   = &powerpackResource{}
	_ resource.ResourceWithImportState = &powerpackResource{}
)

type powerpackResource struct {
	Api  *datadogV2.PowerpackApi
	Auth context.Context
}

type powerpackModel struct {
	ID                types.String              `tfsdk:"id"`
	Description       types.String              `tfsdk:"description"`
	Name              types.String              `tfsdk:"name"`
	Tags              types.List                `tfsdk:"tags"`
	TemplateVariables []*templateVariablesModel `tfsdk:"template_variables"`
	GroupWidget       *groupWidgetModel         `tfsdk:"group_widget"`
}

type templateVariablesModel struct {
	Name     types.String `tfsdk:"name"`
	Defaults types.List   `tfsdk:"defaults"`
}

type groupWidgetModel struct {
	Definition *definitionModel `tfsdk:"definition"`
	Layout     *layoutModel     `tfsdk:"layout"`
}
type definitionModel struct {
	LayoutType types.String    `tfsdk:"layout_type"`
	ShowTitle  types.Bool      `tfsdk:"show_title"`
	Title      types.String    `tfsdk:"title"`
	Type       types.String    `tfsdk:"type"`
	Widgets    []*widgetsModel `tfsdk:"widgets"`
}
type widgetsModel struct {
	Definition *definitionModel `tfsdk:"definition"`
	Layout     *layoutModel     `tfsdk:"layout"`
}
type definitionModel struct {
}
type layoutModel struct {
	Height types.Int64 `tfsdk:"height"`
	Width  types.Int64 `tfsdk:"width"`
	X      types.Int64 `tfsdk:"x"`
	Y      types.Int64 `tfsdk:"y"`
}

type layoutModel struct {
	Height types.Int64 `tfsdk:"height"`
	Width  types.Int64 `tfsdk:"width"`
	X      types.Int64 `tfsdk:"x"`
	Y      types.Int64 `tfsdk:"y"`
}

func NewPowerpackResource() resource.Resource {
	return &powerpackResource{}
}

func (r *powerpackResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetPowerpackApiV2()
	r.Auth = providerData.Auth
}

func (r *powerpackResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "powerpack"
}

func (r *powerpackResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Powerpack resource. This can be used to create and manage Datadog powerpack.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of this powerpack.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the powerpack.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Description: "List of tags to identify this powerpack.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"template_variables": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:    true,
							Description: "The name of the variable.",
						},
						"defaults": schema.ListAttribute{
							Optional:    true,
							Description: "One or many template variable default values within the saved view, which are unioned together using `OR` if more than one is specified.",
							ElementType: types.StringType,
						},
					},
				},
			},
			"group_widget": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"definition": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"layout_type": schema.StringAttribute{
								Optional:    true,
								Description: "Layout type of widgets.",
							},
							"show_title": schema.BoolAttribute{
								Optional:    true,
								Description: "Boolean indicating whether powerpack group title should be visible or not.",
							},
							"title": schema.StringAttribute{
								Optional:    true,
								Description: "Name for the group widget.",
							},
							"type": schema.StringAttribute{
								Optional:    true,
								Description: "Type of widget, must be group.",
							},
						},
						Blocks: map[string]schema.Block{
							"widgets": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{},
									Blocks: map[string]schema.Block{
										"definition": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{},
										},
										"layout": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"height": schema.Int64Attribute{
													Optional:    true,
													Description: "The height of the widget. Should be a non-negative integer.",
												},
												"width": schema.Int64Attribute{
													Optional:    true,
													Description: "The width of the widget. Should be a non-negative integer.",
												},
												"x": schema.Int64Attribute{
													Optional:    true,
													Description: "The position of the widget on the x (horizontal) axis. Should be a non-negative integer.",
												},
												"y": schema.Int64Attribute{
													Optional:    true,
													Description: "The position of the widget on the y (vertical) axis. Should be a non-negative integer.",
												},
											},
										},
									},
								},
							},
						},
					},
					"layout": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"height": schema.Int64Attribute{
								Optional:    true,
								Description: "The height of the widget. Should be a non-negative integer.",
							},
							"width": schema.Int64Attribute{
								Optional:    true,
								Description: "The width of the widget. Should be a non-negative integer.",
							},
							"x": schema.Int64Attribute{
								Optional:    true,
								Description: "The position of the widget on the x (horizontal) axis. Should be a non-negative integer.",
							},
							"y": schema.Int64Attribute{
								Optional:    true,
								Description: "The position of the widget on the y (vertical) axis. Should be a non-negative integer.",
							},
						},
					},
				},
			},
		},
	}
}

func (r *powerpackResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *powerpackResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state powerpackModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetPowerpack(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Powerpack"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *powerpackResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state powerpackModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildPowerpackRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreatePowerpack(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Powerpack"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *powerpackResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state powerpackModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildPowerpackRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdatePowerpack(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Powerpack"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *powerpackResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state powerpackModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeletePowerpack(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting powerpack"))
		return
	}
}

func (r *powerpackResource) updateState(ctx context.Context, state *powerpackModel, resp *datadogV2.PowerpackResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if description, ok := attributes.GetDescriptionOk(); ok {
		state.Description = types.StringValue(*description)
	}

	state.Name = types.StringValue(attributes.GetName())

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		state.Tags, _ = types.ListValueFrom(ctx, types.StringType, *tags)
	}

	if templateVariables, ok := attributes.GetTemplateVariablesOk(); ok && len(*templateVariables) > 0 {
		state.TemplateVariables = []*templateVariablesModel{}
		for _, templateVariablesDd := range *templateVariables {
			templateVariablesTfItem := templateVariablesModel{}
			if defaults, ok := templateVariablesDd.GetDefaultsOk(); ok && len(*defaults) > 0 {
				templateVariablesTfItem.Defaults, _ = types.ListValueFrom(ctx, types.StringType, *defaults)
			}
			if name, ok := templateVariablesDd.GetNameOk(); ok {
				templateVariablesTfItem.Name = types.StringValue(*name)
			}

			state.TemplateVariables = append(state.TemplateVariables, &templateVariablesTfItem)
		}
	}

	groupWidgetTf := groupWidgetModel{}

	state.GroupWidget = &groupWidgetTf
}

func (r *powerpackResource) buildPowerpackRequestBody(ctx context.Context, state *powerpackModel) (*datadogV2.Powerpack, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewPowerpackAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	attributes.SetName(state.Name.ValueString())

	if !state.Tags.IsNull() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	if state.TemplateVariables != nil {
		var templateVariables []datadogV2.PowerpackTemplateVariable
		for _, templateVariablesTFItem := range state.TemplateVariables {
			templateVariablesDDItem := datadogV2.NewPowerpackTemplateVariable()

			templateVariablesDDItem.SetName(templateVariablesTFItem.Name.ValueString())

			if !templateVariablesTFItem.Defaults.IsNull() {
				var defaults []string
				diags.Append(templateVariablesTFItem.Defaults.ElementsAs(ctx, &defaults, false)...)
				templateVariablesDDItem.SetDefaults(defaults)
			}
		}
		attributes.SetTemplateVariables(templateVariables)
	}

	var groupWidget datadogV2.PowerpackGroupWidget

	var definition datadogV2.PowerpackGroupWidgetDefinition

	definition.SetLayoutType(state.GroupWidget.Definition.LayoutType.ValueString())
	if !state.GroupWidget.Definition.ShowTitle.IsNull() {
		definition.SetShowTitle(state.GroupWidget.Definition.ShowTitle.ValueBool())
	}
	if !state.GroupWidget.Definition.Title.IsNull() {
		definition.SetTitle(state.GroupWidget.Definition.Title.ValueString())
	}
	definition.SetType(state.GroupWidget.Definition.Type.ValueString())

	if state.GroupWidget.Definition.Widgets != nil {
		var widgets []datadogV2.PowerpackInnerWidgets
		for _, widgetsTFItem := range state.GroupWidget.Definition.Widgets {
			widgetsDDItem := datadogV2.NewPowerpackInnerWidgets()

			var definition map[string]interface{}

			widgetsDDItem.Definition = definition

			if widgetsTFItem.Layout != nil {
				var layout datadogV2.PowerpackInnerWidgetLayout

				layout.SetHeight(widgetsTFItem.Layout.Height.ValueInt64())
				layout.SetWidth(widgetsTFItem.Layout.Width.ValueInt64())
				layout.SetX(widgetsTFItem.Layout.X.ValueInt64())
				layout.SetY(widgetsTFItem.Layout.Y.ValueInt64())

				widgetsDDItem.Layout = &layout
			}
		}
		definition.SetWidgets(widgets)
	}

	groupWidget.Definition = definition

	if state.GroupWidget.Layout != nil {
		var layout datadogV2.PowerpackGroupWidgetLayout

		layout.SetHeight(state.GroupWidget.Layout.Height.ValueInt64())
		layout.SetWidth(state.GroupWidget.Layout.Width.ValueInt64())
		layout.SetX(state.GroupWidget.Layout.X.ValueInt64())
		layout.SetY(state.GroupWidget.Layout.Y.ValueInt64())

		groupWidget.Layout = &layout
	}

	attributes.GroupWidget = groupWidget

	req := datadogV2.NewPowerpackWithDefaults()
	req.Data = datadogV2.NewPowerpackDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
