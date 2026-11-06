package fwprovider

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRuleActionsEqual(t *testing.T) {
	apiAction := func(actionType datadogV2.SecurityMonitoringRuleCaseActionType, duration *int64, userBehaviorName *string) datadogV2.SecurityMonitoringRuleCaseAction {
		options := datadogV2.NewSecurityMonitoringRuleCaseActionOptions()
		options.Duration = duration
		options.UserBehaviorName = userBehaviorName
		return datadogV2.SecurityMonitoringRuleCaseAction{Type: &actionType, Options: options}
	}
	planAction := func(actionType string, options ...ruleCaseActionOptionsModel) ruleCaseActionModel {
		return ruleCaseActionModel{Type: types.StringValue(actionType), Options: options}
	}
	durationOption := func(duration int64) ruleCaseActionOptionsModel {
		return ruleCaseActionOptionsModel{Duration: types.Int64Value(duration), UserBehaviorName: types.StringNull()}
	}
	userBehaviorOption := func(name string) ruleCaseActionOptionsModel {
		return ruleCaseActionOptionsModel{Duration: types.Int64Null(), UserBehaviorName: types.StringValue(name)}
	}

	tests := []struct {
		name     string
		plan     []ruleCaseActionModel
		api      []datadogV2.SecurityMonitoringRuleCaseAction
		expected bool
	}{
		{
			name: "identical actions",
			plan: []ruleCaseActionModel{
				planAction("block_ip", durationOption(3600)),
				planAction("block_user"),
				planAction("user_behavior", userBehaviorOption("behavior_name")),
			},
			api: []datadogV2.SecurityMonitoringRuleCaseAction{
				apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_IP, datadog.PtrInt64(3600), nil),
				apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_USER, nil, nil),
				apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_USER_BEHAVIOR, nil, datadog.PtrString("behavior_name")),
			},
			expected: true,
		},
		{
			name:     "different order",
			plan:     []ruleCaseActionModel{planAction("block_user"), planAction("block_ip", durationOption(3600))},
			api:      []datadogV2.SecurityMonitoringRuleCaseAction{apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_IP, datadog.PtrInt64(3600), nil), apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_USER, nil, nil)},
			expected: false,
		},
		{
			name:     "different length",
			plan:     []ruleCaseActionModel{planAction("block_user")},
			api:      []datadogV2.SecurityMonitoringRuleCaseAction{apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_USER, nil, nil), apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_IP, nil, nil)},
			expected: false,
		},
		{
			name:     "options omitted in plan but set in API",
			plan:     []ruleCaseActionModel{planAction("block_ip")},
			api:      []datadogV2.SecurityMonitoringRuleCaseAction{apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_IP, datadog.PtrInt64(3600), nil)},
			expected: false,
		},
		{
			name:     "option set in API but not in plan options",
			plan:     []ruleCaseActionModel{planAction("user_behavior", userBehaviorOption("behavior_name"))},
			api:      []datadogV2.SecurityMonitoringRuleCaseAction{apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_USER_BEHAVIOR, datadog.PtrInt64(3600), datadog.PtrString("behavior_name"))},
			expected: false,
		},
		{
			name:     "option set in plan but not in API",
			plan:     []ruleCaseActionModel{planAction("block_ip", durationOption(3600))},
			api:      []datadogV2.SecurityMonitoringRuleCaseAction{apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_IP, nil, nil)},
			expected: false,
		},
		{
			name:     "different option value",
			plan:     []ruleCaseActionModel{planAction("user_behavior", userBehaviorOption("new_name"))},
			api:      []datadogV2.SecurityMonitoringRuleCaseAction{apiAction(datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_USER_BEHAVIOR, nil, datadog.PtrString("behavior_name"))},
			expected: false,
		},
		{
			name:     "empty API options",
			plan:     []ruleCaseActionModel{planAction("block_user")},
			api:      []datadogV2.SecurityMonitoringRuleCaseAction{{Type: datadogV2.SECURITYMONITORINGRULECASEACTIONTYPE_BLOCK_USER.Ptr()}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, defaultRuleActionsEqual(tt.plan, tt.api))
		})
	}
}
