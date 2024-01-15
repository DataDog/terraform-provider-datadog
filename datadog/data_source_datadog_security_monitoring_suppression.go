package datadog

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogSecurityMonitoringSuppressions() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about existing suppression rules, and use them in other resources.",
		ReadContext: dataSourceDatadogSecurityMonitoringSuppressionsRead,
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				// Computed
				"suppression_ids": {
					Description: "List of IDs of suppressions",
					Type:        schema.TypeList,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"suppressions": {
					Description: "List of suppressions",
					Type:        schema.TypeList,
					Elem: &schema.Resource{
						Schema: datadogSecurityMonitoringSuppressionSchema(),
					},
					Computed: true,
				},
			}
		},
	}
}

func dataSourceDatadogSecurityMonitoringSuppressionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().ListSecurityMonitoringSuppressions(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error listing suppressions")
	}

	var suppressionIds []string
	var suppressions []map[string]interface{}
	var diags diag.Diagnostics

	for _, suppression := range response.GetData() {
		if err := utils.CheckForUnparsed(suppression); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping suppression with id: %s", suppression.GetId()),
				Detail:   fmt.Sprintf("security monitoring suppression contains unparsed object: %v", err),
			})
		} else {
			suppressionIds = append(suppressionIds, suppression.GetId())

			suppressionTf := make(map[string]interface{})
			attributes := suppression.Attributes

			suppressionTf["name"] = attributes.GetName()
			suppressionTf["description"] = attributes.GetDescription()
			suppressionTf["enabled"] = attributes.GetEnabled()
			if attributes.ExpirationDate == nil {
				suppressionTf["expiration_date"] = nil
			} else {
				expirationDate := time.UnixMilli(*attributes.ExpirationDate).Format(time.RFC3339)
				suppressionTf["expiration_date"] = &expirationDate
			}
			suppressionTf["rule_query"] = attributes.GetRuleQuery()
			suppressionTf["suppression_query"] = attributes.GetSuppressionQuery()

			suppressions = append(suppressions, suppressionTf)
		}
	}

	d.SetId(buildUniqueId(suppressionIds))
	d.Set("suppression_ids", suppressionIds)
	d.Set("suppressions", suppressions)

	return diags
}
