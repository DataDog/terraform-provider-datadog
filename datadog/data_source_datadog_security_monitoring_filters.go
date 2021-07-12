package datadog

import (
	"context"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogSecurityMonitoringFilters() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about existing security monitoring filters for use in other resources.",
		ReadContext: dataSourceDatadogSecurityFiltersRead,

		Schema: map[string]*schema.Schema{
			// Computed
			"filters_ids": {
				Description: "List of IDs of filters.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"filters": {
				Description: "List of filters.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: securityMonitoringFilterSchema(),
				},
			},
		},
	}
}

func dataSourceDatadogSecurityFiltersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	filterIds := make([]string, 0)
	filters := make([]map[string]interface{}, 0)

	response, httpresp, err := datadogClientV2.SecurityMonitoringApi.ListSecurityFilters(authV2)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error listing filters")
	}

	for _, filter := range response.GetData() {
		// get filter id
		filterIds = append(filterIds, filter.GetId())

		// extract filter
		filterTF := make(map[string]interface{})
		attributes := filter.GetAttributes()

		filterTF["name"] = attributes.GetName()
		filterTF["query"] = attributes.GetQuery()
		filterTF["is_enabled"] = attributes.GetIsEnabled()
		filterTF["filtered_data_type"] = string(attributes.GetFilteredDataType())

		if _, ok := attributes.GetExclusionFiltersOk(); ok {
			exclusionFilters := extractExclusionFiltersTF(attributes)
			filterTF["exclusion_filter"] = exclusionFilters
		}

		filters = append(filters, filterTF)
	}

	d.SetId(buildUniqueId(filterIds))
	d.Set("filters", filters)
	d.Set("filters_ids", filterIds)

	return nil
}

func buildUniqueId(ids []string) string {
	// build a unique id from filters ids
	return strings.Join(ids, "--")
}
