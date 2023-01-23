package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogSensitiveDataScannerGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Sensitive Data Scanner group resource.",
		ReadContext:   resourceDatadogSDSGroupRead,
		CreateContext: resourceDatadogSDSGroupCreate,
		UpdateContext: resourceDatadogSDSGroupUpdate,
		DeleteContext: resourceDatadogSDSGroupDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the Datadog scanning group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the Datadog scanning group.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"product_list": {
				Description: "List of products the scanning group applies.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    4,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_enabled": {
				Description: "Whether or not the scanning group is enabled.",
				Type:        schema.TypeBool,
				Required:    true,
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

func buildTerraformGroupFilter(ddFilter datadogV2.SensitiveDataScannerFilter) *[]map[string]interface{} {
	tfFilter := map[string]interface{}{
		"query": ddFilter.GetQuery(),
	}
	return &[]map[string]interface{}{tfFilter}
}

func buildDatadogGroupFilter(tfFilter map[string]interface{}) *datadogV2.SensitiveDataScannerFilter {
	ddFilter := datadogV2.NewSensitiveDataScannerFilterWithDefaults()
	if tfQuery, exists := tfFilter["query"].(string); exists {
		ddFilter.SetQuery(tfQuery)
	}
	return ddFilter
}

func buildDatadogSDSProductList(tfProductList []interface{}) *[]datadogV2.SensitiveDataScannerProduct {
	var ddProductList []datadogV2.SensitiveDataScannerProduct
	for _, product := range tfProductList {
		if sdsProduct, _ := datadogV2.NewSensitiveDataScannerProductFromValue(product.(string)); sdsProduct != nil {
			ddProductList = append(ddProductList, *sdsProduct)
		}
	}
	return &ddProductList
}

func buildScanningGroupAttributes(d *schema.ResourceData) *datadogV2.SensitiveDataScannerGroupAttributes {
	attributes := &datadogV2.SensitiveDataScannerGroupAttributes{}
	if tfFilter := d.Get("filter").([]interface{}); len(tfFilter) > 0 {
		attributes.SetFilter(*buildDatadogGroupFilter(tfFilter[0].(map[string]interface{})))
	}

	attributes.SetName(d.Get("name").(string))
	attributes.SetDescription(d.Get("description").(string))
	attributes.SetIsEnabled(d.Get("is_enabled").(bool))
	attributes.SetProductList(*buildDatadogSDSProductList(d.Get("product_list").([]interface{})))

	return attributes
}

func buildDatadogScanningGroupCreate(d *schema.ResourceData) *datadogV2.SensitiveDataScannerGroupCreate {
	attributes := buildScanningGroupAttributes(d)

	result := datadogV2.NewSensitiveDataScannerGroupCreateWithDefaults()
	result.SetAttributes(*attributes)
	return result
}

func buildDatadogScanningGroupUpdate(d *schema.ResourceData) *datadogV2.SensitiveDataScannerGroupUpdate {
	attributes := buildScanningGroupAttributes(d)

	result := datadogV2.NewSensitiveDataScannerGroupUpdateWithDefaults()
	result.SetAttributes(*attributes)
	result.SetId(d.Id())

	return result
}

func resourceDatadogSDSGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	resp, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error reading scanning group")
	}

	groupId := d.Id()

	if groupFound := findSensitiveDataScannerGroupHelper(groupId, resp); groupFound != nil {
		return updateSensitiveDataScannerGroupState(d, groupFound.Attributes)
	}

	return nil
}

func resourceDatadogSDSGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupCreateData := buildDatadogScanningGroupCreate(d)

	body := datadogV2.NewSensitiveDataScannerGroupCreateRequestWithDefaults()
	body.SetData(*groupCreateData)

	metaObject := datadogV2.NewSensitiveDataScannerMetaVersionOnlyWithDefaults()
	body.SetMeta(*metaObject)

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	resp, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().CreateScanningGroup(auth, *body)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating scanning group")
	}

	scanningGroup := resp.Data

	d.SetId(*scanningGroup.Id)

	return updateSensitiveDataScannerGroupState(d, scanningGroup.Attributes)
}

func resourceDatadogSDSGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupUpdateData := buildDatadogScanningGroupUpdate(d)

	body := datadogV2.NewSensitiveDataScannerGroupUpdateRequestWithDefaults()
	body.SetData(*groupUpdateData)

	metaObject := datadogV2.NewSensitiveDataScannerMetaVersionOnlyWithDefaults()
	body.SetMeta(*metaObject)

	groupId := d.Id()

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	resp, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().UpdateScanningGroup(auth, groupId, *body)
	// resp is unused
	_ = resp

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating scanning group")
	}

	return updateSensitiveDataScannerGroupState(d, body.Data.Attributes)
}

func resourceDatadogSDSGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	body := datadogV2.NewSensitiveDataScannerGroupDeleteRequestWithDefaults()

	metaObject := datadogV2.NewSensitiveDataScannerMetaVersionOnlyWithDefaults()
	body.SetMeta(*metaObject)

	groupId := d.Id()

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	resp, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().DeleteScanningGroup(auth, groupId, *body)
	// resp is unused
	_ = resp

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting scanning group")
	}

	return nil
}

func updateSensitiveDataScannerGroupState(d *schema.ResourceData, groupAttributes *datadogV2.SensitiveDataScannerGroupAttributes) diag.Diagnostics {
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
	if err := d.Set("filter", buildTerraformGroupFilter(groupAttributes.GetFilter())); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func findSensitiveDataScannerGroupHelper(groupId string, response datadogV2.SensitiveDataScannerGetConfigResponse) *datadogV2.SensitiveDataScannerGroupIncludedItem {
	for _, resource := range response.Included {
		if *resource.SensitiveDataScannerGroupIncludedItem.Id == groupId {
			return resource.SensitiveDataScannerGroupIncludedItem
		}
	}

	return nil
}
