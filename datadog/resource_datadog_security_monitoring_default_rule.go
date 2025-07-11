package datadog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogSecurityMonitoringDefaultRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Security Monitoring Rule API resource for default rules. It can only be imported, you can't create a default rule.",
		CreateContext: resourceDatadogSecurityMonitoringDefaultRuleCreate,
		ReadContext:   resourceDatadogSecurityMonitoringDefaultRuleRead,
		UpdateContext: resourceDatadogSecurityMonitoringDefaultRuleUpdate,
		DeleteContext: resourceDatadogSecurityMonitoringDefaultRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"case": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Cases of the rule, this is used to update notifications.",
					MaxItems:    10,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"status": {
								Type:             schema.TypeString,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
								Required:         true,
								Description:      "Status of the rule case to match.",
							},
							"custom_status": {
								Type:             schema.TypeString,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
								Optional:         true,
								Description:      "Status of the rule case to override.",
							},
							"notifications": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Notification targets for each rule case.",
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},

				"query": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Queries for selecting logs which are part of the rule.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"agent_rule": {
								Type:        schema.TypeList,
								Deprecated:  "`agent_rule` has been deprecated in favor of new Agent Rule resource.",
								Optional:    true,
								Description: "**Deprecated**. It won't be applied anymore.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"agent_rule_id": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "**Deprecated**. It won't be applied anymore.",
										},
										"expression": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "**Deprecated**. It won't be applied anymore.",
										},
									},
								},
							},
							"aggregation": {
								Type:             schema.TypeString,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
								Optional:         true,
								Description:      "The aggregation type. For Signal Correlation rules, it must be event_count.",
								Default:          datadogV2.SECURITYMONITORINGRULEQUERYAGGREGATION_COUNT,
							},
							"distinct_fields": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Field for which the cardinality is measured. Sent as an array.",
								Elem: &schema.Schema{
									Type:             schema.TypeString,
									ValidateDiagFunc: validators.ValidateNonEmptyStrings,
								},
							},
							"group_by_fields": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Fields to group by.",
								Elem: &schema.Schema{
									Type:             schema.TypeString,
									ValidateDiagFunc: validators.ValidateNonEmptyStrings,
								},
							},
							"data_source": {
								Type:             schema.TypeString,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringStandardDataSourceFromValue),
								Optional:         true,
								Description:      "Source of events.",
								Default:          datadogV2.SECURITYMONITORINGSTANDARDDATASOURCE_LOGS,
							},
							"metric": {
								Type:        schema.TypeString,
								Deprecated:  "Configure `metrics` instead. This attribute will be removed in the next major version of the provider.",
								Optional:    true,
								Description: "The target field to aggregate over when using the `sum`, `max`, or `geo_data` aggregations.",
							},
							"metrics": {
								Type:        schema.TypeList,
								Computed:    true,
								Optional:    true,
								Description: "Group of target fields to aggregate over when using the `sum`, `max`, `geo_data`, or `new_value` aggregations. The `sum`, `max`, and `geo_data` aggregations only accept one value in this list, whereas the `new_value` aggregation accepts up to five values.",
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Name of the query. Not compatible with `new_value` aggregations.",
							},
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query to run on logs.",
							},
							"custom_query_extension": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Query extension to append to the logs query.",
							},
						},
					},
				},

				"custom_message": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Custom Message (will override default message) for generated signals.",
				},

				"custom_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The name (will override default name) of the rule.",
				},

				"enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Enable the rule.",
				},

				"filter": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Additional queries to filter matched events before they are processed.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"action": {
								Type:             schema.TypeString,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringFilterActionFromValue),
								Required:         true,
								Description:      "The type of filtering action. Allowed enum values: require, suppress",
							},
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for selecting logs to apply the filtering action.",
							},
						},
					},
				},

				"options": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Options on default rules. Note that only a subset of fields can be updated on default rule options.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"decrease_criticality_based_on_env": {
								Type:        schema.TypeBool,
								Optional:    true,
								Default:     false,
								Description: "If true, signals in non-production environments have a lower severity than what is defined by the rule case, which can reduce noise. The decrement is applied when the environment tag of the signal starts with `staging`, `test`, or `dev`. Only available when the rule type is `log_detection`.",
							},
						},
					},
				},

				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The rule type.",
				},

				"custom_tags": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Custom tags for generated signals.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

func securityMonitoringRuleDeprecationWarning(rule securityMonitoringRuleResponseInterface) diag.Diagnostics {
	var diags diag.Diagnostics

	if deprecationTimestampMs, ok := rule.GetDeprecationDateOk(); ok {
		deprecation := time.UnixMilli(*deprecationTimestampMs)

		warning := diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Rule will be deprecated on %s.", deprecation.Format("Jan _2 2006")),
			Detail: "Please consider deleting the associated resource. " +
				"After the depreciation date, the rule will stop triggering signals. " +
				" Moreover, the API will reject any call to update the rule, which might break your Terraform pipeline. " +
				"The Datadog team performs regular audit of all detection rules to maintain high fidelity signal quality. " +
				"We will be replacing this rule with an improved third party detection rule after the depreciation date.",
		}

		diags = append(diags, warning)
	}

	return diags
}

func resourceDatadogSecurityMonitoringDefaultRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(errors.New("cannot create a default rule, please import it first before making changes"))
}

func resourceDatadogSecurityMonitoringDefaultRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()
	ruleResponse, _, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := utils.CheckForUnparsed(ruleResponse); err != nil {
		return diag.FromErr(err)
	}

	rule := ruleResponse.SecurityMonitoringStandardRuleResponse
	if rule == nil {
		return diag.Errorf("signal rule type is not currently supported")
	}

	d.Set("enabled", *rule.IsEnabled)

	if v, ok := d.GetOk("case"); ok {
		tfCasesRaw := v.([]interface{})
		readNotifications := make([][]string, len(tfCasesRaw))
		for i, tfCaseRaw := range tfCasesRaw {
			tfCase := tfCaseRaw.(map[string]interface{})
			var ruleCase *datadogV2.SecurityMonitoringRuleCase
			tfStatus := datadogV2.SecurityMonitoringRuleSeverity(tfCase["status"].(string))
			for _, rc := range rule.GetCases() {
				if *rc.Status == tfStatus {
					ruleCase = &rc
					break
				}
			}
			if ruleCase == nil {
				return diag.FromErr(errors.New("error: no rule case with status " + string(tfStatus)))
			}
			readNotifications[i] = ruleCase.GetNotifications()
		}

		for i, notification := range readNotifications {
			d.Set(fmt.Sprintf("case.%d.notifications", i), notification)
		}
	}

	ruleFilters := make([]map[string]interface{}, len(rule.GetFilters()))
	for idx, responseRuleFilter := range rule.GetFilters() {
		ruleFilters[idx] = map[string]interface{}{
			"action": responseRuleFilter.GetAction(),
			"query":  responseRuleFilter.GetQuery(),
		}
	}

	d.Set("filter", ruleFilters)

	d.Set("type", rule.GetType())

	responseOptions := rule.GetOptions()
	var ruleOptions []map[string]interface{}

	if *rule.Type == datadogV2.SECURITYMONITORINGRULETYPEREAD_LOG_DETECTION {
		ruleOptions = append(ruleOptions, map[string]interface{}{
			"decrease_criticality_based_on_env": responseOptions.GetDecreaseCriticalityBasedOnEnv(),
		})
	}

	d.Set("options", &ruleOptions)

	defaultTags := make(map[string]bool)
	for _, defaultTag := range rule.GetDefaultTags() {
		defaultTags[defaultTag] = true
	}

	customTags := []string{}
	for _, tag := range rule.GetTags() {
		if _, ok := defaultTags[tag]; !ok {
			customTags = append(customTags, tag)
		}
	}

	d.Set("custom_tags", customTags)

	return securityMonitoringRuleDeprecationWarning(rule)
}

func resourceDatadogSecurityMonitoringDefaultRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ruleID := d.Id()

	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, ruleID)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return diag.FromErr(errors.New("default rule does not exist"))
		}

		return utils.TranslateClientErrorDiag(err, httpResponse, "error fetching default rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	rule := response.SecurityMonitoringStandardRuleResponse
	if rule == nil {
		return diag.Errorf("signal rule type is not currently supported")
	}

	if !rule.GetIsDefault() {
		return diag.FromErr(errors.New("rule is not a default rule"))
	}

	ruleUpdate, shouldUpdate, err := buildSecMonDefaultRuleUpdatePayload(rule, d)

	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if shouldUpdate {
		ruleResponse, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().UpdateSecurityMonitoringRule(auth, ruleID, *ruleUpdate)

		if err != nil {
			diags = append(diags, utils.TranslateClientErrorDiag(err, httpResponse, "error updating security monitoring rule on resource creation")...)
		}

		diags = append(diags, securityMonitoringRuleDeprecationWarning(ruleResponse.SecurityMonitoringStandardRuleResponse)...)
	} else {
		diags = append(diags, securityMonitoringRuleDeprecationWarning(rule)...)
	}

	return diags
}

func buildSecMonDefaultRuleUpdatePayload(currentState *datadogV2.SecurityMonitoringStandardRuleResponse, d *schema.ResourceData) (*datadogV2.SecurityMonitoringRuleUpdatePayload, bool, error) {
	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}
	isSignalCorrelation := isSignalCorrelationSchema(d)

	isEnabled := d.Get("enabled").(bool)
	payload.IsEnabled = &isEnabled

	// Track if any changes are detected
	shouldUpdate := false

	// Compare enabled state
	if currentState.GetIsEnabled() != isEnabled {
		shouldUpdate = true
	}

	matchedCases := 0
	modifiedCases := 0
	tfCasesRaw := d.Get("case").([]interface{})

	updatedRuleCase := make([]datadogV2.SecurityMonitoringRuleCase, len(currentState.GetCases()))
	for i, ruleCase := range currentState.GetCases() {

		updatedRuleCase[i] = datadogV2.SecurityMonitoringRuleCase{
			Condition:     currentState.GetCases()[i].Condition,
			Name:          currentState.GetCases()[i].Name,
			Notifications: currentState.GetCases()[i].Notifications,
			Status:        currentState.GetCases()[i].Status,
			CustomStatus:  currentState.GetCases()[i].CustomStatus,
		}

		if tfCase, ok := findRuleCaseForStatus(tfCasesRaw, ruleCase.GetStatus()); ok {

			// Update rule case notifications when rule added to terraform configuration

			matchedCases++

			tfNotificationsRaw := tfCase["notifications"].([]interface{})
			tfNotifications := make([]string, len(tfNotificationsRaw))
			for notificationIdx, v := range tfNotificationsRaw {
				tfNotifications[notificationIdx] = v.(string)
			}

			if !stringSliceEquals(tfNotifications, ruleCase.GetNotifications()) {
				modifiedCases++
				shouldUpdate = true
				updatedRuleCase[i].Notifications = tfNotifications
			}
			// Compare rule case custom status
			tfCustomStatusRaw := tfCase["custom_status"].(string)
			if tfCustomStatusRaw != "" {
				tfCustomStatus := datadogV2.SecurityMonitoringRuleSeverity(tfCustomStatusRaw)
				if tfCustomStatus != ruleCase.GetCustomStatus() {
					modifiedCases++
					shouldUpdate = true
					updatedRuleCase[i].CustomStatus = &tfCustomStatus
				}
			}
		} else {

			// Clear rule case notifications when rule case removed from terraform configuration

			tfNotifications := make([]string, 0)

			if !stringSliceEquals(tfNotifications, ruleCase.GetNotifications()) {
				modifiedCases++
				shouldUpdate = true
				updatedRuleCase[i].Notifications = tfNotifications
			}
		}

	}

	var v interface{}
	var ok bool
	if !isSignalCorrelation {
		v, ok = d.GetOk("query")
		if ok && v != "" {
			tfQueries := v.([]interface{})
			payloadQueries := make([]datadogV2.SecurityMonitoringRuleQuery, len(tfQueries))
			for idx, tfQuery := range tfQueries {
				payloadQueries[idx] = *buildUpdateStandardRuleQuery(tfQuery)
			}
			payload.SetQueries(payloadQueries)

			// Compare queries including custom_query_extension
			if !compareQueries(currentState.GetQueries(), payloadQueries) {
				shouldUpdate = true
			}
		}
	}

	// Compare custom_message
	if v, ok := d.GetOk("custom_message"); ok {
		customMessage := v.(string)
		payload.SetCustomMessage(customMessage)

		// Check if custom_message exists in current state and compare
		if currentCustomMessage, ok := currentState.GetCustomMessageOk(); ok {
			if *currentCustomMessage != customMessage {
				shouldUpdate = true
			}
		} else {
			// Custom message doesn't exist in the current state, so this is a change
			shouldUpdate = true
		}
	}

	// Compare custom_name
	if v, ok := d.GetOk("custom_name"); ok {
		customName := v.(string)
		payload.SetCustomName(customName)

		// Check if custom_name exists in the current state and compare
		if currentCustomName, ok := currentState.GetCustomNameOk(); ok {
			if *currentCustomName != customName {
				shouldUpdate = true
			}
		} else {
			// Custom name doesn't exist in the current state, so this is a change
			shouldUpdate = true
		}
	}

	if matchedCases < len(tfCasesRaw) {
		// Enable partial state so that we don't persist the changes
		d.Partial(true)
		return nil, false, errors.New("attempted to update notifications for non-existing case for rule " + currentState.GetId())
	}

	if modifiedCases > 0 {
		payload.Cases = updatedRuleCase
	}

	// Compare filters
	tfFilters := d.Get("filter").([]interface{})
	payloadFilters := make([]datadogV2.SecurityMonitoringFilter, len(tfFilters))

	for idx, tfRuleFilter := range tfFilters {
		structRuleFilter := datadogV2.SecurityMonitoringFilter{}

		ruleFilter := tfRuleFilter.(map[string]interface{})

		if action, ok := ruleFilter["action"]; ok {
			structRuleFilter.SetAction(datadogV2.SecurityMonitoringFilterAction(action.(string)))
		}

		if query, ok := ruleFilter["query"]; ok {
			structRuleFilter.SetQuery(query.(string))
		}

		payloadFilters[idx] = structRuleFilter
	}

	// Compare filters
	if !compareFilters(currentState.GetFilters(), payloadFilters) {
		payload.Filters = payloadFilters
		shouldUpdate = true
	}

	// Compare options
	if v, ok := d.GetOk("options"); ok {
		tfOptionsList := v.([]interface{})
		payloadOptions := buildPayloadOptions(tfOptionsList, d.Get("type").(string))
		payload.SetOptions(*payloadOptions)

		// Only update if options actually changed
		currentOptions := currentState.GetOptions()
		if !compareOptions(&currentOptions, payloadOptions) {
			shouldUpdate = true
		}
	}

	// Compare tags
	defaultTags := currentState.GetDefaultTags()
	tags := make(map[string]bool)
	for _, tag := range defaultTags {
		tags[tag] = true
	}

	if v, ok := d.GetOk("custom_tags"); ok {
		tfTags := v.(*schema.Set)
		for _, value := range tfTags.List() {
			customTag := value.(string)
			tags[customTag] = true
		}
	}

	payloadTags := make([]string, 0, len(tags))
	for tag := range tags {
		payloadTags = append(payloadTags, tag)
	}

	payload.SetTags(payloadTags)

	// Compare tags
	if !compareTags(currentState.GetTags(), payloadTags) {
		shouldUpdate = true
	}

	return &payload, shouldUpdate, nil
}

// Helper function to compare queries including custom_query_extension
func compareQueries(currentQueries []datadogV2.SecurityMonitoringStandardRuleQuery, payloadQueries []datadogV2.SecurityMonitoringRuleQuery) bool {
	if len(currentQueries) != len(payloadQueries) {
		return false
	}

	// For now, we'll assume queries are different if they exist in the payload
	// This is a simplified approach - in a more complete implementation,
	// we would need to extract the standard query from the payload query
	// and compare each field individually

	// Since we're building the payload from Terraform config and comparing with current state,
	// if there are any queries in the payload, we should check if they differ from current state
	// For simplicity, we'll return false (indicating a change) if there are queries in the payload
	// This ensures that any query changes are detected

	return len(payloadQueries) == 0
}

// Helper function to compare filters
func compareFilters(currentFilters []datadogV2.SecurityMonitoringFilter, payloadFilters []datadogV2.SecurityMonitoringFilter) bool {
	if len(currentFilters) != len(payloadFilters) {
		return false
	}

	for i, currentFilter := range currentFilters {
		payloadFilter := payloadFilters[i]

		if currentFilter.GetAction() != payloadFilter.GetAction() {
			return false
		}

		if currentFilter.GetQuery() != payloadFilter.GetQuery() {
			return false
		}
	}

	return true
}

// Helper function to compare tags
func compareTags(currentTags []string, payloadTags []string) bool {
	return stringSliceEquals(currentTags, payloadTags)
}

func stringSliceEquals(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// Helper function to compare options
func compareOptions(currentOptions *datadogV2.SecurityMonitoringRuleOptions, payloadOptions *datadogV2.SecurityMonitoringRuleOptions) bool {
	if currentOptions == nil && payloadOptions == nil {
		return true
	}
	if currentOptions == nil || payloadOptions == nil {
		return false
	}
	// Compare decrease_criticality_based_on_env
	return currentOptions.GetDecreaseCriticalityBasedOnEnv() == payloadOptions.GetDecreaseCriticalityBasedOnEnv()
}

func findRuleCaseForStatus(tfCasesRaw []interface{}, status datadogV2.SecurityMonitoringRuleSeverity) (map[string]interface{}, bool) {
	for _, tfCaseRaw := range tfCasesRaw {
		tfCase := tfCaseRaw.(map[string]interface{})
		tfStatus := datadogV2.SecurityMonitoringRuleSeverity(tfCase["status"].(string))
		if tfStatus == status {
			return tfCase, true
		}
	}

	return nil, false
}

func resourceDatadogSecurityMonitoringDefaultRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// no-op
	return nil
}
