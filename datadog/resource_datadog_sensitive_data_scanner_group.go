package datadog

import (
	"context"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var sensitiveDataScannerMutex = sync.Mutex{}

func resourceDatadogSensitiveDataScannerGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Sensitive Data Scanner group resource.",
		ReadContext:   resourceDatadogSensitiveDataScannerGroupRead,
		CreateContext: resourceDatadogSensitiveDataScannerGroupCreate,
		UpdateContext: resourceDatadogSensitiveDataScannerGroupUpdate,
		DeleteContext: resourceDatadogSensitiveDataScannerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
					Type:        schema.TypeSet,
					Required:    true,
					MaxItems:    4,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSensitiveDataScannerProductFromValue),
					},
				},
				"is_enabled": {
					Description: "Whether or not the scanning group is enabled. If the group doesn't contain any rule or if all the rules in it are disabled, the group is force-disabled by our backend",
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
								DiffSuppressFunc: func(_, oldVal, newVal string, d *schema.ResourceData) bool {
									if (oldVal == "" && newVal == "*") || (oldVal == "*" && newVal == "") {
										return true
									}
									return false
								},
							},
						},
					},
				},
			}
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

func buildScanningGroupAttributes(d *schema.ResourceData) *datadogV2.SensitiveDataScannerGroupAttributes {
	attributes := &datadogV2.SensitiveDataScannerGroupAttributes{}

	attributes.SetIsEnabled(d.Get("is_enabled").(bool))
	attributes.SetName(d.Get("name").(string))
	attributes.SetDescription(d.Get("description").(string))

	if description, ok := d.GetOk("description"); ok {
		attributes.SetDescription(description.(string))
	}

	if tfFilter := d.Get("filter").([]interface{}); len(tfFilter) > 0 && tfFilter[0] != nil {
		attributes.SetFilter(*buildDatadogGroupFilter(tfFilter[0].(map[string]interface{})))
	} else {
		filter := datadogV2.NewSensitiveDataScannerFilterWithDefaults()
		filter.SetQuery("*")
		attributes.SetFilter(*filter)
	}
	productList := make([]datadogV2.SensitiveDataScannerProduct, 0)
	if pList, ok := d.GetOk("product_list"); ok {
		for _, s := range pList.(*schema.Set).List() {
			sensitiveDataScannerProductItem, _ := datadogV2.NewSensitiveDataScannerProductFromValue(s.(string))
			productList = append(productList, *sensitiveDataScannerProductItem)
		}
		attributes.SetProductList(productList)
	} else {
		attributes.SetProductList(nil)
	}

	return attributes
}

func resourceDatadogSensitiveDataScannerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	resp, httpResponse, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error calling ListScanningGroups")
	}

	groupId := d.Id()

	if groupFound := findSensitiveDataScannerGroupHelper(groupId, resp); groupFound != nil {
		return updateSensitiveDataScannerGroupState(d, groupFound.Attributes)
	}

	return nil
}

func resourceDatadogSensitiveDataScannerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	sensitiveDataScannerMutex.Lock()
	defer sensitiveDataScannerMutex.Unlock()

	body := buildSensitiveDataScannerGroupCreateRequestBody(d)

	resp, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().CreateScanningGroup(auth, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating SensitiveDataScannerGroup")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateSensitiveDataScannerGroupState(d, resp.Data.Attributes)
}

func buildSensitiveDataScannerGroupCreateRequestBody(d *schema.ResourceData) *datadogV2.SensitiveDataScannerGroupCreateRequest {
	attributes := buildScanningGroupAttributes(d)

	req := datadogV2.NewSensitiveDataScannerGroupCreateRequestWithDefaults()
	req.Data = datadogV2.NewSensitiveDataScannerGroupCreateWithDefaults()
	req.Data.SetAttributes(*attributes)
	req.Meta = datadogV2.NewSensitiveDataScannerMetaVersionOnly()

	return req
}

func resourceDatadogSensitiveDataScannerGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	sensitiveDataScannerMutex.Lock()
	defer sensitiveDataScannerMutex.Unlock()

	id := d.Id()

	body := buildSensitiveDataScannerGroupUpdateRequestBody(d)

	resp, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().UpdateScanningGroup(auth, id, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error updating SensitiveDataScannerGroup")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	return updateSensitiveDataScannerGroupState(d, body.Data.Attributes)
}

func buildSensitiveDataScannerGroupUpdateRequestBody(d *schema.ResourceData) *datadogV2.SensitiveDataScannerGroupUpdateRequest {
	attributes := buildScanningGroupAttributes(d)

	req := datadogV2.NewSensitiveDataScannerGroupUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewSensitiveDataScannerGroupUpdateWithDefaults()
	req.Data.SetAttributes(*attributes)
	req.Data.SetId(d.Id())

	req.Meta = *datadogV2.NewSensitiveDataScannerMetaVersionOnly()

	return req
}

func resourceDatadogSensitiveDataScannerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	sensitiveDataScannerMutex.Lock()
	defer sensitiveDataScannerMutex.Unlock()

	id := d.Id()
	body := datadogV2.NewSensitiveDataScannerGroupDeleteRequestWithDefaults()
	metaVar := datadogV2.NewSensitiveDataScannerMetaVersionOnlyWithDefaults()
	body.SetMeta(*metaVar)

	_, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().DeleteScanningGroup(auth, id, *body)
	if err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResp, "error deleting SensitiveDataScannerGroup")
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
	for _, resource := range response.GetIncluded() {
		if resource.SensitiveDataScannerGroupIncludedItem.GetId() == groupId {
			return resource.SensitiveDataScannerGroupIncludedItem
		}
	}

	return nil
}
