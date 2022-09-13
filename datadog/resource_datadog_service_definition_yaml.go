package datadog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

const serviceDefinitionPath = "/api/v2/services/definitions"

var fieldsWithName = []string{"contacts", "repos", "docs", "links"}

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

func resourceDatadogServiceDefinitionYAML() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog service definition resource. This can be used to create and manage Datadog service definitions in the service catalog using the YAML/JSON definition.",
		CreateContext: resourceDatadogServiceDefinitionCreate,
		ReadContext:   resourceDatadogServiceDefinitionRead,
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
				ValidateFunc: isValidServiceDefinition,
				StateFunc: func(v interface{}) string {
					attrMap, _ := expandYAMLFromString(v.(string))
					prepServiceDefinitionResource(attrMap)
					res, _ := flattenYAMLToString(attrMap)
					return res
				},
				Description: "The YAML/JSON formatted definition of the service",
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
	// this assumes we only support v2
	delete(attrMap, "dd-team") //dd-team is a computed field

	for _, field := range fieldsWithName {
		normalizeArrayField(attrMap, field)
	}

	if tags, ok := attrMap["tags"].([]interface{}); ok {
		if len(tags) == 0 {
			delete(attrMap, "tags")
		} else {
			sort.SliceStable(tags, func(i, j int) bool {
				return tags[i].(string) < tags[j].(string)
			})
		}
	}

	if team, ok := attrMap["team"].(string); ok {
		if team == "" {
			delete(attrMap, "team")
		}
	}
	if extensions, ok := attrMap["extensions"].(map[string]interface{}); ok {
		if len(extensions) == 0 {
			delete(attrMap, "extensions")
		}
	}
	if integrations, ok := attrMap["integrations"].(map[string]interface{}); ok {
		if len(integrations) == 0 {
			delete(attrMap, "integrations")
		}
	}
	return attrMap
}

func normalizeArrayField(attrMap map[string]interface{}, key string) {
	if items, ok := attrMap[key].([]interface{}); ok {
		if len(items) == 0 {
			delete(attrMap, key)
		} else {
			sort.SliceStable(items, func(i, j int) bool {
				name1 := getNameField(items[i])
				name2 := getNameField(items[j])
				return name1 < name2
			})
		}
	}
}

func getNameField(data interface{}) string {
	if stringMap, ok := data.(map[string]interface{}); ok {
		return stringMap["name"].(string)
	}
	return ""
}

func isValidServiceDefinition(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	var attrMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(v), &attrMap); err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid YAML/JSON: %s", k, err))
		return warnings, errors
	}

	if schemaVersion, ok := attrMap["schema-version"].(string); ok {
		if schemaVersion != "v2" {
			errors = append(errors, fmt.Errorf("schema-version must be v2, but %s is used", schemaVersion))
		}
	} else {
		errors = append(errors, fmt.Errorf("schema-version is missing: %q", k))
	}

	if schemaVersion, ok := attrMap["dd-service"].(string); !ok || schemaVersion == "" {
		errors = append(errors, fmt.Errorf("dd-service is missing: %q", k))
	}

	return warnings, errors
}

func resourceDatadogServiceDefinitionRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()
	respByte, resp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", serviceDefinitionPath+"/"+id, nil)
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

func resourceDatadogServiceDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	definition := d.Get("service_definition").(string)

	respByte, resp, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", serviceDefinitionPath, &definition)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error creating service definition")
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
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	definition := d.Get("service_definition").(string)

	respByte, resp, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", serviceDefinitionPath, &definition)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error updating service definition")
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
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()
	_, resp, err := utils.SendRequest(auth, apiInstances.HttpClient, "DELETE", serviceDefinitionPath+"/"+id, nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, resp, "error deleting service definition")
	}

	return nil
}

func updateServiceDefinitionState(d *schema.ResourceData, response sdData) diag.Diagnostics {
	schema := prepServiceDefinitionResource(response.Attributes.Schema)
	serviceDefinition, err := flattenYAMLToString(schema)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("service_definition", serviceDefinition); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
