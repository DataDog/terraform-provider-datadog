package datadog

import (
	"fmt"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

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
	tfArithmeticProcessor:        datadogV1.NewLogsArithmeticProcessorWithDefaults().GetType(),
	tfAttributeRemapperProcessor: datadogV1.NewLogsAttributeRemapperWithDefaults().GetType(),
	tfCategoryProcessor:          datadogV1.NewLogsCategoryProcessorWithDefaults().GetType(),
	tfDateRemapperProcessor:      datadogV1.NewLogsDateRemapperWithDefaults().GetType(),
	tfGeoIPParserProcessor:       datadogV1.NewLogsGeoIPParserWithDefaults().GetType(),
	tfGrokParserProcessor:        datadogV1.NewLogsGrokParserWithDefaults().GetType(),
	tfLookupProcessor:            datadogV1.NewLogsLookupProcessorWithDefaults().GetType(),
	tfMessageRemapperProcessor:   datadogV1.NewLogsMessageRemapperWithDefaults().GetType(),
	tfNestedPipelineProcessor:    datadogV1.NewLogsPipelineProcessorWithDefaults().GetType(),
	tfServiceRemapperProcessor:   datadogV1.NewLogsServiceRemapperWithDefaults().GetType(),
	tfStatusRemapperProcessor:    datadogV1.NewLogsStatusRemapperWithDefaults().GetType(),
	tfStringBuilderProcessor:     datadogV1.NewLogsStringBuilderProcessorWithDefaults().GetType(),
	tfTraceIDRemapperProcessor:   datadogV1.NewLogsTraceRemapperWithDefaults().GetType(),
	tfURLParserProcessor:         datadogV1.NewLogsURLParserWithDefaults().GetType(),
	tfUserAgentParserProcessor:   datadogV1.NewLogsUserAgentParserWithDefaults().GetType(),
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
	datadogV1.NewLogsArithmeticProcessorWithDefaults().GetType():    tfArithmeticProcessor,
	datadogV1.NewLogsAttributeRemapperWithDefaults().GetType():      tfAttributeRemapperProcessor,
	datadogV1.NewLogsCategoryProcessorWithDefaults().GetType():      tfCategoryProcessor,
	datadogV1.NewLogsDateRemapperWithDefaults().GetType():           tfDateRemapperProcessor,
	datadogV1.NewLogsGeoIPParserWithDefaults().GetType():            tfGeoIPParserProcessor,
	datadogV1.NewLogsGrokParserWithDefaults().GetType():             tfGrokParserProcessor,
	datadogV1.NewLogsLookupProcessorWithDefaults().GetType():        tfLookupProcessor,
	datadogV1.NewLogsMessageRemapperWithDefaults().GetType():        tfMessageRemapperProcessor,
	datadogV1.NewLogsPipelineProcessorWithDefaults().GetType():      tfNestedPipelineProcessor,
	datadogV1.NewLogsServiceRemapperWithDefaults().GetType():        tfServiceRemapperProcessor,
	datadogV1.NewLogsStatusRemapperWithDefaults().GetType():         tfStatusRemapperProcessor,
	datadogV1.NewLogsStringBuilderProcessorWithDefaults().GetType(): tfStringBuilderProcessor,
	datadogV1.NewLogsTraceRemapperWithDefaults().GetType():          tfTraceIDRemapperProcessor,
	datadogV1.NewLogsURLParserWithDefaults().GetType():              tfURLParserProcessor,
	datadogV1.NewLogsUserAgentParserWithDefaults().GetType():        tfUserAgentParserProcessor,
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

var lookupProcessor = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":       {Type: schema.TypeString, Optional: true},
			"is_enabled": {Type: schema.TypeBool, Optional: true},
			"source":     {Type: schema.TypeString, Required: true},
			"target":     {Type: schema.TypeString, Required: true},
			"lookup_table": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_lookup": {Type: schema.TypeString, Optional: true},
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	ddPipeline, err := buildDatadogPipeline(d)
	if err != nil {
		return err
	}
	createdPipeline, _, err := datadogClientV1.LogsPipelinesApi.CreateLogsPipeline(authV1).Body(*ddPipeline).Execute()
	if err != nil {
		return translateClientError(err, "failed to create logs pipeline using Datadog API")
	}
	d.SetId(*createdPipeline.Id)
	return resourceDatadogLogsPipelineRead(d, meta)
}

func resourceDatadogLogsPipelineRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	ddPipeline, _, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, d.Id()).Execute()
	if err != nil {
		return translateClientError(err, "failed to get logs pipeline using Datadog API")
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	ddPipeline, err := buildDatadogPipeline(d)
	if err != nil {
		return err
	}
	if _, _, err := datadogClientV1.LogsPipelinesApi.UpdateLogsPipeline(authV1, d.Id()).Body(*ddPipeline).Execute(); err != nil {
		return translateClientError(err, "error updating logs pipeline")
	}
	return resourceDatadogLogsPipelineRead(d, meta)
}

func resourceDatadogLogsPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, err := datadogClientV1.LogsPipelinesApi.DeleteLogsPipeline(authV1, d.Id()).Execute(); err != nil {
		// API returns 400 when the specific pipeline id doesn't exist through DELETE request.
		if strings.Contains(err.Error(), "400 Bad Request") {
			return nil
		}
		return translateClientError(err, "error deleting logs pipeline")
	}
	return nil
}

func resourceDatadogLogsPipelineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, _, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, d.Id()).Execute(); err != nil {
		// API returns 400 when the specific pipeline id doesn't exist through GET request.
		if strings.Contains(err.Error(), "400 Bad Request") {
			return false, nil
		}
		return false, translateClientError(err, "error getting logs pipeline")
	}
	return true, nil
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
	var err error
	switch ddProcessor.LogsProcessorInterface.GetType() {
	case datadogV1.NewLogsArithmeticProcessorWithDefaults().GetType():
		logsArithmeticProcessor := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsArithmeticProcessor)
		tfProcessor = buildTerraformArithmeticProcessor(logsArithmeticProcessor)
	case datadogV1.NewLogsAttributeRemapperWithDefaults().GetType():
		logsAttributeRemapper := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsAttributeRemapper)
		tfProcessor = buildTerraformAttributeRemapper(logsAttributeRemapper)
	case datadogV1.NewLogsCategoryProcessorWithDefaults().GetType():
		logsCategoryProcessor := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsCategoryProcessor)
		tfProcessor = buildTerraformCategoryProcessor(logsCategoryProcessor)
	case datadogV1.NewLogsDateRemapperWithDefaults().GetType():
		logsDateRemapper := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsDateRemapper)
		tfProcessor = buildTerraformDateRemapper(logsDateRemapper)
	case datadogV1.NewLogsMessageRemapperWithDefaults().GetType():
		logsMessageRemapper := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsMessageRemapper)
		tfProcessor = buildTerraformMessageRemapper(logsMessageRemapper)
	case datadogV1.NewLogsServiceRemapperWithDefaults().GetType():
		logsServiceRemapper := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsServiceRemapper)
		tfProcessor = buildTerraformServiceRemapper(logsServiceRemapper)
	case datadogV1.NewLogsStatusRemapperWithDefaults().GetType():
		logsStatusRemapper := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsStatusRemapper)
		tfProcessor = buildTerraformStatusRemapper(logsStatusRemapper)
	case datadogV1.NewLogsTraceRemapperWithDefaults().GetType():
		logsTraceRemapper := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsTraceRemapper)
		tfProcessor = buildTerraformTraceRemapper(logsTraceRemapper)
	case datadogV1.NewLogsGeoIPParserWithDefaults().GetType():
		logsGeoIPParser := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsGeoIPParser)
		tfProcessor = buildTerraformGeoIPParser(logsGeoIPParser)
	case datadogV1.NewLogsGrokParserWithDefaults().GetType():
		logsGrokParser := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsGrokParser)
		tfProcessor = buildTerraformGrokParser(logsGrokParser)
	case datadogV1.NewLogsLookupProcessorWithDefaults().GetType():
		logsLookupProcessor := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsLookupProcessor)
		tfProcessor = buildTerraformLookupProcessor(logsLookupProcessor)
	case datadogV1.NewLogsPipelineProcessorWithDefaults().GetType():
		logsPipelineProcessor := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsPipelineProcessor)
		tfProcessor, err = buildTerraformNestedPipeline(logsPipelineProcessor)
	case datadogV1.NewLogsStringBuilderProcessorWithDefaults().GetType():
		logsStringBuilderProcessor := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsStringBuilderProcessor)
		tfProcessor = buildTerraformStringBuilderProcessor(logsStringBuilderProcessor)
	case datadogV1.NewLogsURLParserWithDefaults().GetType():
		logsURLParser := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsURLParser)
		tfProcessor = buildTerraformURLParser(logsURLParser)
	case datadogV1.NewLogsUserAgentParserWithDefaults().GetType():
		logsUserAgentParser := ddProcessor.LogsProcessorInterface.(*datadogV1.LogsUserAgentParser)
		tfProcessor = buildTerraformUserAgentParser(logsUserAgentParser)
	default:
		err = fmt.Errorf("failed to support datadogV1 processor type, %s", ddProcessor.LogsProcessorInterface.GetType())
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		ddProcessorTypes[ddProcessor.LogsProcessorInterface.GetType()]: []map[string]interface{}{tfProcessor},
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
		"sources":    ddGeoIPParser.Sources,
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
		"sources":    remapper.Sources,
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformDateRemapper(remapper *datadogV1.LogsDateRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.Sources,
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformServiceRemapper(remapper *datadogV1.LogsServiceRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.Sources,
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformStatusRemapper(remapper *datadogV1.LogsStatusRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.Sources,
		"name":       remapper.GetName(),
		"is_enabled": remapper.GetIsEnabled(),
	}
}

func buildTerraformTraceRemapper(remapper *datadogV1.LogsTraceRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":    remapper.Sources,
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

func buildTerraformCategories(ddCategories []datadogV1.LogsCategoryProcessorCategories) []map[string]interface{} {
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
		ddPipeline.SetFilter(buildDatadogFilter(tfFilter[0].(map[string]interface{})))
	}
	ddProcessors, err := buildDatadogProcessors(d.Get("processor").([]interface{}))
	if err != nil {
		return nil, err
	}
	ddPipeline.Processors = ddProcessors
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
	case datadogV1.NewLogsArithmeticProcessorWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogArithmeticProcessor(tfProcessor)
	case datadogV1.NewLogsAttributeRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogAttributeRemapper(tfProcessor)
	case datadogV1.NewLogsCategoryProcessorWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogCategoryProcessor(tfProcessor)
	case datadogV1.NewLogsDateRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogDateRemapperProcessor(tfProcessor)
	case datadogV1.NewLogsMessageRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogMessageRemapper(tfProcessor)
	case datadogV1.NewLogsServiceRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogServiceRemapper(tfProcessor)
	case datadogV1.NewLogsStatusRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogStatusRemapper(tfProcessor)
	case datadogV1.NewLogsTraceRemapperWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogTraceRemapper(tfProcessor)
	case datadogV1.NewLogsGeoIPParserWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogGeoIPParser(tfProcessor)
	case datadogV1.NewLogsGrokParserWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogGrokParser(tfProcessor)
	case datadogV1.NewLogsLookupProcessorWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogLookupProcessor(tfProcessor)
	case datadogV1.NewLogsPipelineProcessorWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface, err = buildDatadogNestedPipeline(tfProcessor)
	case datadogV1.NewLogsStringBuilderProcessorWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface, err = buildDatadogStringBuilderProcessor(tfProcessor)
	case datadogV1.NewLogsURLParserWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogURLParser(tfProcessor)
	case datadogV1.NewLogsUserAgentParserWithDefaults().GetType():
		ddProcessor.LogsProcessorInterface = buildDatadogUserAgentParser(tfProcessor)
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
		ddLookupProcessor.LookupTable = ddLookupTable
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
		ddGeoIPParser.Sources = ddSources
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
		ddCategories := make([]datadogV1.LogsCategoryProcessorCategories, len(tfCategories))
		for i, tfC := range tfCategories {
			tfCategory := tfC.(map[string]interface{})
			ddCategory := datadogV1.LogsCategoryProcessorCategories{}
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
