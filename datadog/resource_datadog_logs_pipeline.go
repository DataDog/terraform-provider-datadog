package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
	"strings"
)

func resourceDatadogLogsPipeline() *schema.Resource {
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
	pipeline, err := buildDatadogLogsPipeline(d)
	if err != nil {
		return err
	}
	createdPipeline, err := meta.(*datadog.Client).CreateLogsPipeline(pipeline)
	if err != nil {
		return fmt.Errorf("failed to create logs pipeline using Datadog API: %s", err.Error())
	}
	d.SetId(*createdPipeline.Id)
	return resourceDatadogLogsPipelineRead(d, meta)
}

func resourceDatadogLogsPipelineRead(d *schema.ResourceData, meta interface{}) error {
	pipeline, err := meta.(*datadog.Client).GetLogsPipeline(d.Id())
	if err != nil {
		return err
	}
	if err = d.Set("name", pipeline.GetName()); err != nil {
		return err
	}
	if err = d.Set("is_enabled", pipeline.GetIsEnabled()); err != nil {
		return err
	}
	if err := setFilter(d, pipeline.Filter); err != nil {
		return err
	}

	if err := setProcessors(d, pipeline.Processors); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	pipeline, err := buildDatadogLogsPipeline(d)
	if err != nil {
		return err
	}
	client := meta.(*datadog.Client)
	if _, err := client.UpdateLogsPipeline(d.Id(), pipeline); err != nil {
		return fmt.Errorf("error updating logs pipeline: (%s)", err.Error())
	}
	return resourceDatadogLogsPipelineRead(d, meta)
}

func resourceDatadogLogsPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	if err := meta.(*datadog.Client).DeleteLogsPipeline(d.Id()); err != nil {
		if strings.Contains(err.Error(), "400 Bad Request") {
			return nil
		}
		return err
	}
	return nil
}

func resourceDatadogLogsPipelineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*datadog.Client)
	if _, err := client.GetLogsPipeline(d.Id()); err != nil {
		if strings.Contains(err.Error(), "400 Bad Request") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func setFilter(d *schema.ResourceData, filter *datadog.FilterConfiguration) error {
	filters := make([]map[string]interface{}, 1, 1)
	tfFilter := make(map[string]interface{})
	tfFilter["query"] = *filter.Query
	filters[0] = tfFilter
	if err := d.Set("filter", filters); err != nil {
		return err
	}
	return nil
}

func convertDDProcessors(tfProcessors []map[string]interface{}, processors []datadog.LogsProcessor) error {
	for i, ddProcessor := range processors {
		tfProcessor := make(map[string]interface{})
		tfPDetails := make(map[string]interface{})
		tfPDetails["name"] = *ddProcessor.Name
		tfPDetails["is_enabled"] = *ddProcessor.IsEnabled
		switch *ddProcessor.Type {
		case datadog.ArithmeticProcessorType:
			arithmeticP := ddProcessor.Definition.(datadog.ArithmeticProcessor)
			tfPDetails["target"] = *arithmeticP.Target
			tfPDetails["is_replace_missing"] = *arithmeticP.IsReplaceMissing
			tfPDetails["expression"] = *arithmeticP.Expression
		case datadog.AttributeRemapperType:
			attributeP := ddProcessor.Definition.(datadog.AttributeRemapper)
			sources := make([]string, len(attributeP.Sources))
			for i, source := range attributeP.Sources {
				sources[i] = source
			}
			tfPDetails["sources"] = sources
			tfPDetails["source_type"] = *attributeP.SourceType
			tfPDetails["target"] = *attributeP.Target
			tfPDetails["target_type"] = *attributeP.TargetType
			tfPDetails["preserve_source"] = *attributeP.PreserveSource
			tfPDetails["override_on_conflict"] = *attributeP.OverrideOnConflict
		case datadog.CategoryProcessorType:
			categoryP := ddProcessor.Definition.(datadog.CategoryProcessor)
			tfPDetails["target"] = *categoryP.Target
			categories := make([]map[string]interface{}, len(categoryP.Categories))
			for i, c := range categoryP.Categories {
				category := make(map[string]interface{})
				category["name"] = *c.Name
				filter := make(map[string]interface{})
				filter["query"] = *c.Filter.Query
				filterList := make([]interface{}, 1, 1)
				filterList[0] = filter
				category["filter"] = filterList
				categories[i] = category
			}
			tfPDetails["category"] = categories
		case datadog.DateRemapperType,
			datadog.MessageRemapperType,
			datadog.ServiceRemapperType,
			datadog.StatusRemapperType,
			datadog.TraceIdRemapperType:
			sourceP := ddProcessor.Definition.(datadog.SourceRemapper)
			sources := make([]string, len(sourceP.Sources))
			for i, source := range sourceP.Sources {
				sources[i] = source
			}
			tfPDetails["sources"] = sources
		case datadog.GrokParserType:
			grokP := ddProcessor.Definition.(datadog.GrokParser)
			tfPDetails["source"] = *grokP.Source
			grok := make(map[string]interface{})
			grok["support_rules"] = *grokP.GrokRule.SupportRules
			grok["match_rules"] = *grokP.GrokRule.MatchRules
			grokList := make([]interface{}, 1, 1)
			grokList[0] = grok
			tfPDetails["grok"] = grokList
		case datadog.NestedPipelineType:
			nestedP := ddProcessor.Definition.(datadog.NestedPipeline)
			filter := make(map[string]interface{})
			filter["query"] = *nestedP.Filter.Query
			filterList := make([]interface{}, 1, 1)
			filterList[0] = filter
			tfPDetails["filter"] = filterList
			ps := make([]map[string]interface{}, len(nestedP.Processors))
			if err := convertDDProcessors(ps, nestedP.Processors); err != nil {
				return nil
			}
			tfPDetails["processor"] = ps
		case datadog.UrlParserType:
			urlP := ddProcessor.Definition.(datadog.UrlParser)
			sources := make([]string, len(urlP.Sources))
			for i, source := range urlP.Sources {
				sources[i] = source
			}
			tfPDetails["sources"] = sources
			tfPDetails["target"] = *urlP.Target
			tfPDetails["normalize_ending_slashes"] = *urlP.NormalizeEndingSlashes
		case datadog.UserAgentParserType:
			userAgentP := ddProcessor.Definition.(datadog.UserAgentParser)
			sources := make([]string, len(userAgentP.Sources))
			for i, source := range userAgentP.Sources {
				sources[i] = source
			}
			tfPDetails["sources"] = sources
			tfPDetails["target"] = *userAgentP.Target
			tfPDetails["is_encoded"] = *userAgentP.IsEncoded
		default:
			return fmt.Errorf("failed to support datadog processor type, %s", *ddProcessor.Type)
		}
		tfPDetailsList := make([]interface{}, 1, 1)
		tfPDetailsList[0] = tfPDetails
		tfProcessor[ddProcessorTypes[*ddProcessor.Type]] = tfPDetailsList
		tfProcessors[i] = tfProcessor
	}
	return nil
}

func setProcessors(d *schema.ResourceData, processors []datadog.LogsProcessor) error {
	tfProcessors := make([]map[string]interface{}, len(processors))
	if err := convertDDProcessors(tfProcessors, processors); err != nil {
		return err
	}
	if err := d.Set("processor", tfProcessors); err != nil {
		return err
	}
	return nil
}

func buildDatadogLogsPipeline(d *schema.ResourceData) (*datadog.LogsPipeline, error) {
	var pipeline datadog.LogsPipeline
	pipeline.SetName(d.Get("name").(string))
	pipeline.SetIsEnabled(d.Get("is_enabled").(bool))
	var ddFilter = datadog.FilterConfiguration{}
	buildFilter(d, &ddFilter)
	pipeline.SetFilter(ddFilter)
	processors, err := buildProcessors(d)
	if err != nil {
		return nil, err
	}
	pipeline.Processors = processors
	return &pipeline, nil
}

func buildProcessors(d *schema.ResourceData) ([]datadog.LogsProcessor, error) {
	tfProcessors := d.Get("processor").([]interface{})
	ddProcessors := make([]datadog.LogsProcessor, len(tfProcessors))
	if err := convertTFProcessors(ddProcessors, tfProcessors); err != nil {
		return ddProcessors, err
	}
	return ddProcessors, nil
}

func convertTFProcessors(ddProcessors []datadog.LogsProcessor, tfProcessors []interface{}) error {
	for i, tfEntry := range tfProcessors {
		tfMap := tfEntry.(map[string]interface{})
		for tfPType, ddPType := range tfProcessorTypes {
			if v, ok := tfMap[tfPType].([]interface{}); ok && len(v) > 0 {
				if tfP, ok := v[0].(map[string]interface{}); ok {
					var ddProcessor = datadog.LogsProcessor{}
					if err := convert(tfP, ddPType, &ddProcessor); err != nil {
						return err
					}
					ddProcessors[i] = ddProcessor
					break
				}
			}
		}
	}
	return nil
}

func convert(tfProcessor map[string]interface{}, processorType string, ddProcessor *datadog.LogsProcessor) error {
	ddProcessor.Type = datadog.String(processorType)
	if name, ok := tfProcessor["name"].(string); ok {
		ddProcessor.Name = &name
	}
	if isEnabled, ok := tfProcessor["is_enabled"].(bool); ok {
		ddProcessor.IsEnabled = &isEnabled
	}
	switch processorType {
	case datadog.ArithmeticProcessorType:
		var arithmeticProcessor = datadog.ArithmeticProcessor{}
		if expression, ok := tfProcessor["expression"].(string); ok {
			arithmeticProcessor.Expression = &expression
		}
		if target, ok := tfProcessor["target"].(string); ok {
			arithmeticProcessor.Target = &target
		}
		if isReplacingMissing, ok := tfProcessor["is_replace_missing"].(bool); ok {
			arithmeticProcessor.IsReplaceMissing = &isReplacingMissing
		}
		ddProcessor.Definition = arithmeticProcessor
	case datadog.AttributeRemapperType:
		var attributeRemapper = datadog.AttributeRemapper{}
		if sources, ok := tfProcessor["sources"].([]interface{}); ok {
			attributeRemapper.Sources = make([]string, len(sources))
			for i, source := range sources {
				attributeRemapper.Sources[i] = source.(string)
			}
		}
		if sourceType, ok := tfProcessor["source_type"].(string); ok {
			attributeRemapper.SourceType = &sourceType
		}
		if target, ok := tfProcessor["target"].(string); ok {
			attributeRemapper.Target = &target
		}
		if targetType, ok := tfProcessor["target_type"].(string); ok {
			attributeRemapper.TargetType = &targetType
		}
		if preserveSource, ok := tfProcessor["preserve_source"].(bool); ok {
			attributeRemapper.PreserveSource = &preserveSource
		}
		if overrideOnConflict, ok := tfProcessor["override_on_conflict"].(bool); ok {
			attributeRemapper.OverrideOnConflict = &overrideOnConflict
		}
		ddProcessor.Definition = attributeRemapper
	case datadog.CategoryProcessorType:
		var categoryProcessor = datadog.CategoryProcessor{}
		if target, ok := tfProcessor["target"].(string); ok {
			categoryProcessor.Target = &target
		}
		if categories, ok := tfProcessor["category"].([]interface{}); ok {
			var ddCategories = make([]datadog.Category, len(categories))
			for i, category := range categories {
				var tfC = category.(map[string]interface{})
				var ddCategory = datadog.Category{}
				if name, ok := tfC["name"].(string); ok {
					ddCategory.Name = &name
				}
				if filter, ok := tfC["filter"].([]interface{}); ok && len(filter) > 0 {
					ddFilter := datadog.FilterConfiguration{}
					convertFilter(filter[0], &ddFilter)
					ddCategory.Filter = &ddFilter
				}
				ddCategories[i] = ddCategory
			}
			categoryProcessor.Categories = ddCategories
		}
		ddProcessor.Definition = categoryProcessor
	case datadog.DateRemapperType,
		datadog.MessageRemapperType,
		datadog.ServiceRemapperType,
		datadog.StatusRemapperType,
		datadog.TraceIdRemapperType:
		var sourceRemapper = datadog.SourceRemapper{}
		if sources, ok := tfProcessor["sources"].([]interface{}); ok {
			sourceRemapper.Sources = make([]string, len(sources))
			for i, source := range sources {
				sourceRemapper.Sources[i] = source.(string)
			}
		}
		ddProcessor.Definition = sourceRemapper
	case datadog.GrokParserType:
		var grokParser = datadog.GrokParser{}
		if source, ok := tfProcessor["source"].(string); ok {
			grokParser.Source = &source
		}
		if grok, ok := tfProcessor["grok"].([]interface{}); ok && len(grok) > 0 {
			var ddGrok = datadog.GrokRule{}
			var tfG = grok[0].(map[string]interface{})
			if support, ok := tfG["support_rules"].(string); ok {
				ddGrok.SupportRules = &support
			}
			if match, ok := tfG["match_rules"].(string); ok {
				ddGrok.MatchRules = &match
			}
			grokParser.GrokRule = &ddGrok
		}
		ddProcessor.Definition = grokParser
	case datadog.NestedPipelineType:
		var nestedPipeline = datadog.NestedPipeline{}
		if filter, ok := tfProcessor["filter"].([]interface{}); ok && len(filter) > 0 {
			ddFilter := datadog.FilterConfiguration{}
			convertFilter(filter[0], &ddFilter)
			nestedPipeline.Filter = &ddFilter
		}
		if processors, ok := tfProcessor["processor"].([]interface{}); ok {
			nestedProcessors := make([]datadog.LogsProcessor, len(processors))
			if err := convertTFProcessors(nestedProcessors, processors); err != nil {
				return err
			}
			nestedPipeline.Processors = nestedProcessors
		}
		ddProcessor.Definition = nestedPipeline
	case datadog.UrlParserType:
		var urlParser = datadog.UrlParser{}
		if sources, ok := tfProcessor["sources"].([]interface{}); ok {
			urlParser.Sources = make([]string, len(sources))
			for i, source := range sources {
				urlParser.Sources[i] = source.(string)
			}
		}
		if target, ok := tfProcessor["target"].(string); ok {
			urlParser.Target = &target
		}
		if normalizedEndingSlashes, ok := tfProcessor["normalize_ending_slashes"].(bool); ok {
			urlParser.NormalizeEndingSlashes = &normalizedEndingSlashes
		}
		ddProcessor.Definition = urlParser
	case datadog.UserAgentParserType:
		var userAgentParser = datadog.UserAgentParser{}
		if sources, ok := tfProcessor["sources"].([]interface{}); ok {
			userAgentParser.Sources = make([]string, len(sources))
			for i, source := range sources {
				userAgentParser.Sources[i] = source.(string)
			}
		}
		if target, ok := tfProcessor["target"].(string); ok {
			userAgentParser.Target = &target
		}
		if isEncoded, ok := tfProcessor["is_encoded"].(bool); ok {
			userAgentParser.IsEncoded = &isEncoded
		}
		ddProcessor.Definition = userAgentParser
	default:
		return fmt.Errorf("failed to recoginize processor type: %s", processorType)
	}

	return nil
}

func buildFilter(d *schema.ResourceData, ddFilter *datadog.FilterConfiguration) {
	if v, ok := d.Get("filter").([]interface{}); ok && len(v) > 0 {
		convertFilter(v[0], ddFilter)
	}
}

func convertFilter(tfFilter interface{}, ddFilter *datadog.FilterConfiguration) {
	if tfF, ok := tfFilter.(map[string]interface{}); ok {
		if query, ok := tfF["query"].(string); ok {
			ddFilter.Query = &query
		}
	}
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
	processorsSchema[tfArithmeticProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getArithmeticProcessor(),
		},
	}
	processorsSchema[tfAttributeRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getAttributeRemapper(),
		},
	}
	processorsSchema[tfCategoryProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getCategoryProcessor(),
		},
	}
	processorsSchema[tfDateRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getSourceRemapper(),
		},
	}
	processorsSchema[tfGrokParserProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getGrokParser(),
		},
	}
	processorsSchema[tfMessageRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getSourceRemapper(),
		},
	}
	processorsSchema[tfServiceRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getSourceRemapper(),
		},
	}
	processorsSchema[tfStatusRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getSourceRemapper(),
		},
	}
	processorsSchema[tfTraceIdRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getSourceRemapper(),
		},
	}
	processorsSchema[tfUrlParserProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getUrlParser(),
		},
	}
	processorsSchema[tfUserAgentParserProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: getUserAgentParser(),
		},
	}
	return processorsSchema
}

func getSourceRemapper() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":       {Type: schema.TypeString, Optional: true},
		"is_enabled": {Type: schema.TypeBool, Optional: true},
		"sources":    {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
	}
}

func getArithmeticProcessor() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":               {Type: schema.TypeString, Optional: true},
		"is_enabled":         {Type: schema.TypeBool, Optional: true},
		"expression":         {Type: schema.TypeString, Required: true},
		"target":             {Type: schema.TypeString, Required: true},
		"is_replace_missing": {Type: schema.TypeBool, Optional: true},
	}
}

func getAttributeRemapper() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":                 {Type: schema.TypeString, Optional: true},
		"is_enabled":           {Type: schema.TypeBool, Optional: true},
		"sources":              {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"source_type":          {Type: schema.TypeString, Required: true},
		"target":               {Type: schema.TypeString, Required: true},
		"target_type":          {Type: schema.TypeString, Required: true},
		"preserve_source":      {Type: schema.TypeBool, Optional: true},
		"override_on_conflict": {Type: schema.TypeBool, Optional: true},
	}
}

func getCategoryProcessor() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":       {Type: schema.TypeString, Optional: true},
		"is_enabled": {Type: schema.TypeBool, Optional: true},
		"target":     {Type: schema.TypeString, Required: true},
		"category": {Type: schema.TypeList, Required: true, Elem: &schema.Resource{
			Schema: getCategorySchema(),
		}},
	}
}

func getCategorySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func getGrokParser() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":       {Type: schema.TypeString, Optional: true},
		"is_enabled": {Type: schema.TypeBool, Optional: true},
		"source":     {Type: schema.TypeString, Required: true},
		"grok": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Required: true,
			Elem:     &schema.Resource{Schema: getGrokSchema()},
		},
	}
}

func getGrokSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"support_rules": {Type: schema.TypeString, Required: true},
		"match_rules":   {Type: schema.TypeString, Required: true},
	}
}

func getUrlParser() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":                     {Type: schema.TypeString, Optional: true},
		"is_enabled":               {Type: schema.TypeBool, Optional: true},
		"sources":                  {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"target":                   {Type: schema.TypeString, Required: true},
		"normalize_ending_slashes": {Type: schema.TypeBool, Optional: true},
	}
}

func getUserAgentParser() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":       {Type: schema.TypeString, Optional: true},
		"is_enabled": {Type: schema.TypeBool, Optional: true},
		"sources":    {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"target":     {Type: schema.TypeString, Required: true},
		"is_encoded": {Type: schema.TypeBool, Optional: true},
	}
}

const (
	tfArithmeticProcessor        = "arithmetic_processor"
	tfAttributeRemapperProcessor = "attribute_remapper"
	tfCategoryProcessor          = "category_processor"
	tfDateRemapperProcessor      = "date_remapper"
	tfGrokParserProcessor        = "grok_parser"
	tfMessageRemapperProcessor   = "message_remapper"
	tfNestedPipelineProcessor    = "pipeline"
	tfServiceRemapperProcessor   = "service_remapper"
	tfStatusRemapperProcessor    = "status_remapper"
	tfTraceIdRemapperProcessor   = "trace_id_remapper"
	tfUrlParserProcessor         = "url_parser"
	tfUserAgentParserProcessor   = "user_agent_parser"
)

var tfProcessorTypes = map[string]string{
	tfArithmeticProcessor:        datadog.ArithmeticProcessorType,
	tfAttributeRemapperProcessor: datadog.AttributeRemapperType,
	tfCategoryProcessor:          datadog.CategoryProcessorType,
	tfDateRemapperProcessor:      datadog.DateRemapperType,
	tfGrokParserProcessor:        datadog.GrokParserType,
	tfMessageRemapperProcessor:   datadog.MessageRemapperType,
	tfNestedPipelineProcessor:    datadog.NestedPipelineType,
	tfServiceRemapperProcessor:   datadog.ServiceRemapperType,
	tfStatusRemapperProcessor:    datadog.StatusRemapperType,
	tfTraceIdRemapperProcessor:   datadog.TraceIdRemapperType,
	tfUrlParserProcessor:         datadog.UrlParserType,
	tfUserAgentParserProcessor:   datadog.UserAgentParserType,
}

var ddProcessorTypes = map[string]string{
	datadog.ArithmeticProcessorType: tfArithmeticProcessor,
	datadog.AttributeRemapperType:   tfAttributeRemapperProcessor,
	datadog.CategoryProcessorType:   tfCategoryProcessor,
	datadog.DateRemapperType:        tfDateRemapperProcessor,
	datadog.GrokParserType:          tfGrokParserProcessor,
	datadog.MessageRemapperType:     tfMessageRemapperProcessor,
	datadog.NestedPipelineType:      tfNestedPipelineProcessor,
	datadog.ServiceRemapperType:     tfServiceRemapperProcessor,
	datadog.StatusRemapperType:      tfStatusRemapperProcessor,
	datadog.TraceIdRemapperType:     tfTraceIdRemapperProcessor,
	datadog.UrlParserType:           tfUrlParserProcessor,
	datadog.UserAgentParserType:     tfUserAgentParserProcessor,
}
