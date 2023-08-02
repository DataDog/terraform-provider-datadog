package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

func resourceDatadogSensitiveDataScannerRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog SensitiveDataScannerRule resource. This can be used to create and manage Datadog sensitive_data_scanner_rule.",
		ReadContext:   resourceDatadogSensitiveDataScannerRuleRead,
		CreateContext: resourceDatadogSensitiveDataScannerRuleCreate,
		UpdateContext: resourceDatadogSensitiveDataScannerRuleUpdate,
		DeleteContext: resourceDatadogSensitiveDataScannerRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Name of the rule.",
				},
				"description": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Description of the rule.",
				},
				"group_id": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
					Description: "Id of the scanning group the rule belongs to.",
				},
				"standard_pattern_id": {
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
					Description: "Id of the standard pattern the rule refers to. If provided, then pattern must not be provided.",
				},
				"excluded_namespaces": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Attributes excluded from the scan. If namespaces is provided, it has to be a sub-path of the namespaces array.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"namespaces": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Attributes included in the scan. If namespaces is empty or missing, all attributes except excluded_namespaces are scanned. If both are missing the whole event is scanned.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"is_enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether or not the rule is enabled.",
				},
				"pattern": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Not included if there is a relationship to a standard pattern.",
				},
				"tags": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "List of tags.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"text_replacement": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Object describing how the scanned event will be replaced. Defaults to `type: none`",
					DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
						old, new := d.GetChange("text_replacement.0.type")
						oldS := old.(string)
						newS := new.(string)
						if (oldS == "" && newS == "none") || (oldS == "none" && newS == "") || (oldS == "none" && newS == "none") {
							return true
						}
						return false
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"number_of_chars": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Required if type == 'partial_replacement_from_beginning' or 'partial_replacement_from_end'. It must be > 0.",
							},
							"replacement_string": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Required if type == 'replacement_string'.",
							},
							"type": {
								Type:             schema.TypeString,
								Required:         true,
								Description:      "Type of the replacement text. None means no replacement. hash means the data will be stubbed. replacement_string means that one can chose a text to replace the data. partial_replacement_from_beginning allows a user to partially replace the data from the beginning, and partial_replacement_from_end on the other hand, allows to replace data from the end.",
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSensitiveDataScannerTextReplacementTypeFromValue),
							},
						},
					},
				},
			}
		},
	}
}

func resourceDatadogSensitiveDataScannerRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResp, "error calling ListScanningGroups")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	ruleId := d.Id()

	if ruleFound := findSensitiveDataScannerRuleHelper(ruleId, resp); ruleFound != nil {

		if err := d.Set("group_id", *ruleFound.Relationships.Group.Data.Id); err != nil {
			return diag.FromErr(err)
		}
		if standardPattern, ok := ruleFound.Relationships.GetStandardPatternOk(); ok {
			if standardPattern.Data.Id != nil {
				if err := d.Set("standard_pattern_id", *standardPattern.Data.Id); err != nil {
					return diag.FromErr(err)
				}
			}
		}
		return updateSensitiveDataScannerRuleState(d, ruleFound.Attributes)
	}

	return nil
}

func resourceDatadogSensitiveDataScannerRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	sensitiveDataScannerMutex.Lock()
	defer sensitiveDataScannerMutex.Unlock()

	attributes := buildSensitiveDataScannerRuleAttributes(d)

	req := datadogV2.NewSensitiveDataScannerRuleCreateRequestWithDefaults()
	req.Data = *datadogV2.NewSensitiveDataScannerRuleCreateWithDefaults()

	relationships := datadogV2.NewSensitiveDataScannerRuleRelationshipsWithDefaults()

	groupData := datadogV2.NewSensitiveDataScannerGroupDataWithDefaults()
	if groupId, ok := d.GetOk("group_id"); ok {
		groupRelationship := datadogV2.NewSensitiveDataScannerGroup()
		groupRelationship.SetId(groupId.(string))
		groupData.SetData(*groupRelationship)
		relationships.SetGroup(*groupData)
	}

	spData := datadogV2.NewSensitiveDataScannerStandardPatternData()
	if spId, ok := d.GetOk("standard_pattern_id"); ok {
		spRelationship := datadogV2.NewSensitiveDataScannerStandardPattern()
		spRelationship.SetId(spId.(string))
		spData.SetData(*spRelationship)
		relationships.SetStandardPattern(*spData)

	}

	req.Data.SetAttributes(*attributes)
	req.Data.SetRelationships(*relationships)

	resp, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().CreateScanningRule(auth, *req)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating SensitiveDataScannerRule")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	if err := d.Set("group_id", *resp.Data.Relationships.Group.Data.Id); err != nil {
		return diag.FromErr(err)
	}
	if standardPattern, ok := resp.GetData().Relationships.GetStandardPatternOk(); ok {
		if standardPattern.Data.Id != nil {
			if err := d.Set("standard_pattern_id", *standardPattern.Data.Id); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return updateSensitiveDataScannerRuleState(d, resp.Data.Attributes)
}

func buildSensitiveDataScannerRuleAttributes(d *schema.ResourceData) *datadogV2.SensitiveDataScannerRuleAttributes {
	attributes := datadogV2.NewSensitiveDataScannerRuleAttributesWithDefaults()

	if description, ok := d.GetOk("description"); ok {
		attributes.SetDescription(description.(string))
	}

	namespaces := []string{}
	for _, s := range d.Get("namespaces").([]interface{}) {
		namespaces = append(namespaces, s.(string))
	}
	attributes.SetNamespaces(namespaces)

	excludedNamespaces := []string{}
	for _, s := range d.Get("excluded_namespaces").([]interface{}) {
		if s == nil {
			// sdkv2 treats empty strings in list as nils so
			// append an empty string
			excludedNamespaces = append(excludedNamespaces, "")
		} else {
			excludedNamespaces = append(excludedNamespaces, s.(string))
		}
	}
	attributes.SetExcludedNamespaces(excludedNamespaces)

	if isEnabled := d.Get("is_enabled"); isEnabled != nil {
		attributes.SetIsEnabled(isEnabled.(bool))
	}

	if name, ok := d.GetOk("name"); ok {
		attributes.SetName(name.(string))
	}

	if pattern, ok := d.GetOk("pattern"); ok {
		attributes.SetPattern(pattern.(string))
	}
	tags := []string{}
	for _, s := range d.Get("tags").([]interface{}) {
		tags = append(tags, s.(string))
	}
	attributes.SetTags(tags)

	var textReplacement datadogV2.SensitiveDataScannerTextReplacement
	if _, ok := d.GetOk("text_replacement"); ok {
		if numberOfChars, ok := d.GetOk("text_replacement.0.number_of_chars"); ok {
			textReplacement.SetNumberOfChars(int64(numberOfChars.(int)))
		}

		if replacementString, ok := d.GetOk("text_replacement.0.replacement_string"); ok {
			textReplacement.SetReplacementString(replacementString.(string))
		}

		if typeVar, ok := d.GetOk("text_replacement.0.type"); ok {
			typeVarItem, _ := datadogV2.NewSensitiveDataScannerTextReplacementTypeFromValue(typeVar.(string))
			textReplacement.SetType(*typeVarItem)
		}
	} else {
		textReplacement.Type = datadogV2.SENSITIVEDATASCANNERTEXTREPLACEMENTTYPE_NONE.Ptr()
	}

	attributes.SetTextReplacement(textReplacement)

	return attributes
}

func resourceDatadogSensitiveDataScannerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	sensitiveDataScannerMutex.Lock()
	defer sensitiveDataScannerMutex.Unlock()

	id := d.Id()

	attributes := buildSensitiveDataScannerRuleAttributes(d)

	req := datadogV2.NewSensitiveDataScannerRuleUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewSensitiveDataScannerRuleUpdateWithDefaults()
	req.Data.SetAttributes(*attributes)
	req.Data.SetId(id)

	resp, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().UpdateScanningRule(auth, id, *req)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error updating SensitiveDataScannerRule")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	return updateSensitiveDataScannerRuleState(d, req.Data.Attributes)
}

func resourceDatadogSensitiveDataScannerRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	sensitiveDataScannerMutex.Lock()
	defer sensitiveDataScannerMutex.Unlock()

	id := d.Id()
	body := datadogV2.NewSensitiveDataScannerRuleDeleteRequestWithDefaults()

	_, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().DeleteScanningRule(auth, id, *body)
	if err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResp, "error deleting SensitiveDataScannerRule")
	}

	return nil
}

func updateSensitiveDataScannerRuleState(d *schema.ResourceData, ruleAttributes *datadogV2.SensitiveDataScannerRuleAttributes) diag.Diagnostics {
	if err := d.Set("name", ruleAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", ruleAttributes.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_enabled", ruleAttributes.GetIsEnabled()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("excluded_namespaces", ruleAttributes.GetExcludedNamespaces()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("namespaces", ruleAttributes.GetNamespaces()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pattern", ruleAttributes.GetPattern()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", ruleAttributes.GetTags()); err != nil {
		return diag.FromErr(err)
	}

	if tR, ok := ruleAttributes.GetTextReplacementOk(); ok && tR != nil {
		textReplacement := make(map[string]interface{})
		textReplacementList := make([]map[string]interface{}, 0, 1)

		if numberOfChars, ok := tR.GetNumberOfCharsOk(); ok {
			textReplacement["number_of_chars"] = numberOfChars
		}
		if replacementString, ok := tR.GetReplacementStringOk(); ok {
			textReplacement["replacement_string"] = replacementString
		}
		if replacementType, ok := tR.GetTypeOk(); ok {
			textReplacement["type"] = *replacementType
		}
		textReplacementList = append(textReplacementList, textReplacement)
		if err := d.Set("text_replacement", textReplacementList); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func findSensitiveDataScannerRuleHelper(ruleId string, response datadogV2.SensitiveDataScannerGetConfigResponse) *datadogV2.SensitiveDataScannerRuleIncludedItem {
	for _, resource := range response.GetIncluded() {
		if resource.SensitiveDataScannerRuleIncludedItem.GetId() == ruleId {
			return resource.SensitiveDataScannerRuleIncludedItem
		}
	}

	return nil
}
