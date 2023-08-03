package datadog

import (
	"context"
	"regexp"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceDatadogSyntheticsGlobalVariable() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve a Datadog Synthetics global variable (to be used in Synthetics tests).",
		ReadContext: dataSourceDatadogSyntheticsGlobalVariableRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Description:  "The synthetics global variable name to search for. Must only match one global variable.",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
				},
				"tags": {
					Description: "A list of tags assigned to the Synthetics global variable.",
					Type:        schema.TypeList,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
			}
		},
	}
}

func dataSourceDatadogSyntheticsGlobalVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	globalVariables, httpresp, err := apiInstances.GetSyntheticsApiV1().ListGlobalVariables(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics global variables")
	}

	searchedName := d.Get("name").(string)
	var matchedGlobalVariables []datadogV1.SyntheticsGlobalVariable
	for _, globalVariable := range globalVariables.Variables {
		if globalVariable.Name == searchedName {
			matchedGlobalVariables = append(matchedGlobalVariables, globalVariable)
		}
	}

	if len(matchedGlobalVariables) == 0 {
		return diag.Errorf("Couldn't find synthetics global variable named %s", searchedName)
	} else if len(matchedGlobalVariables) > 1 {
		return diag.Errorf("Found multiple synthetics global variables named %s", searchedName)
	}

	if err := utils.CheckForUnparsed(matchedGlobalVariables); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(matchedGlobalVariables[0].GetId())
	d.Set("tags", matchedGlobalVariables[0].GetTags())

	return nil
}
