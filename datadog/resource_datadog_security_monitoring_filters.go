package datadog

import (
	"context"
	"errors"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
	_ "gopkg.in/warnings.v0"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogSecurityMonitoringFilter() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Security Monitoring Rule API resource for security filters.",
		CreateContext: resourceDatadogSecurityMonitoringFilterCreate,
		ReadContext:   resourceDatadogSecurityMonitoringFilterRead,
		UpdateContext: resourceDatadogSecurityMonitoringFilterUpdate,
		DeleteContext: resourceDatadogSecurityMonitoringFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the security filter.",
			},
			"query": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The query of the security filter.",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the security filter is enabled.",
			},
			"exclusion_filters": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Exclusion filters to exclude some logs from the security filter.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Exclusion filter name.",
						},
						"query": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "Exclusion filter query. Logs that match this query are excluded from the security filter.",
						},
					},
				},
			},

			"filtered_data_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The filtered data type.",
				Default: "logs",
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityFilterFilteredDataTypeFromValue),
			},
		},
	}
}

func resourceDatadogSecurityMonitoringFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(errors.New("error"))
}

func resourceDatadogSecurityMonitoringFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	id := d.Id()
	filterResponse, httpResponse, err := datadogClientV2.SecurityMonitoringApi.GetSecurityFilter(authV2, id)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	updateResourceDataFilterFromResponse(d, filterResponse)

	// handle warning
	if filterResponse.HasMeta() {
		filterMeta := filterResponse.GetMeta()
		warning := filterMeta.GetWarning()
		diagnostic := diag.Diagnostic{Severity: diag.Warning, Summary: warning}
		return diag.Diagnostics{diagnostic}
	}
	return nil
}

func resourceDatadogSecurityMonitoringFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(errors.New("error"))
}

func resourceDatadogSecurityMonitoringFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(errors.New("error"))
}


func updateResourceDataFilterFromResponse(d *schema.ResourceData, filterResponse datadogV2.SecurityFilterResponse) {
	data := filterResponse.GetData()
	attributes := data.GetAttributes()
	d.Set("name", attributes.GetName())
	d.Set("query", attributes.GetQuery())
	d.Set("is_enabled", attributes.GetIsEnabled())
	d.Set("filtered_data_type", attributes.GetFilteredDataType())

	if exclusionFilters, ok := attributes.GetExclusionFiltersOk(); ok {
		exclusionFiltersTF := make([]map[string]interface{}, len(*exclusionFilters))
		for idx := range attributes.GetExclusionFilters() {
			exclusionFilterTF := make(map[string]interface{})
			exclusionFilter := attributes.GetExclusionFilters()[idx]
			exclusionFilterTF["name"] = exclusionFilter.GetName()
			exclusionFilterTF["query"] = exclusionFilter.GetQuery()
			exclusionFiltersTF[idx] = exclusionFilterTF
		}
		d.Set("exclusion_filters", exclusionFiltersTF)
	}
}
