package datadog

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"net/http"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var monitorComputedFields = []string{
	"id",
	"author_handle",
	"author_name",
	"classification",
	"created",
	"creator",
	"created_at",
	"deleted",
	"modified",
	"modified_at",
	"org_id",
	"overall_state",
	"overall_state_modified",
	"restricted_roles",
	"url",
}

const monitorPath = "/api/v1/monitor"

func resourceDatadogMonitorJSON() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog monitor JSON resource. This can be used to create and manage Datadog monitors using the JSON definition.",
		CreateContext: resourceDatadogMonitorJSONCreate,
		ReadContext:   resourceDatadogMonitorJSONRead,
		UpdateContext: resourceDatadogMonitorJSONUpdate,
		DeleteContext: resourceDatadogMonitorJSONDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"monitor": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					// Remove computed fields when comparing diffs
					attrMap, _ := structure.ExpandJsonFromString(v.(string))
					for _, f := range monitorComputedFields {
						delete(attrMap, f)
					}
					res, _ := structure.FlattenJsonToString(attrMap)
					return res
				},
				Description: "The JSON formatted definition of the Monitor.",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The URL of the monitor.",
			},
		},
	}
}

func resourceDatadogMonitorJSONRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()
	respByte, httpResp, err := utils.SendRequest(authV1, datadogClientV1, "GET", monitorPath+"/"+id, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	return updateMonitorJSONState(d, respMap)
}

func resourceDatadogMonitorJSONCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	monitor := d.Get("monitor").(string)

	respByte, httpresp, err := utils.SendRequest(authV1, datadogClientV1, "POST", monitorPath, &monitor)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating resource")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	id, ok := respMap["id"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving id from response"))
	}
	stringId := fmt.Sprintf("%.0f", id)
	d.SetId(stringId)

	var httpResponse *http.Response
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, httpResponse, err = utils.SendRequest(authV1, datadogClientV1, "GET", monitorPath+"/"+stringId, nil)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return resource.RetryableError(fmt.Errorf("monitor not created yet"))
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return updateMonitorJSONState(d, respMap)
}

func resourceDatadogMonitorJSONUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	monitor := d.Get("monitor").(string)
	id := d.Id()

	respByte, httpresp, err := utils.SendRequest(authV1, datadogClientV1, "PUT", monitorPath+"/"+id, &monitor)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating monitor")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	return updateMonitorJSONState(d, respMap)
}

func resourceDatadogMonitorJSONDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()

	_, httpresp, err := utils.SendRequest(authV1, datadogClientV1, "DELETE", monitorPath+"/"+id, nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting monitor")
	}

	return nil
}

func updateMonitorJSONState(d *schema.ResourceData, monitor map[string]interface{}) diag.Diagnostics {
	if v, ok := monitor["url"]; ok {
		if err := d.Set("url", v.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Remove computed fields from the object
	for _, f := range monitorComputedFields {
		delete(monitor, f)
	}

	monitorString, err := structure.FlattenJsonToString(monitor)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("monitor", monitorString); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
