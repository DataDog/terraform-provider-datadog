package datadog

import (
	"context"
	"errors"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogSecurityMonitoringDefaultRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Security Monitoring Rule API resource for default rules.",
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
				MaxItems:    5,
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
		},
	}
}

func resourceDatadogSecurityMonitoringDefaultRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(errors.New("cannot create a default rule, please import it first before making changes"))
}

func resourceDatadogSecurityMonitoringDefaultRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	id := d.Id()
	ruleResponse, _, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := utils.CheckForUnparsed(ruleResponse); err != nil {
		return diag.FromErr(err)
	}

	d.Set("enabled", *ruleResponse.IsEnabled)

	if v, ok := d.GetOk("case"); ok {
		tfCasesRaw := v.([]interface{})
		readNotifications := make([][]string, len(tfCasesRaw))
		for i, tfCaseRaw := range tfCasesRaw {
			tfCase := tfCaseRaw.(map[string]interface{})
			var ruleCase *datadogV2.SecurityMonitoringRuleCase
			tfStatus := datadogV2.SecurityMonitoringRuleSeverity(tfCase["status"].(string))
			for _, rc := range ruleResponse.GetCases() {
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

	ruleFilters := make([]map[string]interface{}, len(ruleResponse.GetFilters()))
	for idx, responseRuleFilter := range ruleResponse.GetFilters() {
		ruleFilters[idx] = map[string]interface{}{
			"action": responseRuleFilter.GetAction(),
			"query":  responseRuleFilter.GetQuery(),
		}
	}

	d.Set("filter", ruleFilters)

	return nil
}

func resourceDatadogSecurityMonitoringDefaultRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleID := d.Id()

	response, httpResponse, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, ruleID)

	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return diag.FromErr(errors.New("default rule does not exist"))
		}

		return utils.TranslateClientErrorDiag(err, httpResponse, "error fetching default rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	if !response.GetIsDefault() {
		return diag.FromErr(errors.New("rule is not a default rule"))
	}

	ruleUpdate, shouldUpdate, err := buildSecMonDefaultRuleUpdatePayload(response, d)

	if err != nil {
		return diag.FromErr(err)
	}

	if shouldUpdate {
		if _, httpResponse, err := datadogClientV2.SecurityMonitoringApi.UpdateSecurityMonitoringRule(authV2, ruleID, *ruleUpdate); err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error updating security monitoring rule on resource creation")
		}
	}

	return nil
}

func buildSecMonDefaultRuleUpdatePayload(currentState datadogV2.SecurityMonitoringRuleResponse, d *schema.ResourceData) (*datadogV2.SecurityMonitoringRuleUpdatePayload, bool, error) {
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
				updatedRuleCase[i].Notifications = &tfNotifications
			}

		} else {

			// Clear rule case notifications when rule case removed from terraform configuration

			tfNotifications := make([]string, 0)

			if !stringSliceEquals(tfNotifications, ruleCase.GetNotifications()) {
				modifiedCases++
				updatedRuleCase[i].Notifications = &tfNotifications
			}
		}

	}

	if matchedCases < len(tfCasesRaw) {
		// Enable partial state so that we don't persist the changes
		d.Partial(true)
		return nil, false, errors.New("attempted to update notifications for non-existing case for rule " + currentState.GetId())
	}

	if modifiedCases > 0 {
		payload.Cases = &updatedRuleCase
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
	payload.Filters = &payloadFilters

	return &payload, true, nil
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
