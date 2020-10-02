package datadog

import (
	"regexp"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceDatadogSyntheticsGlobalVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogSyntheticsGlobalVariableCreate,
		Read:   resourceDatadogSyntheticsGlobalVariableRead,
		Update: resourceDatadogSyntheticsGlobalVariableUpdate,
		Delete: resourceDatadogSyntheticsGlobalVariableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDatadogSyntheticsGlobalVariableCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsGlobalVariable := buildSyntheticsGlobalVariableStruct(d)
	createdSyntheticsGlobalVariable, _, err := datadogClientV1.SyntheticsApi.CreateGlobalVariable(authV1).Body(*syntheticsGlobalVariable).Execute()
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return translateClientError(err, "error creating synthetics global variable")
	}

	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsGlobalVariable.GetId())

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsGlobalVariableRead(d, meta)
}

func resourceDatadogSyntheticsGlobalVariableRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsGlobalVariable, httpresp, err := datadogClientV1.SyntheticsApi.GetGlobalVariable(authV1, d.Id()).Execute()

	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error getting synthetics global variable")
	}

	return updateSyntheticsGlobalVariableLocalState(d, &syntheticsGlobalVariable)
}

func resourceDatadogSyntheticsGlobalVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsGlobalVariable := buildSyntheticsGlobalVariableStruct(d)
	if _, _, err := datadogClientV1.SyntheticsApi.EditGlobalVariable(authV1, d.Id()).Body(*syntheticsGlobalVariable).Execute(); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		translateClientError(err, "error updating synthetics global variable")
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsGlobalVariableRead(d, meta)
}

func resourceDatadogSyntheticsGlobalVariableDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, err := datadogClientV1.SyntheticsApi.DeleteGlobalVariable(authV1, d.Id()).Execute(); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return translateClientError(err, "error deleting synthetics global variable")
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
	syntheticsGlobalVariableValue.SetSecure(false)

	syntheticsGlobalVariable.SetValue(syntheticsGlobalVariableValue)

	return syntheticsGlobalVariable
}

func updateSyntheticsGlobalVariableLocalState(d *schema.ResourceData, syntheticsGlobalVariable *datadogV1.SyntheticsGlobalVariable) error {
	d.Set("name", syntheticsGlobalVariable.GetName())
	d.Set("description", syntheticsGlobalVariable.GetDescription())

	syntheticsGlobalVariableValue := syntheticsGlobalVariable.GetValue()
	d.Set("value", syntheticsGlobalVariableValue.GetValue())

	d.Set("tags", syntheticsGlobalVariable.Tags)

	return nil
}
