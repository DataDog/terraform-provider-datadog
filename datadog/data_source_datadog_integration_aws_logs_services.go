package datadog

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogIntegrationAWSLogsServices() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve all AWS log ready services.",
		ReadContext: dataSourceDatadogIntegrationAWSLogsServicesRead,
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				// Computed
				"aws_logs_services": {
					Description: "List of AWS log ready services.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The id of the AWS log service.",
							},
							"label": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The name of the AWS log service.",
							},
						},
					},
				},
			}
		},
	}
}

func dataSourceDatadogIntegrationAWSLogsServicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	awsLogsServices, httpresp, err := apiInstances.GetAWSLogsIntegrationApiV1().ListAWSLogsServices(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying AWS logs services")
	}

	diags := diag.Diagnostics{}
	tfLogsServices := make([]map[string]interface{}, 0)
	for _, awsLogsService := range awsLogsServices {
		if err := utils.CheckForUnparsed(awsLogsService); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping integration aws logs services with id: %s", awsLogsService.GetId()),
				Detail:   fmt.Sprintf("aws logs service contains unparsed object: %v", err),
			})
			continue
		}

		// extract agent rule
		awsLogsServiceTF := make(map[string]interface{})
		if awsLogsServiceId, ok := awsLogsService.GetIdOk(); ok {
			awsLogsServiceTF["id"] = *awsLogsServiceId
		} else {
			continue
		}
		if awsLogsServiceLabel, ok := awsLogsService.GetLabelOk(); ok {
			awsLogsServiceTF["label"] = *awsLogsServiceLabel
		}
		tfLogsServices = append(tfLogsServices, awsLogsServiceTF)
	}

	sort.SliceStable(tfLogsServices, func(i, j int) bool {
		return tfLogsServices[i]["id"].(string) < tfLogsServices[j]["id"].(string)
	})

	if err := d.Set("aws_logs_services", tfLogsServices); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("aws-logs-services")

	return diags
}
