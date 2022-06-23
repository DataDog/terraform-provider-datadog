package datadog

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var timeFormat = "2006-01-02T15:04:05Z07:00"

func getMarkdownCellSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"definition": {
			Description: "Text in a notebook is formatted with [Markdown](https://daringfireball.net/projects/markdown/), which enables the use of headings, subheadings, links, images, lists, and code blocks.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Description:      "Type of the markdown cell.",
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookMarkdownCellDefinitionTypeFromValue),
						Default:          datadogV1.NOTEBOOKMARKDOWNCELLDEFINITIONTYPE_MARKDOWN,
					},
					"text": {
						Description: "The markdown content.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
	}
}

func getWidgetBasedNotebookCellSchema(definitionSchema map[string]*schema.Schema) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"definition": {
			Description: "The definition for a Timeseries widget.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem:        &schema.Resource{Schema: definitionSchema},
		},
		"split_by": {
			Description: "Object describing how to split the graph to display multiple visualizations per request.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"keys": {
						Type:        schema.TypeList,
						Required:    true,
						Description: "Keys to split on.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"tags": {
						Type:        schema.TypeSet,
						Required:    true,
						Description: "Tags to split on.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		"graph_size": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The size of the graph.",
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookGraphSizeFromValue),
		},
		"time": {
			Description: "Timeframe for the notebook cell. When 'null', the notebook global time is used.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: getTimeSchema(),
			},
		},
	}

}

func getNotebookCellAttributesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"markdown_cell": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The attributes of a notebook `markdown` cell.",
			Elem: &schema.Resource{
				Schema: getMarkdownCellSchema(),
			},
		},
		"timeseries_cell": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The attributes of a notebook `timeseries` cell.",
			Elem: &schema.Resource{
				Schema: getWidgetBasedNotebookCellSchema(getTimeseriesDefinitionSchema()),
			},
		},
		"toplist_cell": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The attributes of a notebook `toplist` cell.",
			Elem: &schema.Resource{
				Schema: getWidgetBasedNotebookCellSchema(getToplistDefinitionSchema()),
			},
		},
		"heatmap_cell": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The attributes of a notebook `heatmap` cell.",
			Elem: &schema.Resource{
				Schema: getWidgetBasedNotebookCellSchema(getHeatmapDefinitionSchema()),
			},
		},
		"distribution_cell": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The attributes of a notebook `distribution` cell.",
			Elem: &schema.Resource{
				Schema: getWidgetBasedNotebookCellSchema(getHeatmapDefinitionSchema()),
			},
		},
		"log_stream_cell": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The attributes of a notebook `logstream` cell.",
			Elem: &schema.Resource{
				Schema: getWidgetBasedNotebookCellSchema(getLogStreamDefinitionSchema()),
			},
		},
	}
}

func getTimeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"notebook_relative_time": {
			Description: "Relative timeframe.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"live_span": getWidgetLiveSpanSchema(),
				},
			},
		},
		"notebook_absolute_time": {
			Description: "Relative timeframe.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"start": {
						Type:     schema.TypeString,
						Required: true,
						DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
							oldTime, _ := time.Parse(timeFormat, old)
							newTime, _ := time.Parse(timeFormat, new)

							return oldTime.Equal(newTime)
						},
						Description: "The start time.",
					},
					"end": {
						Type:     schema.TypeString,
						Required: true,
						DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
							oldTime, _ := time.Parse(timeFormat, old)
							newTime, _ := time.Parse(timeFormat, new)

							return oldTime.Equal(newTime)
						},
						Description: "The end time.",
					},
					"live": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Indicates whether the timeframe should be shifted to end at the current time.",
					},
				},
			},
		},
	}
}

func getMetadataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"is_template": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Indicates whether the timeframe should be shifted to end at the current time.",
		},
		"take_snapshots": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether or not the notebook takes snapshot image backups of the notebook's fixed-time graphs.",
		},
		"type": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookMetadataTypeFromValue),
			Description:      "Metadata type of the notebook.",
		},
	}
}

func resourceDatadogNotebook() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource for interacting with the logs_metric API",
		CreateContext: resourceDatadogNotebookCreate,
		ReadContext:   resourceDatadogNotebookRead,
		UpdateContext: resourceDatadogNotebookUpdate,
		DeleteContext: resourceDatadogNotebookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"cells": {
				Description: "List of cells to display in the notebook.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Notebook cell ID.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"attributes": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The attributes of a notebook cell in create cell request.",
							Elem: &schema.Resource{
								Schema: getNotebookCellAttributesSchema(),
							},
						},
						"type": {
							Type:             schema.TypeString,
							Optional:         true,
							Description:      "Type of the Notebook Cell resource.",
							ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookCellResourceTypeFromValue),
							Default:          datadogV1.NOTEBOOKCELLRESOURCETYPE_NOTEBOOK_CELLS,
						},
					},
				},
			},
			"time": {
				Description: "Notebook global timeframe.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: getTimeSchema(),
				},
			},
			"status": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookStatusFromValue),
				Description:      "Publication status of the notebook.",
				Default:          datadogV1.NOTEBOOKSTATUS_PUBLISHED,
			},
			"metadata": {
				Type:        schema.TypeList,
				Description: "Metadata associated with the notebook.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: getMetadataSchema(),
				},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the notebook.",
			},
			"type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewNotebookResourceTypeFromValue),
				Description:      "Publication status of the notebook.",
				Default:          datadogV1.NOTEBOOKRESOURCETYPE_NOTEBOOKS,
			},
		},
	}
}

func updateNotebookState(d *schema.ResourceData, resource *datadogV1.NotebookResponseData) diag.Diagnostics {
	cells := buildTerraformCells(&resource.Attributes.Cells)
	if err := d.Set("cells", cells); err != nil {
		return diag.FromErr(err)
	}

	_time := buildTerraformTime(&resource.Attributes.Time)
	if err := d.Set("time", _time); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("status", resource.Attributes.GetStatus()); err != nil {
		return diag.FromErr(err)
	}

	metadata := buildTerraformMetadata(resource.Attributes.Metadata)
	if err := d.Set("metadata", metadata); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", resource.Attributes.GetName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogNotebookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	resultNotebookCreateData, err := buildCreateDatadogNotebook(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building Notebook object: %w", err))
	}

	ddObject := datadogV1.NewNotebookCreateRequest(*resultNotebookCreateData)

	response, httpResponse, err := datadogClient.NotebooksApi.CreateNotebook(auth, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating Notebook")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}
	id := strconv.FormatInt(response.GetData().Id, 10)
	d.SetId(id)

	return updateNotebookState(d, response.Data)
}

func resourceDatadogNotebookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	response, httpresp, err := datadogClient.NotebooksApi.GetNotebook(auth, id)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting dashboard")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateNotebookState(d, response.Data)
}

func resourceDatadogNotebookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}
	resultNotebookUpdateData, err := buildDatadogNotebookUpdate(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error building Notebook object: %w", err))
	}

	ddObject := datadogV1.NewNotebookUpdateRequest(*resultNotebookUpdateData)

	response, httpResponse, err := datadogClient.NotebooksApi.UpdateNotebook(auth, id, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating Notebook")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateNotebookState(d, response.Data)
}

func resourceDatadogNotebookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	httpResponse, err := datadogClient.NotebooksApi.DeleteNotebook(auth, id)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting Notebook")
	}

	return nil
}

func buildDatadogUpdateNotebookCellAttributes(tfAttributes map[string]interface{}) (*datadogV1.NotebookCellUpdateRequestAttributes, error) {
	var definition datadogV1.NotebookCellUpdateRequestAttributes

	if markdown, ok := tfAttributes["markdown_cell"].([]interface{}); ok && len(markdown) > 0 {
		if tfMarkdownCellAttributes, ok := markdown[0].(map[string]interface{}); ok {
			ddMarkdownCellAttribute := datadogV1.NotebookMarkdownCellAttributesAsNotebookCellUpdateRequestAttributes(buildDatadogMarkdownCellDefinition(tfMarkdownCellAttributes))
			definition = ddMarkdownCellAttribute
		}
	} else if timeseries, ok := tfAttributes["timeseries_cell"].([]interface{}); ok && len(timeseries) > 0 {
		if tfTimeseriesCellAttributes, ok := timeseries[0].(map[string]interface{}); ok {
			ddTimeseriesCellAttribute := datadogV1.NotebookTimeseriesCellAttributesAsNotebookCellUpdateRequestAttributes(buildDatadogTimeseriesCellDefinition(tfTimeseriesCellAttributes))
			definition = ddTimeseriesCellAttribute
		}
	} else if toplist, ok := tfAttributes["toplist_cell"].([]interface{}); ok && len(toplist) > 0 {
		if tfTimeseriesCellAttributes, ok := toplist[0].(map[string]interface{}); ok {
			ddToplistCellAttribute := datadogV1.NotebookToplistCellAttributesAsNotebookCellUpdateRequestAttributes(buildDatadogToplistCellDefinition(tfTimeseriesCellAttributes))
			definition = ddToplistCellAttribute
		}
	} else if heatmap, ok := tfAttributes["heatmap_cell"].([]interface{}); ok && len(heatmap) > 0 {
		if tfHeatmapCellAttributes, ok := heatmap[0].(map[string]interface{}); ok {
			ddHeatmapCellAttribute := datadogV1.NotebookHeatMapCellAttributesAsNotebookCellUpdateRequestAttributes(buildDatadogHeatmapCellDefinition(tfHeatmapCellAttributes))
			definition = ddHeatmapCellAttribute
		}
	} else if distribution, ok := tfAttributes["distribution_cell"].([]interface{}); ok && len(distribution) > 0 {
		if tfDistributionCellAttributes, ok := distribution[0].(map[string]interface{}); ok {
			ddDistributionCellAttribute := datadogV1.NotebookDistributionCellAttributesAsNotebookCellUpdateRequestAttributes(buildDatadogDistributionCellDefinition(tfDistributionCellAttributes))
			definition = ddDistributionCellAttribute
		}
	} else if logstream, ok := tfAttributes["log_stream_cell"].([]interface{}); ok && len(logstream) > 0 {
		if tfLogStreamCellAttributes, ok := logstream[0].(map[string]interface{}); ok {
			ddLogStreamCellAttribute := datadogV1.NotebookLogStreamCellAttributesAsNotebookCellUpdateRequestAttributes(buildDatadogLogStreamCellDefinition(tfLogStreamCellAttributes))
			definition = ddLogStreamCellAttribute
		}
	}

	return &definition, nil
}

func buildDatadogUpdateNotebookCell(tfNotebookCell map[string]interface{}) (*datadogV1.NotebookUpdateCell, error) {
	var ddNotebookCell datadogV1.NotebookUpdateCell
	if cellId, ok := tfNotebookCell["id"].(string); ok {
		ddNotebookUpdateCell := datadogV1.NotebookCellUpdateRequest{}

		ddNotebookUpdateCell.SetId(cellId)
		if tfCellType, ok := tfNotebookCell["type"].(string); ok {
			ddCellType, _ := datadogV1.NewNotebookCellResourceTypeFromValue(tfCellType)
			ddNotebookUpdateCell.SetType(*ddCellType)
		}

		if tfCellAttributes, ok := tfNotebookCell["attributes"].([]interface{}); ok {
			ddCellAttributes, _ := buildDatadogUpdateNotebookCellAttributes(tfCellAttributes[0].(map[string]interface{}))
			ddNotebookUpdateCell.SetAttributes(*ddCellAttributes)
		}

		ddNotebookCell = datadogV1.NotebookCellUpdateRequestAsNotebookUpdateCell(&ddNotebookUpdateCell)
	} else {
		ddNotebookCreateCell, _ := buildDatadogCreateNotebookCell(tfNotebookCell)
		ddNotebookCell = datadogV1.NotebookCellCreateRequestAsNotebookUpdateCell(ddNotebookCreateCell)
	}

	return &ddNotebookCell, nil
}

func buildDatadogUpdateNotebookCells(terraformNotebookCells []interface{}) (*[]datadogV1.NotebookUpdateCell, error) {
	ddCells := make([]datadogV1.NotebookUpdateCell, len(terraformNotebookCells))

	for i, notebookCell := range terraformNotebookCells {
		datadogCell, err := buildDatadogUpdateNotebookCell(notebookCell.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		ddCells[i] = *datadogCell
	}

	return &ddCells, nil
}

func buildDatadogNotebookUpdate(d *schema.ResourceData) (*datadogV1.NotebookUpdateData, error) {
	notebookUpdateDataAttributes := datadogV1.NotebookUpdateDataAttributes{}

	if tfNotebookCells, ok := d.Get("cells").([]interface{}); ok {
		ddCells, _ := buildDatadogUpdateNotebookCells(tfNotebookCells)
		notebookUpdateDataAttributes.SetCells(*ddCells)
	}

	if tfNotebookTime, ok := d.Get("time").([]interface{}); ok && len(tfNotebookTime) > 0 {
		ddTime := buildDatadogNotebookTime(tfNotebookTime[0].(map[string]interface{}))
		notebookUpdateDataAttributes.SetTime(*ddTime)
	}

	if tfNotebookStatus, ok := d.Get("status").(string); ok {
		ddStatus, _ := datadogV1.NewNotebookStatusFromValue(tfNotebookStatus)
		notebookUpdateDataAttributes.SetStatus(*ddStatus)
	}

	if tfNotebookMetadata, ok := d.Get("metadata").([]interface{}); ok && len(tfNotebookMetadata) > 0 {
		ddMetadata := buildDatadogNotebookMetadata(tfNotebookMetadata[0].(map[string]interface{}))
		notebookUpdateDataAttributes.SetMetadata(*ddMetadata)
	}

	if tfNotebookName, ok := d.Get("name").(string); ok {
		notebookUpdateDataAttributes.SetName(tfNotebookName)
	}

	notebookType, _ := datadogV1.NewNotebookResourceTypeFromValue(d.Get("type").(string))
	notebookUpdateData := datadogV1.NewNotebookUpdateData(notebookUpdateDataAttributes, *notebookType)

	return notebookUpdateData, nil
}

func buildDatadogMarkdownCellDefinition(tfCellAttribute map[string]interface{}) *datadogV1.NotebookMarkdownCellAttributes {
	cellAttributes := datadogV1.NotebookMarkdownCellAttributes{}
	tfDefinition := tfCellAttribute["definition"].([]interface{})[0].(map[string]interface{})

	ddDefinition := datadogV1.NotebookMarkdownCellDefinition{}
	if tfType, ok := tfDefinition["type"]; ok {
		ddType, _ := datadogV1.NewNotebookMarkdownCellDefinitionTypeFromValue(tfType.(string))
		ddDefinition.SetType(*ddType)
	}
	if tfType, ok := tfDefinition["text"]; ok {
		ddDefinition.SetText(tfType.(string))
	}

	cellAttributes.SetDefinition(ddDefinition)

	return &cellAttributes
}

func buildDatadogTimeseriesCellDefinition(tfCellAttribute map[string]interface{}) *datadogV1.NotebookTimeseriesCellAttributes {
	cellAttributes := datadogV1.NotebookTimeseriesCellAttributes{}

	tfDefinition := tfCellAttribute["definition"].([]interface{})[0].(map[string]interface{})
	cellAttributes.SetDefinition(*buildDatadogTimeseriesDefinition(tfDefinition))

	if tfSplitBy, ok := tfCellAttribute["split_by"].([]interface{}); ok && len(tfSplitBy) > 0 {
		cellAttributes.SetSplitBy(*buildDDSplitBy(tfSplitBy[0].(map[string]interface{})))
	}

	if tfGraphSize, ok := tfCellAttribute["graph_size"].(string); ok && len(tfGraphSize) > 0 {
		graphSize, _ := datadogV1.NewNotebookGraphSizeFromValue(tfGraphSize)
		cellAttributes.SetGraphSize(*graphSize)
	}

	if tfTime, ok := tfCellAttribute["time"].([]interface{}); ok && len(tfTime) > 0 {
		ddTime := buildDatadogNotebookCellTime(tfTime[0].(map[string]interface{}))
		cellAttributes.SetTime(*ddTime)
	}

	return &cellAttributes
}

func buildDatadogToplistCellDefinition(tfCellAttribute map[string]interface{}) *datadogV1.NotebookToplistCellAttributes {
	cellAttributes := datadogV1.NotebookToplistCellAttributes{}

	tfDefinition := tfCellAttribute["definition"].([]interface{})[0].(map[string]interface{})
	cellAttributes.SetDefinition(*buildDatadogToplistDefinition(tfDefinition))

	if tfSplitBy, ok := tfCellAttribute["split_by"].([]interface{}); ok && len(tfSplitBy) > 0 {
		cellAttributes.SetSplitBy(*buildDDSplitBy(tfSplitBy[0].(map[string]interface{})))
	}

	if tfGraphSize, ok := tfCellAttribute["graph_size"].(string); ok && len(tfGraphSize) > 0 {
		graphSize, _ := datadogV1.NewNotebookGraphSizeFromValue(tfGraphSize)
		cellAttributes.SetGraphSize(*graphSize)
	}

	if tfTime, ok := tfCellAttribute["time"].([]interface{}); ok && len(tfTime) > 0 {
		ddTime := buildDatadogNotebookCellTime(tfTime[0].(map[string]interface{}))
		cellAttributes.SetTime(*ddTime)
	}

	return &cellAttributes
}

func buildDatadogHeatmapCellDefinition(tfCellAttribute map[string]interface{}) *datadogV1.NotebookHeatMapCellAttributes {
	cellAttributes := datadogV1.NotebookHeatMapCellAttributes{}

	tfDefinition := tfCellAttribute["definition"].([]interface{})[0].(map[string]interface{})
	cellAttributes.SetDefinition(*buildDatadogHeatmapDefinition(tfDefinition))

	if tfSplitBy, ok := tfCellAttribute["split_by"].([]interface{}); ok && len(tfSplitBy) > 0 {
		cellAttributes.SetSplitBy(*buildDDSplitBy(tfSplitBy[0].(map[string]interface{})))
	}

	if tfGraphSize, ok := tfCellAttribute["graph_size"].(string); ok && len(tfGraphSize) > 0 {
		graphSize, _ := datadogV1.NewNotebookGraphSizeFromValue(tfGraphSize)
		cellAttributes.SetGraphSize(*graphSize)
	}

	if tfTime, ok := tfCellAttribute["time"].([]interface{}); ok && len(tfTime) > 0 {
		ddTime := buildDatadogNotebookCellTime(tfTime[0].(map[string]interface{}))
		cellAttributes.SetTime(*ddTime)
	}

	return &cellAttributes
}

func buildDatadogDistributionCellDefinition(tfCellAttribute map[string]interface{}) *datadogV1.NotebookDistributionCellAttributes {
	cellAttributes := datadogV1.NotebookDistributionCellAttributes{}

	tfDefinition := tfCellAttribute["definition"].([]interface{})[0].(map[string]interface{})
	cellAttributes.SetDefinition(*buildDatadogDistributionDefinition(tfDefinition))

	if tfSplitBy, ok := tfCellAttribute["split_by"].([]interface{}); ok && len(tfSplitBy) > 0 {
		cellAttributes.SetSplitBy(*buildDDSplitBy(tfSplitBy[0].(map[string]interface{})))
	}

	if tfGraphSize, ok := tfCellAttribute["graph_size"].(string); ok && len(tfGraphSize) > 0 {
		graphSize, _ := datadogV1.NewNotebookGraphSizeFromValue(tfGraphSize)
		cellAttributes.SetGraphSize(*graphSize)
	}

	if tfTime, ok := tfCellAttribute["time"].([]interface{}); ok && len(tfTime) > 0 {
		ddTime := buildDatadogNotebookCellTime(tfTime[0].(map[string]interface{}))
		cellAttributes.SetTime(*ddTime)
	}

	return &cellAttributes
}

func buildDatadogLogStreamCellDefinition(tfCellAttribute map[string]interface{}) *datadogV1.NotebookLogStreamCellAttributes {
	cellAttributes := datadogV1.NotebookLogStreamCellAttributes{}

	tfDefinition := tfCellAttribute["definition"].([]interface{})[0].(map[string]interface{})
	cellAttributes.SetDefinition(*buildDatadogLogStreamDefinition(tfDefinition))

	if tfGraphSize, ok := tfCellAttribute["graph_size"].(string); ok && len(tfGraphSize) > 0 {
		graphSize, _ := datadogV1.NewNotebookGraphSizeFromValue(tfGraphSize)
		cellAttributes.SetGraphSize(*graphSize)
	}

	if tfTime, ok := tfCellAttribute["time"].([]interface{}); ok && len(tfTime) > 0 {
		ddTime := buildDatadogNotebookCellTime(tfTime[0].(map[string]interface{}))
		cellAttributes.SetTime(*ddTime)
	}

	return &cellAttributes
}

func buildDDSplitBy(tfSplitBy map[string]interface{}) *datadogV1.NotebookSplitBy {
	var keys []string
	var tags []string

	for _, tag := range tfSplitBy["tags"].(*schema.Set).List() {
		tags = append(tags, tag.(string))
	}

	for _, key := range tfSplitBy["keys"].([]interface{}) {
		keys = append(keys, key.(string))
	}

	ddSplitBy := datadogV1.NotebookSplitBy{
		Tags: tags,
		Keys: keys,
	}
	return &ddSplitBy
}

func buildDatadogCreateNotebookCellAttributes(tfAttributes map[string]interface{}) (*datadogV1.NotebookCellCreateRequestAttributes, error) {
	var definition datadogV1.NotebookCellCreateRequestAttributes

	if markdown, ok := tfAttributes["markdown_cell"].([]interface{}); ok && len(markdown) > 0 {
		if tfMarkdownCellAttributes, ok := markdown[0].(map[string]interface{}); ok {
			ddMarkdownCellAttribute := datadogV1.NotebookMarkdownCellAttributesAsNotebookCellCreateRequestAttributes(buildDatadogMarkdownCellDefinition(tfMarkdownCellAttributes))
			definition = ddMarkdownCellAttribute
		}
	} else if timeseries, ok := tfAttributes["timeseries_cell"].([]interface{}); ok && len(timeseries) > 0 {
		if tfTimeseriesCellAttributes, ok := timeseries[0].(map[string]interface{}); ok {
			ddTimeseriesCellAttribute := datadogV1.NotebookTimeseriesCellAttributesAsNotebookCellCreateRequestAttributes(buildDatadogTimeseriesCellDefinition(tfTimeseriesCellAttributes))
			definition = ddTimeseriesCellAttribute
		}
	} else if toplist, ok := tfAttributes["toplist_cell"].([]interface{}); ok && len(toplist) > 0 {
		if tfToplistCellAttributes, ok := toplist[0].(map[string]interface{}); ok {
			ddToplistCellAttribute := datadogV1.NotebookToplistCellAttributesAsNotebookCellCreateRequestAttributes(buildDatadogToplistCellDefinition(tfToplistCellAttributes))
			definition = ddToplistCellAttribute
		}
	} else if heatmap, ok := tfAttributes["heatmap_cell"].([]interface{}); ok && len(heatmap) > 0 {
		if tfHeatmapCellAttributes, ok := heatmap[0].(map[string]interface{}); ok {
			ddHeatmapCellAttribute := datadogV1.NotebookHeatMapCellAttributesAsNotebookCellCreateRequestAttributes(buildDatadogHeatmapCellDefinition(tfHeatmapCellAttributes))
			definition = ddHeatmapCellAttribute
		}
	} else if distribution, ok := tfAttributes["distribution_cell"].([]interface{}); ok && len(distribution) > 0 {
		if tfDistributionCellAttributes, ok := distribution[0].(map[string]interface{}); ok {
			ddDistributionCellAttribute := datadogV1.NotebookDistributionCellAttributesAsNotebookCellCreateRequestAttributes(buildDatadogDistributionCellDefinition(tfDistributionCellAttributes))
			definition = ddDistributionCellAttribute
		}
	} else if logstream, ok := tfAttributes["log_stream_cell"].([]interface{}); ok && len(logstream) > 0 {
		if tfLogStreamCellAttributes, ok := logstream[0].(map[string]interface{}); ok {
			ddLogStreamCellAttribute := datadogV1.NotebookLogStreamCellAttributesAsNotebookCellCreateRequestAttributes(buildDatadogLogStreamCellDefinition(tfLogStreamCellAttributes))
			definition = ddLogStreamCellAttribute
		}
	}

	return &definition, nil
}

func buildDatadogCreateNotebookCell(tfNotebookCell map[string]interface{}) (*datadogV1.NotebookCellCreateRequest, error) {
	ddNotebookCell := datadogV1.NewNotebookCellCreateRequestWithDefaults()

	if tfCellType, ok := tfNotebookCell["type"].(string); ok {
		ddCellType, _ := datadogV1.NewNotebookCellResourceTypeFromValue(tfCellType)
		ddNotebookCell.SetType(*ddCellType)
	}

	if tfCellAttributes, ok := tfNotebookCell["attributes"].([]interface{}); ok {
		ddCellAttributes, _ := buildDatadogCreateNotebookCellAttributes(tfCellAttributes[0].(map[string]interface{}))
		ddNotebookCell.SetAttributes(*ddCellAttributes)
	}

	return ddNotebookCell, nil
}

func buildDatadogCreateNotebookCells(terraformNotebookCells []interface{}) (*[]datadogV1.NotebookCellCreateRequest, error) {
	ddCells := make([]datadogV1.NotebookCellCreateRequest, len(terraformNotebookCells))

	for i, notebookCell := range terraformNotebookCells {
		datadogCell, err := buildDatadogCreateNotebookCell(notebookCell.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		ddCells[i] = *datadogCell
	}

	return &ddCells, nil
}

func buildDatadogNotebookTime(tfTime map[string]interface{}) *datadogV1.NotebookGlobalTime {
	var notebookGlobalTime datadogV1.NotebookGlobalTime
	if tfNotebookRelativeTime, ok := tfTime["notebook_relative_time"].([]interface{}); ok && len(tfNotebookRelativeTime) > 0 {
		tfRelativeTime := tfNotebookRelativeTime[0].(map[string]interface{})
		liveSpan, _ := datadogV1.NewWidgetLiveSpanFromValue(tfRelativeTime["live_span"].(string))
		ddRelativeTime := datadogV1.NotebookRelativeTime{
			LiveSpan: *liveSpan,
		}
		notebookGlobalTime = datadogV1.NotebookRelativeTimeAsNotebookGlobalTime(&ddRelativeTime)
	} else if tfNotebookAbsoluteTime, ok := tfTime["notebook_absolute_time"].([]interface{}); ok && len(tfNotebookAbsoluteTime) > 0 {
		tfAbsoluteTime := tfNotebookAbsoluteTime[0].(map[string]interface{})
		start, _ := time.Parse(timeFormat, tfAbsoluteTime["start"].(string))
		end, _ := time.Parse(timeFormat, tfAbsoluteTime["end"].(string))

		ddAbsoluteTime := datadogV1.NotebookAbsoluteTime{
			Start: start,
			End:   end,
		}

		if live := tfAbsoluteTime["live"].(bool); ok {
			ddAbsoluteTime.SetLive(live)
		}

		notebookGlobalTime = datadogV1.NotebookAbsoluteTimeAsNotebookGlobalTime(&ddAbsoluteTime)

	}

	return &notebookGlobalTime
}

func buildDatadogNotebookCellTime(tfTime map[string]interface{}) *datadogV1.NotebookCellTime {
	var notebookGlobalTime datadogV1.NotebookCellTime
	if tfNotebookRelativeTime, ok := tfTime["notebook_relative_time"].([]interface{}); ok && len(tfNotebookRelativeTime) > 0 {
		tfRelativeTime := tfNotebookRelativeTime[0].(map[string]interface{})
		liveSpan, _ := datadogV1.NewWidgetLiveSpanFromValue(tfRelativeTime["live_span"].(string))
		ddRelativeTime := datadogV1.NotebookRelativeTime{
			LiveSpan: *liveSpan,
		}
		notebookGlobalTime = datadogV1.NotebookRelativeTimeAsNotebookCellTime(&ddRelativeTime)
	} else if tfNotebookAbsoluteTime, ok := tfTime["notebook_absolute_time"].([]interface{}); ok && len(tfNotebookAbsoluteTime) > 0 {
		tfAbsoluteTime := tfNotebookAbsoluteTime[0].(map[string]interface{})
		start, _ := time.Parse(timeFormat, tfAbsoluteTime["start"].(string))
		end, _ := time.Parse(timeFormat, tfAbsoluteTime["end"].(string))

		ddAbsoluteTime := datadogV1.NotebookAbsoluteTime{
			Start: start,
			End:   end,
		}

		if live := tfAbsoluteTime["live"].(bool); ok {
			ddAbsoluteTime.SetLive(live)
		}

		notebookGlobalTime = datadogV1.NotebookAbsoluteTimeAsNotebookCellTime(&ddAbsoluteTime)

	}

	return &notebookGlobalTime
}

func buildDatadogNotebookMetadata(tfMetadata map[string]interface{}) *datadogV1.NotebookMetadata {
	notebookMetadata := datadogV1.NotebookMetadata{
		IsTemplate:    datadogV1.PtrBool(tfMetadata["is_template"].(bool)),
		TakeSnapshots: datadogV1.PtrBool(tfMetadata["take_snapshots"].(bool)),
	}

	if metadataType, ok := tfMetadata["type"].(string); ok {
		ddType, _ := datadogV1.NewNotebookMetadataTypeFromValue(metadataType)
		notebookMetadata.SetType(*ddType)
	}

	return &notebookMetadata
}

func buildTerraformNotebookMarkdownCellAttributes(attribute datadogV1.NotebookMarkdownCellAttributes) []map[string]interface{} {
	definition := []map[string]interface{}{{
		"type": attribute.Definition.GetType(),
		"text": attribute.Definition.GetText(),
	}}

	tfAttributes := []map[string]interface{}{{
		"markdown_cell": []map[string]interface{}{{"definition": definition}},
	}}

	return tfAttributes
}

func buildTerraformNotebookTimeseriesCellAttributes(attribute datadogV1.NotebookTimeseriesCellAttributes) []map[string]interface{} {
	tCell := map[string]interface{}{}

	definition := []interface{}{
		buildTerraformTimeseriesDefinition(attribute.Definition, utils.NewResourceDataKey(&schema.ResourceData{}, fmt.Sprintf(""))),
	}
	tCell["definition"] = definition

	if attribute.SplitBy != nil {
		tCell["split_by"] = []map[string]interface{}{{
			"keys": attribute.SplitBy.GetKeys(),
			"tags": attribute.SplitBy.GetTags(),
		}}
	}

	if ddTime, ok := attribute.GetTimeOk(); ok {
		tCell["time"] = buildTerraformCellTime(ddTime)
	}

	if graphSize, ok := attribute.GetGraphSizeOk(); ok {
		tCell["graph_size"] = *graphSize
	}

	tfAttributes := []map[string]interface{}{{
		"timeseries_cell": []map[string]interface{}{tCell},
	}}

	return tfAttributes
}

func buildTerraformNotebookToplistCellAttributes(attribute datadogV1.NotebookToplistCellAttributes) []map[string]interface{} {
	tCell := map[string]interface{}{}

	definition := []interface{}{
		buildTerraformToplistDefinition(attribute.Definition, utils.NewResourceDataKey(&schema.ResourceData{}, fmt.Sprintf(""))),
	}
	tCell["definition"] = definition

	if attribute.SplitBy != nil {
		tCell["split_by"] = []map[string]interface{}{{
			"keys": attribute.SplitBy.GetKeys(),
			"tags": attribute.SplitBy.GetTags(),
		}}
	}

	if ddTime, ok := attribute.GetTimeOk(); ok {
		tCell["time"] = buildTerraformCellTime(ddTime)
	}

	if graphSize, ok := attribute.GetGraphSizeOk(); ok {
		tCell["graph_size"] = *graphSize
	}

	tfAttributes := []map[string]interface{}{{
		"toplist_cell": []map[string]interface{}{tCell},
	}}

	return tfAttributes
}

func buildTerraformNotebookHeatmapCellAttributes(attribute datadogV1.NotebookHeatMapCellAttributes) []map[string]interface{} {
	tCell := map[string]interface{}{}

	definition := []interface{}{
		buildTerraformHeatmapDefinition(attribute.Definition, utils.NewResourceDataKey(&schema.ResourceData{}, fmt.Sprintf(""))),
	}
	tCell["definition"] = definition

	if attribute.SplitBy != nil {
		tCell["split_by"] = []map[string]interface{}{{
			"keys": attribute.SplitBy.GetKeys(),
			"tags": attribute.SplitBy.GetTags(),
		}}
	}

	if ddTime, ok := attribute.GetTimeOk(); ok {
		tCell["time"] = buildTerraformCellTime(ddTime)
	}

	if graphSize, ok := attribute.GetGraphSizeOk(); ok {
		tCell["graph_size"] = *graphSize
	}

	tfAttributes := []map[string]interface{}{{
		"heatmap_cell": []map[string]interface{}{tCell},
	}}

	return tfAttributes
}

func buildTerraformNotebookDistributionCellAttributes(attribute datadogV1.NotebookDistributionCellAttributes) []map[string]interface{} {
	tCell := map[string]interface{}{}

	definition := []interface{}{
		buildTerraformDistributionDefinition(attribute.Definition, utils.NewResourceDataKey(&schema.ResourceData{}, fmt.Sprintf(""))),
	}
	tCell["definition"] = definition

	if attribute.SplitBy != nil {
		tCell["split_by"] = []map[string]interface{}{{
			"keys": attribute.SplitBy.GetKeys(),
			"tags": attribute.SplitBy.GetTags(),
		}}
	}

	if ddTime, ok := attribute.GetTimeOk(); ok {
		tCell["time"] = buildTerraformCellTime(ddTime)
	}

	if graphSize, ok := attribute.GetGraphSizeOk(); ok {
		tCell["graph_size"] = *graphSize
	}

	tfAttributes := []map[string]interface{}{{
		"distribution_cell": []map[string]interface{}{tCell},
	}}

	return tfAttributes
}

func buildTerraformNotebookLogStreamCellAttributes(attribute datadogV1.NotebookLogStreamCellAttributes) []map[string]interface{} {
	tCell := map[string]interface{}{}

	definition := []interface{}{
		buildTerraformLogStreamDefinition(attribute.Definition, utils.NewResourceDataKey(&schema.ResourceData{}, fmt.Sprintf(""))),
	}
	tCell["definition"] = definition

	if ddTime, ok := attribute.GetTimeOk(); ok {
		tCell["time"] = buildTerraformCellTime(ddTime)
	}

	if graphSize, ok := attribute.GetGraphSizeOk(); ok {
		tCell["graph_size"] = *graphSize
	}

	tfAttributes := []map[string]interface{}{{
		"log_stream_cell": []map[string]interface{}{tCell},
	}}

	return tfAttributes
}

func buildTerraformNotebookCell(datadogDefinition datadogV1.NotebookCellResponse) *map[string]interface{} {
	terraformNotebookCell := map[string]interface{}{}

	terraformNotebookCell["id"] = datadogDefinition.GetId()
	terraformNotebookCell["type"] = datadogDefinition.GetType()

	if datadogDefinition.Attributes.NotebookMarkdownCellAttributes != nil {
		terraformAttributeDefinition := buildTerraformNotebookMarkdownCellAttributes(*datadogDefinition.Attributes.NotebookMarkdownCellAttributes)
		terraformNotebookCell["attributes"] = terraformAttributeDefinition
	} else if datadogDefinition.Attributes.NotebookTimeseriesCellAttributes != nil {
		terraformAttributeDefinition := buildTerraformNotebookTimeseriesCellAttributes(*datadogDefinition.Attributes.NotebookTimeseriesCellAttributes)
		terraformNotebookCell["attributes"] = terraformAttributeDefinition
	} else if datadogDefinition.Attributes.NotebookToplistCellAttributes != nil {
		terraformAttributeDefinition := buildTerraformNotebookToplistCellAttributes(*datadogDefinition.Attributes.NotebookToplistCellAttributes)
		terraformNotebookCell["attributes"] = terraformAttributeDefinition
	} else if datadogDefinition.Attributes.NotebookHeatMapCellAttributes != nil {
		terraformAttributeDefinition := buildTerraformNotebookHeatmapCellAttributes(*datadogDefinition.Attributes.NotebookHeatMapCellAttributes)
		terraformNotebookCell["attributes"] = terraformAttributeDefinition
	} else if datadogDefinition.Attributes.NotebookDistributionCellAttributes != nil {
		terraformAttributeDefinition := buildTerraformNotebookDistributionCellAttributes(*datadogDefinition.Attributes.NotebookDistributionCellAttributes)
		terraformNotebookCell["attributes"] = terraformAttributeDefinition
	} else if datadogDefinition.Attributes.NotebookLogStreamCellAttributes != nil {
		terraformAttributeDefinition := buildTerraformNotebookLogStreamCellAttributes(*datadogDefinition.Attributes.NotebookLogStreamCellAttributes)
		terraformNotebookCell["attributes"] = terraformAttributeDefinition
	}

	return &terraformNotebookCell
}

func buildTerraformCells(datadogNotebookCells *[]datadogV1.NotebookCellResponse) []map[string]interface{} {
	terraformNotebookCell := make([]map[string]interface{}, len(*datadogNotebookCells))
	for i, datadogCell := range *datadogNotebookCells {
		terraformWidget := buildTerraformNotebookCell(datadogCell)
		terraformNotebookCell[i] = *terraformWidget
	}
	return terraformNotebookCell

}

func buildTerraformCellTime(notebookCellTime *datadogV1.NotebookCellTime) []map[string]interface{} {
	var tfNotebookCellTime map[string]interface{}

	if notebookCellTime.NotebookRelativeTime != nil {
		tfNotebookCellTime = map[string]interface{}{
			"notebook_relative_time": []map[string]interface{}{{
				"live_span": notebookCellTime.NotebookRelativeTime.LiveSpan,
			}},
		}
	} else if notebookCellTime.NotebookAbsoluteTime != nil {
		tfNotebookCellTime = map[string]interface{}{
			"notebook_absolute_time": []map[string]interface{}{{
				"start": notebookCellTime.NotebookAbsoluteTime.Start.Format(timeFormat),
				"end":   notebookCellTime.NotebookAbsoluteTime.End.Format(timeFormat),
				"live":  notebookCellTime.NotebookAbsoluteTime.Live,
			}},
		}
	}

	return []map[string]interface{}{tfNotebookCellTime}
}

func buildTerraformTime(globalTime *datadogV1.NotebookGlobalTime) []map[string]interface{} {
	var tfNotebookTime map[string]interface{}

	if globalTime.NotebookRelativeTime != nil {
		tfNotebookTime = map[string]interface{}{
			"notebook_relative_time": []map[string]interface{}{{
				"live_span": globalTime.NotebookRelativeTime.LiveSpan,
			}},
		}
	} else if globalTime.NotebookAbsoluteTime != nil {
		tfNotebookTime = map[string]interface{}{
			"notebook_absolute_time": []map[string]interface{}{{
				"start": globalTime.NotebookAbsoluteTime.Start.Format(timeFormat),
				"end":   globalTime.NotebookAbsoluteTime.End.Format(timeFormat),
				"live":  globalTime.NotebookAbsoluteTime.Live,
			}},
		}
	}

	return []map[string]interface{}{tfNotebookTime}
}

func buildTerraformMetadata(metadata *datadogV1.NotebookMetadata) []map[string]interface{} {
	tfNotebookMetadata := map[string]interface{}{}

	if isTemplate, ok := metadata.GetIsTemplateOk(); ok {
		tfNotebookMetadata["is_template"] = isTemplate
	}

	if takeSnapshots, ok := metadata.GetTakeSnapshotsOk(); ok {
		tfNotebookMetadata["take_snapshots"] = takeSnapshots
	}

	if notebookType, ok := metadata.GetTypeOk(); ok {
		tfNotebookMetadata["type"] = notebookType
	}

	return []map[string]interface{}{tfNotebookMetadata}
}

func buildCreateDatadogNotebook(d *schema.ResourceData) (*datadogV1.NotebookCreateData, error) {
	notebookCreateDataAttributes := datadogV1.NotebookCreateDataAttributes{}

	if tfNotebookCells, ok := d.Get("cells").([]interface{}); ok {
		ddCells, _ := buildDatadogCreateNotebookCells(tfNotebookCells)
		notebookCreateDataAttributes.SetCells(*ddCells)
	}

	if tfNotebookTime, ok := d.Get("time").([]interface{}); ok && len(tfNotebookTime) > 0 {
		ddTime := buildDatadogNotebookTime(tfNotebookTime[0].(map[string]interface{}))
		notebookCreateDataAttributes.SetTime(*ddTime)
	}

	if tfNotebookStatus, ok := d.Get("status").(string); ok {
		ddStatus, _ := datadogV1.NewNotebookStatusFromValue(tfNotebookStatus)
		notebookCreateDataAttributes.SetStatus(*ddStatus)
	}

	if tfNotebookMetadata, ok := d.Get("metadata").([]interface{}); ok && len(tfNotebookMetadata) > 0 {
		ddMetadata := buildDatadogNotebookMetadata(tfNotebookMetadata[0].(map[string]interface{}))
		notebookCreateDataAttributes.SetMetadata(*ddMetadata)
	}

	if tfNotebookName, ok := d.Get("name").(string); ok {
		notebookCreateDataAttributes.SetName(tfNotebookName)
	}

	notebookType, _ := datadogV1.NewNotebookResourceTypeFromValue(d.Get("type").(string))

	notebookCreateData := datadogV1.NewNotebookCreateData(notebookCreateDataAttributes, *notebookType)

	return notebookCreateData, nil
}
