package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/dashboardmapping"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogPowerpackV2() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog powerpack resource.",
		CreateContext: resourceDatadogPowerpackV2Create,
		ReadContext:   resourceDatadogPowerpackV2Read,
		UpdateContext: resourceDatadogPowerpackV2Update,
		DeleteContext: resourceDatadogPowerpackV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: buildPowerpackV2Schema,
	}
}

func buildPowerpackV2Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name for the powerpack.",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description of the powerpack.",
		},
		"live_span": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The timeframe to use when displaying the powerpack.",
		},
		"show_title": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Whether or not title should be displayed in the powerpack.",
		},
		"tags": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "List of tags to identify this powerpack.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"template_variables": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of template variables for this powerpack.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The name of the powerpack template variable.",
					},
					"defaults": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "One or many default values for powerpack template variables on load.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		"layout": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Description: "The layout of the powerpack on a free-form dashboard.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"x": {
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "The position of the widget on the x (horizontal) axis.",
					},
					"y": {
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "The position of the widget on the y (vertical) axis.",
					},
					"width": {
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "The width of the widget.",
					},
					"height": {
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "The height of the widget.",
					},
				},
			},
		},
		"widget": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of widgets to display in the powerpack.",
			Elem: &schema.Resource{
				Schema: dashboardmapping.AllWidgetSDKv2Schema(true), // excludePowerpackOnly=true
			},
		},
	}
	return s
}

func resourceDatadogPowerpackV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ppk, err := buildPowerpackPayload(d)
	if err != nil {
		return diag.Errorf("error building powerpack: %s", err)
	}

	created, httpResp, apiErr := apiInstances.GetPowerpackApiV2().CreatePowerpack(auth, *ppk)
	if apiErr != nil {
		return diag.Errorf("error creating powerpack: %s", utils.TranslateClientError(apiErr, httpResp, ""))
	}
	if apiErr = utils.CheckForUnparsed(created); apiErr != nil {
		return diag.Errorf("unparsed create response: %s", apiErr)
	}

	d.SetId(*created.Data.Id)

	// Retry-read so state reflects what the API confirmed.
	var getPPK datadogV2.PowerpackResponse
	var getResp *http.Response
	retryErr := retryPowerpack(ctx, func() error {
		var err error
		getPPK, getResp, err = apiInstances.GetPowerpackApiV2().GetPowerpack(auth, *created.Data.Id)
		if err != nil {
			if getResp != nil {
				return fmt.Errorf("powerpack not ready yet")
			}
			return fmt.Errorf("non-retryable: %w", err)
		}
		if err = utils.CheckForUnparsed(getPPK); err != nil {
			return fmt.Errorf("non-retryable: %w", err)
		}
		return nil
	})
	if retryErr != nil {
		return diag.Errorf("error waiting for powerpack: %s", retryErr)
	}

	return setPowerpackState(d, &getPPK)
}

func resourceDatadogPowerpackV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ppk, httpResp, apiErr := apiInstances.GetPowerpackApiV2().GetPowerpack(auth, d.Id())
	if apiErr != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting powerpack: %s", utils.TranslateClientError(apiErr, httpResp, ""))
	}
	if apiErr = utils.CheckForUnparsed(ppk); apiErr != nil {
		return diag.Errorf("unparsed get response: %s", apiErr)
	}

	return setPowerpackState(d, &ppk)
}

func resourceDatadogPowerpackV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ppk, err := buildPowerpackPayload(d)
	if err != nil {
		return diag.Errorf("error building powerpack: %s", err)
	}

	updated, httpResp, apiErr := apiInstances.GetPowerpackApiV2().UpdatePowerpack(auth, d.Id(), *ppk)
	if apiErr != nil {
		return diag.Errorf("error updating powerpack: %s", utils.TranslateClientError(apiErr, httpResp, ""))
	}
	if apiErr = utils.CheckForUnparsed(updated); apiErr != nil {
		return diag.Errorf("unparsed update response: %s", apiErr)
	}

	return setPowerpackState(d, &updated)
}

func resourceDatadogPowerpackV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	httpResp, apiErr := apiInstances.GetPowerpackApiV2().DeletePowerpack(auth, d.Id())
	if apiErr != nil {
		return diag.Errorf("error deleting powerpack: %s", utils.TranslateClientError(apiErr, httpResp, ""))
	}
	return nil
}

// buildPowerpackPayload constructs a datadogV2.Powerpack from ResourceData.
func buildPowerpackPayload(d *schema.ResourceData) (*datadogV2.Powerpack, error) {
	attrs := datadogV2.NewPowerpackAttributesWithDefaults()

	if v, ok := d.GetOk("name"); ok {
		attrs.SetName(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		attrs.SetDescription(v.(string))
	}

	// Tags
	tags := make([]string, 0)
	if v, ok := d.GetOk("tags"); ok {
		for _, t := range v.(*schema.Set).List() {
			tags = append(tags, t.(string))
		}
	}
	attrs.SetTags(tags)

	// Template variables
	attrs.SetTemplateVariables(buildPowerpackV2TemplateVariables(d))

	// Group widget
	var groupWidget datadogV2.PowerpackGroupWidget
	var definition datadogV2.PowerpackGroupWidgetDefinition
	definition.SetLayoutType("ordered")
	definition.SetType("group")

	if v, ok := d.GetOk("show_title"); ok {
		definition.SetShowTitle(v.(bool))
	}
	if v, ok := d.GetOk("name"); ok {
		definition.SetTitle(v.(string))
	}

	// Layout
	if v, ok := d.GetOk("layout"); ok {
		if layouts := v.([]interface{}); len(layouts) > 0 {
			lm := layouts[0].(map[string]interface{})
			layout := datadogV2.NewPowerpackGroupWidgetLayout(
				int64(lm["height"].(int)),
				int64(lm["width"].(int)),
				int64(lm["x"].(int)),
				int64(lm["y"].(int)),
			)
			groupWidget.SetLayout(*layout)
		}
	}

	// Widgets
	powerpackWidgets, err := buildPowerpackWidgets(d)
	if err != nil {
		return nil, err
	}
	definition.SetWidgets(powerpackWidgets)
	groupWidget.Definition = definition

	// Live span
	if v, ok := d.GetOk("live_span"); ok && v.(string) != "" {
		liveSpan, err := datadogV2.NewWidgetLiveSpanFromValue(v.(string))
		if err != nil {
			return nil, fmt.Errorf("invalid live_span: %s", v.(string))
		}
		groupWidget.LiveSpan = liveSpan
	}

	attrs.GroupWidget = groupWidget

	req := datadogV2.NewPowerpackWithDefaults()
	req.Data = datadogV2.NewPowerpackDataWithDefaults()
	req.Data.SetType("powerpack")
	req.Data.SetAttributes(*attrs)
	return req, nil
}

func buildPowerpackV2TemplateVariables(d *schema.ResourceData) []datadogV2.PowerpackTemplateVariable {
	result := make([]datadogV2.PowerpackTemplateVariable, 0)
	tvList, ok := d.GetOk("template_variables")
	if !ok {
		return result
	}
	for _, elem := range tvList.([]interface{}) {
		tv := elem.(map[string]interface{})
		ppkTV := datadogV2.NewPowerpackTemplateVariable(tv["name"].(string))
		if defaults, ok := tv["defaults"].([]interface{}); ok && len(defaults) > 0 {
			strs := make([]string, len(defaults))
			for i, d := range defaults {
				strs[i] = d.(string)
			}
			ppkTV.SetDefaults(strs)
		}
		result = append(result, *ppkTV)
	}
	return result
}

func buildPowerpackWidgets(d *schema.ResourceData) ([]datadogV2.PowerpackInnerWidgets, error) {
	widgetList, ok := d.GetOk("widget")
	if !ok {
		return []datadogV2.PowerpackInnerWidgets{}, nil
	}
	rawWidgets := widgetList.([]interface{})
	result := make([]datadogV2.PowerpackInnerWidgets, 0, len(rawWidgets))
	for _, w := range rawWidgets {
		wMap, ok := w.(map[string]interface{})
		if !ok {
			continue
		}
		widgetJSON, err := json.Marshal(dashboardmapping.BuildWidgetEngineJSONFromMap(wMap))
		if err != nil {
			return nil, fmt.Errorf("error marshaling widget: %w", err)
		}
		var ppkWidget datadogV2.PowerpackInnerWidgets
		if err := ppkWidget.UnmarshalJSON(widgetJSON); err != nil {
			return nil, fmt.Errorf("error unmarshaling widget: %w", err)
		}
		ppkWidget.AdditionalProperties = nil
		result = append(result, ppkWidget)
	}
	return result, nil
}

// setPowerpackState sets ResourceData fields from a PowerpackResponse.
func setPowerpackState(d *schema.ResourceData, ppk *datadogV2.PowerpackResponse) diag.Diagnostics {
	var diags diag.Diagnostics
	if ppk.Data == nil {
		return diag.Errorf("empty powerpack response")
	}

	attrs := ppk.Data.Attributes

	if err := d.Set("name", attrs.GetName()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", attrs.GetDescription()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("tags", attrs.GetTags()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("show_title", attrs.GroupWidget.Definition.GetShowTitle()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("live_span", string(attrs.GroupWidget.GetLiveSpan())); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Template variables
	tvs := flattenPowerpackV2TemplateVariables(attrs.GetTemplateVariables())
	if err := d.Set("template_variables", tvs); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Layout
	if v, ok := attrs.GroupWidget.GetLayoutOk(); ok && v != nil {
		layout := []map[string]interface{}{{
			"x":      int((*v).GetX()),
			"y":      int((*v).GetY()),
			"width":  int((*v).GetWidth()),
			"height": int((*v).GetHeight()),
		}}
		if err := d.Set("layout", layout); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// Widgets — unmarshal each PowerpackInnerWidget to map, then flatten via engine
	rawWidgets := attrs.GroupWidget.Definition.Widgets
	flatWidgets := make([]interface{}, 0, len(rawWidgets))
	for _, ppkWidget := range rawWidgets {
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
		flatWidgets = append(flatWidgets, flattened)
	}
	if err := d.Set("widget", flatWidgets); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func flattenPowerpackV2TemplateVariables(tvs []datadogV2.PowerpackTemplateVariable) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(tvs))
	for _, tv := range tvs {
		m := map[string]interface{}{
			"name":     tv.GetName(),
			"defaults": tv.GetDefaults(),
		}
		result = append(result, m)
	}
	return result
}

// retryPowerpack retries fn up to 3 times, stopping on context cancellation,
// deadline, or a "non-retryable:" error prefix.
func retryPowerpack(ctx context.Context, fn func() error) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if len(err.Error()) > 14 && err.Error()[:14] == "non-retryable:" {
			return err
		}
		lastErr = err
	}
	return lastErr
}
