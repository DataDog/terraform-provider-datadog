package datadog

import (
	"context"
	"regexp"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatadogSyntheticsGlobalVariable() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog synthetics global variable resource. This can be used to create and manage Datadog synthetics global variables.",
		CreateContext: resourceDatadogSyntheticsGlobalVariableCreate,
		ReadContext:   resourceDatadogSyntheticsGlobalVariableRead,
		UpdateContext: resourceDatadogSyntheticsGlobalVariableUpdate,
		DeleteContext: resourceDatadogSyntheticsGlobalVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Synthetics global variable name.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
			},
			"description": {
				Description: "Description of the global variable.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tags": {
				Description: "A list of tags to associate with your synthetics global variable.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"value": {
				Description: "The value of the global variable.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"secure": {
				Description: "If set to true, the value of the global variable is hidden. Defaults to `false`.",
				Default:     false,
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"parse_test_id": {
				Description: "Id of the Synthetics test to use for a variable from test.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"parse_test_options": {
				Description: "ID of the Synthetics test to use a source of the global variable value.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Description: "Required when type = `http_header`. Defines the header to use to extract the value",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"type": {
							Description:      "Defines the source to use to extract the value.",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsGlobalVariableParseTestOptionsTypeFromValue),
						},
						"parser": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Description:      "Type of parser to extract the value.",
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsGlobalVariableParserTypeFromValue),
									},
									"value": {
										Description: "Value for the parser to use, required for type `json_path` or `regex`.",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
			"restricted_roles": {
				Description: "A list of role identifiers to associate with the Synthetics global variable.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
		},
	}
}

func resourceDatadogSyntheticsGlobalVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsGlobalVariable := buildSyntheticsGlobalVariableStruct(d)
	createdSyntheticsGlobalVariable, httpResponse, err := datadogClientV1.SyntheticsApi.CreateGlobalVariable(authV1, *syntheticsGlobalVariable)
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating synthetics global variable")
	}
	if err := utils.CheckForUnparsed(createdSyntheticsGlobalVariable); err != nil {
		return diag.FromErr(err)
	}

	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsGlobalVariable.GetId())

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsGlobalVariableRead(ctx, d, meta)
}

func resourceDatadogSyntheticsGlobalVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsGlobalVariable, httpresp, err := datadogClientV1.SyntheticsApi.GetGlobalVariable(authV1, d.Id())

	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics global variable")
	}
	if err := utils.CheckForUnparsed(syntheticsGlobalVariable); err != nil {
		return diag.FromErr(err)
	}

	return updateSyntheticsGlobalVariableLocalState(d, &syntheticsGlobalVariable)
}

func resourceDatadogSyntheticsGlobalVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsGlobalVariable := buildSyntheticsGlobalVariableStruct(d)
	if _, httpResponse, err := datadogClientV1.SyntheticsApi.EditGlobalVariable(authV1, d.Id(), *syntheticsGlobalVariable); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		utils.TranslateClientErrorDiag(err, httpResponse, "error updating synthetics global variable")
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsGlobalVariableRead(ctx, d, meta)
}

func resourceDatadogSyntheticsGlobalVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if httpResponse, err := datadogClientV1.SyntheticsApi.DeleteGlobalVariable(authV1, d.Id()); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting synthetics global variable")
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

func buildSyntheticsGlobalVariableStruct(d *schema.ResourceData) *datadogV1.SyntheticsGlobalVariable {
	syntheticsGlobalVariable := datadogV1.NewSyntheticsGlobalVariableWithDefaults()

	syntheticsGlobalVariable.SetName(d.Get("name").(string))

	if description, ok := d.GetOk("description"); ok {
		syntheticsGlobalVariable.SetDescription(description.(string))
	}

	tags := make([]string, 0)
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.([]interface{}) {
			tags = append(tags, s.(string))
		}
	}
	syntheticsGlobalVariable.SetTags(tags)

	syntheticsGlobalVariableValue := datadogV1.SyntheticsGlobalVariableValue{}

	syntheticsGlobalVariableValue.SetValue(d.Get("value").(string))
	syntheticsGlobalVariableValue.SetSecure(d.Get("secure").(bool))

	syntheticsGlobalVariable.SetValue(syntheticsGlobalVariableValue)

	if parseTestID, ok := d.GetOk("parse_test_id"); ok {
		if _, ok := d.GetOk("parse_test_options.0"); ok {
			syntheticsGlobalVariable.SetParseTestPublicId(parseTestID.(string))

			parseTestOptions := datadogV1.SyntheticsGlobalVariableParseTestOptions{}
			parseTestOptions.SetType(datadogV1.SyntheticsGlobalVariableParseTestOptionsType(d.Get("parse_test_options.0.type").(string)))

			if field, ok := d.GetOk("parse_test_options.0.field"); ok {
				parseTestOptions.SetField(field.(string))
			}

			parser := datadogV1.SyntheticsVariableParser{}
			parser.SetType(datadogV1.SyntheticsGlobalVariableParserType(d.Get("parse_test_options.0.parser.0.type").(string)))

			if value, ok := d.GetOk("parse_test_options.0.parser.0.value"); ok {
				parser.SetValue(value.(string))
			}

			parseTestOptions.SetParser(parser)

			syntheticsGlobalVariable.SetParseTestOptions(parseTestOptions)
		}
	}

	if restrictedRolesSet, ok := d.GetOk("restricted_roles"); ok {
		restrictedRoles := buildDatadogRestrictedRoles(restrictedRolesSet.(*schema.Set))
		attributes := datadogV1.SyntheticsGlobalVariableAttributes{
			RestrictedRoles: restrictedRoles,
		}
		syntheticsGlobalVariable.SetAttributes(attributes)
	}

	return syntheticsGlobalVariable
}

func updateSyntheticsGlobalVariableLocalState(d *schema.ResourceData, syntheticsGlobalVariable *datadogV1.SyntheticsGlobalVariable) diag.Diagnostics {
	d.Set("name", syntheticsGlobalVariable.GetName())
	d.Set("description", syntheticsGlobalVariable.GetDescription())

	syntheticsGlobalVariableValue := syntheticsGlobalVariable.GetValue()

	if syntheticsGlobalVariableValue.GetSecure() {
		// if the global variable is secure we need to get the value
		// from the config since it will not be returned by the api
		d.Set("value", d.Get("value").(string))
	} else {
		d.Set("value", syntheticsGlobalVariableValue.GetValue())
	}

	d.Set("secure", syntheticsGlobalVariableValue.GetSecure())

	d.Set("tags", syntheticsGlobalVariable.Tags)

	if syntheticsGlobalVariable.HasParseTestPublicId() {
		d.Set("parse_test_id", syntheticsGlobalVariable.GetParseTestPublicId())

		localParseTestOptions := make(map[string]interface{})
		localParser := make(map[string]string)

		parseTestOptions := syntheticsGlobalVariable.GetParseTestOptions()
		parser := parseTestOptions.GetParser()

		if v, ok := parser.GetTypeOk(); ok {
			localParser["type"] = string(*v)
		}

		localParser["value"] = parser.GetValue()

		localParseTestOptions["type"] = parseTestOptions.GetType()
		if v, ok := parseTestOptions.GetFieldOk(); ok {
			localParseTestOptions["field"] = string(*v)
		}
		localParseTestOptions["parser"] = []map[string]string{localParser}

		d.Set("parse_test_options", []map[string]interface{}{localParseTestOptions})
	}

	if syntheticsGlobalVariable.HasAttributes() {
		attributes := syntheticsGlobalVariable.GetAttributes()
		variableRestrictedRoles := attributes.GetRestrictedRoles()
		restrictedRoles := buildTerraformRestrictedRoles(&variableRestrictedRoles)
		d.Set("restricted_roles", restrictedRoles)
	}

	return nil
}
