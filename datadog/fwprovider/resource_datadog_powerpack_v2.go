package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/dashboardmapping"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &powerpackV2Resource{}
	_ resource.ResourceWithImportState = &powerpackV2Resource{}
)

// NewPowerpackV2Resource returns a new framework resource for datadog_powerpack_v2.
func NewPowerpackV2Resource() resource.Resource {
	return &powerpackV2Resource{}
}

type powerpackV2Resource struct {
	ApiInstances *utils.ApiInstances
	Auth         context.Context
}

type powerpackV2ResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	LiveSpan          types.String `tfsdk:"live_span"`
	ShowTitle         types.Bool   `tfsdk:"show_title"`
	Tags              types.Set    `tfsdk:"tags"`
	TemplateVariables types.List   `tfsdk:"template_variables"`
	Widget            types.List   `tfsdk:"widget"`
	Layout            types.List   `tfsdk:"layout"`
}

func (r *powerpackV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	providerData, ok := req.ProviderData.(*FrameworkProvider)
	if !ok {
		return
	}
	r.ApiInstances = providerData.DatadogApiInstances
	r.Auth = providerData.Auth
}

func (r *powerpackV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), req, resp)
}

func (r *powerpackV2Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "powerpack_v2"
}

func (r *powerpackV2Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	widgetAttrs, widgetBlocks := dashboardmapping.AllWidgetFWBlocks(true) // excludePowerpackOnly=true

	resp.Schema = schema.Schema{
		Description: "Provides a Datadog powerpack resource (v2, FieldSpec engine). This can be used to create and manage Datadog powerpacks.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name for the powerpack.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the powerpack.",
			},
			"live_span": schema.StringAttribute{
				Optional:    true,
				Description: "The timeframe to use when displaying the powerpack.",
			},
			"show_title": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether or not title should be displayed in the powerpack.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Description: "List of tags to identify this powerpack.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"template_variables": schema.ListNestedBlock{
				Description: "The list of template variables for this powerpack.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the powerpack template variable.",
						},
						"defaults": schema.ListAttribute{
							Optional:    true,
							Description: "One or many default values for powerpack template variables on load.",
							ElementType: types.StringType,
						},
					},
				},
			},
			"widget": schema.ListNestedBlock{
				Description: "The list of widgets to display in the powerpack.",
				NestedObject: schema.NestedBlockObject{
					Attributes: widgetAttrs,
					Blocks:     widgetBlocks,
				},
			},
			"layout": schema.ListNestedBlock{
				Description: "The layout of the powerpack on a free-form dashboard.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"x": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The position of the widget on the x (horizontal) axis.",
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"y": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The position of the widget on the y (vertical) axis.",
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"width": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The width of the widget.",
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"height": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The height of the widget.",
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *powerpackV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan powerpackV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	powerpackPayload, err := r.buildPowerpack(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building powerpack", err.Error())
		return
	}

	ppk, httpresp, apiErr := r.ApiInstances.GetPowerpackApiV2().CreatePowerpack(r.Auth, *powerpackPayload)
	if apiErr != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(apiErr, httpresp, ""), "error creating powerpack"))
		return
	}
	if apiErr = utils.CheckForUnparsed(ppk); apiErr != nil {
		resp.Diagnostics.AddError("Unparsed response", apiErr.Error())
		return
	}
	plan.ID = types.StringValue(*ppk.Data.Id)

	var getPowerpackResponse datadogV2.PowerpackResponse
	var httpResponse *http.Response
	retryErr := retryContextDashboard(ctx, func() error {
		var gerr error
		getPowerpackResponse, httpResponse, gerr = r.ApiInstances.GetPowerpackApiV2().GetPowerpack(r.Auth, *ppk.Data.Id)
		if gerr != nil {
			if httpResponse != nil {
				return fmt.Errorf("powerpack not created yet")
			}
			return fmt.Errorf("non-retryable: %w", gerr)
		}
		if gerr = utils.CheckForUnparsed(getPowerpackResponse); gerr != nil {
			return fmt.Errorf("non-retryable: %w", gerr)
		}
		return nil
	})
	if retryErr != nil {
		resp.Diagnostics.AddError("Error waiting for powerpack", retryErr.Error())
		return
	}

	stDiags := r.setStateFromResp(ctx, &getPowerpackResponse, &plan)
	resp.Diagnostics.Append(stDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *powerpackV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state powerpackV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppk, httpResponse, apiErr := r.ApiInstances.GetPowerpackApiV2().GetPowerpack(r.Auth, state.ID.ValueString())
	if apiErr != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(apiErr, httpResponse, ""), "error getting powerpack"))
		return
	}
	if apiErr = utils.CheckForUnparsed(ppk); apiErr != nil {
		resp.Diagnostics.AddError("Unparsed response", apiErr.Error())
		return
	}

	stDiags := r.setStateFromResp(ctx, &ppk, &state)
	resp.Diagnostics.Append(stDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *powerpackV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan powerpackV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	powerpackPayload, err := r.buildPowerpack(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building powerpack", err.Error())
		return
	}

	updated, httpResponse, apiErr := r.ApiInstances.GetPowerpackApiV2().UpdatePowerpack(r.Auth, plan.ID.ValueString(), *powerpackPayload)
	if apiErr != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(apiErr, httpResponse, ""), "error updating powerpack"))
		return
	}
	if apiErr = utils.CheckForUnparsed(updated); apiErr != nil {
		resp.Diagnostics.AddError("Unparsed response", apiErr.Error())
		return
	}

	stDiags := r.setStateFromResp(ctx, &updated, &plan)
	resp.Diagnostics.Append(stDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *powerpackV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state powerpackV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpresp, apiErr := r.ApiInstances.GetPowerpackApiV2().DeletePowerpack(r.Auth, state.ID.ValueString())
	if apiErr != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(apiErr, httpresp, ""), "error deleting powerpack"))
		return
	}
}

// buildPowerpack constructs a datadogV2.Powerpack from the Terraform model.
func (r *powerpackV2Resource) buildPowerpack(ctx context.Context, m *powerpackV2ResourceModel) (*datadogV2.Powerpack, error) {
	attributes := datadogV2.NewPowerpackAttributesWithDefaults()

	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		attributes.SetDescription(m.Description.ValueString())
	}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		attributes.SetName(m.Name.ValueString())
	}

	// Tags
	tags := make([]string, 0)
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		for _, elem := range m.Tags.Elements() {
			if sv, ok := elem.(types.String); ok {
				tags = append(tags, sv.ValueString())
			}
		}
	}
	attributes.SetTags(tags)

	// Template variables
	tvs := r.buildTemplateVariables(m.TemplateVariables)
	attributes.SetTemplateVariables(tvs)

	// Group widget definition
	var groupWidget datadogV2.PowerpackGroupWidget
	var definition datadogV2.PowerpackGroupWidgetDefinition
	definition.SetLayoutType("ordered")
	definition.SetType("group")

	if !m.ShowTitle.IsNull() && !m.ShowTitle.IsUnknown() {
		definition.SetShowTitle(m.ShowTitle.ValueBool())
	}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		definition.SetTitle(m.Name.ValueString())
	}

	// Layout
	if !m.Layout.IsNull() && !m.Layout.IsUnknown() && len(m.Layout.Elements()) > 0 {
		layoutObj, ok := m.Layout.Elements()[0].(types.Object)
		if ok {
			layoutAttrs := layoutObj.Attributes()
			var x, y, w, h int64
			if v, ok := layoutAttrs["x"].(types.Int64); ok && !v.IsNull() {
				x = v.ValueInt64()
			}
			if v, ok := layoutAttrs["y"].(types.Int64); ok && !v.IsNull() {
				y = v.ValueInt64()
			}
			if v, ok := layoutAttrs["width"].(types.Int64); ok && !v.IsNull() {
				w = v.ValueInt64()
			}
			if v, ok := layoutAttrs["height"].(types.Int64); ok && !v.IsNull() {
				h = v.ValueInt64()
			}
			layout := datadogV2.NewPowerpackGroupWidgetLayout(h, w, x, y)
			groupWidget.SetLayout(*layout)
		}
	}

	// Widgets
	widgetElems := m.Widget.Elements()
	powerpackWidgets := make([]datadogV2.PowerpackInnerWidgets, 0, len(widgetElems))
	for _, elem := range widgetElems {
		wObj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		widgetMap := dashboardmapping.BuildWidgetEngineJSON(wObj.Attributes())
		widgetJSON, err := json.Marshal(widgetMap)
		if err != nil {
			return nil, fmt.Errorf("error marshaling widget: %w", err)
		}
		var ppkWidget datadogV2.PowerpackInnerWidgets
		if err := ppkWidget.UnmarshalJSON(widgetJSON); err != nil {
			return nil, fmt.Errorf("error unmarshaling widget: %w", err)
		}
		ppkWidget.AdditionalProperties = nil
		powerpackWidgets = append(powerpackWidgets, ppkWidget)
	}
	definition.SetWidgets(powerpackWidgets)
	groupWidget.Definition = definition

	// Live span
	if !m.LiveSpan.IsNull() && !m.LiveSpan.IsUnknown() && m.LiveSpan.ValueString() != "" {
		liveSpan, err := datadogV2.NewWidgetLiveSpanFromValue(m.LiveSpan.ValueString())
		if err != nil {
			return nil, fmt.Errorf("invalid live_span: %s", m.LiveSpan.ValueString())
		}
		groupWidget.LiveSpan = liveSpan
	}

	attributes.GroupWidget = groupWidget

	req := datadogV2.NewPowerpackWithDefaults()
	req.Data = datadogV2.NewPowerpackDataWithDefaults()
	req.Data.SetType("powerpack")
	req.Data.SetAttributes(*attributes)
	return req, nil
}

// setStateFromResp populates the model from the powerpack API response.
func (r *powerpackV2Resource) setStateFromResp(ctx context.Context, ppk *datadogV2.PowerpackResponse, m *powerpackV2ResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if ppk.Data == nil {
		diags.AddError("Empty powerpack response", "data field is nil")
		return diags
	}

	attrs := ppk.Data.Attributes
	m.Description = types.StringValue(attrs.GetDescription())
	m.Name = types.StringValue(attrs.GetName())

	// Tags
	tagsSet, d := types.SetValueFrom(ctx, types.StringType, attrs.GetTags())
	diags.Append(d...)
	m.Tags = tagsSet

	// Live span
	m.LiveSpan = types.StringValue(string(attrs.GroupWidget.GetLiveSpan()))

	// Show title
	m.ShowTitle = types.BoolValue(attrs.GroupWidget.Definition.GetShowTitle())

	// Template variables
	tvList := r.flattenTemplateVariables(ctx, attrs.GetTemplateVariables())
	m.TemplateVariables = tvList

	// Layout
	if v, ok := attrs.GroupWidget.GetLayoutOk(); ok && v != nil {
		layoutAttrTypes := map[string]attr.Type{
			"x": types.Int64Type, "y": types.Int64Type,
			"width": types.Int64Type, "height": types.Int64Type,
		}
		layoutVals := map[string]attr.Value{
			"x":      types.Int64Value((*v).GetX()),
			"y":      types.Int64Value((*v).GetY()),
			"width":  types.Int64Value((*v).GetWidth()),
			"height": types.Int64Value((*v).GetHeight()),
		}
		layoutObj, d := types.ObjectValue(layoutAttrTypes, layoutVals)
		diags.Append(d...)
		layoutObjType := types.ObjectType{AttrTypes: layoutAttrTypes}
		layoutList, d := types.ListValue(layoutObjType, []attr.Value{layoutObj})
		diags.Append(d...)
		m.Layout = layoutList
	}

	// Widgets
	rawWidgets := attrs.GroupWidget.Definition.Widgets
	widgetList := r.flattenPowerpackWidgets(rawWidgets)
	m.Widget = widgetList

	return diags
}

func (r *powerpackV2Resource) buildTemplateVariables(tvList types.List) []datadogV2.PowerpackTemplateVariable {
	if tvList.IsNull() || tvList.IsUnknown() {
		return []datadogV2.PowerpackTemplateVariable{}
	}
	result := make([]datadogV2.PowerpackTemplateVariable, 0)
	for _, elem := range tvList.Elements() {
		tvObj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		tvAttrs := tvObj.Attributes()
		ppkTV := datadogV2.NewPowerpackTemplateVariable("")
		if v, ok := tvAttrs["name"].(types.String); ok && !v.IsNull() {
			ppkTV.SetName(v.ValueString())
		}
		if defaultsList, ok := tvAttrs["defaults"].(types.List); ok && !defaultsList.IsNull() {
			defaults := make([]string, 0)
			for _, d := range defaultsList.Elements() {
				if sv, ok := d.(types.String); ok {
					defaults = append(defaults, sv.ValueString())
				}
			}
			if len(defaults) > 0 {
				ppkTV.SetDefaults(defaults)
			}
		}
		result = append(result, *ppkTV)
	}
	return result
}

func (r *powerpackV2Resource) flattenTemplateVariables(ctx context.Context, ppkTVs []datadogV2.PowerpackTemplateVariable) types.List {
	tvAttrTypes := map[string]attr.Type{
		"name":     types.StringType,
		"defaults": types.ListType{ElemType: types.StringType},
	}
	objType := types.ObjectType{AttrTypes: tvAttrTypes}

	elems := make([]attr.Value, 0, len(ppkTVs))
	for _, tv := range ppkTVs {
		defaults := tv.GetDefaults()
		defaultVals := make([]attr.Value, len(defaults))
		for i, d := range defaults {
			defaultVals[i] = types.StringValue(d)
		}
		defaultList, _ := types.ListValue(types.StringType, defaultVals)
		tvVals := map[string]attr.Value{
			"name":     types.StringValue(tv.GetName()),
			"defaults": defaultList,
		}
		tvObj, _ := types.ObjectValue(tvAttrTypes, tvVals)
		elems = append(elems, tvObj)
	}

	list, _ := types.ListValue(objType, elems)
	return list
}

func (r *powerpackV2Resource) flattenPowerpackWidgets(ppkWidgets []datadogV2.PowerpackInnerWidgets) types.List {
	widgetAttrTypes := dashboardmapping.AllWidgetAttrTypes(true) // excludePowerpackOnly=true
	widgetObjType := types.ObjectType{AttrTypes: widgetAttrTypes}

	elems := make([]attr.Value, 0, len(ppkWidgets))
	for _, ppkWidget := range ppkWidgets {
		widgetJSON, err := ppkWidget.MarshalJSON()
		if err != nil {
			continue
		}
		var widgetData map[string]interface{}
		if err := json.Unmarshal(widgetJSON, &widgetData); err != nil {
			continue
		}

		flattened := dashboardmapping.FlattenWidgetEngineJSON(widgetData)
		if flattened == nil {
			flattened = map[string]interface{}{}
		}

		// Flatten widget_layout from JSON "layout"
		if layout, ok := widgetData["layout"].(map[string]interface{}); ok {
			layoutState := map[string]interface{}{}
			for _, key := range []string{"x", "y", "width", "height"} {
				if v, ok := layout[key]; ok {
					switch iv := v.(type) {
					case float64:
						layoutState[key] = int(iv)
					case int:
						layoutState[key] = iv
					}
				}
			}
			if len(layoutState) > 0 {
				flattened["widget_layout"] = []interface{}{layoutState}
			}
		}

		attrVals := dashboardmapping.MapToAttrValues(flattened, widgetAttrTypes)
		widgetObj, objDiags := types.ObjectValue(widgetAttrTypes, attrVals)
		if objDiags.HasError() {
			continue
		}
		elems = append(elems, widgetObj)
	}

	list, _ := types.ListValue(widgetObjType, elems)
	return list
}
