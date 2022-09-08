package datadog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

const serviceDefinitionPath = "/api/v2/services/definitions"

type sdAttribute struct {
	Schema map[string]interface{} `json:"schema"`
}

type sdData struct {
	Attributes sdAttribute `json:"attributes"`
}

type createSDResponse struct {
	Data []sdData `json:"data"`
}

type getSDResponse struct {
	Data sdData `json:"data"`
}

func resourceDatadogServiceDefinition() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog service definition resource. This can be used to create and manage Datadog service definitions in the service catalog using the YAML/JSON definition.",
		CreateContext: resourceDatadogServiceDefinitionInsert,
		ReadContext:   resourceDatadogServiceDefinitionGet,
		UpdateContext: resourceDatadogServiceDefinitionUpdate,
		DeleteContext: resourceDatadogServiceDefinitionDelete,
		CustomizeDiff: customdiff.ForceNewIfChange("service_definition", func(ctx context.Context, old, new, meta interface{}) bool {
			oldAttrMap, _ := expandYAMLFromString(old.(string))
			newAttrMap, _ := expandYAMLFromString(new.(string))

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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"service_definition": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: stringIsYAML,
				StateFunc: func(v interface{}) string {
					attrMap, _ := expandYAMLFromString(v.(string))
					prepServiceDefinitionResource(attrMap)
					res, _ := flattenYAMLToString(attrMap)
					return res
				},
				Description: "Service Definition YAML (Single)",
			},
		},
	}
}

func expandYAMLFromString(yamlString string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := yaml.Unmarshal([]byte(yamlString), &result)

	return result, err
}

func flattenYAMLToString(input map[string]interface{}) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	result, err := yaml.Marshal(input)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func prepServiceDefinitionResource(attrMap map[string]interface{}) map[string]interface{} {
	return attrMap
}

func stringIsYAML(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	var j interface{}
	if err := yaml.Unmarshal([]byte(v), &j); err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid YAML/JSON: %s", k, err))
	}

	return warnings, errors
}

func resourceDatadogServiceDefinitionGet(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()
	respByte, resp, err := utils.SendRequest(authV1, datadogClientV1, "GET", serviceDefinitionPath+"/"+id, nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, fmt.Sprintf("error retrieving service definition %s", id))
	}

	if resp.StatusCode == 404 {
		d.SetId("")
		return nil
	}

	var response getSDResponse

	err = json.Unmarshal(respByte, &response)
	if err != nil {
		return diag.FromErr(err)
	}

	return updateServiceDefinitionState(d, response.Data)
}

func resourceDatadogServiceDefinitionInsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	definition := d.Get("service_definition").(string)

	respByte, resp, err := utils.SendRequest(authV1, datadogClientV1, "POST", serviceDefinitionPath, &definition)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error creating service definition")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return utils.TranslateClientErrorDiag(err, resp, "error creating service definition, received an error from the Service Catalog endpoint")
	}

	var response createSDResponse
	err = json.Unmarshal(respByte, &response)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(response.Data) != 1 {
		return diag.FromErr(errors.New("error retrieving data from response"))
	}

	d.SetId(response.Data[0].Attributes.Schema["dd-service"].(string))
	return updateServiceDefinitionState(d, response.Data[0])
}

func resourceDatadogServiceDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	definition := d.Get("service_definition").(string)

	respByte, resp, err := utils.SendRequest(authV1, datadogClientV1, "POST", serviceDefinitionPath, &definition)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error updating service definition")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return utils.TranslateClientErrorDiag(err, resp, "error updating service definition, received an error from the Service Catalog endpoint")
	}

	var response createSDResponse
	err = json.Unmarshal(respByte, &response)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(response.Data) != 1 {
		return diag.FromErr(errors.New("error retrieving data from response"))
	}

	return updateServiceDefinitionState(d, response.Data[0])
}

func resourceDatadogServiceDefinitionDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	id := d.Id()
	_, resp, err := utils.SendRequest(authV1, datadogClientV1, "DELETE", serviceDefinitionPath+"/"+id, nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error deleting service definition")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return utils.TranslateClientErrorDiag(err, resp, "error deleting service definition, received an error from the Service Catalog endpoint")
	}
	return nil
}

func updateServiceDefinitionState(d *schema.ResourceData, response sdData) diag.Diagnostics {
	serviceString, err := flattenYAMLToString(response.Attributes.Schema)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("service_definition", serviceString); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
