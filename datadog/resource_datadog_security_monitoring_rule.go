package datadog

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
)

func resourceDatadogSecurityMonitoringRule() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule use `datadog_security_default_rule` instead.",
		Create:      resourceDatadogSecurityMonitoringRuleCreate,
		Read:        resourceDatadogSecurityMonitoringRuleRead,
		Update:      resourceDatadogSecurityMonitoringRuleUpdate,
		Delete:      resourceDatadogSecurityMonitoringRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
						Type:         schema.TypeString,
						ValidateFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
						Required:     true,
						Description:  "Severity of the Security Signal.",
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

		"options": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Options on rules.",

			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"evaluation_window": {
						Type:         schema.TypeInt,
						ValidateFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleEvaluationWindowFromValue),
						Required:     true,
						Description:  "A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time.",
					},

					"keep_alive": {
						Type:         schema.TypeInt,
						ValidateFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleKeepAliveFromValue),
						Required:     true,
						Description:  "Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window.",
					},

					"max_signal_duration": {
						Type:         schema.TypeInt,
						ValidateFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleMaxSignalDurationFromValue),
						Required:     true,
						Description:  "A signal will “close” regardless of the query being matched once the time exceeds the maximum duration. This time is calculated from the first seen timestamp.",
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
						Type:         schema.TypeString,
						ValidateFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
						Optional:     true,
						Description:  "The aggregation type.",
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
	}
}

func resourceDatadogSecurityMonitoringRuleCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleCreate, err := buildCreatePayload(d)
	if err != nil {
		return err
	}
	response, _, err := datadogClientV2.SecurityMonitoringApi.CreateSecurityMonitoringRule(authV2).Body(ruleCreate).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error creating security monitoring rule")
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

	if v, ok := d.GetOk("options"); ok {
		tfOptionsList := v.([]interface{})
		payloadOptions := buildCreatePayloadOptions(tfOptionsList)
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

func buildCreatePayloadOptions(tfOptionsList []interface{}) *datadogV2.SecurityMonitoringRuleOptions {
	payloadOptions := datadogV2.NewSecurityMonitoringRuleOptions()
	var tfOptions map[string]interface{}
	if tfOptionsList[0] == nil {
		tfOptions = make(map[string]interface{})
	} else {
		tfOptions = tfOptionsList[0].(map[string]interface{})
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
	return payloadOptions
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

func resourceDatadogSecurityMonitoringRuleRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	id := d.Id()
	ruleResponse, httpResponse, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, id).Execute()
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
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

	options := make(map[string]interface{})
	getOptions := ruleResponse.GetOptions()
	if evaluationWindow, ok := getOptions.GetEvaluationWindowOk(); ok {
		options["evaluation_window"] = *evaluationWindow
	}
	if keepAlive, ok := getOptions.GetKeepAliveOk(); ok {
		options["keep_alive"] = *keepAlive
	}
	if maxSignalDuration, ok := getOptions.GetMaxSignalDurationOk(); ok {
		options["max_signal_duration"] = *maxSignalDuration
	}
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
}

func resourceDatadogSecurityMonitoringRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleUpdate := buildUpdatePayload(d)
	response, _, err := datadogClientV2.SecurityMonitoringApi.UpdateSecurityMonitoringRule(authV2, d.Id()).Body(ruleUpdate).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error updating security monitoring rule")
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

	if v, ok := d.GetOk("message"); ok {
		message := v.(string)
		payload.Message = &message
	}

	if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		payload.Name = &name
	}

	if v, ok := d.GetOk("options"); ok {
		payloadOptions := datadogV2.NewSecurityMonitoringRuleOptions()
		tfOptions := v.([]interface{})
		options := tfOptions[0].(map[string]interface{})
		if v, ok := options["evaluation_window"]; ok {
			evaluationWindow := datadogV2.SecurityMonitoringRuleEvaluationWindow(v.(int))
			payloadOptions.EvaluationWindow = &evaluationWindow
		}
		if v, ok := options["keep_alive"]; ok {
			keepAlive := datadogV2.SecurityMonitoringRuleKeepAlive(v.(int))
			payloadOptions.KeepAlive = &keepAlive
		}
		if v, ok := options["max_signal_duration"]; ok {
			maxSignalDuration := datadogV2.SecurityMonitoringRuleMaxSignalDuration(v.(int))
			payloadOptions.MaxSignalDuration = &maxSignalDuration
		}
		payload.Options = payloadOptions
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

	return payload
}

func resourceDatadogSecurityMonitoringRuleDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if _, err := datadogClientV2.SecurityMonitoringApi.DeleteSecurityMonitoringRule(authV2, d.Id()).Execute(); err != nil {
		return utils.TranslateClientError(err, "error deleting security monitoring rule")
	}

	return nil
}
