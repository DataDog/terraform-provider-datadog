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

		Schema: map[string]*schema.Schema{
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
						"notifications": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "Notification targets for each rule case.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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
				"After the deprecation date, the rule will stop triggering signals. " +
				"Moreover, the API will reject any call to update the rule, which might break your Terraform pipeline.",
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

	isEnabled := d.Get("enabled").(bool)
	payload.IsEnabled = &isEnabled

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
				updatedRuleCase[i].Notifications = tfNotifications
			}

		} else {

			// Clear rule case notifications when rule case removed from terraform configuration

			tfNotifications := make([]string, 0)

			if !stringSliceEquals(tfNotifications, ruleCase.GetNotifications()) {
				modifiedCases++
				updatedRuleCase[i].Notifications = tfNotifications
			}
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
	payload.Filters = payloadFilters

	payload.Options = buildDefaultRulePayloadOptions(d)

	return &payload, true, nil
}

func buildDefaultRulePayloadOptions(d *schema.ResourceData) *datadogV2.SecurityMonitoringRuleOptions {
	tfOptions := extractMapFromInterface(d.Get("options").([]interface{}))

	if len(tfOptions) == 0 {
		return nil
	}

	payloadOptions := datadogV2.NewSecurityMonitoringRuleOptions()
	ruleType := d.Get("type").(string)

	if v, ok := tfOptions["decrease_criticality_based_on_env"]; ok && ruleType == string(datadogV2.SECURITYMONITORINGRULETYPECREATE_LOG_DETECTION) {
		payloadOptions.SetDecreaseCriticalityBasedOnEnv(v.(bool))
	}

	return payloadOptions
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
