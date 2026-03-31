package datadog

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogSensitiveDataScannerStandardPattern() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing sensitive data scanner standard pattern. You can look up a pattern directly by its stable standard pattern ID or by exact name.",
		ReadContext: dataSourceDatadogSensitiveDataScannerStandardPatternRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"filter": {
					Description:  "Case-insensitive substring of the Datadog standard pattern name to retrieve.",
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{"filter", "standard_pattern_id"},
				},
				"standard_pattern_id": {
					Description:  "Stable ID of the Datadog standard pattern to retrieve. This can be set directly to avoid Terraform configs breaking when Datadog renames a standard pattern.",
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{"filter", "standard_pattern_id"},
				},
				// Computed
				"included_keywords": {
					Description: "List of recommended keywords to improve rule accuracy.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"name": {
					Description: "Name of the standard pattern.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"description": {
					Description: "Description of the standard pattern.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"pattern": {
					Description: "Regex to match, optionally documented for older standard rules. ",
					Type:        schema.TypeString,
					Computed:    true,
					Deprecated:  "Refer to the description field to understand what the rule does.",
				},
				"tags": {
					Description: "List of tags.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

func dataSourceDatadogSensitiveDataScannerStandardPatternRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().ListStandardPatterns(auth)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error listing standard patterns")
	}

	foundStandardPatterns := make([]datadogV2.SensitiveDataScannerStandardPatternsResponseItem, 0)
	matchDescription := ""
	if standardPatternID, ok := d.GetOk("standard_pattern_id"); ok {
		matchDescription = fmt.Sprintf("id %s", standardPatternID.(string))
		for _, resource := range resp.GetData() {
			if resource.GetId() == standardPatternID.(string) {
				foundStandardPatterns = append(foundStandardPatterns, resource)
			}
		}
	} else {
		searchedName := d.Get("filter").(string)
		matchDescription = fmt.Sprintf("name %s", searchedName)
		for _, resource := range resp.GetData() {
			if resource.Attributes == nil || resource.Attributes.Name == nil {
				continue
			}
			if strings.Contains(strings.ToLower(resource.Attributes.GetName()), strings.ToLower(searchedName)) {
				foundStandardPatterns = append(foundStandardPatterns, resource)
			}
		}
	}

	if len(foundStandardPatterns) == 0 {
		return diag.Errorf("Couldn't find the standard pattern with %s", matchDescription)
	}
	if len(foundStandardPatterns) > 1 {
		return diag.Errorf("Found more than one standard pattern with %s", matchDescription)
	}
	d.SetId(foundStandardPatterns[0].GetId())

	return dataSourceSensitiveDataScannerStandardPatternUpdate(d, &foundStandardPatterns[0])

}

func dataSourceSensitiveDataScannerStandardPatternUpdate(d *schema.ResourceData, standardPattern *datadogV2.SensitiveDataScannerStandardPatternsResponseItem) diag.Diagnostics {
	if err := d.Set("included_keywords", standardPattern.Attributes.GetIncludedKeywords()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", standardPattern.Attributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", standardPattern.Attributes.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pattern", standardPattern.Attributes.GetPattern()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", standardPattern.Attributes.GetTags()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
