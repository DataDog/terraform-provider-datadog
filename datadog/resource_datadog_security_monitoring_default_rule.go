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
							Type:         schema.TypeString,
							ValidateFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
							Required:     true,
							Description:  "Status of the rule case to match.",
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
	ruleResponse, _, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, id).Execute()
	if err != nil {
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

	return nil
}

func resourceDatadogSecurityMonitoringDefaultRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ruleID := d.Id()

	response, httpResponse, err := datadogClientV2.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, ruleID).Execute()

	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return diag.FromErr(errors.New("default rule does not exist"))
		}

		return utils.TranslateClientErrorDiag(err, "error fetching default rule")
	}

	if !response.GetIsDefault() {
		return diag.FromErr(errors.New("rule is not a default rule"))
	}

	ruleUpdate, shouldUpdate, err := buildSecMonDefaultRuleUpdatePayload(response, d)

	if err != nil {
		return diag.FromErr(err)
	}

	if shouldUpdate {
		if _, _, err := datadogClientV2.SecurityMonitoringApi.UpdateSecurityMonitoringRule(authV2, ruleID).Body(*ruleUpdate).Execute(); err != nil {
			return utils.TranslateClientErrorDiag(err, "error updating security monitoring rule on resource creation")
		}
	}

	return nil
}

func buildSecMonDefaultRuleUpdatePayload(currentState datadogV2.SecurityMonitoringRuleResponse, d *schema.ResourceData) (*datadogV2.SecurityMonitoringRuleUpdatePayload, bool, error) {
	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}

	isEnabled := d.Get("enabled").(bool)
	payload.IsEnabled = &isEnabled

	if v, ok := d.GetOk("cases"); ok {
		matchedCases := 0
		modifiedCases := 0
		tfCases := v.([]map[string]interface{})

		updatedRuleCase := make([]datadogV2.SecurityMonitoringRuleCase, len(currentState.GetCases()))
		for i, ruleCase := range currentState.GetCases() {
			var updatedNotifications []string
			if tfCase, ok := findRuleCaseForStatus(tfCases, ruleCase.GetStatus()); ok {
				matchedCases++

				tfNotificationsRaw := tfCase["notifications"].([]interface{})
				tfNotifications := make([]string, len(tfNotificationsRaw))
				for notificationIdx, v := range tfNotificationsRaw {
					tfNotifications[notificationIdx] = v.(string)
				}

				if !stringSliceEquals(tfNotifications, ruleCase.GetNotifications()) {
					modifiedCases++
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

func findRuleCaseForStatus(cases []map[string]interface{}, status datadogV2.SecurityMonitoringRuleSeverity) (map[string]interface{}, bool) {
	for _, tfCase := range cases {
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
