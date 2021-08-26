package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogSecurityMonitoringRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule use `datadog_security_default_rule` instead.",
		CreateContext: resourceDatadogSecurityMonitoringRuleCreate,
		ReadContext:   resourceDatadogSecurityMonitoringRuleRead,
		UpdateContext: resourceDatadogSecurityMonitoringRuleUpdate,
		DeleteContext: resourceDatadogSecurityMonitoringRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: datadogSecurityMonitoringRuleSchema(),
	}
}

func datadogSecurityMonitoringRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"case": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Cases for generating signals.",
			MaxItems:    5,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the case.",
					},
					"condition": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "A rule case contains logical operations (`>`,`>=`, `&&`, `||`) to determine if a signal should be generated based on the event counts in the previously defined queries.",
					},
					"notifications": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Notification targets for each rule case.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"status": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
						Required:         true,
						Description:      "Severity of the Security Signal.",
					},
				},
			},
		},

		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether the rule is enabled.",
		},

		"message": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Message for generated signals.",
		},

		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the rule.",
		},

		"has_extended_title": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether the notifications include the triggering group-by values in their title.",
		},

		"options": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Options on rules.",

			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"detection_method": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleDetectionMethodFromValue),
						Optional:         true,
						Description:      "The detection method.",
						Default:          "threshold",
					},

					"evaluation_window": {
						Type:             schema.TypeInt,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleEvaluationWindowFromValue),
						Required:         true,
						Description:      "A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time.",
					},

					"keep_alive": {
						Type:             schema.TypeInt,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleKeepAliveFromValue),
						Required:         true,
						Description:      "Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window.",
					},

					"max_signal_duration": {
						Type:             schema.TypeInt,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleMaxSignalDurationFromValue),
						Required:         true,
						Description:      "A signal will “close” regardless of the query being matched once the time exceeds the maximum duration. This time is calculated from the first seen timestamp.",
					},

					"new_value_options": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "New value rules specific options.",

						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"learning_duration": {
									Type:             schema.TypeInt,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleNewValueOptionsLearningDurationFromValue),
									Required:         true,
									Description:      "The duration in days during which values are learned, and after which signals will be generated for values that weren't learned. If set to 0, a signal will be generated for all new values after the first value is learned.",
								},
								"forget_after": {
									Type:             schema.TypeInt,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleNewValueOptionsForgetAfterFromValue),
									Required:         true,
									Description:      "The duration in days after which a learned value is forgotten.",
								},
							},
						},
					},
				},
			},
		},

		"query": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Queries for selecting logs which are part of the rule.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aggregation": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
						Optional:         true,
						Description:      "The aggregation type.",
					},
					"distinct_fields": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Field for which the cardinality is measured. Sent as an array.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"group_by_fields": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Fields to group by.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"metric": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The target field to aggregate over when using the sum or max aggregations.",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the query.",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query to run on logs.",
					},
				},
			},
		},

		"tags": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Tags for generated signals.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},

		"filter": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Additional queries to filter matched events before they are processed.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for selecting logs to apply the filtering action.",
					},
					"action": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      "The type of filtering action.",
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringFilterActionFromValue),
					},
				},
			},
		},
	}
}

func resourceDatadogSecurityMonitoringRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleCreate, err := buildCreatePayload(d)
	if err != nil {
		return diag.FromErr(err)
	}
	response, httpResponse, err := datadogClientV2.SecurityMonitoringApi.CreateSecurityMonitoringRule(authV2, ruleCreate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating security monitoring rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.GetId())

	return nil
}

func buildCreatePayload(d *schema.ResourceData) (datadogV2.SecurityMonitoringRuleCreatePayload, error) {
	payload := datadogV2.SecurityMonitoringRuleCreatePayload{}
	payload.Cases = buildCreatePayloadCases(d)

	payload.IsEnabled = d.Get("enabled").(bool)
	payload.Message = d.Get("message").(string)
	payload.Name = d.Get("name").(string)
	payload.SetHasExtendedTitle(d.Get("has_extended_title").(bool))

	if v, ok := d.GetOk("options"); ok {
		tfOptionsList := v.([]interface{})
		payloadOptions := buildPayloadOptions(tfOptionsList)
		payload.Options = *payloadOptions
	}

	payload.Queries = buildCreatePayloadQueries(d)

	if v, ok := d.GetOk("tags"); ok {
		tfTags := v.([]interface{})
		tags := make([]string, len(tfTags))
		for i, value := range tfTags {
			tags[i] = value.(string)
		}
		payload.Tags = &tags
	}

	if v, ok := d.GetOk("filter"); ok {
		tfFilterList := v.([]interface{})
		payload.SetFilters(buildPayloadFilters(tfFilterList))
	}

	return payload, nil
}

func buildCreatePayloadCases(d *schema.ResourceData) []datadogV2.SecurityMonitoringRuleCaseCreate {
	tfCases := d.Get("case").([]interface{})
	payloadCases := make([]datadogV2.SecurityMonitoringRuleCaseCreate, len(tfCases))

	for idx, ruleCaseIf := range tfCases {
		ruleCase := ruleCaseIf.(map[string]interface{})
		status := datadogV2.SecurityMonitoringRuleSeverity(ruleCase["status"].(string))
		structRuleCase := datadogV2.NewSecurityMonitoringRuleCaseCreate(status)
		if v, ok := ruleCase["name"]; ok {
			name := v.(string)
			structRuleCase.SetName(name)
		}
		if v, ok := ruleCase["condition"]; ok {
			condition := v.(string)
			structRuleCase.SetCondition(condition)
		}
		if v, ok := ruleCase["notifications"]; ok {
			tfNotifications := v.([]interface{})
			notifications := make([]string, len(tfNotifications))
			for i, value := range tfNotifications {
				notifications[i] = value.(string)
			}
			structRuleCase.SetNotifications(notifications)
		}
		payloadCases[idx] = *structRuleCase
	}
	return payloadCases
}

func buildPayloadOptions(tfOptionsList []interface{}) *datadogV2.SecurityMonitoringRuleOptions {
	payloadOptions := datadogV2.NewSecurityMonitoringRuleOptions()
	tfOptions := extractMapFromInterface(tfOptionsList)

	if v, ok := tfOptions["detection_method"]; ok {
		detectionMethod := datadogV2.SecurityMonitoringRuleDetectionMethod(v.(string))
		payloadOptions.DetectionMethod = &detectionMethod
	}
	if v, ok := tfOptions["evaluation_window"]; ok {
		evaluationWindow := datadogV2.SecurityMonitoringRuleEvaluationWindow(v.(int))
		payloadOptions.EvaluationWindow = &evaluationWindow
	}
	if v, ok := tfOptions["keep_alive"]; ok {
		keepAlive := datadogV2.SecurityMonitoringRuleKeepAlive(v.(int))
		payloadOptions.KeepAlive = &keepAlive
	}
	if v, ok := tfOptions["max_signal_duration"]; ok {
		maxSignalDuration := datadogV2.SecurityMonitoringRuleMaxSignalDuration(v.(int))
		payloadOptions.MaxSignalDuration = &maxSignalDuration
	}

	if v, ok := tfOptions["new_value_options"]; ok {
		tfNewValueOptionsList := v.([]interface{})
		if payloadNewValueOptions, ok := buildPayloadNewValueOptions(tfNewValueOptionsList); ok {
			payloadOptions.NewValueOptions = payloadNewValueOptions
		}
	}

	return payloadOptions
}

func buildPayloadNewValueOptions(tfOptionsList []interface{}) (*datadogV2.SecurityMonitoringRuleNewValueOptions, bool) {
	payloadNewValueRulesOptions := datadogV2.NewSecurityMonitoringRuleNewValueOptions()
	tfOptions := extractMapFromInterface(tfOptionsList)
	hasPayload := false
	if v, ok := tfOptions["learning_duration"]; ok {
		hasPayload = true
		learningDuration := datadogV2.SecurityMonitoringRuleNewValueOptionsLearningDuration(v.(int))
		payloadNewValueRulesOptions.LearningDuration = &learningDuration
	}
	if v, ok := tfOptions["forget_after"]; ok {
		hasPayload = true
		forgetAfter := datadogV2.SecurityMonitoringRuleNewValueOptionsForgetAfter(v.(int))
		payloadNewValueRulesOptions.ForgetAfter = &forgetAfter
	}
	return payloadNewValueRulesOptions, hasPayload
}

func extractMapFromInterface(tfOptionsList []interface{}) map[string]interface{} {
	var tfOptions map[string]interface{}
	if len(tfOptionsList) == 0 || tfOptionsList[0] == nil {
		tfOptions = make(map[string]interface{})
	} else {
		tfOptions = tfOptionsList[0].(map[string]interface{})
	}
	return tfOptions
}

func buildCreatePayloadQueries(d *schema.ResourceData) []datadogV2.SecurityMonitoringRuleQueryCreate {
	tfQueries := d.Get("query").([]interface{})
	payloadQueries := make([]datadogV2.SecurityMonitoringRuleQueryCreate, len(tfQueries))
	for idx, tfQuery := range tfQueries {
		query := tfQuery.(map[string]interface{})
		payloadQuery := datadogV2.SecurityMonitoringRuleQueryCreate{}

		if v, ok := query["aggregation"]; ok {
			aggregation := datadogV2.SecurityMonitoringRuleQueryAggregation(v.(string))
			payloadQuery.Aggregation = &aggregation
		}

		if v, ok := query["group_by_fields"]; ok {
			tfGroupByFields := v.([]interface{})
			groupByFields := make([]string, len(tfGroupByFields))
			for i, value := range tfGroupByFields {
				groupByFields[i] = value.(string)
			}
			payloadQuery.GroupByFields = &groupByFields
		}

		if v, ok := query["distinct_fields"]; ok {
			tfDistinctFields := v.([]interface{})
			distinctFields := make([]string, len(tfDistinctFields))
			for i, value := range tfDistinctFields {
				distinctFields[i] = value.(string)
			}
			payloadQuery.DistinctFields = &distinctFields
		}

		if v, ok := query["metric"]; ok {
			metric := v.(string)
			payloadQuery.Metric = &metric
		}

		if v, ok := query["name"]; ok {
			name := v.(string)
			payloadQuery.Name = &name
		}

		payloadQuery.Query = query["query"].(string)

		payloadQueries[idx] = payloadQuery
	}
	return payloadQueries
}

func buildPayloadFilters(tfFilters []interface{}) []datadogV2.SecurityMonitoringFilter {
	payloadFilters := make([]datadogV2.SecurityMonitoringFilter, len(tfFilters))
	for idx, tfFilter := range tfFilters {
		filter := tfFilter.(map[string]interface{})
		payloadFilter := datadogV2.SecurityMonitoringFilter{}

		action := datadogV2.SecurityMonitoringFilterAction(filter["action"].(string))
		payloadFilter.SetAction(action)

		payloadFilter.SetQuery(filter["query"].(string))

		payloadFilters[idx] = payloadFilter
	}
	return payloadFilters
}

func resourceDatadogSecurityMonitoringRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	id := d.Id()
	ruleResponse, httpResponse, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, id)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := utils.CheckForUnparsed(ruleResponse); err != nil {
		return diag.FromErr(err)
	}

	updateResourceDataFromResponse(d, ruleResponse)

	return nil
}

func updateResourceDataFromResponse(d *schema.ResourceData, ruleResponse datadogV2.SecurityMonitoringRuleResponse) {
	ruleCases := make([]interface{}, len(ruleResponse.GetCases()))
	for idx := range ruleResponse.GetCases() {
		ruleCase := make(map[string]interface{})
		responseRuleCase := ruleResponse.GetCases()[idx]

		if name, ok := responseRuleCase.GetNameOk(); ok {
			ruleCase["name"] = *name
		}
		if condition, ok := responseRuleCase.GetConditionOk(); ok {
			ruleCase["condition"] = *condition
		}
		if notifications, ok := responseRuleCase.GetNotificationsOk(); ok {
			ruleCase["notifications"] = *notifications
		}
		ruleCase["status"] = responseRuleCase.GetStatus()

		ruleCases[idx] = ruleCase
	}

	d.Set("case", ruleCases)
	d.Set("message", ruleResponse.GetMessage())
	d.Set("name", ruleResponse.GetName())
	d.Set("has_extended_title", ruleResponse.GetHasExtendedTitle())
	d.Set("enabled", ruleResponse.GetIsEnabled())

	options := extractTfOptions(ruleResponse.GetOptions())

	d.Set("options", []map[string]interface{}{options})

	ruleQueries := make([]map[string]interface{}, len(ruleResponse.GetQueries()))
	for idx := range ruleResponse.GetQueries() {
		ruleQuery := make(map[string]interface{})
		responseRuleQuery := ruleResponse.GetQueries()[idx]
		if aggregation, ok := responseRuleQuery.GetAggregationOk(); ok {
			ruleQuery["aggregation"] = *aggregation
		}
		if distinctFields, ok := responseRuleQuery.GetDistinctFieldsOk(); ok {
			ruleQuery["distinct_fields"] = *distinctFields
		}
		if groupByFields, ok := responseRuleQuery.GetGroupByFieldsOk(); ok {
			ruleQuery["group_by_fields"] = *groupByFields
		}
		if metric, ok := responseRuleQuery.GetMetricOk(); ok {
			ruleQuery["metric"] = *metric
		}
		if name, ok := responseRuleQuery.GetNameOk(); ok {
			ruleQuery["name"] = *name
		}
		if query, ok := responseRuleQuery.GetQueryOk(); ok {
			ruleQuery["query"] = *query
		}

		ruleQueries[idx] = ruleQuery
	}
	d.Set("query", ruleQueries)

	if _, ok := ruleResponse.GetFiltersOk(); ok {
		filters := extractFiltersFromRuleResponse(ruleResponse)
		d.Set("filter", filters)
	}

}

func extractFiltersFromRuleResponse(ruleResponse datadogV2.SecurityMonitoringRuleResponse) []interface{} {
	filters := make([]interface{}, len(ruleResponse.GetFilters()))
	for idx, responseFilter := range ruleResponse.GetFilters() {
		filter := make(map[string]interface{})
		if query, ok := responseFilter.GetQueryOk(); ok {
			filter["query"] = *query
		}
		if action, ok := responseFilter.GetActionOk(); ok {
			filter["action"] = *action
		}

		filters[idx] = filter
	}
	return filters
}

func extractTfOptions(options datadogV2.SecurityMonitoringRuleOptions) map[string]interface{} {
	tfOptions := make(map[string]interface{})
	if evaluationWindow, ok := options.GetEvaluationWindowOk(); ok {
		tfOptions["evaluation_window"] = *evaluationWindow
	}
	if keepAlive, ok := options.GetKeepAliveOk(); ok {
		tfOptions["keep_alive"] = *keepAlive
	}
	if maxSignalDuration, ok := options.GetMaxSignalDurationOk(); ok {
		tfOptions["max_signal_duration"] = *maxSignalDuration
	}
	if detectionMethod, ok := options.GetDetectionMethodOk(); ok {
		tfOptions["detection_method"] = *detectionMethod
	}
	if newValueOptions, ok := options.GetNewValueOptionsOk(); ok {
		tfNewValueOptions := make(map[string]interface{})
		tfNewValueOptions["forget_after"] = int(newValueOptions.GetForgetAfter())
		tfNewValueOptions["learning_duration"] = int(newValueOptions.GetLearningDuration())
		tfOptions["new_value_options"] = []map[string]interface{}{tfNewValueOptions}
	}
	return tfOptions
}

func resourceDatadogSecurityMonitoringRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleUpdate := buildUpdatePayload(d)
	response, httpResponse, err := datadogClientV2.SecurityMonitoringApi.UpdateSecurityMonitoringRule(authV2, d.Id(), ruleUpdate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating security monitoring rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	updateResourceDataFromResponse(d, response)

	return nil
}

func buildUpdatePayload(d *schema.ResourceData) datadogV2.SecurityMonitoringRuleUpdatePayload {
	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}

	tfCases := d.Get("case").([]interface{})
	payloadCases := make([]datadogV2.SecurityMonitoringRuleCase, len(tfCases))

	for idx, tfRuleCase := range tfCases {
		structRuleCase := datadogV2.SecurityMonitoringRuleCase{}

		ruleCase := tfRuleCase.(map[string]interface{})
		status := datadogV2.SecurityMonitoringRuleSeverity(ruleCase["status"].(string))
		structRuleCase.Status = &status

		if name, ok := ruleCase["name"]; ok {
			structRuleCase.SetName(name.(string))
		}
		if condition, ok := ruleCase["condition"]; ok {
			structRuleCase.SetCondition(condition.(string))
		}
		if v, ok := ruleCase["notifications"]; ok {
			tfNotifications := v.([]interface{})
			notifications := make([]string, len(tfNotifications))
			for i, value := range tfNotifications {
				notifications[i] = value.(string)
			}
			structRuleCase.SetNotifications(notifications)
		}
		payloadCases[idx] = structRuleCase
	}
	payload.Cases = &payloadCases

	payload.SetIsEnabled(d.Get("enabled").(bool))
	payload.SetHasExtendedTitle(d.Get("has_extended_title").(bool))

	if v, ok := d.GetOk("message"); ok {
		message := v.(string)
		payload.Message = &message
	}

	if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		payload.Name = &name
	}

	if v, ok := d.GetOk("options"); ok {
		payload.Options = buildPayloadOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("query"); ok {
		tfQueries := v.([]interface{})
		payloadQueries := make([]datadogV2.SecurityMonitoringRuleQuery, len(tfQueries))
		for idx, tfQuery := range tfQueries {
			query := tfQuery.(map[string]interface{})
			payloadQuery := datadogV2.SecurityMonitoringRuleQuery{}

			if v, ok := query["aggregation"]; ok {
				aggregation := datadogV2.SecurityMonitoringRuleQueryAggregation(v.(string))
				payloadQuery.Aggregation = &aggregation
			}

			if v, ok := query["group_by_fields"]; ok {
				tfGroupByFields := v.([]interface{})
				groupByFields := make([]string, len(tfGroupByFields))
				for i, value := range tfGroupByFields {
					groupByFields[i] = value.(string)
				}
				payloadQuery.GroupByFields = &groupByFields
			}

			if v, ok := query["distinct_fields"]; ok {
				tfDistinctFields := v.([]interface{})
				distinctFields := make([]string, len(tfDistinctFields))
				for i, field := range tfDistinctFields {
					distinctFields[i] = field.(string)
				}
				payloadQuery.DistinctFields = &distinctFields
			}

			if v, ok := query["metric"]; ok {
				metric := v.(string)
				payloadQuery.Metric = &metric
			}

			if v, ok := query["name"]; ok {
				name := v.(string)
				payloadQuery.Name = &name
			}

			queryQuery := query["query"].(string)
			payloadQuery.Query = &queryQuery

			payloadQueries[idx] = payloadQuery
		}
		payload.Queries = &payloadQueries
	}

	if v, ok := d.GetOk("tags"); ok {
		tfTags := v.([]interface{})
		tags := make([]string, len(tfTags))
		for i, value := range tfTags {
			tags[i] = value.(string)
		}
		payload.Tags = &tags
	}

	if v, ok := d.GetOk("filter"); ok {
		tfFilters := v.([]interface{})
		payload.SetFilters(buildPayloadFilters(tfFilters))
	}

	return payload
}

func resourceDatadogSecurityMonitoringRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if httpResponse, err := datadogClientV2.SecurityMonitoringApi.DeleteSecurityMonitoringRule(authV2, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting security monitoring rule")
	}

	return nil
}
