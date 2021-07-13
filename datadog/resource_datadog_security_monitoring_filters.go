package datadog

import (
	"context"
	"errors"

	_ "gopkg.in/warnings.v0"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const securityFilterType = "security_filters"

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

		Schema: securityMonitoringFilterSchema(),
	}
}

func securityMonitoringFilterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		"exclusion_filter": {
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
						Type:        schema.TypeString,
						Required:    true,
						Description: "Exclusion filter query. Logs that match this query are excluded from the security filter.",
					},
				},
			},
		},

		"filtered_data_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The filtered data type.",
			Default:          "logs",
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityFilterFilteredDataTypeFromValue),
		},
	}
}

func resourceDatadogSecurityMonitoringFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	filterCreate := buildSecMonFilterCreatePayload(d)

	response, httpResponse, err := datadogClientV2.SecurityMonitoringApi.CreateSecurityFilter(authV2, *filterCreate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating security monitoring filter")
	}

	// store the resource id
	data := response.GetData()
	d.SetId(data.GetId())

	return nil
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
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	filterId := d.Id()

	response, diagnostics, failed := fetchFilterFromApi(datadogClientV2, authV2, filterId)
	if failed {
		return diagnostics
	}

	filterUpdate := buildSecMonFilterUpdatePayload(response, d)

	if _, httpResponse, err := datadogClientV2.SecurityMonitoringApi.UpdateSecurityFilter(authV2, filterId, *filterUpdate); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating security monitoring filter")
	}

	return nil
}

func fetchFilterFromApi(datadogClientV2 *datadogV2.APIClient, authV2 context.Context, filterId string) (datadogV2.SecurityFilterResponse, diag.Diagnostics, bool) {
	response, httpResponse, err := datadogClientV2.SecurityMonitoringApi.GetSecurityFilter(authV2, filterId)

	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return datadogV2.SecurityFilterResponse{}, diag.FromErr(errors.New("default rule does not exist")), true
		}

		return datadogV2.SecurityFilterResponse{}, utils.TranslateClientErrorDiag(err, httpResponse, "error fetching filter"), true
	}
	return response, nil, false
}

func resourceDatadogSecurityMonitoringFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	filterId := d.Id()

	_, diagnostics, failed := fetchFilterFromApi(datadogClientV2, authV2, filterId)
	if failed {
		return diagnostics
	}

	if httpResponse, err := datadogClientV2.SecurityMonitoringApi.DeleteSecurityFilter(authV2, filterId); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting security monitoring filter")
	}

	return nil
}

func updateResourceDataFilterFromResponse(d *schema.ResourceData, filterResponse datadogV2.SecurityFilterResponse) {
	data := filterResponse.GetData()
	attributes := data.GetAttributes()
	d.Set("name", attributes.GetName())
	d.Set("query", attributes.GetQuery())
	d.Set("is_enabled", attributes.GetIsEnabled())
	d.Set("filtered_data_type", attributes.GetFilteredDataType())

	if _, ok := attributes.GetExclusionFiltersOk(); ok {
		exclusionFiltersTF := extractExclusionFiltersTF(attributes)
		d.Set("exclusion_filter", exclusionFiltersTF)
	}
}

func extractExclusionFiltersTF(attributes datadogV2.SecurityFilterAttributes) []map[string]interface{} {
	exclusionFiltersTF := make([]map[string]interface{}, len(attributes.GetExclusionFilters()))
	for idx := range attributes.GetExclusionFilters() {
		exclusionFilterTF := make(map[string]interface{})
		exclusionFilter := attributes.GetExclusionFilters()[idx]
		exclusionFilterTF["name"] = exclusionFilter.GetName()
		exclusionFilterTF["query"] = exclusionFilter.GetQuery()
		exclusionFiltersTF[idx] = exclusionFilterTF
	}
	return exclusionFiltersTF
}

func buildSecMonFilterUpdatePayload(currentState datadogV2.SecurityFilterResponse, d *schema.ResourceData) *datadogV2.SecurityFilterUpdateRequest {
	payload := datadogV2.SecurityFilterUpdateRequest{}
	payload.Data.Type = securityFilterType
	// set the version from current state
	data := currentState.GetData()
	attributes := data.GetAttributes()
	version := attributes.GetVersion()
	payload.Data.Attributes.SetVersion(version)

	isEnabled, name, filteredDataType, query, filters := extractFilterAttributedFromResource(d)

	payload.Data.Attributes.SetIsEnabled(isEnabled)
	payload.Data.Attributes.SetName(name)
	payload.Data.Attributes.SetFilteredDataType(filteredDataType)
	payload.Data.Attributes.SetQuery(query)
	payload.Data.Attributes.SetExclusionFilters(filters)

	return &payload
}

func buildSecMonFilterCreatePayload(d *schema.ResourceData) *datadogV2.SecurityFilterCreateRequest {
	payload := datadogV2.SecurityFilterCreateRequest{}
	payload.Data.Type = securityFilterType

	isEnabled, name, filteredDataType, query, filters := extractFilterAttributedFromResource(d)

	payload.Data.Attributes.SetIsEnabled(isEnabled)
	payload.Data.Attributes.SetName(name)
	payload.Data.Attributes.SetFilteredDataType(filteredDataType)
	payload.Data.Attributes.SetQuery(query)
	payload.Data.Attributes.SetExclusionFilters(filters)

	return &payload
}

func extractFilterAttributedFromResource(d *schema.ResourceData) (bool, string, datadogV2.SecurityFilterFilteredDataType, string, []datadogV2.SecurityFilterExclusionFilter) {
	isEnabled := d.Get("is_enabled").(bool)
	name := d.Get("name").(string)
	filteredDataTypeString := d.Get("filtered_data_type").(string)
	filteredDataType := datadogV2.SecurityFilterFilteredDataType(filteredDataTypeString)
	query := d.Get("query").(string)

	var filters []datadogV2.SecurityFilterExclusionFilter
	if v, ok := d.GetOk("exclusion_filter"); ok {
		tfFilters := v.([]interface{})

		filters = make([]datadogV2.SecurityFilterExclusionFilter, len(tfFilters))
		for i, tfFiler := range tfFilters {
			filter := tfFiler.(map[string]interface{})
			filters[i].SetName(filter["name"].(string))
			filters[i].SetQuery(filter["query"].(string))
		}
	} else {
		filters = make([]datadogV2.SecurityFilterExclusionFilter, 0)
	}
	return isEnabled, name, filteredDataType, query, filters
}
