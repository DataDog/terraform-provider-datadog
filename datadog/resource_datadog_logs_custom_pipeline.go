package datadog

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

const (
	tfArithmeticProcessor        = "arithmetic_processor"
	tfAttributeRemapperProcessor = "attribute_remapper"
	tfCategoryProcessor          = "category_processor"
	tfDateRemapperProcessor      = "date_remapper"
	tfGeoIPParserProcessor       = "geo_ip_parser"
	tfGrokParserProcessor        = "grok_parser"
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
	tfArithmeticProcessor:        datadog.NewLogsArithmeticProcessorWithDefaults().GetType(),
	tfAttributeRemapperProcessor: datadog.NewLogsRemapperWithDefaults().GetType(),
	tfCategoryProcessor:          datadog.NewLogsCategoryProcessorWithDefaults().GetType(),
	tfDateRemapperProcessor:      datadog.NewLogsDateRemapperWithDefaults().GetType(),
	tfGeoIPParserProcessor:       datadog.NewLogsGeoIPParserWithDefaults().GetType(),
	tfGrokParserProcessor:        datadog.NewLogsGrokParserWithDefaults().GetType(),
	tfMessageRemapperProcessor:   datadog.NewLogsMessageRemapperWithDefaults().GetType(),
	tfNestedPipelineProcessor:    "pipeline",
	tfServiceRemapperProcessor:   datadog.NewLogsServiceRemapperWithDefaults().GetType(),
	tfStatusRemapperProcessor:    datadog.NewLogsStatusRemapperWithDefaults().GetType(),
	tfStringBuilderProcessor:     datadog.NewLogsStringBuilderProcessorWithDefaults().GetType(),
	tfTraceIDRemapperProcessor:   datadog.NewLogsTraceRemapperWithDefaults().GetType(),
	tfURLParserProcessor:         datadog.NewLogsURLParserWithDefaults().GetType(),
	tfUserAgentParserProcessor:   datadog.NewLogsUserAgentParserWithDefaults().GetType(),
}

var tfProcessors = map[string]*schema.Schema{
	tfArithmeticProcessor:        arithmeticProcessor,
	tfAttributeRemapperProcessor: attributeRemapper,
	tfCategoryProcessor:          categoryProcessor,
	tfDateRemapperProcessor:      dateRemapper,
	tfGeoIPParserProcessor:       geoIPParser,
	tfGrokParserProcessor:        grokParser,
	tfMessageRemapperProcessor:   messageRemapper,
	tfServiceRemapperProcessor:   serviceRemapper,
	tfStatusRemapperProcessor:    statusRemmaper,
	tfStringBuilderProcessor:     stringBuilderProcessor,
	tfTraceIDRemapperProcessor:   traceIDRemapper,
	tfURLParserProcessor:         urlParser,
	tfUserAgentParserProcessor:   userAgentParser,
}

var ddProcessorTypes = map[string]string{
	datadog.NewLogsArithmeticProcessorWithDefaults().GetType():    	tfArithmeticProcessor,
	datadog.NewLogsRemapperWithDefaults().GetType():      			tfAttributeRemapperProcessor,
	datadog.NewLogsCategoryProcessorWithDefaults().GetType():      	tfCategoryProcessor,
	datadog.NewLogsDateRemapperWithDefaults().GetType():           	tfDateRemapperProcessor,
	datadog.NewLogsGeoIPParserWithDefaults().GetType():            	tfGeoIPParserProcessor,
	datadog.NewLogsGrokParserWithDefaults().GetType():             	tfGrokParserProcessor,
	datadog.NewLogsMessageRemapperWithDefaults().GetType():        	tfMessageRemapperProcessor,
	"pipeline":         											tfNestedPipelineProcessor,
	datadog.NewLogsServiceRemapperWithDefaults().GetType():        	tfServiceRemapperProcessor,
	datadog.NewLogsStatusRemapperWithDefaults().GetType():         	tfStatusRemapperProcessor,
	datadog.NewLogsStringBuilderProcessorWithDefaults().GetType(): 	tfStringBuilderProcessor,
	datadog.NewLogsTraceRemapperWithDefaults().GetType():        	tfTraceIDRemapperProcessor,
	datadog.NewLogsURLParserWithDefaults().GetType():              	tfURLParserProcessor,
	datadog.NewLogsUserAgentParserWithDefaults().GetType():        	tfUserAgentParserProcessor,
}

var arithmeticProcessor = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":               {Type: schema.TypeString, Optional: true},
			"is_enabled":         {Type: schema.TypeBool, Optional: true},
			"expression":         {Type: schema.TypeString, Required: true},
			"target":             {Type: schema.TypeString, Required: true},
			"is_replace_missing": {Type: schema.TypeBool, Optional: true},
		},
	},
}

var attributeRemapper = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":                 {Type: schema.TypeString, Optional: true},
			"is_enabled":           {Type: schema.TypeBool, Optional: true},
			"sources":              {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"source_type":          {Type: schema.TypeString, Required: true},
			"target":               {Type: schema.TypeString, Required: true},
			"target_type":          {Type: schema.TypeString, Required: true},
			"preserve_source":      {Type: schema.TypeBool, Optional: true},
			"override_on_conflict": {Type: schema.TypeBool, Optional: true},
		},
	},
}

var categoryProcessor = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Type: schema.TypeString, Optional: true},
			"is_enabled": {Type: schema.TypeBool, Optional: true},
			"target":     {Type: schema.TypeString, Required: true},
			"category": {Type: schema.TypeList, Required: true, Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"filter": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"query": {Type: schema.TypeString, Required: true},
							},
						},
					},
					"name": {Type: schema.TypeString, Required: true},
				},
			}},
		},
	},
}

var dateRemapper = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var geoIPParser = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Type: schema.TypeString, Optional: true},
			"is_enabled": {Type: schema.TypeBool, Optional: true},
			"sources":    {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"target":     {Type: schema.TypeString, Required: true},
		},
	},
}

var grokParser = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Type: schema.TypeString, Optional: true},
			"is_enabled": {Type: schema.TypeBool, Optional: true},
			"source":     {Type: schema.TypeString, Required: true},
			"samples": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"grok": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"support_rules": {Type: schema.TypeString, Required: true},
						"match_rules":   {Type: schema.TypeString, Required: true},
					},
				},
			},
		},
	},
}

var messageRemapper = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var serviceRemapper = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var statusRemmaper = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var stringBuilderProcessor = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":               {Type: schema.TypeString, Optional: true},
			"is_enabled":         {Type: schema.TypeBool, Optional: true},
			"template":           {Type: schema.TypeString, Required: true},
			"target":             {Type: schema.TypeString, Required: true},
			"is_replace_missing": {Type: schema.TypeBool, Optional: true},
		},
	},
}

var traceIDRemapper = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: sourceRemapper,
	},
}

var sourceRemapper = map[string]*schema.Schema{
	"name":       {Type: schema.TypeString, Optional: true},
	"is_enabled": {Type: schema.TypeBool, Optional: true},
	"sources":    {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
}

var urlParser = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":                     {Type: schema.TypeString, Optional: true},
			"is_enabled":               {Type: schema.TypeBool, Optional: true},
			"sources":                  {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"target":                   {Type: schema.TypeString, Required: true},
			"normalize_ending_slashes": {Type: schema.TypeBool, Optional: true},
		},
	},
}

var userAgentParser = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Type: schema.TypeString, Optional: true},
			"is_enabled": {Type: schema.TypeBool, Optional: true},
			"sources":    {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"target":     {Type: schema.TypeString, Required: true},
			"is_encoded": {Type: schema.TypeBool, Optional: true},
		},
	},
}

func resourceDatadogLogsCustomPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogLogsPipelineCreate,
		Update: resourceDatadogLogsPipelineUpdate,
		Read:   resourceDatadogLogsPipelineRead,
		Delete: resourceDatadogLogsPipelineDelete,
		Exists: resourceDatadogLogsPipelineExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: getPipelineSchema(false),
	}
}

func resourceDatadogLogsPipelineCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	ddPipeline, err := buildDatadogPipeline(d)
	if err != nil {
		return err
	}
	createdPipeline, _, err := client.LogsPipelinesApi.CreateLogsPipeline(auth).Body(*ddPipeline).Execute()
	if err != nil {
		return translateClientError(err,"failed to create logs pipeline using Datadog API")
	}
	d.SetId(*createdPipeline.Id)
	return resourceDatadogLogsPipelineRead(d, meta)
}

func resourceDatadogLogsPipelineRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	ddPipeline, _, err := client.LogsPipelinesApi.GetLogsPipeline(auth, d.Id()).Execute()
	if err != nil {
		return translateClientError(err,"failed to get logs pipeline using Datadog API")
	}
	if err = d.Set("name", ddPipeline.GetName()); err != nil {
		return err
	}
	if err = d.Set("is_enabled", ddPipeline.GetIsEnabled()); err != nil {
		return err
	}
	if err := d.Set("filter", buildTerraformFilter(ddPipeline.Filter)); err != nil {
		return err
	}
	tfProcessors, err := buildTerraformProcessors(ddPipeline.GetProcessors())
	if err != nil {
		return err
	}
	if err := d.Set("processor", tfProcessors); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	ddPipeline, err := buildDatadogPipeline(d)
	if err != nil {
		return err
	}
	if _, _, err := client.LogsPipelinesApi.UpdateLogsPipeline(auth, d.Id()).Body(*ddPipeline).Execute(); err != nil {
		return translateClientError(err,"error updating logs pipeline")
	}
	return resourceDatadogLogsPipelineRead(d, meta)
}

func resourceDatadogLogsPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	if  _, err := client.LogsPipelinesApi.DeleteLogsPipeline(auth, d.Id()).Execute(); err != nil {
		// API returns 400 when the specific pipeline id doesn't exist through DELETE request.
		if strings.Contains(err.Error(), "400 Bad Request") {
			return nil
		}
		return translateClientError(err,"error deleting logs pipeline")
	}
	return nil
}

func resourceDatadogLogsPipelineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	if _, _, err := client.LogsPipelinesApi.GetLogsPipeline(auth, d.Id()).Execute(); err != nil {
		// API returns 400 when the specific pipeline id doesn't exist through GET request.
		if strings.Contains(err.Error(), "400 Bad Request") {
			return false, nil
		}
		return false, translateClientError(err,"error getting logs pipeline")
	}
	return true, nil
}

func buildTerraformProcessors(ddProcessors []datadog.LogsProcessor) ([]map[string]interface{}, error) {
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

func buildTerraformProcessor(ddProcessor datadog.LogsProcessor) (map[string]interface{}, error) {
	tfProcessor := make(map[string]interface{})
	var err error
	switch ddProcessor.LogsProcessorInterface.GetType() {
	case datadog.NewLogsArithmeticProcessorWithDefaults().GetType():
		logsArithmeticProcessor := ddProcessor.LogsProcessorInterface.(*datadog.LogsArithmeticProcessor)
		tfProcessor = buildTerraformArithmeticProcessor(logsArithmeticProcessor)
	case datadog.NewLogsRemapperWithDefaults().GetType():
		logsAttributeRemapper := ddProcessor.LogsProcessorInterface.(*datadog.LogsRemapper)
		tfProcessor = buildTerraformAttributeRemapper(logsAttributeRemapper)
	case datadog.NewLogsCategoryProcessorWithDefaults().GetType():
		logsCategoryProcessor := ddProcessor.LogsProcessorInterface.(*datadog.LogsCategoryProcessor)
		tfProcessor = buildTerraformCategoryProcessor(logsCategoryProcessor)
	case datadog.NewLogsDateRemapperWithDefaults().GetType():
		logsDateRemapper := ddProcessor.LogsProcessorInterface.(*datadog.LogsDateRemapper)
		tfProcessor = buildTerraformDateRemapper(logsDateRemapper)
	case datadog.NewLogsMessageRemapperWithDefaults().GetType():
		logsMessageRemapper := ddProcessor.LogsProcessorInterface.(*datadog.LogsMessageRemapper)
		tfProcessor = buildTerraformMessageRemapper(logsMessageRemapper)
	case datadog.NewLogsServiceRemapperWithDefaults().GetType():
		logsServiceRemapper := ddProcessor.LogsProcessorInterface.(*datadog.LogsServiceRemapper)
		tfProcessor = buildTerraformServiceRemapper(logsServiceRemapper)
	case datadog.NewLogsStatusRemapperWithDefaults().GetType():
		logsStatusRemapper := ddProcessor.LogsProcessorInterface.(*datadog.LogsStatusRemapper)
		tfProcessor = buildTerraformStatusRemapper(logsStatusRemapper)
	case datadog.NewLogsTraceRemapperWithDefaults().GetType():
		logsTraceRemapper := ddProcessor.LogsProcessorInterface.(*datadog.LogsTraceRemapper)
		tfProcessor = buildTerraformTraceRemapper(logsTraceRemapper)
	case datadog.NewLogsGeoIPParserWithDefaults().GetType():
		logsGeoIPParser := ddProcessor.LogsProcessorInterface.(*datadog.LogsGeoIPParser)
		tfProcessor = buildTerraformGeoIPParser(logsGeoIPParser)
	case datadog.NewLogsGrokParserWithDefaults().GetType():
		logsGrokParser := ddProcessor.LogsProcessorInterface.(*datadog.LogsGrokParser)
		tfProcessor = buildTerraformGrokParser(logsGrokParser)
	case "pipeline":
		logsPipeline := ddProcessor.LogsProcessorInterface.(*datadog.LogsPipeline)
		tfProcessor, err = buildTerraformNestedPipeline(logsPipeline)
	case datadog.NewLogsStringBuilderProcessorWithDefaults().GetType():
		logsStringBuilderProcessor := ddProcessor.LogsProcessorInterface.(*datadog.LogsStringBuilderProcessor)
		tfProcessor = buildTerraformStringBuilderProcessor(logsStringBuilderProcessor)
	case datadog.NewLogsURLParserWithDefaults().GetType():
		logsURLParser := ddProcessor.LogsProcessorInterface.(*datadog.LogsURLParser)
		tfProcessor = buildTerraformURLParser(logsURLParser)
	case datadog.NewLogsUserAgentParserWithDefaults().GetType():
		logsUserAgentParser := ddProcessor.LogsProcessorInterface.(*datadog.LogsUserAgentParser)
		tfProcessor = buildTerraformUserAgentParser(logsUserAgentParser)
	default:
		err = fmt.Errorf("failed to support datadog processor type, %s", ddProcessor.LogsProcessorInterface.GetType())
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		ddProcessorTypes[ddProcessor.LogsProcessorInterface.GetType()]: []map[string]interface{}{tfProcessor},
	}, nil
}

func buildTerraformUserAgentParser(ddUserAgent *datadog.LogsUserAgentParser) map[string]interface{} {
	return map[string]interface{}{
		"sources":    	ddUserAgent.Sources,
		"target":     	ddUserAgent.GetTarget(),
		"is_encoded": 	ddUserAgent.GetIsEncoded(),
		"name":			ddUserAgent.GetName(),
		"is_enabled":	ddUserAgent.GetIsEnabled(),
	}
}

func buildTerraformURLParser(ddURL *datadog.LogsURLParser) map[string]interface{} {
	return map[string]interface{}{
		"sources":                  ddURL.Sources,
		"target":                   ddURL.GetTarget(),
		"normalize_ending_slashes": ddURL.GetNormalizeEndingSlashes(),
		"name":						ddURL.GetName(),
		"is_enabled":				ddURL.GetIsEnabled(),
	}
}

func buildTerraformNestedPipeline(ddNested *datadog.LogsPipeline) (map[string]interface{}, error) {
	tfProcessors, err := buildTerraformProcessors(ddNested.GetProcessors())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"filter":    	buildTerraformFilter(ddNested.Filter),
		"processor": 	tfProcessors,
		"name":			ddNested.GetName(),
		"is_enabled":	ddNested.GetIsEnabled(),
	}, nil
}

func buildTerraformStringBuilderProcessor(ddStringBuilder *datadog.LogsStringBuilderProcessor) map[string]interface{} {
	return map[string]interface{}{
		"template":           	ddStringBuilder.GetTemplate(),
		"target":             	ddStringBuilder.GetTarget(),
		"is_replace_missing": 	ddStringBuilder.GetIsReplaceMissing(),
		"name":					ddStringBuilder.GetName(),
		"is_enabled":			ddStringBuilder.GetIsEnabled(),
	}
}

func buildTerraformGeoIPParser(ddGeoIPParser *datadog.LogsGeoIPParser) map[string]interface{} {
	return map[string]interface{}{
		"sources": 		ddGeoIPParser.Sources,
		"target":  		ddGeoIPParser.GetTarget(),
		"name":			ddGeoIPParser.GetName(),
		"is_enabled":	ddGeoIPParser.GetIsEnabled(),
	}
}

func buildTerraformGrokParser(ddGrok *datadog.LogsGrokParser) map[string]interface{} {
	return map[string]interface{}{
		"samples": 		ddGrok.Samples,
		"source":  		ddGrok.GetSource(),
		"grok":    		buildTerraformGrokRule(&ddGrok.Grok),
		"name":			ddGrok.GetName(),
		"is_enabled":	ddGrok.GetIsEnabled(),
	}
}

func buildTerraformGrokRule(ddGrokRule *datadog.LogsGrokParserRules) []map[string]interface{} {
	tfGrokRule := map[string]interface{}{
		"support_rules": 	ddGrokRule.GetSupportRules(),
		"match_rules":   	ddGrokRule.GetMatchRules(),
	}
	return []map[string]interface{}{tfGrokRule}
}

func buildTerraformMessageRemapper(remapper *datadog.LogsMessageRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources": 		remapper.Sources,
		"name":			remapper.GetName(),
		"is_enabled":	remapper.GetIsEnabled(),
	}
}

func buildTerraformDateRemapper(remapper *datadog.LogsDateRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources": 		remapper.Sources,
		"name":			remapper.GetName(),
		"is_enabled":	remapper.GetIsEnabled(),
	}
}

func buildTerraformServiceRemapper(remapper *datadog.LogsServiceRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources": 		remapper.Sources,
		"name":			remapper.GetName(),
		"is_enabled":	remapper.GetIsEnabled(),
	}
}

func buildTerraformStatusRemapper(remapper *datadog.LogsStatusRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources": 		remapper.Sources,
		"name":			remapper.GetName(),
		"is_enabled":	remapper.GetIsEnabled(),
	}
}

func buildTerraformTraceRemapper(remapper *datadog.LogsTraceRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources": 		remapper.Sources,
		"name":			remapper.GetName(),
		"is_enabled":	remapper.GetIsEnabled(),
	}
}

func buildTerraformCategoryProcessor(ddCategory *datadog.LogsCategoryProcessor) map[string]interface{} {
	return map[string]interface{}{
		"target":   	ddCategory.GetTarget(),
		"category": 	buildTerraformCategories(ddCategory.Categories),
		"name":			ddCategory.GetName(),
		"is_enabled":	ddCategory.GetIsEnabled(),
	}
}

func buildTerraformCategories(ddCategories []datadog.LogsCategoryProcessorCategories) []map[string]interface{} {
	tfCategories := make([]map[string]interface{}, len(ddCategories))
	for i, ddCategory := range ddCategories {
		tfCategories[i] = map[string]interface{}{
			"name":   ddCategory.GetName(),
			"filter": buildTerraformFilter(ddCategory.Filter),
		}
	}
	return tfCategories
}

func buildTerraformAttributeRemapper(ddAttribute *datadog.LogsRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":              ddAttribute.Sources,
		"source_type":          ddAttribute.GetSourceType(),
		"target":               ddAttribute.GetTarget(),
		"target_type":          ddAttribute.GetTargetType(),
		"preserve_source":      ddAttribute.GetPreserveSource(),
		"override_on_conflict": ddAttribute.GetOverrideOnConflict(),
		"name":				  	ddAttribute.GetName(),
		"is_enabled":		  	ddAttribute.GetIsEnabled(),
	}
}

func buildTerraformArithmeticProcessor(ddArithmetic *datadog.LogsArithmeticProcessor) map[string]interface{} {

	return map[string]interface{}{
		"target":             ddArithmetic.GetTarget(),
		"is_replace_missing": ddArithmetic.GetIsReplaceMissing(),
		"expression":         ddArithmetic.GetExpression(),
		"name":				  ddArithmetic.GetName(),
		"is_enabled":		  ddArithmetic.GetIsEnabled(),
	}
}

func buildTerraformFilter(ddFilter *datadog.LogsFilter) []map[string]interface{} {
	tfFilter := map[string]interface{}{
		"query": ddFilter.GetQuery(),
	}
	return []map[string]interface{}{tfFilter}
}

func buildDatadogPipeline(d *schema.ResourceData) (*datadog.LogsPipeline, error) {
	var ddPipeline datadog.LogsPipeline
	ddPipeline.SetName(d.Get("name").(string))
	ddPipeline.SetIsEnabled(d.Get("is_enabled").(bool))
	if tfFilter := d.Get("filter").([]interface{}); len(tfFilter) > 0 {
		ddPipeline.SetFilter(buildDatadogFilter(tfFilter[0].(map[string]interface{})))
	}
	ddProcessors, err := buildDatadogProcessors(d.Get("processor").([]interface{}))
	if err != nil {
		return nil, err
	}
	ddPipeline.Processors = ddProcessors
	return &ddPipeline, nil
}

func buildDatadogProcessors(tfProcessors []interface{}) (*[]datadog.LogsProcessor, error) {
	ddProcessors := make([]datadog.LogsProcessor, len(tfProcessors))
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

func buildDatadogProcessor(ddProcessorType string, tfProcessor map[string]interface{}) (datadog.LogsProcessor, error) {
	var ddProcessor = datadog.LogsProcessor{}
	var err error
	switch ddProcessorType {
	case datadog.NewLogsArithmeticProcessorWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogArithmeticProcessor(tfProcessor)
	case datadog.NewLogsRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogAttributeRemapper(tfProcessor)
	case datadog.NewLogsCategoryProcessorWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogCategoryProcessor(tfProcessor)
	case datadog.NewLogsDateRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogAttributeRemapper(tfProcessor)
	case datadog.NewLogsMessageRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogMessageRemapper(tfProcessor)
	case datadog.NewLogsServiceRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogServiceRemapper(tfProcessor)
	case datadog.NewLogsStatusRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogStatusRemapper(tfProcessor)
	case datadog.NewLogsTraceRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogTraceRemapper(tfProcessor)
	case datadog.NewLogsGeoIPParserWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogGeoIPParser(tfProcessor)
	case datadog.NewLogsGrokParserWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogGrokParser(tfProcessor)
	case datadog.NewLogsPipelineWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface, err = buildDatadogNestedPipeline(tfProcessor)
	case "pipeline":
		ddProcessor.LogsProcessorInterface = buildDatadogStringBuilderProcessor(tfProcessor)
	case datadog.NewLogsURLParserWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogURLParser(tfProcessor)
	case datadog.NewLogsUserAgentParserWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogUserAgentParser(tfProcessor)
	default:
		err = fmt.Errorf("failed to recoginize processor type: %s", ddProcessorType)
	}

	return ddProcessor, err
}

func buildDatadogURLParser(tfProcessor map[string]interface{}) *datadog.LogsURLParser {
	ddURLParser := datadog.LogsURLParser{}
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
	return &ddURLParser
}

func buildDatadogUserAgentParser(tfProcessor map[string]interface{}) *datadog.LogsUserAgentParser {
	ddUserAgentParser := datadog.LogsUserAgentParser{}
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
	return &ddUserAgentParser
}

func buildDatadogNestedPipeline(tfProcessor map[string]interface{}) (*datadog.LogsPipeline, error) {
	ddNestedPipeline := datadog.LogsPipeline{}
	if tfFilter, exist := tfProcessor["filter"].([]interface{}); exist && len(tfFilter) > 0 {
		ddNestedPipeline.SetFilter(buildDatadogFilter(tfFilter[0].(map[string]interface{})))
	}
	if tfProcessors, exists := tfProcessor["processor"].([]interface{}); exists && len(tfProcessors) > 0 {
		ddProcessors, err := buildDatadogProcessors(tfProcessors)
		if err != nil {
			return &ddNestedPipeline, err
		}
		ddNestedPipeline.Processors = ddProcessors
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddNestedPipeline.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddNestedPipeline.SetIsEnabled(tfIsEnabled)
	}
	return &ddNestedPipeline, nil
}

func buildDatadogStringBuilderProcessor(tfProcessor map[string]interface{}) *datadog.LogsStringBuilderProcessor {
	ddStringBuilder := datadog.LogsStringBuilderProcessor{}
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
	return &ddStringBuilder
}

func buildDatadogGeoIPParser(tfProcessor map[string]interface{}) *datadog.LogsGeoIPParser {
	ddGeoIPParser := datadog.LogsGeoIPParser{}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddGeoIPParser.SetTarget(tfTarget)
	}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddGeoIPParser.Sources = ddSources
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddGeoIPParser.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddGeoIPParser.SetIsEnabled(tfIsEnabled)
	}
	return &ddGeoIPParser
}

func buildDatadogGrokParser(tfProcessor map[string]interface{}) *datadog.LogsGrokParser {
	ddGrokParser := datadog.LogsGrokParser{}
	if tfSource, exists := tfProcessor["source"].(string); exists {
		ddGrokParser.SetSource(tfSource)
	}
	if tfSamples, exists := tfProcessor["samples"].([]interface{}); exists && len(tfSamples) > 0 {
		ddSamples := make([]string, len(tfSamples))
		for i, tfSample := range tfSamples {
			ddSamples[i] = tfSample.(string)
		}
		ddGrokParser.Samples = &ddSamples
	}
	if tfGrok, exists := tfProcessor["grok"].([]interface{}); exists && len(tfGrok) > 0 {
		ddGrok := datadog.LogsGrokParserRules{}
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
	return &ddGrokParser
}

func buildDatadogMessageRemapper(tfProcessor map[string]interface{}) *datadog.LogsMessageRemapper {
	ddRemapper := datadog.LogsMessageRemapper{}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddRemapper.Sources = ddSources
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddRemapper.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddRemapper.SetIsEnabled(tfIsEnabled)
	}
	return &ddRemapper
}

func buildDatadogServiceRemapper(tfProcessor map[string]interface{}) *datadog.LogsServiceRemapper {
	ddRemapper := datadog.LogsServiceRemapper{}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddRemapper.Sources = ddSources
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddRemapper.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddRemapper.SetIsEnabled(tfIsEnabled)
	}
	return &ddRemapper
}

func buildDatadogStatusRemapper(tfProcessor map[string]interface{}) *datadog.LogsStatusRemapper {
	ddRemapper := datadog.LogsStatusRemapper{}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddRemapper.Sources = ddSources
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddRemapper.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddRemapper.SetIsEnabled(tfIsEnabled)
	}
	return &ddRemapper
}

func buildDatadogTraceRemapper(tfProcessor map[string]interface{}) *datadog.LogsTraceRemapper {
	ddRemapper := datadog.LogsTraceRemapper{}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddRemapper.Sources = &ddSources
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddRemapper.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddRemapper.SetIsEnabled(tfIsEnabled)
	}
	return &ddRemapper
}

func buildDatadogCategoryProcessor(tfProcessor map[string]interface{}) *datadog.LogsCategoryProcessor {
	ddCategory := datadog.LogsCategoryProcessor{}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddCategory.SetTarget(tfTarget)
	}
	if tfCategories, exists := tfProcessor["category"].([]interface{}); exists {
		ddCategories := make([]datadog.LogsCategoryProcessorCategories, len(tfCategories))
		for i, tfC := range tfCategories {
			tfCategory := tfC.(map[string]interface{})
			ddCategory := datadog.LogsCategoryProcessorCategories{}
			if tfName, exist := tfCategory["name"].(string); exist {
				ddCategory.SetName(tfName)
			}
			if tfFilter, exist := tfCategory["filter"].([]interface{}); exist && len(tfFilter) > 0 {
				ddCategory.SetFilter(buildDatadogFilter(tfFilter[0].(map[string]interface{})))
			}

			ddCategories[i] = ddCategory
		}
		ddCategory.Categories = ddCategories
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddCategory.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddCategory.SetIsEnabled(tfIsEnabled)
	}
	return &ddCategory
}

func buildDatadogAttributeRemapper(tfProcessor map[string]interface{}) *datadog.LogsRemapper {
	ddAttribute := datadog.LogsRemapper{}
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
	return &ddAttribute
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

func buildDatadogArithmeticProcessor(tfProcessor map[string]interface{}) *datadog.LogsArithmeticProcessor {
	ddArithmetic := datadog.LogsArithmeticProcessor{}
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
	return &ddArithmetic
}

func buildDatadogFilter(tfFilter map[string]interface{}) datadog.LogsFilter {
	ddFilter := datadog.LogsFilter{}
	if tfQuery, exists := tfFilter["query"].(string); exists {
		ddFilter.SetQuery(tfQuery)
	}
	return ddFilter
}

func getPipelineSchema(isNested bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":       {Type: schema.TypeString, Required: true},
		"is_enabled": {Type: schema.TypeBool, Optional: true},
		"filter": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query": {Type: schema.TypeString, Required: true},
				},
			},
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
