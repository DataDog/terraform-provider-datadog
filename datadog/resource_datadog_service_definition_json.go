package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const serviceDefinitionPath = "/api/v2/services/definitions"

type responseDataAttributes struct {
	Schema map[string]interface{}
}

type responseData struct {
	Attributes responseDataAttributes
}

type responseListData struct {
	Data []responseData
}

type responseSingleData struct {
	Data responseData
}

func resourceDatadogServiceDefinitionJSON() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog service definition JSON resource. This can be used to create and manage Datadog service definitions in the service catalog using the JSON definition.",
		CreateContext: resourceDatadogServiceDefinitionJSONCreate,
		ReadContext:   resourceDatadogServiceDefinitionJSONRead,
		UpdateContext: resourceDatadogServiceDefinitionJSONCreate,
		DeleteContext: resourceDatadogServiceDefinitionJSONDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.ForceNewIfChange("definition", func(ctx context.Context, old, new, meta interface{}) bool {
			oldAttrMap, _ := structure.ExpandJsonFromString(old.(string))
			newAttrMap, _ := structure.ExpandJsonFromString(new.(string))

			oldName, ok := oldAttrMap["dd-service"].(string)
			if !ok {
				return true
			}

			newName, ok := newAttrMap["dd-service"].(string)
			if !ok {
				return true
			}

			return oldName != newName
		}),
		Schema: map[string]*schema.Schema{
			"definition": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(validation.StringIsJSON, func(val interface{}, key string) (warns []string, errs []error) {
					definitionMap, _ := structure.ExpandJsonFromString(val.(string))
					ddServiceInterface, ok := definitionMap["dd-service"]
					if !ok {
						errs = append(errs, fmt.Errorf("the definition must include a field called dd-service"))
						return
					}
					ddServiceString, ok := ddServiceInterface.(string)
					if !ok {
						errs = append(errs, fmt.Errorf("the field dd-service must be a string"))
						return
					}
					if ddServiceString != strings.ToLower(ddServiceString) {
						errs = append(errs, fmt.Errorf("the field dd-service must be all lower case"))
					}
					return
				}),
				StateFunc: func(v interface{}) string {
					// Unmarshal and Marshal the definition to get consistent formatting
					definitionMap, _ := structure.ExpandJsonFromString(v.(string))
					definition, _ := structure.FlattenJsonToString(definitionMap)
					return definition
				},
				Description: "The JSON formatted definition of the service.",
			},
		},
	}
}

func resourceDatadogServiceDefinitionJSONRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()
	respByte, httpResp, err := utils.SendRequest(authV1, datadogClientV1, "GET", serviceDefinitionPath+"/"+id, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var responseData responseSingleData
	err = json.Unmarshal(respByte, &responseData)
	if err != nil {
		return diag.FromErr(err)
	}

	return updateServiceDefinitionJSONState(d, responseData.Data)
}

func resourceDatadogServiceDefinitionJSONCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	definition := d.Get("definition").(string)
	definitionMap, _ := structure.ExpandJsonFromString(definition)
	id := definitionMap["dd-service"]

	if d.Id() == "" {
		_, httpResp, _ := utils.SendRequest(authV1, datadogClientV1, "GET", serviceDefinitionPath+"/"+id.(string), nil)
		if httpResp != nil && httpResp.StatusCode != 404 {
			return diag.FromErr(fmt.Errorf("a service with name '%s' already exists", id))
		}
	}

	respByte, httpresp, err := utils.SendRequest(authV1, datadogClientV1, "POST", serviceDefinitionPath, &definition)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating resource")
	}

	var responseData responseListData
	err = json.Unmarshal(respByte, &responseData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(responseData.Data[0].Attributes.Schema["dd-service"].(string))

	return updateServiceDefinitionJSONState(d, responseData.Data[0])
}

func resourceDatadogServiceDefinitionJSONDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	id := d.Id()
	_, httpResp, err := utils.SendRequest(authV1, datadogClientV1, "DELETE", serviceDefinitionPath+"/"+id, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return nil
}

func updateServiceDefinitionJSONState(d *schema.ResourceData, response responseData) diag.Diagnostics {
	serviceString, err := structure.FlattenJsonToString(response.Attributes.Schema)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("definition", serviceString); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
