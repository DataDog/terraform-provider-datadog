package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogLogsArchivesOrder() *schema.Resource {
	return &schema.Resource{
		Description: "Get the current order of your logs archives.",
		ReadContext: dataSourceDatadogLogsArchivesOrderRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				// Computed values
				"archive_ids": {
					Description: "The archive IDs list. The order of archive IDs in this attribute defines the overall archive order for logs.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

func dataSourceDatadogLogsArchivesOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	logsArchiveOrder, httpresp, err := apiInstances.GetLogsArchivesApiV2().GetLogsArchiveOrder(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying the order of your logs archives")
	}

	if err := d.Set("archive_ids", logsArchiveOrder.Data.Attributes.ArchiveIds); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("logs-archive-order")

	return nil
}
