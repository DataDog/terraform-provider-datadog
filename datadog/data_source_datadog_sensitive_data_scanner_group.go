package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogSensitiveDataScannerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing sensitive data scanner group.",
		ReadContext: dataSourceDatadogSensitiveDataScannerGroupRead,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Description: "Id of the Datadog scanning group to search for.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Computed
			"name": {
				Description: "Name of the Datadog scanning group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description of the Datadog scanning group.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"product_list": {
				Description: "List of products the scanning group applies.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_enabled": {
				Description: "Whether or not the scanning group is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"filter": {
				Description: "Filter object the scanning group applies.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Description: "Query to filter the events.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDatadogSensitiveDataScannerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error listing scanning groups")
	}

	searchedId := d.Get("group_id").(string)

	for _, resource := range resp.Included {
		if *resource.SensitiveDataScannerGroupIncludedItem.Id == searchedId {
			attributes := resource.SensitiveDataScannerGroupIncludedItem.Attributes
			return dataSourceSensitiveDataScannerGroupUpdate(d, attributes)
		}
	}

	return diag.Errorf("Couldn't find the scanning group with id %s", searchedId)

}

func dataSourceSensitiveDataScannerGroupUpdate(d *schema.ResourceData, groupAttributes *datadogV2.SensitiveDataScannerGroupAttributes) diag.Diagnostics {
	if err := d.Set("name", groupAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", groupAttributes.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_enabled", groupAttributes.GetIsEnabled()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("product_list", groupAttributes.GetProductList()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("filter", groupAttributes.GetFilter()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
