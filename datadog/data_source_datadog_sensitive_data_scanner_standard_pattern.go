package datadog

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogSensitiveDataScannerStandardPattern() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing sensitive data scanner standard pattern.",
		ReadContext: dataSourceDatadogSensitiveDataScannerStandardPatternRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"filter": {
					Description: "Filter all the Datadog standard patterns by name.",
					Type:        schema.TypeString,
					Required:    true,
				},
				// Computed
				"name": {
					Description: "Name of the standard pattern.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"pattern": {
					Description: "Regex that the standard pattern applies.",
					Type:        schema.TypeString,
					Computed:    true,
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

	searchedName := d.Get("filter").(string)

	foundStandardPatterns := make([]datadogV2.SensitiveDataScannerStandardPatternsResponseItem, 0)
	for _, resource := range resp.GetData() {
		if strings.Contains(strings.ToLower(*resource.Attributes.Name), strings.ToLower(searchedName)) {
			foundStandardPatterns = append(foundStandardPatterns, resource)
		}
	}

	if len(foundStandardPatterns) == 0 {
		return diag.Errorf("Couldn't find the standard pattern with name %s", searchedName)
	}
	if len(foundStandardPatterns) > 1 {
		return diag.Errorf("Your query returned more than one result, please try a more specific search criteria")
	}
	d.SetId(foundStandardPatterns[0].GetId())

	return dataSourceSensitiveDataScannerStandardPatternUpdate(d, &foundStandardPatterns[0])

}

func dataSourceSensitiveDataScannerStandardPatternUpdate(d *schema.ResourceData, standardPattern *datadogV2.SensitiveDataScannerStandardPatternsResponseItem) diag.Diagnostics {
	if err := d.Set("name", standardPattern.Attributes.GetName()); err != nil {
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
