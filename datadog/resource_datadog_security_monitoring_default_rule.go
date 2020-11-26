package datadog

import (
	"errors"
	"fmt"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogSecurityMonitoringDefaultRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogSecurityMonitoringDefaultRuleCreate,
		Read:   resourceDatadogSecurityMonitoringDefaultRuleRead,
		Update: resourceDatadogSecurityMonitoringDefaultRuleUpdate,
		Delete: resourceDatadogSecurityMonitoringDefaultRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"rule_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the rule.",
				ForceNew:    true,
			},

			"case": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Cases of the rule, this is used to update notifications.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Status of the rule case to match.",
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
				Description: "Enable the rule.",
			},

			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable the rule.",
			},
		},
	}
}

func resourceDatadogSecurityMonitoringDefaultRuleCreate(d *schema.ResourceData, meta interface{}) error {
	// create only updates an existing rule
	err := resourceDatadogSecurityMonitoringDefaultRuleUpdate(d, meta)
	if err != nil {
		return err
	}

	d.SetId(d.Get("rule_id").(string))

	return nil
}

func resourceDatadogSecurityMonitoringDefaultRuleRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	id := d.Id()
	ruleResponse, _, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, id).Execute()
	if err != nil {
		return err
	}

	_, isEnabled := d.GetOk("enabled")
	_, isDisabled := d.GetOk("disabled")
	if isEnabled && isDisabled {
		return errors.New("can not set a rule to both enabled and disabled")
	}
	if isEnabled || isDisabled {
		d.Set("enabled", *ruleResponse.IsEnabled)
		d.Set("disabled", !*ruleResponse.IsEnabled)
	}

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
				return errors.New("error: no rule case with status " + string(tfStatus))
			}
			readNotifications[i] = ruleCase.GetNotifications()
		}

		for i, notification := range readNotifications {
			d.Set(fmt.Sprintf("case.%d.notifications", i), notification)
		}
	}

	return nil
}

func resourceDatadogSecurityMonitoringDefaultRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleId := d.Get("rule_id").(string)

	response, httpResponse, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, ruleId).Execute()

	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return errors.New("default rule does not exist")
		}

		return translateClientError(err, "error fetching default rule")
	}

	if !response.GetIsDefault() {
		return errors.New("rule is not a default rule")
	}

	ruleUpdate, shouldUpdate, err := buildSecMonDefaultRuleUpdatePayload(response, d)

	if err != nil {
		return err
	}

	if shouldUpdate {
		_, _, err := datadogClientV2.SecurityMonitoringApi.UpdateSecurityMonitoringRule(authV2, ruleId).Body(*ruleUpdate).Execute()
		if err != nil {
			return translateClientError(err, "error updating security monitoring rule on resource creation")
		}
	}

	return nil
}

func buildSecMonDefaultRuleUpdatePayload(currentState datadogV2.SecurityMonitoringRuleResponse, d *schema.ResourceData) (*datadogV2.SecurityMonitoringRuleUpdatePayload, bool, error) {
	modified := false
	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}

	_, isEnabled := d.GetOk("enabled")
	_, isDisabled := d.GetOk("disabled")

	if isEnabled && isDisabled {
		return nil, false, errors.New("can not set a rule to both enabled and disabled")
	}

	if isEnabled && !currentState.GetIsEnabled() {
		modified = true
		enabled := true
		payload.IsEnabled = &enabled
	}
	if isDisabled && currentState.GetIsEnabled() {
		modified = true
		disabled := false
		payload.IsEnabled = &disabled
	}

	if v, ok := d.GetOk("cases"); ok {
		matchedCases := 0
		modifiedCases := 0
		tfCases := v.([]map[string]interface{})

		updatedRuleCase := make([]datadogV2.SecurityMonitoringRuleCase, len(currentState.GetCases()))
		for i, ruleCase := range currentState.GetCases() {
			var updatedNotifications []string
			if tfCase, ok := findRuleCaseForStatus(tfCases, ruleCase.GetStatus()); ok {
				matchedCases += 1

				tfNotificationsRaw := tfCase["notifications"].([]interface{})
				tfNotifications := make([]string, len(tfNotificationsRaw))
				for notificationIdx, v := range tfNotificationsRaw {
					tfNotifications[notificationIdx] = v.(string)
				}

				if !stringSliceEquals(tfNotifications, ruleCase.GetNotifications()) {
					modified = true
					modifiedCases += 1
					updatedNotifications = tfNotifications
				}
			}

			updatedRuleCase[i] = datadogV2.SecurityMonitoringRuleCase{
				Condition:     currentState.GetCases()[i].Condition,
				Name:          currentState.GetCases()[i].Name,
				Notifications: currentState.GetCases()[i].Notifications,
				Status:        currentState.GetCases()[i].Status,
			}
			if updatedNotifications != nil {
				updatedRuleCase[i].Notifications = &updatedNotifications
			}
		}

		if matchedCases < len(tfCases) {
			return nil, false, errors.New("attempted to update notifications for non-existing case for rule " + currentState.GetId())
		}

		if modifiedCases > 0 {
			payload.Cases = &updatedRuleCase
		}
	}

	if modified {
		return &payload, true, nil
	} else {

		return nil, false, nil
	}
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

func findRuleCaseForStatus(cases []map[string]interface{}, status datadogV2.SecurityMonitoringRuleSeverity) (map[string]interface{}, bool) {
	for _, tfCase := range cases {
		tfStatus := datadogV2.SecurityMonitoringRuleSeverity(tfCase["status"].(string))
		if tfStatus == status {
			return tfCase, true
		}
	}

	return nil, false
}

func resourceDatadogSecurityMonitoringDefaultRuleDelete(d *schema.ResourceData, meta interface{}) error {
	// no-op
	return nil
}
