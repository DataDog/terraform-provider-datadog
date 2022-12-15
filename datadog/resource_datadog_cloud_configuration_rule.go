package datadog

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

const nameField = "name"
const messageField = "message"
const enabledField = "enabled"
const policyField = "policy"
const resourceTypesField = "resource_types"
const severityField = "severity"
const notificationsField = "notifications"
const groupByField = "group_by"
const tagsField = "tags"

func resourceDatadogCloudConfigurationRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Cloud Configuration Rule resource.",
		CreateContext: cloudConfigurationRuleCreateContext,
		DeleteContext: resourceDatadogSecurityMonitoringRuleDelete,
		Schema:        cloudConfigurationRuleSchema(),
	}
}

func cloudConfigurationRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		nameField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the cloud configuration rule.",
		},
		messageField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The message associated to the rule that will be shown in findings and signals.",
		},
		enabledField: {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Whether the cloud configuration rule is enabled.",
		},
		policyField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Policy written in rego format.",
		},
		resourceTypesField: {
			Type:             schema.TypeList,
			Required:         true,
			Description:      "Resource types to be checked by the rule. Must have at least one element.",
			ValidateDiagFunc: validators.ValidateNonEmptyStringList,
			Elem:             &schema.Schema{Type: schema.TypeString},
		},
		severityField: {
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
			Required:         true,
			Description:      "Severity of the rule and associated signals.",
		},
		notificationsField: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Notification targets for signals. Defaults to empty list.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		groupByField: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Fields to group by when generating signals, e.g. @resource. Defaults to empty list.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		tagsField: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Tags of the rule, propagated to findings and signals. Defaults to empty list.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func cloudConfigurationRuleCreateContext(ctx context.Context, d *schema.ResourceData, metadata interface{}) diag.Diagnostics {
	providerConf := metadata.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ruleCreate := buildCloudConfigurationRulePayload(d)

	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().CreateSecurityMonitoringRule(auth, ruleCreate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating security monitoring rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	if response.SecurityMonitoringStandardRuleResponse != nil {
		d.SetId(response.SecurityMonitoringStandardRuleResponse.GetId())
	} else {
		return diag.FromErr(fmt.Errorf("SecurityMonitoringStandardRuleResponse is empty"))
	}

	return nil
}

func buildCloudConfigurationRulePayload(d *schema.ResourceData) datadogV2.SecurityMonitoringRuleCreatePayload {
	payload := datadogV2.NewCloudConfigurationRuleCreatePayloadWithDefaults()
	payload.SetName(d.Get(nameField).(string))
	payload.SetMessage(d.Get(messageField).(string))
	payload.SetIsEnabled(d.Get(enabledField).(bool))
	payload.SetOptions(buildComplianceRuleOptions(d))
	payload.SetComplianceSignalOptions(builComplianceSignalOptions(d))
	payload.SetCases(buildCases(d))
	payload.SetTags(getStringSlice(d, tagsField))
	payload.SetType(datadogV2.CLOUDCONFIGURATIONRULETYPE_CLOUD_CONFIGURATION)

	return datadogV2.CloudConfigurationRuleCreatePayloadAsSecurityMonitoringRuleCreatePayload(payload)
}

func buildComplianceRuleOptions(d *schema.ResourceData) datadogV2.CloudConfigurationRuleOptions {
	regoPolicy := d.Get(policyField).(string)
	resourceTypes := getStringSlice(d, resourceTypesField)
	isComplexrule := len(resourceTypes) > 1

	complianceRuleOptions := datadogV2.NewCloudConfigurationComplianceRuleOptions(*datadogV2.NewCloudConfigurationRegoRule(regoPolicy, resourceTypes))
	complianceRuleOptions.SetComplexRule(isComplexrule)

	return *datadogV2.NewCloudConfigurationRuleOptions(*complianceRuleOptions)
}

func buildCases(d *schema.ResourceData) []datadogV2.CloudConfigurationRuleCaseCreate {
	notifications := getStringSlice(d, notificationsField)
	severity := d.Get(severityField).(string)

	ruleCase := datadogV2.NewCloudConfigurationRuleCaseCreate(datadogV2.SecurityMonitoringRuleSeverity(severity))
	ruleCase.SetNotifications(notifications)

	return []datadogV2.CloudConfigurationRuleCaseCreate{*ruleCase}
}

func builComplianceSignalOptions(d *schema.ResourceData) datadogV2.CloudConfigurationRuleComplianceSignalOptions {
	groupByFields := getStringSlice(d, groupByField)

	signalOptions := datadogV2.NewCloudConfigurationRuleComplianceSignalOptions()
	signalOptions.SetUserActivationStatus(len(groupByFields) > 1)
	signalOptions.SetUserGroupByFields(groupByFields)

	return *signalOptions
}

func getStringSlice(d *schema.ResourceData, key string) []string {
	if v, ok := d.GetOk(key); ok {
		values := v.([]interface{})
		stringValues := make([]string, len(values))
		for i, value := range values {
			stringValues[i] = value.(string)
		}
		return stringValues
	}
	return []string{}
}
