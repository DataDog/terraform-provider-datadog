package datadog

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogAwsLogsServices() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve all AWS logs services id's.",
		ReadContext: dataSourceDatadogAwsLogsServicesRead,
		Schema: map[string]*schema.Schema{
			"aws_logs_services_ids": {
				Description: "List of aws logs services id's",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDatadogAwsLogsServicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	awsLogsServices, httpresp, err := apiInstances.GetAWSLogsIntegrationApiV1().ListAWSLogsServices(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying aws logs services")
	}
	if err := utils.CheckForUnparsed(awsLogsServices); err != nil {
		return diag.FromErr(err)
	}

	tfLogsServiceIds := make([]string, 0)

	for _, service := range awsLogsServices {
		tfLogsServiceIds = append(tfLogsServiceIds, service.GetId())
	}

	sort.SliceStable(tfLogsServiceIds, func(i, j int) bool {
		return tfLogsServiceIds[i] < tfLogsServiceIds[j]
	})

	if err := d.Set("aws_logs_services_ids", tfLogsServiceIds); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("aws-logs-services")

	return nil
}
