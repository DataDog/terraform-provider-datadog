package datadog

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var logCustomPipelineMutex = sync.Mutex{}

const (
	tfArithmeticProcessor        = "arithmetic_processor"
	tfAttributeRemapperProcessor = "attribute_remapper"
	tfCategoryProcessor          = "category_processor"
	tfDateRemapperProcessor      = "date_remapper"
	tfGeoIPParserProcessor       = "geo_ip_parser"
	tfGrokParserProcessor        = "grok_parser"
	tfLookupProcessor            = "lookup_processor"
	tfMessageRemapperProcessor   = "message_remapper"
	tfNestedPipelineProcessor    = "pipeline"
	tfServiceRemapperProcessor   = "service_remapper"
	tfStatusRemapperProcessor    = "status_remapper"
	tfStringBuilderProcessor     = "string_builder_processor"
	tfTraceIDRemapperProcessor   = "trace_id_remapper"
	tfURLParserProcessor         = "url_parser"
	tfUserAgentParserProcessor   = "user_agent_parser"
)

var tfProcessorTypes = map[string]string{
	tfArithmeticProcessor:        string(datadogV1.LOGSARITHMETICPROCESSORTYPE_ARITHMETIC_PROCESSOR),
	tfAttributeRemapperProcessor: string(datadogV1.LOGSATTRIBUTEREMAPPERTYPE_ATTRIBUTE_REMAPPER),
	tfCategoryProcessor:          string(datadogV1.LOGSCATEGORYPROCESSORTYPE_CATEGORY_PROCESSOR),
	tfDateRemapperProcessor:      string(datadogV1.LOGSDATEREMAPPERTYPE_DATE_REMAPPER),
	tfGeoIPParserProcessor:       string(datadogV1.LOGSGEOIPPARSERTYPE_GEO_IP_PARSER),
	tfGrokParserProcessor:        string(datadogV1.LOGSGROKPARSERTYPE_GROK_PARSER),
	tfLookupProcessor:            string(datadogV1.LOGSLOOKUPPROCESSORTYPE_LOOKUP_PROCESSOR),
	tfMessageRemapperProcessor:   string(datadogV1.LOGSMESSAGEREMAPPERTYPE_MESSAGE_REMAPPER),
	tfNestedPipelineProcessor:    string(datadogV1.LOGSPIPELINEPROCESSORTYPE_PIPELINE),
	tfServiceRemapperProcessor:   string(datadogV1.LOGSSERVICEREMAPPERTYPE_SERVICE_REMAPPER),
	tfStatusRemapperProcessor:    string(datadogV1.LOGSSTATUSREMAPPERTYPE_STATUS_REMAPPER),
	tfStringBuilderProcessor:     string(datadogV1.LOGSSTRINGBUILDERPROCESSORTYPE_STRING_BUILDER_PROCESSOR),
	tfTraceIDRemapperProcessor:   string(datadogV1.LOGSTRACEREMAPPERTYPE_TRACE_ID_REMAPPER),
	tfURLParserProcessor:         string(datadogV1.LOGSURLPARSERTYPE_URL_PARSER),
	tfUserAgentParserProcessor:   string(datadogV1.LOGSUSERAGENTPARSERTYPE_USER_AGENT_PARSER),
}

var tfProcessors = map[string]*schema.Schema{
	tfArithmeticProcessor:        arithmeticProcessor,
	tfAttributeRemapperProcessor: attributeRemapper,
	tfCategoryProcessor:          categoryProcessor,
	tfDateRemapperProcessor:      dateRemapper,
	tfGeoIPParserProcessor:       geoIPParser,
	tfGrokParserProcessor:        grokParser,
	tfLookupProcessor:            lookupProcessor,
	tfMessageRemapperProcessor:   messageRemapper,
	tfServiceRemapperProcessor:   serviceRemapper,
	tfStatusRemapperProcessor:    statusRemmaper,
	tfStringBuilderProcessor:     stringBuilderProcessor,
	tfTraceIDRemapperProcessor:   traceIDRemapper,
	tfURLParserProcessor:         urlParser,
	tfUserAgentParserProcessor:   userAgentParser,
}

var ddProcessorTypes = map[string]string{
	string(datadogV1.LOGSARITHMETICPROCESSORTYPE_ARITHMETIC_PROCESSOR):        tfArithmeticProcessor,
	string(datadogV1.LOGSATTRIBUTEREMAPPERTYPE_ATTRIBUTE_REMAPPER):            tfAttributeRemapperProcessor,
	string(datadogV1.LOGSCATEGORYPROCESSORTYPE_CATEGORY_PROCESSOR):            tfCategoryProcessor,
	string(datadogV1.LOGSDATEREMAPPERTYPE_DATE_REMAPPER):                      tfDateRemapperProcessor,
	string(datadogV1.LOGSGEOIPPARSERTYPE_GEO_IP_PARSER):                       tfGeoIPParserProcessor,
	string(datadogV1.LOGSGROKPARSERTYPE_GROK_PARSER):                          tfGrokParserProcessor,
	string(datadogV1.LOGSLOOKUPPROCESSORTYPE_LOOKUP_PROCESSOR):                tfLookupProcessor,
	string(datadogV1.LOGSMESSAGEREMAPPERTYPE_MESSAGE_REMAPPER):                tfMessageRemapperProcessor,
	string(datadogV1.LOGSPIPELINEPROCESSORTYPE_PIPELINE):                      tfNestedPipelineProcessor,
	string(datadogV1.LOGSSERVICEREMAPPERTYPE_SERVICE_REMAPPER):                tfServiceRemapperProcessor,
	string(datadogV1.LOGSSTATUSREMAPPERTYPE_STATUS_REMAPPER):                  tfStatusRemapperProcessor,
	string(datadogV1.LOGSSTRINGBUILDERPROCESSORTYPE_STRING_BUILDER_PROCESSOR): tfStringBuilderProcessor,
	string(datadogV1.LOGSTRACEREMAPPERTYPE_TRACE_ID_REMAPPER):                 tfTraceIDRemapperProcessor,
	string(datadogV1.LOGSURLPARSERTYPE_URL_PARSER):                            tfURLParserProcessor,
	string(datadogV1.LOGSUSERAGENTPARSERTYPE_USER_AGENT_PARSER):               tfUserAgentParserProcessor,
}

var arithmeticProcessor = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Arithmetic Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#arithmetic-processor)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Your pipeline name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"is_enabled": {
				Description: "Boolean value to enable your pipeline.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"expression": {
				Description: "Arithmetic operation between one or more log attributes.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target": {
				Description: "Name of the attribute that contains the result of the arithmetic operation.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"is_replace_missing": {
				Description: "If true, it replaces all missing attributes of expression by 0, false skips the operation if an attribute is missing.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	},
}

var attributeRemapper = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Attribute Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#remapper)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":        {Description: "Name of the processor", Type: schema.TypeString, Optional: true},
			"is_enabled":  {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
			"sources":     {Description: "List of source attributes or tags.", Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"source_type": {Description: "Defines where the sources are from (log `attribute` or `tag`).", Type: schema.TypeString, Required: true},
			"target":      {Description: "Final attribute or tag name to remap the sources.", Type: schema.TypeString, Required: true},
			"target_type": {Description: "Defines if the target is a log `attribute` or `tag`.", Type: schema.TypeString, Required: true},
			"target_format": {
				Description:  "If the `target_type` of the remapper is `attribute`, try to cast the value to a new specific type. If the cast is not possible, the original type is kept. `string`, `integer`, or `double` are the possible types. If the `target_type` is `tag`, this parameter may not be specified.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"auto", "string", "integer", "double"}, false),
			},
			"preserve_source":      {Description: "Remove or preserve the remapped source element.", Type: schema.TypeBool, Optional: true},
			"override_on_conflict": {Description: "Override the target element if already set.", Type: schema.TypeBool, Optional: true},
		},
	},
}

var categoryProcessor = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Category Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#category-processor)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Description: "Name of the category", Type: schema.TypeString, Optional: true},
			"is_enabled": {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
			"target":     {Description: "Name of the target attribute whose value is defined by the matching category.", Type: schema.TypeString, Required: true},
			"category": {Description: "List of filters to match or exclude a log with their corresponding name to assign a custom value to the log.", Type: schema.TypeList, Required: true, Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"filter": {
						Type:     schema.TypeList,
						Required: true,
						MaxItems: 1,
						Elem:     getFilterSchema(),
					},
					"name": {Type: schema.TypeString, Required: true},
				},
			}},
		},
	},
}

var dateRemapper = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Date Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-date-remapper)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var geoIPParser = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Date GeoIP Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#geoip-parser)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Description: "Name of the processor.", Type: schema.TypeString, Optional: true},
			"is_enabled": {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
			"sources":    {Description: "List of source attributes.", Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"target":     {Description: "Name of the parent attribute that contains all the extracted details from the sources.", Type: schema.TypeString, Required: true},
		},
	},
}

var grokParser = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Grok Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#grok-parser)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Description: "Name of the processor", Type: schema.TypeString, Optional: true},
			"is_enabled": {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
			"source":     {Description: "Name of the log attribute to parse.", Type: schema.TypeString, Required: true},
			"samples": {
				Description: "List of sample logs for this parser. It can save up to 5 samples. Each sample takes up to 5000 characters.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"grok": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"support_rules": {Description: "Support rules for your grok parser.", Type: schema.TypeString, Required: true},
						"match_rules":   {Description: "Match rules for your grok parser.", Type: schema.TypeString, Required: true},
					},
				},
			},
		},
	},
}

var lookupProcessor = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Lookup Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#lookup-processor)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Description: "Name of the processor", Type: schema.TypeString, Optional: true},
			"is_enabled": {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
			"source":     {Description: "Name of the source attribute used to do the lookup.", Type: schema.TypeString, Required: true},
			"target":     {Description: "Name of the attribute that contains the result of the lookup.", Type: schema.TypeString, Required: true},
			"lookup_table": {
				Description: "List of entries of the lookup table using `key,value` format.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"default_lookup": {Description: "Default lookup value to use if there is no entry in the lookup table for the value of the source attribute.", Type: schema.TypeString, Optional: true},
		},
	},
}

var messageRemapper = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Message Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-message-remapper)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var serviceRemapper = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Service Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#service-remapper)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var statusRemmaper = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Status Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-status-remapper)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var stringBuilderProcessor = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "String Builder Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#string-builder-processor)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":               {Description: "The name of the processor.", Type: schema.TypeString, Optional: true},
			"is_enabled":         {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
			"template":           {Description: "The formula with one or more attributes and raw text.", Type: schema.TypeString, Required: true},
			"target":             {Description: "The name of the attribute that contains the result of the template.", Type: schema.TypeString, Required: true},
			"is_replace_missing": {Description: "If it replaces all missing attributes of template by an empty string.", Type: schema.TypeBool, Optional: true},
		},
	},
}

var traceIDRemapper = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "Trace ID Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#trace-remapper)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var sourceRemapper = map[string]*schema.Schema{
	"name":       {Description: "Name of the processor.", Type: schema.TypeString, Optional: true},
	"is_enabled": {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
	"sources":    {Description: "List of source attributes.", Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
}

var urlParser = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "URL Parser Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#url-parser)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":                     {Description: "Name of the processor", Type: schema.TypeString, Optional: true},
			"is_enabled":               {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
			"sources":                  {Description: "List of source attributes.", Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"target":                   {Description: "Name of the parent attribute that contains all the extracted details from the sources.", Type: schema.TypeString, Required: true},
			"normalize_ending_slashes": {Description: "Normalize the ending slashes or not.", Type: schema.TypeBool, Optional: true},
		},
	},
}

var userAgentParser = &schema.Schema{
	Type:        schema.TypeList,
	MaxItems:    1,
	Description: "User-Agent Parser Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#user-agent-parser)",
	Optional:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Description: "Name of the processor", Type: schema.TypeString, Optional: true},
			"is_enabled": {Description: "If the processor is enabled or not.", Type: schema.TypeBool, Optional: true},
			"sources":    {Description: "List of source attributes.", Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"target":     {Description: "Name of the parent attribute that contains all the extracted details from the sources.", Type: schema.TypeString, Required: true},
			"is_encoded": {Description: "If the source attribute is URL encoded or not.", Type: schema.TypeBool, Optional: true},
		},
	},
}

func resourceDatadogLogsCustomPipeline() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatadogLogsPipelineCreate,
		UpdateContext: resourceDatadogLogsPipelineUpdate,
		ReadContext:   resourceDatadogLogsPipelineRead,
		DeleteContext: resourceDatadogLogsPipelineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/v1/logs-pipelines/) resource, which is used to create and manage Datadog logs custom pipelines. Each `datadog_logs_custom_pipeline` resource defines a complete pipeline. The order of the pipelines is maintained in a different resource: `datadog_logs_pipeline_order`. When creating a new pipeline, you need to **explicitly** add this pipeline to the `datadog_logs_pipeline_order` resource to track the pipeline. Similarly, when a pipeline needs to be destroyed, remove its references from the `datadog_logs_pipeline_order` resource.",
		Schema:      getPipelineSchema(false),
	}
}

func resourceDatadogLogsPipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	logCustomPipelineMutex.Lock()
	defer logCustomPipelineMutex.Unlock()

	ddPipeline, err := buildDatadogPipeline(d)
	if err != nil {
		return diag.FromErr(err)
	}
	createdPipeline, httpResponse, err := datadogClientV1.LogsPipelinesApi.CreateLogsPipeline(authV1, *ddPipeline)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "failed to create logs pipeline using Datadog API")
	}
	if err := utils.CheckForUnparsed(createdPipeline); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*createdPipeline.Id)
	return updateLogsCustomPipelineState(d, &createdPipeline)
}

func updateLogsCustomPipelineState(d *schema.ResourceData, pipeline *datadogV1.LogsPipeline) diag.Diagnostics {
	if err := d.Set("name", pipeline.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_enabled", pipeline.GetIsEnabled()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("filter", buildTerraformFilter(pipeline.Filter)); err != nil {
		return diag.FromErr(err)
	}
	tfProcessors, err := buildTerraformProcessors(pipeline.GetProcessors())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("processor", tfProcessors); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogLogsPipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	ddPipeline, httpresp, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 400 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "failed to get logs pipeline using Datadog API")
	}
	if err := utils.CheckForUnparsed(ddPipeline); err != nil {
		return diag.FromErr(err)
	}
	return updateLogsCustomPipelineState(d, &ddPipeline)
}

func resourceDatadogLogsPipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	logCustomPipelineMutex.Lock()
	defer logCustomPipelineMutex.Unlock()

	ddPipeline, err := buildDatadogPipeline(d)
	if err != nil {
		return diag.FromErr(err)
	}
	updatedPipeline, httpResponse, err := datadogClientV1.LogsPipelinesApi.UpdateLogsPipeline(authV1, d.Id(), *ddPipeline)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating logs pipeline")
	}
	if err := utils.CheckForUnparsed(updatedPipeline); err != nil {
		return diag.FromErr(err)
	}
	return updateLogsCustomPipelineState(d, &updatedPipeline)
}

func resourceDatadogLogsPipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	logCustomPipelineMutex.Lock()
	defer logCustomPipelineMutex.Unlock()

	if httpResponse, err := datadogClientV1.LogsPipelinesApi.DeleteLogsPipeline(authV1, d.Id()); err != nil {
		// API returns 400 when the specific pipeline id doesn't exist through DELETE request.
		if strings.Contains(err.Error(), "400 Bad Request") {
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting logs pipeline")
	}
	return nil
}

func buildTerraformProcessors(ddProcessors []datadogV1.LogsProcessor) ([]map[string]interface{}, error) {
	tfProcessors := make([]map[string]interface{}, len(ddProcessors))
	for i, ddProcessor := range ddProcessors {
		tfProcessor, err := buildTerraformProcessor(ddProcessor)
		if err != nil {
			return nil, err
		}
		tfProcessors[i] = tfProcessor
	}

	return tfProcessors, nil
}

func buildTerraformProcessor(ddProcessor datadogV1.LogsProcessor) (map[string]interface{}, error) {
	tfProcessor := make(map[string]interface{})
	var processorType string
	var err error
	if ddProcessor.LogsArithmeticProcessor != nil {
		tfProcessor = buildTerraformArithmeticProcessor(ddProcessor.LogsArithmeticProcessor)
		processorType = string(datadogV1.LOGSARITHMETICPROCESSORTYPE_ARITHMETIC_PROCESSOR)
	} else if ddProcessor.LogsAttributeRemapper != nil {
		tfProcessor = buildTerraformAttributeRemapper(ddProcessor.LogsAttributeRemapper)
		processorType = string(datadogV1.LOGSATTRIBUTEREMAPPERTYPE_ATTRIBUTE_REMAPPER)
	} else if ddProcessor.LogsCategoryProcessor != nil {
		tfProcessor = buildTerraformCategoryProcessor(ddProcessor.LogsCategoryProcessor)
		processorType = string(datadogV1.LOGSCATEGORYPROCESSORTYPE_CATEGORY_PROCESSOR)
	} else if ddProcessor.LogsDateRemapper != nil {
		tfProcessor = buildTerraformDateRemapper(ddProcessor.LogsDateRemapper)
		processorType = string(datadogV1.LOGSDATEREMAPPERTYPE_DATE_REMAPPER)
	} else if ddProcessor.LogsMessageRemapper != nil {
		tfProcessor = buildTerraformMessageRemapper(ddProcessor.LogsMessageRemapper)
		processorType = string(datadogV1.LOGSMESSAGEREMAPPERTYPE_MESSAGE_REMAPPER)
	} else if ddProcessor.LogsServiceRemapper != nil {
		tfProcessor = buildTerraformServiceRemapper(ddProcessor.LogsServiceRemapper)
		processorType = string(datadogV1.LOGSSERVICEREMAPPERTYPE_SERVICE_REMAPPER)
	} else if ddProcessor.LogsStatusRemapper != nil {
		tfProcessor = buildTerraformStatusRemapper(ddProcessor.LogsStatusRemapper)
		processorType = string(datadogV1.LOGSSTATUSREMAPPERTYPE_STATUS_REMAPPER)
	} else if ddProcessor.LogsTraceRemapper != nil {
		tfProcessor = buildTerraformTraceRemapper(ddProcessor.LogsTraceRemapper)
		processorType = string(datadogV1.LOGSTRACEREMAPPERTYPE_TRACE_ID_REMAPPER)
	} else if ddProcessor.LogsGeoIPParser != nil {
		tfProcessor = buildTerraformGeoIPParser(ddProcessor.LogsGeoIPParser)
		processorType = string(datadogV1.LOGSGEOIPPARSERTYPE_GEO_IP_PARSER)
	} else if ddProcessor.LogsGrokParser != nil {
		tfProcessor = buildTerraformGrokParser(ddProcessor.LogsGrokParser)
		processorType = string(datadogV1.LOGSGROKPARSERTYPE_GROK_PARSER)
	} else if ddProcessor.LogsLookupProcessor != nil {
		tfProcessor = buildTerraformLookupProcessor(ddProcessor.LogsLookupProcessor)
		processorType = string(datadogV1.LOGSLOOKUPPROCESSORTYPE_LOOKUP_PROCESSOR)
	} else if ddProcessor.LogsPipelineProcessor != nil {
		tfProcessor, err = buildTerraformNestedPipeline(ddProcessor.LogsPipelineProcessor)
		processorType = string(datadogV1.LOGSPIPELINEPROCESSORTYPE_PIPELINE)
	} else if ddProcessor.LogsStringBuilderProcessor != nil {
		tfProcessor = buildTerraformStringBuilderProcessor(ddProcessor.LogsStringBuilderProcessor)
		processorType = string(datadogV1.LOGSSTRINGBUILDERPROCESSORTYPE_STRING_BUILDER_PROCESSOR)
	} else if ddProcessor.LogsURLParser != nil {
		tfProcessor = buildTerraformURLParser(ddProcessor.LogsURLParser)
		processorType = string(datadogV1.LOGSURLPARSERTYPE_URL_PARSER)
	} else if ddProcessor.LogsUserAgentParser != nil {
		tfProcessor = buildTerraformUserAgentParser(ddProcessor.LogsUserAgentParser)
		processorType = string(datadogV1.LOGSUSERAGENTPARSERTYPE_USER_AGENT_PARSER)
	} else {
		err = fmt.Errorf("failed to support datadogV1 processor type, %s", ddProcessor.GetActualInstance())
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		ddProcessorTypes[processorType]: []map[string]interface{}{tfProcessor},
	}, nil
}

func buildTerraformUserAgentParser(ddUserAgent *datadogV1.LogsUserAgentParser) map[string]interface{} {
	return map[string]interface{}{
		"sources":    ddUserAgent.Sources,
		"target":     ddUserAgent.GetTarget(),
		"is_encoded": ddUserAgent.GetIsEncoded(),
		"name":       ddUserAgent.GetName(),
		"is_enabled": ddUserAgent.GetIsEnabled(),
	}
}

func buildTerraformURLParser(ddURL *datadogV1.LogsURLParser) map[string]interface{} {
	return map[string]interface{}{
		"sources":                  ddURL.Sources,
		"target":                   ddURL.GetTarget(),
		"normalize_ending_slashes": ddURL.GetNormalizeEndingSlashes(),
		"name":                     ddURL.GetName(),
		"is_enabled":               ddURL.GetIsEnabled(),
	}
}

func buildTerraformLookupProcessor(ddLookup *datadogV1.LogsLookupProcessor) map[string]interface{} {
	tfProcessor := map[string]interface{}{
		"source":       ddLookup.GetSource(),
		"target":       ddLookup.GetTarget(),
		"lookup_table": ddLookup.GetLookupTable(),
		"name":         ddLookup.GetName(),
		"is_enabled":   ddLookup.GetIsEnabled(),
	}

	if ddLookup.HasDefaultLookup() {
		tfProcessor["default_lookup"] = ddLookup.GetDefaultLookup()
	}

	return tfProcessor
}

func buildTerraformNestedPipeline(ddNested *datadogV1.LogsPipelineProcessor) (map[string]interface{}, error) {
	tfProcessors, err := buildTerraformProcessors(ddNested.GetProcessors())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"filter":     buildTerraformFilter(ddNested.Filter),
		"processor":  tfProcessors,
		"name":       ddNested.GetName(),
		"is_enabled": ddNested.GetIsEnabled(),
	}, nil
}

func buildTerraformStringBuilderProcessor(ddStringBuilder *datadogV1.LogsStringBuilderProcessor) map[string]interface{} {
	return map[string]interface{}{
		"template":           ddStringBuilder.GetTemplate(),
		"target":             ddStringBuilder.GetTarget(),
		"is_replace_missing": ddStringBuilder.GetIsReplaceMissing(),
		"name":               ddStringBuilder.GetName(),
		"is_enabled":         ddStringBuilder.GetIsEnabled(),
	}
}

func buildTerraformGeoIPParser(ddGeoIPParser *datadogV1.LogsGeoIPParser) map[string]interface{} {
	return map[string]interface{}{
		"sources":    ddGeoIPParser.GetSources(),
		"target":     ddGeoIPParser.GetTarget(),
		"name":       ddGeoIPParser.GetName(),
		"is_enabled": ddGeoIPParser.GetIsEnabled(),
	}
}

func buildTerraformGrokParser(ddGrok *datadogV1.LogsGrokParser) map[string]interface{} {
	return map[string]interface{}{
		"samples":    ddGrok.GetSamples(),
		"source":     ddGrok.GetSource(),
		"grok":       buildTerraformGrokRule(&ddGrok.Grok),
		"name":       ddGrok.GetName(),
		"is_enabled": ddGrok.GetIsEnabled(),
	}
}

func buildTerraformGrokRule(ddGrokRule *datadogV1.LogsGrokParserRules) []map[string]interface{} {
	tfGrokRule := map[string]interface{}{
		"support_rules": ddGrokRule.GetSupportRules(),
		"match_rules":   ddGrokRule.GetMatchRules(),
	}
	return []map[string]interface{}{tfGrokRule}
}

func buildTerraformMessageRemapper(remapper *datadogV1.LogsMessageRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.GetSources(),
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformDateRemapper(remapper *datadogV1.LogsDateRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.GetSources(),
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformServiceRemapper(remapper *datadogV1.LogsServiceRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.GetSources(),
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformStatusRemapper(remapper *datadogV1.LogsStatusRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.GetSources(),
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformTraceRemapper(remapper *datadogV1.LogsTraceRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.GetSources(),
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformCategoryProcessor(ddCategory *datadogV1.LogsCategoryProcessor) map[string]interface{} {
	return map[string]interface{}{
		"target":     ddCategory.GetTarget(),
		"category":   buildTerraformCategories(ddCategory.Categories),
		"name":       ddCategory.GetName(),
		"is_enabled": ddCategory.GetIsEnabled(),
	}
}

func buildTerraformCategories(ddCategories []datadogV1.LogsCategoryProcessorCategory) []map[string]interface{} {
	tfCategories := make([]map[string]interface{}, len(ddCategories))
	for i, ddCategory := range ddCategories {
		tfCategories[i] = map[string]interface{}{
			"name":   ddCategory.GetName(),
			"filter": buildTerraformFilter(ddCategory.Filter),
		}
	}
	return tfCategories
}

func buildTerraformAttributeRemapper(ddAttribute *datadogV1.LogsAttributeRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":              ddAttribute.Sources,
		"source_type":          ddAttribute.GetSourceType(),
		"target":               ddAttribute.GetTarget(),
		"target_type":          ddAttribute.GetTargetType(),
		"target_format":        ddAttribute.GetTargetFormat(),
		"preserve_source":      ddAttribute.GetPreserveSource(),
		"override_on_conflict": ddAttribute.GetOverrideOnConflict(),
		"name":                 ddAttribute.GetName(),
		"is_enabled":           ddAttribute.GetIsEnabled(),
	}
}

func buildTerraformArithmeticProcessor(ddArithmetic *datadogV1.LogsArithmeticProcessor) map[string]interface{} {

	return map[string]interface{}{
		"target":             ddArithmetic.GetTarget(),
		"is_replace_missing": ddArithmetic.GetIsReplaceMissing(),
		"expression":         ddArithmetic.GetExpression(),
		"name":               ddArithmetic.GetName(),
		"is_enabled":         ddArithmetic.GetIsEnabled(),
	}
}

func buildTerraformFilter(ddFilter *datadogV1.LogsFilter) []map[string]interface{} {
	tfFilter := map[string]interface{}{
		"query": ddFilter.GetQuery(),
	}
	return []map[string]interface{}{tfFilter}
}

func buildDatadogPipeline(d *schema.ResourceData) (*datadogV1.LogsPipeline, error) {
	var ddPipeline datadogV1.LogsPipeline
	ddPipeline.SetName(d.Get("name").(string))
	ddPipeline.SetIsEnabled(d.Get("is_enabled").(bool))
	if tfFilter := d.Get("filter").([]interface{}); len(tfFilter) > 0 {
		filter, ok := tfFilter[0].(map[string]interface{})
		if !ok {
			filter = make(map[string]interface{})
		}
		ddPipeline.SetFilter(buildDatadogFilter(filter))
	}
	ddProcessors, err := buildDatadogProcessors(d.Get("processor").([]interface{}))
	if err != nil {
		return nil, err
	}
	ddPipeline.SetProcessors(*ddProcessors)
	return &ddPipeline, nil
}

func buildDatadogProcessors(tfProcessors []interface{}) (*[]datadogV1.LogsProcessor, error) {
	ddProcessors := make([]datadogV1.LogsProcessor, len(tfProcessors))
	for i, tfProcessor := range tfProcessors {
		for tfProcessorType, ddProcessorType := range tfProcessorTypes {
			tfProcessorMap := tfProcessor.(map[string]interface{})
			if tfProcessorDetails, exists := tfProcessorMap[tfProcessorType].([]interface{}); exists && len(tfProcessorDetails) > 0 {
				ddProcessor, err := buildDatadogProcessor(ddProcessorType, tfProcessorDetails[0].(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				ddProcessors[i] = ddProcessor
				break
			}
		}
	}
	return &ddProcessors, nil
}

func buildDatadogProcessor(ddProcessorType string, tfProcessor map[string]interface{}) (datadogV1.LogsProcessor, error) {
	var ddProcessor = datadogV1.LogsProcessor{}
	var err error
	switch ddProcessorType {
	case string(datadogV1.LOGSARITHMETICPROCESSORTYPE_ARITHMETIC_PROCESSOR):
		ddProcessor = datadogV1.LogsArithmeticProcessorAsLogsProcessor(buildDatadogArithmeticProcessor(tfProcessor))
	case string(datadogV1.LOGSATTRIBUTEREMAPPERTYPE_ATTRIBUTE_REMAPPER):
		ddProcessor = datadogV1.LogsAttributeRemapperAsLogsProcessor(buildDatadogAttributeRemapper(tfProcessor))
	case string(datadogV1.LOGSCATEGORYPROCESSORTYPE_CATEGORY_PROCESSOR):
		ddProcessor = datadogV1.LogsCategoryProcessorAsLogsProcessor(buildDatadogCategoryProcessor(tfProcessor))
	case string(datadogV1.LOGSDATEREMAPPERTYPE_DATE_REMAPPER):
		ddProcessor = datadogV1.LogsDateRemapperAsLogsProcessor(buildDatadogDateRemapperProcessor(tfProcessor))
	case string(datadogV1.LOGSMESSAGEREMAPPERTYPE_MESSAGE_REMAPPER):
		ddProcessor = datadogV1.LogsMessageRemapperAsLogsProcessor(buildDatadogMessageRemapper(tfProcessor))
	case string(datadogV1.LOGSSERVICEREMAPPERTYPE_SERVICE_REMAPPER):
		ddProcessor = datadogV1.LogsServiceRemapperAsLogsProcessor(buildDatadogServiceRemapper(tfProcessor))
	case string(datadogV1.LOGSSTATUSREMAPPERTYPE_STATUS_REMAPPER):
		ddProcessor = datadogV1.LogsStatusRemapperAsLogsProcessor(buildDatadogStatusRemapper(tfProcessor))
	case string(datadogV1.LOGSTRACEREMAPPERTYPE_TRACE_ID_REMAPPER):
		ddProcessor = datadogV1.LogsTraceRemapperAsLogsProcessor(buildDatadogTraceRemapper(tfProcessor))
	case string(datadogV1.LOGSGEOIPPARSERTYPE_GEO_IP_PARSER):
		ddProcessor = datadogV1.LogsGeoIPParserAsLogsProcessor(buildDatadogGeoIPParser(tfProcessor))
	case string(datadogV1.LOGSGROKPARSERTYPE_GROK_PARSER):
		ddProcessor = datadogV1.LogsGrokParserAsLogsProcessor(buildDatadogGrokParser(tfProcessor))
	case string(datadogV1.LOGSLOOKUPPROCESSORTYPE_LOOKUP_PROCESSOR):
		ddProcessor = datadogV1.LogsLookupProcessorAsLogsProcessor(buildDatadogLookupProcessor(tfProcessor))
	case string(datadogV1.LOGSPIPELINEPROCESSORTYPE_PIPELINE):
		ddNestedPipeline, err := buildDatadogNestedPipeline(tfProcessor)
		if err != nil {
			return ddProcessor, err
		}
		ddProcessor = datadogV1.LogsPipelineProcessorAsLogsProcessor(ddNestedPipeline)
	case string(datadogV1.LOGSSTRINGBUILDERPROCESSORTYPE_STRING_BUILDER_PROCESSOR):
		ddStringBuilderProcessor, err := buildDatadogStringBuilderProcessor(tfProcessor)
		if err != nil {
			return ddProcessor, err
		}
		ddProcessor = datadogV1.LogsStringBuilderProcessorAsLogsProcessor(ddStringBuilderProcessor)
	case string(datadogV1.LOGSURLPARSERTYPE_URL_PARSER):
		ddProcessor = datadogV1.LogsURLParserAsLogsProcessor(buildDatadogURLParser(tfProcessor))
	case string(datadogV1.LOGSUSERAGENTPARSERTYPE_USER_AGENT_PARSER):
		ddProcessor = datadogV1.LogsUserAgentParserAsLogsProcessor(buildDatadogUserAgentParser(tfProcessor))
	default:
		err = fmt.Errorf("failed to recoginize processor type: %s", ddProcessorType)
	}

	return ddProcessor, err
}

func buildDatadogURLParser(tfProcessor map[string]interface{}) *datadogV1.LogsURLParser {
	ddURLParser := datadogV1.NewLogsURLParserWithDefaults()
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddURLParser.Sources = ddSources
	}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddURLParser.SetTarget(tfTarget)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddURLParser.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddURLParser.SetIsEnabled(tfIsEnabled)
	}
	if tfNormalizeEndingSlashes, exists := tfProcessor["normalize_ending_slashes"].(bool); exists {
		ddURLParser.SetNormalizeEndingSlashes(tfNormalizeEndingSlashes)
	}
	return ddURLParser
}

func buildDatadogUserAgentParser(tfProcessor map[string]interface{}) *datadogV1.LogsUserAgentParser {
	ddUserAgentParser := datadogV1.NewLogsUserAgentParserWithDefaults()
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddUserAgentParser.Sources = ddSources
	}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddUserAgentParser.SetTarget(tfTarget)
	}
	if tfIsEncoded, exists := tfProcessor["is_encoded"].(bool); exists {
		ddUserAgentParser.SetIsEncoded(tfIsEncoded)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddUserAgentParser.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddUserAgentParser.SetIsEnabled(tfIsEnabled)
	}
	return ddUserAgentParser
}

func buildDatadogLookupProcessor(tfProcessor map[string]interface{}) *datadogV1.LogsLookupProcessor {
	ddLookupProcessor := datadogV1.NewLogsLookupProcessorWithDefaults()
	if tfSource, exists := tfProcessor["source"].(string); exists {
		ddLookupProcessor.SetSource(tfSource)
	}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddLookupProcessor.SetTarget(tfTarget)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddLookupProcessor.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddLookupProcessor.SetIsEnabled(tfIsEnabled)
	}
	if tfLookupTable, exists := tfProcessor["lookup_table"].([]interface{}); exists && len(tfLookupTable) > 0 {
		ddLookupTable := make([]string, len(tfLookupTable))
		for i, tfLookupLine := range tfLookupTable {
			ddLookupTable[i] = tfLookupLine.(string)
		}
		ddLookupProcessor.SetLookupTable(ddLookupTable)
	}
	if tfDefaultLookup, exists := tfProcessor["default_lookup"].(string); exists && len(tfDefaultLookup) > 0 {
		ddLookupProcessor.SetDefaultLookup(tfDefaultLookup)
	}
	return ddLookupProcessor
}

func buildDatadogNestedPipeline(tfProcessor map[string]interface{}) (*datadogV1.LogsPipelineProcessor, error) {
	ddNestedPipeline := datadogV1.NewLogsPipelineProcessorWithDefaults()
	if tfFilter, exist := tfProcessor["filter"].([]interface{}); exist && len(tfFilter) > 0 {
		ddNestedPipeline.SetFilter(buildDatadogFilter(tfFilter[0].(map[string]interface{})))
	}
	if tfProcessors, exists := tfProcessor["processor"].([]interface{}); exists && len(tfProcessors) > 0 {
		ddProcessors, err := buildDatadogProcessors(tfProcessors)
		if err != nil {
			return ddNestedPipeline, err
		}
		ddNestedPipeline.Processors = ddProcessors
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddNestedPipeline.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddNestedPipeline.SetIsEnabled(tfIsEnabled)
	}
	return ddNestedPipeline, nil
}

func buildDatadogStringBuilderProcessor(tfProcessor map[string]interface{}) (*datadogV1.LogsStringBuilderProcessor, error) {
	ddStringBuilder := datadogV1.NewLogsStringBuilderProcessorWithDefaults()
	if tfTemplate, exists := tfProcessor["template"].(string); exists {
		ddStringBuilder.SetTemplate(tfTemplate)
	}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddStringBuilder.SetTarget(tfTarget)
	}
	if tfReplaceMissing, exists := tfProcessor["is_replace_missing"].(bool); exists {
		ddStringBuilder.SetIsReplaceMissing(tfReplaceMissing)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddStringBuilder.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddStringBuilder.SetIsEnabled(tfIsEnabled)
	}
	return ddStringBuilder, nil
}

func buildDatadogGeoIPParser(tfProcessor map[string]interface{}) *datadogV1.LogsGeoIPParser {
	ddGeoIPParser := datadogV1.NewLogsGeoIPParserWithDefaults()
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddGeoIPParser.SetTarget(tfTarget)
	}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddGeoIPParser.SetSources(ddSources)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddGeoIPParser.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddGeoIPParser.SetIsEnabled(tfIsEnabled)
	}
	return ddGeoIPParser
}

func buildDatadogGrokParser(tfProcessor map[string]interface{}) *datadogV1.LogsGrokParser {
	ddGrokParser := datadogV1.NewLogsGrokParserWithDefaults()
	if tfSource, exists := tfProcessor["source"].(string); exists {
		ddGrokParser.SetSource(tfSource)
	}
	if tfSamples, exists := tfProcessor["samples"].([]interface{}); exists && len(tfSamples) > 0 {
		ddSamples := make([]string, len(tfSamples))
		for i, tfSample := range tfSamples {
			ddSamples[i] = tfSample.(string)
		}
		ddGrokParser.SetSamples(ddSamples)
	}
	if tfGrok, exists := tfProcessor["grok"].([]interface{}); exists && len(tfGrok) > 0 {
		ddGrok := datadogV1.LogsGrokParserRules{}
		tfGrokRule := tfGrok[0].(map[string]interface{})
		if tfSupportRule, exist := tfGrokRule["support_rules"].(string); exist {
			ddGrok.SetSupportRules(tfSupportRule)
		}
		if tfMatchRule, exist := tfGrokRule["match_rules"].(string); exist {
			ddGrok.SetMatchRules(tfMatchRule)
		}
		ddGrokParser.SetGrok(ddGrok)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddGrokParser.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddGrokParser.SetIsEnabled(tfIsEnabled)
	}
	return ddGrokParser
}

func buildDatadogMessageRemapper(tfProcessor map[string]interface{}) *datadogV1.LogsMessageRemapper {
	ddRemapper := datadogV1.NewLogsMessageRemapperWithDefaults()
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddRemapper.SetSources(ddSources)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddRemapper.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddRemapper.SetIsEnabled(tfIsEnabled)
	}
	return ddRemapper
}

func buildDatadogServiceRemapper(tfProcessor map[string]interface{}) *datadogV1.LogsServiceRemapper {
	ddRemapper := datadogV1.NewLogsServiceRemapperWithDefaults()
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddRemapper.Sources = ddSources
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddRemapper.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddRemapper.SetIsEnabled(tfIsEnabled)
	}
	return ddRemapper
}

func buildDatadogStatusRemapper(tfProcessor map[string]interface{}) *datadogV1.LogsStatusRemapper {
	ddRemapper := datadogV1.NewLogsStatusRemapperWithDefaults()
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddRemapper.Sources = ddSources
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddRemapper.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddRemapper.SetIsEnabled(tfIsEnabled)
	}
	return ddRemapper
}

func buildDatadogTraceRemapper(tfProcessor map[string]interface{}) *datadogV1.LogsTraceRemapper {
	ddRemapper := datadogV1.NewLogsTraceRemapperWithDefaults()
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddRemapper.Sources = &ddSources
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddRemapper.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddRemapper.SetIsEnabled(tfIsEnabled)
	}
	return ddRemapper
}

func buildDatadogCategoryProcessor(tfProcessor map[string]interface{}) *datadogV1.LogsCategoryProcessor {
	ddCategory := datadogV1.NewLogsCategoryProcessorWithDefaults()
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddCategory.SetTarget(tfTarget)
	}
	if tfCategories, exists := tfProcessor["category"].([]interface{}); exists {
		ddCategories := make([]datadogV1.LogsCategoryProcessorCategory, len(tfCategories))
		for i, tfC := range tfCategories {
			tfCategory := tfC.(map[string]interface{})
			ddCategory := datadogV1.LogsCategoryProcessorCategory{}
			if tfName, exist := tfCategory["name"].(string); exist {
				ddCategory.SetName(tfName)
			}
			if tfFilter, exist := tfCategory["filter"].([]interface{}); exist && len(tfFilter) > 0 {
				ddCategory.SetFilter(buildDatadogFilter(tfFilter[0].(map[string]interface{})))
			}

			ddCategories[i] = ddCategory
		}
		ddCategory.SetCategories(ddCategories)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddCategory.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddCategory.SetIsEnabled(tfIsEnabled)
	}
	return ddCategory
}

func buildDatadogAttributeRemapper(tfProcessor map[string]interface{}) *datadogV1.LogsAttributeRemapper {
	ddAttribute := datadogV1.NewLogsAttributeRemapperWithDefaults()

	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddAttribute.Sources = ddSources
	}
	if tfSourceType, exists := tfProcessor["source_type"].(string); exists {
		ddAttribute.SetSourceType(tfSourceType)
	}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddAttribute.SetTarget(tfTarget)
	}
	if tfTargetType, exists := tfProcessor["target_type"].(string); exists {
		ddAttribute.SetTargetType(tfTargetType)
	}
	if tfTargetFormat, exists := tfProcessor["target_format"].(string); exists && (tfTargetFormat != "") {
		ddAttribute.SetTargetFormat(datadogV1.TargetFormatType(tfTargetFormat))
	}
	if tfPreserveSource, exists := tfProcessor["preserve_source"].(bool); exists {
		ddAttribute.SetPreserveSource(tfPreserveSource)
	}
	if tfOverrideOnConflict, exists := tfProcessor["override_on_conflict"].(bool); exists {
		ddAttribute.SetOverrideOnConflict(tfOverrideOnConflict)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddAttribute.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddAttribute.SetIsEnabled(tfIsEnabled)
	}
	return ddAttribute
}

func buildDatadogSources(tfProcessor map[string]interface{}) []string {
	if tfSources, exists := tfProcessor["sources"].([]interface{}); exists && len(tfSources) > 0 {
		ddSources := make([]string, len(tfSources))
		for i, tfSource := range tfSources {
			ddSources[i] = tfSource.(string)
		}
		return ddSources
	}
	return nil
}

func buildDatadogArithmeticProcessor(tfProcessor map[string]interface{}) *datadogV1.LogsArithmeticProcessor {
	ddArithmetic := datadogV1.NewLogsArithmeticProcessorWithDefaults()
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddArithmetic.SetTarget(tfTarget)
	}
	if tfExpression, exists := tfProcessor["expression"].(string); exists {
		ddArithmetic.SetExpression(tfExpression)
	}
	if tfIsReplaceMissing, exists := tfProcessor["is_replace_missing"].(bool); exists {
		ddArithmetic.SetIsReplaceMissing(tfIsReplaceMissing)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddArithmetic.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddArithmetic.SetIsEnabled(tfIsEnabled)
	}
	return ddArithmetic
}

func buildDatadogDateRemapperProcessor(tfProcessor map[string]interface{}) *datadogV1.LogsDateRemapper {
	ddDate := datadogV1.NewLogsDateRemapperWithDefaults()
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddDate.SetSources(ddSources)
	}

	if tfName, exists := tfProcessor["name"].(string); exists {
		ddDate.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddDate.SetIsEnabled(tfIsEnabled)
	}
	return ddDate
}

func buildDatadogFilter(tfFilter map[string]interface{}) datadogV1.LogsFilter {
	ddFilter := datadogV1.LogsFilter{}
	var query string
	if tfQuery, exists := tfFilter["query"].(string); exists {
		query = tfQuery
	}
	ddFilter.SetQuery(query)
	return ddFilter
}

func getPipelineSchema(isNested bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":       {Type: schema.TypeString, Required: true},
		"is_enabled": {Type: schema.TypeBool, Optional: true},
		"filter": {
			Type:     schema.TypeList,
			Required: true,
			Elem:     getFilterSchema(),
		},
		"processor": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getProcessorSchema(isNested),
			},
		},
	}
}

func getProcessorSchema(isNested bool) map[string]*schema.Schema {
	var processorsSchema = make(map[string]*schema.Schema)
	if !isNested {
		processorsSchema[tfNestedPipelineProcessor] = &schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getPipelineSchema(!isNested),
			},
		}
	}
	for tfProcessorType, processor := range tfProcessors {
		processorsSchema[tfProcessorType] = processor
	}
	return processorsSchema
}

func getFilterSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"query": {
				Description: "Filter criteria of the category.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
