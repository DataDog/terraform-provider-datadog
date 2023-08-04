package datadog

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogServiceLevelObjectives() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about multiple SLOs for use in other resources.",
		ReadContext: dataSourceDatadogServiceLevelObjectivesRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"ids": {
					Description: "An array of SLO IDs to limit the search.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"name_query": {
					Description: "Filter results based on SLO names.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"tags_query": {
					Description: "Filter results based on a single SLO tag.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"metrics_query": {
					Description: "Filter results based on SLO numerator and denominator.",
					Type:        schema.TypeString,
					Optional:    true,
				},

				// Computed values
				"slos": {
					Description: "List of SLOs",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Description: "ID of the Datadog service level objective",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"name": {
								Description: "Name of the Datadog service level objective",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"type": {
								Description: "The type of the service level objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/service-level-objectives/#create-a-slo-object). Available options to choose from are: `metric` and `monitor`.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
			}
		},
	}
}

func dataSourceDatadogServiceLevelObjectivesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	var idsPtr *string
	var nameQueryPtr *string
	var tagsQueryPtr *string
	var metricsQueryPtr *string

	reqParams := datadogV1.NewListSLOsOptionalParameters()
	if v, ok := d.GetOk("ids"); ok {
		ids := strings.Join(expandStringList(v.([]interface{})), ",")
		idsPtr = &ids
		reqParams.WithIds(ids)
	}
	if v, ok := d.GetOk("name_query"); ok {
		nameQuery := v.(string)
		nameQueryPtr = &nameQuery
		reqParams.WithQuery(nameQuery)
	}
	if v, ok := d.GetOk("tags_query"); ok {
		tagsQuery := v.(string)
		tagsQueryPtr = &tagsQuery
		reqParams.WithTagsQuery(tagsQuery)
	}
	if v, ok := d.GetOk("metrics_query"); ok {
		metricsQuery := v.(string)
		metricsQueryPtr = &metricsQuery
		reqParams.WithMetricsQuery(metricsQuery)
	}

	slosResp, httpresp, err := apiInstances.GetServiceLevelObjectivesApiV1().ListSLOs(auth, *reqParams)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying service level objectives")
	}
	if len(slosResp.GetData()) == 0 {
		return diag.Errorf("your query returned no result, please try a less specific search criteria")
	}

	diags := diag.Diagnostics{}
	slos := make([]map[string]interface{}, 0, len(slosResp.GetData()))
	for _, slo := range slosResp.GetData() {
		if err := utils.CheckForUnparsed(slo); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping service level objective with id: %s", slo.GetId()),
				Detail:   fmt.Sprintf("service level objective contains unparsed object: %v", err),
			})
			continue
		}

		slos = append(slos, map[string]interface{}{
			"id":   slo.GetId(),
			"name": slo.GetName(),
			"type": slo.GetType(),
		})
	}
	if err := d.Set("slos", slos); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(computeSLOsDataSourceID(idsPtr, nameQueryPtr, tagsQueryPtr, metricsQueryPtr))

	return diags
}

func computeSLOsDataSourceID(ids *string, nameQuery *string, tagsQuery *string, metricsQuery *string) string {
	// Key for hashing
	var b strings.Builder
	if ids != nil {
		b.WriteString(*ids)
	}
	b.WriteRune('|')
	if nameQuery != nil {
		b.WriteString(*nameQuery)
	}
	b.WriteRune('|')
	if tagsQuery != nil {
		b.WriteString(*tagsQuery)
	}
	b.WriteRune('|')
	if metricsQuery != nil {
		b.WriteString(*metricsQuery)
	}

	keyStr := b.String()
	h := sha256.New()
	log.Println("HASHKEY", keyStr)
	h.Write([]byte(keyStr))

	return fmt.Sprintf("%x", h.Sum(nil))
}
