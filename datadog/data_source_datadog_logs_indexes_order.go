package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogLogsIndexesOrder() *schema.Resource {
	return &schema.Resource{
		Description: "Get the current order of your log indexes.",
		ReadContext: dataSourceDatadogLogsIndexesOrderRead,

		Schema: map[string]*schema.Schema{
			// Computed values
			"index_names": {
				Description: "Array of strings identifying by their name(s) the index(es) of your organization. Logs are tested against the query filter of each index one by one, following the order of the array. Logs are eventually stored in the first matching index.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDatadogLogsIndexesOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	logsIndexesOrder, httpresp, err := datadogClientV1.LogsIndexesApi.GetLogsIndexOrder(authV1)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying the order of your log indexes")
	}
	if err := utils.CheckForUnparsed(logsIndexesOrder); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("index_names", logsIndexesOrder.GetIndexNames()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("logs-index-order")

	return nil
}
