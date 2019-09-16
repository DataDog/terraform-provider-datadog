package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
	"strings"
)

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

var arithmeticProcessor = map[string]*schema.Schema{
	"name":               {Type: schema.TypeString, Optional: true},
	"is_enabled":         {Type: schema.TypeBool, Optional: true},
	"expression":         {Type: schema.TypeString, Required: true},
	"target":             {Type: schema.TypeString, Required: true},
	"is_replace_missing": {Type: schema.TypeBool, Optional: true},
}

var attributeRemapper = map[string]*schema.Schema{
	"name":                 {Type: schema.TypeString, Optional: true},
	"is_enabled":           {Type: schema.TypeBool, Optional: true},
	"sources":              {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
	"source_type":          {Type: schema.TypeString, Required: true},
	"target":               {Type: schema.TypeString, Required: true},
	"target_type":          {Type: schema.TypeString, Required: true},
	"preserve_source":      {Type: schema.TypeBool, Optional: true},
	"override_on_conflict": {Type: schema.TypeBool, Optional: true},
}

var categoryProcessor = map[string]*schema.Schema{
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
}

var grokParser = map[string]*schema.Schema{
	"name":       {Type: schema.TypeString, Optional: true},
	"is_enabled": {Type: schema.TypeBool, Optional: true},
	"source":     {Type: schema.TypeString, Required: true},
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
}

var sourceRemapper = map[string]*schema.Schema{
	"name":       {Type: schema.TypeString, Optional: true},
	"is_enabled": {Type: schema.TypeBool, Optional: true},
	"sources":    {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
}

var urlParser = map[string]*schema.Schema{
	"name":                     {Type: schema.TypeString, Optional: true},
	"is_enabled":               {Type: schema.TypeBool, Optional: true},
	"sources":                  {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
	"target":                   {Type: schema.TypeString, Required: true},
	"normalize_ending_slashes": {Type: schema.TypeBool, Optional: true},
}

var userAgentParser = map[string]*schema.Schema{
	"name":       {Type: schema.TypeString, Optional: true},
	"is_enabled": {Type: schema.TypeBool, Optional: true},
	"sources":    {Type: schema.TypeList, Required: true, Elem: &schema.Schema{Type: schema.TypeString}},
	"target":     {Type: schema.TypeString, Required: true},
	"is_encoded": {Type: schema.TypeBool, Optional: true},
}

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
	ddPipeline, err := buildDatadogPipeline(d)
	if err != nil {
		return err
	}
	createdPipeline, err := meta.(*datadog.Client).CreateLogsPipeline(ddPipeline)
	if err != nil {
		return fmt.Errorf("failed to create logs pipeline using Datadog API: %s", err.Error())
	}
	d.SetId(*createdPipeline.Id)
	return resourceDatadogLogsPipelineRead(d, meta)
}

func resourceDatadogLogsPipelineRead(d *schema.ResourceData, meta interface{}) error {
	ddPipeline, err := meta.(*datadog.Client).GetLogsPipeline(d.Id())
	if err != nil {
		return err
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
	tfProcessors, err := buildTerraformProcessors(ddPipeline.Processors)
	if err != nil {
		return err
	}
	if err := d.Set("processor", tfProcessors); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	ddPipeline, err := buildDatadogPipeline(d)
	if err != nil {
		return err
	}
	client := meta.(*datadog.Client)
	if _, err := client.UpdateLogsPipeline(d.Id(), ddPipeline); err != nil {
		return fmt.Errorf("error updating logs pipeline: (%s)", err.Error())
	}
	return resourceDatadogLogsPipelineRead(d, meta)
}

func resourceDatadogLogsPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	if err := meta.(*datadog.Client).DeleteLogsPipeline(d.Id()); err != nil {
		// API returns 400 when the specific pipeline id doesn't exist through DELETE request.
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
		// API returns 400 when the specific pipeline id doesn't exist through GET request.
		if strings.Contains(err.Error(), "400 Bad Request") {
			return false, nil
		}
		return false, err
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
	switch *ddProcessor.Type {
	case datadog.ArithmeticProcessorType:
		tfProcessor = buildTerraformArithmeticProcessor(ddProcessor.Definition.(datadog.ArithmeticProcessor))
	case datadog.AttributeRemapperType:
		tfProcessor = buildTerraformAttributeRemapper(ddProcessor.Definition.(datadog.AttributeRemapper))
	case datadog.CategoryProcessorType:
		tfProcessor = buildTerraformCategoryProcessor(ddProcessor.Definition.(datadog.CategoryProcessor))
	case datadog.DateRemapperType,
		datadog.MessageRemapperType,
		datadog.ServiceRemapperType,
		datadog.StatusRemapperType,
		datadog.TraceIdRemapperType:
		tfProcessor = buildTerraformSourceRemapper(ddProcessor.Definition.(datadog.SourceRemapper))
	case datadog.GrokParserType:
		tfProcessor = buildTerraformGrokParser(ddProcessor.Definition.(datadog.GrokParser))
	case datadog.NestedPipelineType:
		tfProcessor, err = buildTerraformNestedPipeline(ddProcessor.Definition.(datadog.NestedPipeline))
	case datadog.UrlParserType:
		tfProcessor = buildTerraformUrlParser(ddProcessor.Definition.(datadog.UrlParser))
	case datadog.UserAgentParserType:
		tfProcessor = buildTerraformUserAgentParser(ddProcessor.Definition.(datadog.UserAgentParser))
	default:
		err = fmt.Errorf("failed to support datadog processor type, %s", *ddProcessor.Type)
	}
	if err != nil {
		return nil, err
	}
	tfProcessor["name"] = ddProcessor.GetName()
	tfProcessor["is_enabled"] = ddProcessor.GetIsEnabled()
	return map[string]interface{}{
		ddProcessorTypes[*ddProcessor.Type]: []map[string]interface{}{tfProcessor},
	}, nil
}

func buildTerraformUserAgentParser(ddUserAgent datadog.UserAgentParser) map[string]interface{} {
	return map[string]interface{}{
		"sources":    ddUserAgent.Sources,
		"target":     ddUserAgent.GetTarget(),
		"is_encoded": ddUserAgent.GetIsEncoded(),
	}
}

func buildTerraformUrlParser(ddUrl datadog.UrlParser) map[string]interface{} {
	return map[string]interface{}{
		"sources":                  ddUrl.Sources,
		"target":                   ddUrl.GetTarget(),
		"normalize_ending_slashes": ddUrl.GetNormalizeEndingSlashes(),
	}
}

func buildTerraformNestedPipeline(ddNested datadog.NestedPipeline) (map[string]interface{}, error) {
	tfProcessors, err := buildTerraformProcessors(ddNested.Processors)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"filter":    buildTerraformFilter(ddNested.Filter),
		"processor": tfProcessors,
	}, nil
}

func buildTerraformGrokParser(ddGrok datadog.GrokParser) map[string]interface{} {
	return map[string]interface{}{
		"source": ddGrok.GetSource(),
		"grok":   buildTerraformGrokRule(ddGrok.GrokRule),
	}
}

func buildTerraformGrokRule(ddGrokRule *datadog.GrokRule) []map[string]interface{} {
	tfGrokRule := map[string]interface{}{
		"support_rules": ddGrokRule.GetSupportRules(),
		"match_rules":   ddGrokRule.GetMatchRules(),
	}
	return []map[string]interface{}{tfGrokRule}
}

func buildTerraformSourceRemapper(ddSource datadog.SourceRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources": ddSource.Sources,
	}
}

func buildTerraformCategoryProcessor(ddCategory datadog.CategoryProcessor) map[string]interface{} {
	return map[string]interface{}{
		"target":   ddCategory.GetTarget(),
		"category": buildTerraformCategories(ddCategory.Categories),
	}
}

func buildTerraformCategories(ddCategories []datadog.Category) []map[string]interface{} {
	tfCategories := make([]map[string]interface{}, len(ddCategories))
	for i, ddCategory := range ddCategories {
		tfCategories[i] = map[string]interface{}{
			"name":   ddCategory.GetName(),
			"filter": buildTerraformFilter(ddCategory.Filter),
		}
	}
	return tfCategories
}

func buildTerraformAttributeRemapper(ddAttribute datadog.AttributeRemapper) map[string]interface{} {
	return map[string]interface{}{
		"sources":              ddAttribute.Sources,
		"source_type":          ddAttribute.GetSourceType(),
		"target":               ddAttribute.GetTarget(),
		"target_type":          ddAttribute.GetTargetType(),
		"preserve_source":      ddAttribute.GetPreserveSource(),
		"override_on_conflict": ddAttribute.GetOverrideOnConflict(),
	}
}

func buildTerraformArithmeticProcessor(ddArithmetic datadog.ArithmeticProcessor) map[string]interface{} {
	return map[string]interface{}{
		"target":             ddArithmetic.GetTarget(),
		"is_replace_missing": ddArithmetic.GetIsReplaceMissing(),
		"expression":         ddArithmetic.GetExpression(),
	}
}

func buildTerraformFilter(ddFilter *datadog.FilterConfiguration) []map[string]interface{} {
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

func buildDatadogProcessors(tfProcessors []interface{}) ([]datadog.LogsProcessor, error) {
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
	return ddProcessors, nil
}

func buildDatadogProcessor(ddProcessorType string, tfProcessor map[string]interface{}) (datadog.LogsProcessor, error) {
	var ddProcessor = datadog.LogsProcessor{}
	var err error
	switch ddProcessorType {
	case datadog.ArithmeticProcessorType:
		ddProcessor.Definition = buildDatadogArithmeticProcessor(tfProcessor)
	case datadog.AttributeRemapperType:
		ddProcessor.Definition = buildDatadogAttributeRemapper(tfProcessor)
	case datadog.CategoryProcessorType:
		ddProcessor.Definition = buildDatadogCategoryProcessor(tfProcessor)
	case datadog.DateRemapperType,
		datadog.MessageRemapperType,
		datadog.ServiceRemapperType,
		datadog.StatusRemapperType,
		datadog.TraceIdRemapperType:
		ddProcessor.Definition = buildDatadogSourceRemapper(tfProcessor)
	case datadog.GrokParserType:
		ddProcessor.Definition = buildDatadogGrokParser(tfProcessor)
	case datadog.NestedPipelineType:
		ddProcessor.Definition, err = buildDatadogNestedPipeline(tfProcessor)
	case datadog.UrlParserType:
		ddProcessor.Definition = buildDatadogUrlParser(tfProcessor)
	case datadog.UserAgentParserType:
		ddProcessor.Definition = buildDatadogUserAgentParser(tfProcessor)
	default:
		err = fmt.Errorf("failed to recoginize processor type: %s", ddProcessorType)
	}
	if tfName, exists := tfProcessor["name"].(string); exists {
		ddProcessor.SetName(tfName)
	}
	if tfIsEnabled, exists := tfProcessor["is_enabled"].(bool); exists {
		ddProcessor.SetIsEnabled(tfIsEnabled)
	}
	ddProcessor.SetType(ddProcessorType)
	return ddProcessor, err
}

func buildDatadogUrlParser(tfProcessor map[string]interface{}) datadog.UrlParser {
	ddUrlParser := datadog.UrlParser{}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddUrlParser.Sources = ddSources
	}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddUrlParser.SetTarget(tfTarget)
	}
	if tfNormalizeEndingSlashes, exists := tfProcessor["normalize_ending_slashes"].(bool); exists {
		ddUrlParser.SetNormalizeEndingSlashes(tfNormalizeEndingSlashes)
	}
	return ddUrlParser
}

func buildDatadogUserAgentParser(tfProcessor map[string]interface{}) datadog.UserAgentParser {
	ddUserAgentParser := datadog.UserAgentParser{}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddUserAgentParser.Sources = ddSources
	}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddUserAgentParser.SetTarget(tfTarget)
	}
	if tfIsEncoded, exists := tfProcessor["is_encoded"].(bool); exists {
		ddUserAgentParser.SetIsEncoded(tfIsEncoded)
	}
	return ddUserAgentParser
}

func buildDatadogNestedPipeline(tfProcessor map[string]interface{}) (datadog.NestedPipeline, error) {
	ddNestedPipeline := datadog.NestedPipeline{}
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
	return ddNestedPipeline, nil
}

func buildDatadogGrokParser(tfProcessor map[string]interface{}) datadog.GrokParser {
	ddGrokParser := datadog.GrokParser{}
	if tfSource, exists := tfProcessor["source"].(string); exists {
		ddGrokParser.SetSource(tfSource)
	}
	if tfGrok, exists := tfProcessor["grok"].([]interface{}); exists && len(tfGrok) > 0 {
		ddGrok := datadog.GrokRule{}
		tfGrokRule := tfGrok[0].(map[string]interface{})
		if tfSupportRule, exist := tfGrokRule["support_rules"].(string); exist {
			ddGrok.SetSupportRules(tfSupportRule)
		}
		if tfMatchRule, exist := tfGrokRule["match_rules"].(string); exist {
			ddGrok.SetMatchRules(tfMatchRule)
		}
		ddGrokParser.GrokRule = &ddGrok
	}
	return ddGrokParser
}

func buildDatadogSourceRemapper(tfProcessor map[string]interface{}) datadog.SourceRemapper {
	ddSourceRemapper := datadog.SourceRemapper{}
	if ddSources := buildDatadogSources(tfProcessor); ddSources != nil {
		ddSourceRemapper.Sources = ddSources
	}
	return ddSourceRemapper
}

func buildDatadogCategoryProcessor(tfProcessor map[string]interface{}) datadog.CategoryProcessor {
	ddCategory := datadog.CategoryProcessor{}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddCategory.SetTarget(tfTarget)
	}
	if tfCategories, exists := tfProcessor["category"].([]interface{}); exists {
		ddCategories := make([]datadog.Category, len(tfCategories))
		for i, tfC := range tfCategories {
			tfCategory := tfC.(map[string]interface{})
			ddCategory := datadog.Category{}
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
	return ddCategory
}

func buildDatadogAttributeRemapper(tfProcessor map[string]interface{}) datadog.AttributeRemapper {
	ddAttribute := datadog.AttributeRemapper{}
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

func buildDatadogArithmeticProcessor(tfProcessor map[string]interface{}) datadog.ArithmeticProcessor {
	ddArithmetic := datadog.ArithmeticProcessor{}
	if tfTarget, exists := tfProcessor["target"].(string); exists {
		ddArithmetic.SetTarget(tfTarget)
	}
	if tfExpression, exists := tfProcessor["expression"].(string); exists {
		ddArithmetic.SetExpression(tfExpression)
	}
	if tfIsReplaceMissing, exists := tfProcessor["is_replace_missing"].(bool); exists {
		ddArithmetic.SetIsReplaceMissing(tfIsReplaceMissing)
	}
	return ddArithmetic
}

func buildDatadogFilter(tfFilter map[string]interface{}) datadog.FilterConfiguration {
	ddFilter := datadog.FilterConfiguration{}
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
	processorsSchema[tfArithmeticProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: arithmeticProcessor,
		},
	}
	processorsSchema[tfAttributeRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: attributeRemapper,
		},
	}
	processorsSchema[tfCategoryProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: categoryProcessor,
		},
	}
	processorsSchema[tfDateRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: sourceRemapper,
		},
	}
	processorsSchema[tfGrokParserProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: grokParser,
		},
	}
	processorsSchema[tfMessageRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: sourceRemapper,
		},
	}
	processorsSchema[tfServiceRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: sourceRemapper,
		},
	}
	processorsSchema[tfStatusRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: sourceRemapper,
		},
	}
	processorsSchema[tfTraceIdRemapperProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: sourceRemapper,
		},
	}
	processorsSchema[tfUrlParserProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: urlParser,
		},
	}
	processorsSchema[tfUserAgentParserProcessor] = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: userAgentParser,
		},
	}
	return processorsSchema
}
