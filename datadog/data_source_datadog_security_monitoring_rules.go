package datadog

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
)

func dataSourceDatadogSecurityMonitoringRules() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about existing security monitoring rules for use in other resources.",
		Read:        dataSourceDatadogSecurityMonitoringRulesRead,

		Schema: map[string]*schema.Schema{
			// Filters
			"name_filter": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "A rule name to limit the search",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"tags_filter": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of tags to limit the search",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"default_only_filter": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Limit the search to default rules",
			},
			"user_only_filter": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Limit the search to user rules",
			},

			// Computed
			"rule_ids": {
				Description: "List of IDs of the matched rules.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"rules": {
				Description: "List of rules.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: datadogSecurityMonitoringRuleSchema(),
				},
			},
		},
	}
}

func dataSourceDatadogSecurityMonitoringRulesRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	var nameFilter *string
	var defaultFilter *bool
	var tagFilter map[string]bool

	if v, ok := d.GetOk("name_filter"); ok {
		filter := v.(string)
		nameFilter = &filter
	}

	_, filterDefault := d.GetOk("default_only_filter")
	_, filterUser := d.GetOk("user_only_filter")
	if filterDefault && filterUser {
		return errors.New("error: cannot filter both default and user rules")
	}
	if filterDefault {
		filter := true
		defaultFilter = &filter
	}
	if filterUser {
		filter := false
		defaultFilter = &filter
	}

	if v, ok := d.GetOk("tags_filter"); ok {
		filter := v.([]interface{})
		tagFilter = make(map[string]bool)
		for _, tag := range filter {
			tagFilter[tag.(string)] = true
		}
	}

	ruleIds := make([]string, 0)
	rules := make([]map[string]interface{}, 0)
	page := int64(0)

	for {
		response, _, err := datadogClientV2.SecurityMonitoringApi.ListSecurityMonitoringRules(authV2).
			PageNumber(page).
			PageSize(100).
			Execute()

		if err != nil {
			return translateClientError(err, "error listing rules")
		}

		for _, rule := range response.GetData() {
			if !matchesSecMonRuleFilters(rule, nameFilter, defaultFilter, tagFilter) {
				continue
			}
			ruleIds = append(ruleIds, rule.GetId())
			rules = append(rules, buildSecurityMonitoringTfRule(rule))
		}

		totalCount := *response.Meta.GetPage().TotalCount
		if totalCount-1 <= page*100 {
			break
		}
		page += 1
	}

	d.SetId(computeSecMonDataSourceRulesId(nameFilter, defaultFilter, tagFilter))
	d.Set("rules", rules)
	d.Set("rule_ids", ruleIds)

	return nil
}

func computeSecMonDataSourceRulesId(nameFilter *string, defaultFilter *bool, tagFilter map[string]bool) string {
	// Sort tags to make key unique
	tags := make([]string, len(tagFilter))
	idx := 0
	for tag := range tagFilter {
		tags[idx] = tag
		idx += 1
	}
	sort.Strings(tags)

	// Key for hashing
	var b strings.Builder
	if nameFilter != nil {
		b.WriteString(*nameFilter)
	}
	b.WriteRune('|')
	if defaultFilter != nil {
		if *defaultFilter {
			b.WriteRune('1')
		} else {
			b.WriteRune('0')
		}
	}
	b.WriteRune('|')
	for _, tag := range tags {
		b.WriteString(tag)
		b.WriteRune(',')
	}
	keyStr := b.String()

	h := sha256.New()
	log.Println("HASHKEY", keyStr)
	h.Write([]byte(keyStr))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func buildSecurityMonitoringTfRule(rule datadogV2.SecurityMonitoringRuleResponse) map[string]interface{} {
	tfRule := make(map[string]interface{})

	cases := make([]map[string]interface{}, len(rule.GetCases()))
	for i, ruleCase := range rule.GetCases() {
		tfRuleCase := make(map[string]interface{})
		tfRuleCase["name"] = ruleCase.GetName()
		tfRuleCase["condition"] = ruleCase.GetCondition()
		tfRuleCase["status"] = ruleCase.Status
		if notifications, ok := ruleCase.GetNotificationsOk(); ok {
			tfRuleCase["notifications"] = notifications
		}
		cases[i] = tfRuleCase
	}
	tfRule["case"] = cases

	tfRule["enabled"] = rule.GetIsEnabled()
	tfRule["message"] = rule.GetMessage()
	tfRule["name"] = rule.GetName()

	tfOptions := make(map[string]interface{})
	options := rule.GetOptions()
	tfOptions["evaluation_window"] = int(options.GetEvaluationWindow())
	tfOptions["keep_alive"] = int(options.GetKeepAlive())
	tfOptions["max_signal_duration"] = int(options.GetMaxSignalDuration())
	tfRule["options"] = []map[string]interface{}{tfOptions}

	tfQueries := make([]map[string]interface{}, len(rule.GetQueries()))
	for i, query := range rule.GetQueries() {
		tfQuery := make(map[string]interface{})
		if aggregation, ok := query.GetAggregationOk(); ok {
			tfQuery["aggregation"] = string(*aggregation)
		}
		if distinctFields, ok := query.GetDistinctFieldsOk(); ok {
			tfQuery["distinct_fields"] = *distinctFields
		}
		if groupByFields, ok := query.GetGroupByFieldsOk(); ok {
			tfQuery["group_by_fields"] = *groupByFields
		}
		if metric, ok := query.GetMetricOk(); ok {
			tfQuery["metric"] = *metric
		}
		if name, ok := query.GetNameOk(); ok {
			tfQuery["name"] = *name
		}
		tfQuery["query"] = query.GetQuery()
		tfQueries[i] = tfQuery
	}
	tfRule["query"] = tfQueries

	if tags, ok := rule.GetTagsOk(); ok {
		tfRule["tags"] = *tags
	}

	return tfRule
}

func matchesSecMonRuleFilters(
	rule datadogV2.SecurityMonitoringRuleResponse,
	nameFilter *string,
	defaultFilter *bool,
	tagFilter map[string]bool) bool {

	if nameFilter != nil {
		name := *rule.Name
		if !strings.Contains(name, *nameFilter) {
			return false
		}
	}
	if defaultFilter != nil {
		if *rule.IsDefault != *defaultFilter {
			return false
		}
	}
	if tagFilter != nil {
		matchedTagCount := 0
		for _, tag := range *rule.Tags {
			if _, ok := tagFilter[tag]; ok {
				matchedTagCount += 1
			}
		}
		if matchedTagCount < len(tagFilter) {
			return false
		}
	}

	return true
}
