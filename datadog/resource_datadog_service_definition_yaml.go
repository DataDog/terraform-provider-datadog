package datadog

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

const serviceDefinitionPath = "/api/v2/services/definitions"
const rawServiceDefinitionPath = "/api/v2/services/raw_definitions"

var fieldsWithName = []string{"contacts", "repos", "docs", "links"}

type sdAttribute struct {
	Schema map[string]interface{} `json:"schema"`
	Meta   sdMeta                 `json:"meta"`
}

type sdMeta struct {
	IngestedSchemaVersion string `json:"ingested-schema-version"`
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

type getSDRawResponse struct {
	RawData sdRawData `json:"data"`
}

type sdRawAttribute struct {
	RawContent string `json:"raw-content"`
}

type sdRawData struct {
	Attributes sdRawAttribute `json:"attributes"`
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

			oldName, ok := getServiceName(oldAttrMap)
			if !ok {
				return true
			}

			newName, ok := getServiceName(newAttrMap)
			if !ok {
				return true
			}
			return utils.NormalizeTag(oldName) != utils.NormalizeTag(newName)
		}),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
			}
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
	if isBackstageSchema(attrMap) {
		// Don't prepare 3rd party schemas
		return attrMap
	}

	// this assumes we only support >= v2
	delete(attrMap, "dd-team") //dd-team is a computed field

	for _, field := range fieldsWithName {
		normalizeArrayField(attrMap, field)
	}

	if service, ok := getServiceName(attrMap); ok {
		attrMap["dd-service"] = utils.NormalizeTag(service)
	}

	if tags, ok := attrMap["tags"].([]interface{}); ok {
		if len(tags) == 0 {
			delete(attrMap, "tags")
		} else {
			normalizedTags := make([]string, 0)
			for _, tag := range tags {
				normalizedTags = append(normalizedTags, utils.NormalizeTag(tag.(string)))
			}
			sort.SliceStable(normalizedTags, func(i, j int) bool {
				return normalizedTags[i] < normalizedTags[j]
			})
			attrMap["tags"] = normalizedTags
		}
	}

	if contacts, ok := attrMap["contacts"].([]interface{}); ok {
		if len(contacts) == 0 {
			delete(attrMap, "contacts")
		} else {
			sortedContacts := make([]map[string]interface{}, 0)
			for _, contact := range contacts {
				sortedContacts = append(sortedContacts, contact.(map[string]interface{}))
			}

			sort.SliceStable(sortedContacts, func(i, j int) bool {
				typeLVal, typeLOk := sortedContacts[i]["type"]
				typeRVal, typeROk := sortedContacts[j]["type"]
				if typeLVal != typeRVal && typeLOk && typeROk {
					return typeLVal.(string) < typeRVal.(string)
				}
				contactLVal, contactLOk := sortedContacts[i]["contact"]
				contactRVal, contactROk := sortedContacts[j]["contact"]
				if contactLVal != contactRVal && contactLOk && contactROk {
					return contactLVal.(string) < contactRVal.(string)
				}
				nameLVal, nameLOk := sortedContacts[i]["name"]
				nameRVal, nameROk := sortedContacts[j]["name"]
				if nameLOk && nameROk {
					return nameLVal.(string) < nameRVal.(string)
				}
				return false
			})
			attrMap["contacts"] = sortedContacts
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
		if name, ok := stringMap["name"]; ok {
			return name.(string)
		}
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

	if isDatadogServiceSchema(attrMap) {
		return isValidDatadogServiceDefinition(attrMap, k)
	} else if isBackstageSchema(attrMap) {
		return isValidBackstageServiceDefinition(attrMap, k)
	} else {
		errors = append(errors, fmt.Errorf("Must be a supported service schema: %s", k))
	}
	return warnings, errors
}

func isValidDatadogServiceDefinition(attrMap map[string]interface{}, k string) (warnings []string, errors []error) {
	if schemaVersion, ok := attrMap["schema-version"].(string); ok {
		if schemaVersion != "v2" && schemaVersion != "v2.1" {
			errors = append(errors, fmt.Errorf("schema-version must be >= v2, but %s is used", schemaVersion))
		}
	} else {
		errors = append(errors, fmt.Errorf("schema-version is missing: %q", k))
	}

	if schemaVersion, ok := attrMap["dd-service"].(string); !ok || schemaVersion == "" {
		errors = append(errors, fmt.Errorf("dd-service is missing: %q", k))
	}

	return warnings, errors
}

func isValidBackstageServiceDefinition(attrMap map[string]interface{}, k string) (warnings []string, errors []error) {
	if apiVersion, ok := attrMap["apiVersion"].(string); ok {
		if apiVersion != "backstage.io/v1alpha1" {
			errors = append(errors, fmt.Errorf("apiVersion must be backstage.io/v1alpha1, but %s is used", apiVersion))
		}
	} else {
		errors = append(errors, fmt.Errorf("apiVersion is missing: %q", k))
	}

	if kind, ok := attrMap["kind"].(string); ok {
		if kind != "Component" {
			errors = append(errors, fmt.Errorf("kind must be Component, but %s is used", kind))
		}
	} else {
		errors = append(errors, fmt.Errorf("kind is missing: %q", k))
	}

	if spec, ok := attrMap["spec"].(map[string]interface{}); ok {
		if _, okay := spec["type"].(string); !okay {
			errors = append(errors, fmt.Errorf("spec.type is missing: %q", k))
		}
	} else {
		errors = append(errors, fmt.Errorf("spec is missing: %q", k))
	}

	if metadata, ok := attrMap["metadata"].(map[string]interface{}); ok {
		if _, okay := metadata["name"].(string); !okay {
			errors = append(errors, fmt.Errorf("metadata.name is missing: %q", k))
		}
	} else {
		errors = append(errors, fmt.Errorf("metadata is missing: %q", k))
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
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return utils.TranslateClientErrorDiag(err, resp, fmt.Sprintf("error retrieving service definition %s", id))
	}

	var response getSDResponse

	err = json.Unmarshal(respByte, &response)
	if err != nil {
		return diag.FromErr(err)
	}

	schema := response.Data.Attributes.Schema
	if strings.HasPrefix(response.Data.Attributes.Meta.IngestedSchemaVersion, "backstage.io") {
		rawResp, resp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", rawServiceDefinitionPath+"/"+id, nil)
		if err != nil {
			if resp.StatusCode == 404 {
				d.SetId("")
				return nil
			}

			return utils.TranslateClientErrorDiag(err, resp, fmt.Sprintf("error retrieving service definition %s", id))
		}
		var rawResponse getSDRawResponse
		err = json.Unmarshal(rawResp, &rawResponse)
		if err != nil {
			return diag.FromErr(err)
		}

		rawContent, err := base64.StdEncoding.DecodeString(rawResponse.RawData.Attributes.RawContent)
		if err != nil {
			return diag.FromErr(err)
		}

		rawSchema, err := expandYAMLFromString(string(rawContent))
		if err != nil {
			return diag.FromErr(err)
		}
		schema = rawSchema
	}

	return updateServiceDefinitionState(d, schema)
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

	raw, err := expandYAMLFromString(definition)
	if err != nil {
		return diag.FromErr(err)
	}

	if isBackstageSchema(raw) {
		return updateServiceDefinitionState(d, raw)
	}
	return updateServiceDefinitionState(d, response.Data[0].Attributes.Schema)
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

	raw, err := expandYAMLFromString(definition)
	if err != nil {
		return diag.FromErr(err)
	}

	if isBackstageSchema(raw) {
		return updateServiceDefinitionState(d, raw)
	}
	return updateServiceDefinitionState(d, response.Data[0].Attributes.Schema)
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

func updateServiceDefinitionState(d *schema.ResourceData, attrMap map[string]interface{}) diag.Diagnostics {
	schema := prepServiceDefinitionResource(attrMap)
	serviceDefinition, err := flattenYAMLToString(schema)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("service_definition", serviceDefinition); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func getServiceName(attrMap map[string]interface{}) (string, bool) {
	if isDatadogServiceSchema(attrMap) {
		service, ok := attrMap["dd-service"].(string)
		return service, ok
	} else if isBackstageSchema(attrMap) {
		if metadata, ok := attrMap["metadata"].(map[string]interface{}); ok {
			service, okay := metadata["name"].(string)
			return service, okay
		}
	}
	return "", false
}

func isDatadogServiceSchema(attrMap map[string]interface{}) bool {
	if _, ok := attrMap["schema-version"]; ok {
		return true
	}
	return false
}

func isBackstageSchema(attrMap map[string]interface{}) bool {
	if apiVersion, ok := attrMap["apiVersion"].(string); ok {
		return strings.HasPrefix(apiVersion, "backstage.io")
	}
	return false
}
