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
const resourceTypeField = "resource_type"
const relatedResourceTypesField = "related_resource_types"
const severityField = "severity"
const notificationsField = "notifications"
const groupByField = "group_by"
const tagsField = "tags"
const filterField = "filter"
const queryField = "query"
const actionField = "action"

func resourceDatadogCloudConfigurationRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Cloud Configuration Rule resource.",
		CreateContext: cloudConfigurationRuleCreateContext,
		ReadContext:   cloudConfigurationRuleReadContext,
		UpdateContext: cloudConfigurationRuleUpdateContext,
		DeleteContext: resourceDatadogSecurityMonitoringRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return cloudConfigurationRuleSchema()
		},
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
			Description: "Policy written in Rego format.",
		},
		resourceTypeField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Main resource type to be checked by the rule.",
		},
		relatedResourceTypesField: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Related resource types to be checked by the rule. Defaults to empty list.",
			MinItems:    0,
			MaxItems:    9,
			Elem:        &schema.Schema{Type: schema.TypeString},
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
		filterField: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Additional queries to filter matched events before they are processed. Defaults to empty list",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					queryField: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for selecting logs to apply the filtering action.",
					},
					actionField: {
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

func cloudConfigurationRuleCreateContext(ctx context.Context, d *schema.ResourceData, metadata interface{}) diag.Diagnostics {
	providerConf := metadata.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ruleCreate := buildRuleCreatePayload(d)

	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().CreateSecurityMonitoringRule(auth, *ruleCreate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating security monitoring rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	if response.SecurityMonitoringStandardRuleResponse == nil {
		return diag.FromErr(fmt.Errorf("SecurityMonitoringStandardRuleResponse is empty"))
	}

	d.SetId(response.SecurityMonitoringStandardRuleResponse.GetId())
	updateResourceDataFromResponse(d, response.SecurityMonitoringStandardRuleResponse)
	return nil
}

func buildRuleCreatePayload(d *schema.ResourceData) *datadogV2.SecurityMonitoringRuleCreatePayload {
	payload := datadogV2.NewCloudConfigurationRuleCreatePayloadWithDefaults()
	payload.SetName(d.Get(nameField).(string))
	payload.SetMessage(d.Get(messageField).(string))
	payload.SetIsEnabled(d.Get(enabledField).(bool))
	payload.SetOptions(buildRuleCreationOptions(d))
	payload.SetComplianceSignalOptions(*buildComplianceSignalOptions(d))
	payload.SetCases(*buildRuleCreationCases(d))
	payload.SetTags(utils.GetStringSlice(d, tagsField))
	payload.SetType(datadogV2.CLOUDCONFIGURATIONRULETYPE_CLOUD_CONFIGURATION)
	payload.SetFilters(buildFiltersFromResourceData(d))

	createPayload := datadogV2.CloudConfigurationRuleCreatePayloadAsSecurityMonitoringRuleCreatePayload(payload)
	return &createPayload
}

func buildRuleCreationOptions(d *schema.ResourceData) datadogV2.CloudConfigurationRuleOptions {
	return *datadogV2.NewCloudConfigurationRuleOptions(*buildComplianceRuleOptions(d))
}

func buildRuleCreationCases(d *schema.ResourceData) *[]datadogV2.CloudConfigurationRuleCaseCreate {
	notifications := utils.GetStringSlice(d, notificationsField)
	severity := d.Get(severityField).(string)

	ruleCase := datadogV2.NewCloudConfigurationRuleCaseCreate(datadogV2.SecurityMonitoringRuleSeverity(severity))
	ruleCase.SetNotifications(notifications)

	return &[]datadogV2.CloudConfigurationRuleCaseCreate{*ruleCase}
}

func buildComplianceSignalOptions(d *schema.ResourceData) *datadogV2.CloudConfigurationRuleComplianceSignalOptions {
	groupByFields := utils.GetStringSlice(d, groupByField)

	signalOptions := datadogV2.NewCloudConfigurationRuleComplianceSignalOptions()
	signalOptions.SetUserActivationStatus(len(groupByFields) > 0)
	signalOptions.SetUserGroupByFields(groupByFields)

	return signalOptions
}

func cloudConfigurationRuleUpdateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ruleUpdate := buildRuleUpdatePayload(d)
	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().UpdateSecurityMonitoringRule(auth, d.Id(), *ruleUpdate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating security monitoring rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	if response.SecurityMonitoringStandardRuleResponse != nil {
		updateResourceDataFromResponse(d, response.SecurityMonitoringStandardRuleResponse)
	}

	return nil
}

func buildRuleUpdatePayload(d *schema.ResourceData) *datadogV2.SecurityMonitoringRuleUpdatePayload {
	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}
	payload.SetName(d.Get(nameField).(string))
	payload.SetMessage(d.Get(messageField).(string))
	payload.SetIsEnabled(d.Get(enabledField).(bool))
	payload.SetOptions(buildRuleUpdateOptions(d))
	payload.SetComplianceSignalOptions(*buildComplianceSignalOptions(d))
	payload.SetCases(*buildRuleUpdateCases(d))
	payload.SetTags(utils.GetStringSlice(d, tagsField))
	payload.SetFilters(buildFiltersFromResourceData(d))
	return &payload
}

func buildRuleUpdateOptions(d *schema.ResourceData) datadogV2.SecurityMonitoringRuleOptions {
	options := datadogV2.NewSecurityMonitoringRuleOptions()
	options.SetComplianceRuleOptions(*buildComplianceRuleOptions(d))
	return *options
}

func buildComplianceRuleOptions(d *schema.ResourceData) *datadogV2.CloudConfigurationComplianceRuleOptions {
	regoPolicy := d.Get(policyField).(string)
	mainResourceType, resourceTypes, isComplexRule := getAllResourceTypes(d)
	complianceRuleOptions := datadogV2.NewCloudConfigurationComplianceRuleOptions()
	complianceRuleOptions.SetResourceType(mainResourceType)
	complianceRuleOptions.SetRegoRule(*datadogV2.NewCloudConfigurationRegoRule(regoPolicy, resourceTypes))
	complianceRuleOptions.SetComplexRule(isComplexRule)
	return complianceRuleOptions
}

func getAllResourceTypes(d *schema.ResourceData) (string, []string, bool) {
	mainResourceType := d.Get(resourceTypeField).(string)
	relatedResourceTypes := utils.GetStringSlice(d, relatedResourceTypesField)

	resourceTypes := make([]string, 0)
	resourceTypes = append(resourceTypes, mainResourceType)
	resourceTypes = append(resourceTypes, relatedResourceTypes...)

	return mainResourceType, resourceTypes, len(relatedResourceTypes) > 0
}

func buildRuleUpdateCases(d *schema.ResourceData) *[]datadogV2.SecurityMonitoringRuleCase {
	notifications := utils.GetStringSlice(d, notificationsField)
	severity := d.Get(severityField).(string)

	ruleCase := datadogV2.NewSecurityMonitoringRuleCase()
	ruleCase.SetStatus(datadogV2.SecurityMonitoringRuleSeverity(severity))
	ruleCase.SetNotifications(notifications)

	return &[]datadogV2.SecurityMonitoringRuleCase{*ruleCase}
}

func cloudConfigurationRuleReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()
	ruleResponse, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, id)
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

	standardRuleResponse := ruleResponse.SecurityMonitoringStandardRuleResponse
	if standardRuleResponse != nil {
		if standardRuleResponse.GetType() == datadogV2.SECURITYMONITORINGRULETYPEREAD_CLOUD_CONFIGURATION {
			updateResourceDataFromResponse(d, standardRuleResponse)
		} else {
			return diag.Errorf("Rule with id %s is not a cloud_configuration rule. This terraform resource can only manage `cloud_configuration` rules", d.Id())
		}
	}
	return nil
}

func updateResourceDataFromResponse(d *schema.ResourceData, ruleResponse *datadogV2.SecurityMonitoringStandardRuleResponse) {
	d.Set(messageField, ruleResponse.GetMessage())
	d.Set(nameField, ruleResponse.GetName())
	d.Set(enabledField, ruleResponse.GetIsEnabled())

	resourceType := ruleResponse.GetOptions().ComplianceRuleOptions.ResourceType
	regoRule := ruleResponse.GetOptions().ComplianceRuleOptions.GetRegoRule()
	d.Set(resourceTypeField, resourceType)
	d.Set(relatedResourceTypesField, getRelatedResourceTypes(*resourceType, regoRule.ResourceTypes))
	d.Set(policyField, regoRule.Policy)

	ruleCase := ruleResponse.GetCases()[0]
	d.Set(severityField, ruleCase.Status)
	d.Set(notificationsField, ruleCase.GetNotifications())
	d.Set(groupByField, ruleResponse.ComplianceSignalOptions.GetUserGroupByFields())
	d.Set(tagsField, ruleResponse.GetTags())

	if filters, ok := ruleResponse.GetFiltersOk(); ok {
		d.Set(filterField, extractFiltersFromRuleResponse(*filters))
	}
}

func getRelatedResourceTypes(mainResourceType string, resourceTypes []string) []string {
	relatedResourceTypes := make([]string, 0)
	for _, resourceType := range resourceTypes {
		if resourceType != mainResourceType {
			relatedResourceTypes = append(relatedResourceTypes, resourceType)
		}
	}
	return relatedResourceTypes
}

func buildFiltersFromResourceData(d *schema.ResourceData) []datadogV2.SecurityMonitoringFilter {
	if filters, ok := d.GetOk(filterField); ok {
		filterList := filters.([]interface{})
		return buildPayloadFilters(filterList)
	}
	return []datadogV2.SecurityMonitoringFilter{}
}
