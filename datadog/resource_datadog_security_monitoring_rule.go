package datadog

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
)

func resourceDatadogSecurityMonitoringRule() *schema.Resource {
	return &schema.Resource{
		Exists: resourceDatadogSecurityMonitoringRuleExists,
		Create: resourceDatadogSecurityMonitoringRuleCreate,
		Read:   resourceDatadogSecurityMonitoringRuleRead,
		Update: resourceDatadogSecurityMonitoringRuleUpdate,
		Delete: resourceDatadogSecurityMonitoringRuleDelete,
		//CustomizeDiff: resourceDatadogSecurityMonitoringCustomizeDiff,
		//Importer: &schema.ResourceImporter{
		//	State: resourceDatadogSecurityMonitoringImport,
		//},

		Schema: map[string]*schema.Schema{
			"case": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Cases for generating signals.",
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
							Type:        schema.TypeString,
							Required:    true,
							Description: "Severity of the Security Signal.",
						},
					},
				},
			},

			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
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
				MaxItems: 	 1,
				Description: "Options on rules.",

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"evaluation_window": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time.",
						},

						"keep_alive": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window.",
						},

						"max_signal_duration": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "A signal will “close” regardless of the query being matched once the time exceeds the maximum duration. This time is calculated from the first seen timestamp.",
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
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The aggregation type.",
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
		},
	}
}

//func resourceDatadogSecurityMonitoringImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
//	return nil, nil
//}

func resourceDatadogSecurityMonitoringRuleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	id := d.Id()
	_, httpResponse, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, id).Execute()
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return false, nil
		}
		return false, translateClientError(err, "error checking security monitoring rule exists")
	}
	return true, nil
}

func resourceDatadogSecurityMonitoringRuleCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleCreate := buildCreatePayload(d)
	response, _, err := datadogClientV2.SecurityMonitoringApi.CreateSecurityMonitoringRule(authV2).Body(ruleCreate).Execute()
	if err != nil {
		return translateClientError(err, "error creating security monitoring rule")
	}

	d.SetId(response.GetId())

	return nil
}

func buildCreatePayload(d *schema.ResourceData) datadogV2.SecurityMonitoringRuleCreatePayload {
	payload := datadogV2.SecurityMonitoringRuleCreatePayload{}
	cases := d.Get("case").([]interface{})
	payloadCases := make([]datadogV2.SecurityMonitoringRuleCaseCreate, len(cases))

	for idx, ruleCaseIf := range cases {
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
	payload.Cases = payloadCases

	if v, ok := d.GetOk("is_enabled"); ok {
		payload.IsEnabled = v.(bool)
	} else {
		payload.IsEnabled = true
	}

	payload.Message = d.Get("message").(string)
	payload.Name = d.Get("name").(string)

	if v, ok := d.GetOk("options"); ok {
		payloadOptions := datadogV2.NewSecurityMonitoringRuleOptions()
		tfOptions := (v.([]interface{})[0]).(map[string]interface{})
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
		payload.Options = *payloadOptions
	}

	tfQueries := d.Get("query").([]interface{})
	payloadQueries := make([]datadogV2.SecurityMonitoringRuleQueryCreate, len(tfQueries))
	for idx, tfQuery := range tfQueries {
		query := tfQuery.(map[string]interface{})
		payloadQuery := datadogV2.SecurityMonitoringRuleQueryCreate{}

		if v, ok := query["aggregation"]; ok {
			aggregation := datadogV2.SecurityMonitoringRuleQueryAggregation(v.(string))
			payloadQuery.Aggregation = &aggregation
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
	payload.Queries = payloadQueries

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

func resourceDatadogSecurityMonitoringRuleRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	id := d.Id()
	ruleResponse, _, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, id).Execute()
	if err != nil {
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
	d.Set("options", options)

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
		if query, ok := responseRuleQuery.GetNameOk(); ok {
			ruleQuery["query"] = *query
		}

		ruleQueries[idx] = ruleQuery
	}
}

func resourceDatadogSecurityMonitoringRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleUpdate := buildUpdatePayload(d)
	response, _, err := datadogClientV2.SecurityMonitoringApi.UpdateSecurityMonitoringRule(authV2, d.Id()).Body(ruleUpdate).Execute()
	if err != nil {
		return translateClientError(err, "error updating security monitoring rule")
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

	if v, ok := d.GetOk("isEnabled"); ok {
		isEnabled := v.(bool)
		payload.IsEnabled = &isEnabled
	}

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

	_, err := datadogClientV2.SecurityMonitoringApi.DeleteSecurityMonitoringRule(authV2, d.Id()).Execute()

	if err != nil {
		return translateClientError(err, "error deleting security monitoring rule")
	}

	return nil
}

//func resourceDatadogSecurityMonitoringCustomizeDiff(d *schema.ResourceDiff, meta interface{}) error {
//	return nil
//}

