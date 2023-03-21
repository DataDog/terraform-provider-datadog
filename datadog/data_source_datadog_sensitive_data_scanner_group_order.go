package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceDatadogSDSGroupOrder() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog Sensitive Data Scanner Group Order API data source. This can be used to retrieve the order of Datadog Sensitive Data Scanner Groups.",
		ReadContext: datasourceDatadogSDSGroupOrderRead,
		Schema: map[string]*schema.Schema{
			"groups": {
				Description: "The list of Sensitive Data Scanner group IDs, in order. Groups are applied sequentially following the order of the list.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func datasourceDatadogSDSGroupOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	ddGroupList, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting logs index list")
	}
	if err := utils.CheckForUnparsed(ddGroupList); err != nil {
		return diag.FromErr(err)
	}
	groupItems := ddGroupList.Data.Relationships.Groups.Data
	tfList := make([]string, len(groupItems))
	for i, ddGroup := range groupItems {
		tfList[i] = ddGroup.GetId()
	}
	if err := d.Set("groups", tfList); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(ddGroupList.Data.GetId())
	return nil
}
