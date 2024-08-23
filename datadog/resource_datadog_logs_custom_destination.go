package datadog

import (
	"context"
	"fmt"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var logCustomDestinationMutex = sync.Mutex{}

var customDestinationSchema = map[string]*schema.Schema{
	"name": {
		Description: "The custom destination name.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"query": {
		Description: "The custom destination query filter. Logs matching this query are forwarded to the destination.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"enabled": {
		Description: "Whether logs matching this custom destination should be forwarded or not.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"forward_tags": {
		Description: "Whether tags from the forwarded logs should be forwarded or not.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"forward_tags_restriction_list": {
		Description: `List of [keys of tags](https://docs.datadoghq.com/getting_started/tagging/#define-tags) to be filtered.
		An empty list represents no restriction is in place and either all or no tags will be
		forwarded depending on ` + "`forward_tags_restriction_list_type`" + ` parameter.`,
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 10,
		Elem:     &schema.Schema{Type: schema.TypeString},
	},
	"forward_tags_restriction_list_type": {
		Description: `How ` + "`forward_tags_restriction_list`" + ` parameter should be interpreted.
        If ` + "`ALLOW_LIST`" + `, then only tags whose keys on the forwarded logs match the ones on the restriction list
        are forwarded.
        ` + "`BLOCK_LIST`" + ` works the opposite way. It does not forward the tags matching the ones on the list.`,
		Type:             schema.TypeString,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewCustomDestinationAttributeTagsRestrictionListTypeFromValue),
	},

	"http_destination":          httpDestination,
	"splunk_destination":        splunkDestination,
	"elasticsearch_destination": elasticsearchDestination,
}

var httpDestination = &schema.Schema{
	Type:          schema.TypeList,
	MaxItems:      1,
	Description:   "The HTTP destination.",
	Optional:      true,
	ConflictsWith: []string{"splunk_destination", "elasticsearch_destination"},
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Description: "The destination for which logs will be forwarded to. Must have HTTPS scheme and forwarding back to Datadog is not allowed.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"basic_auth":         httpBasicAuth,
			"custom_header_auth": httpCustomHeaderAuth,
		},
	},
}

var httpBasicAuth = &schema.Schema{
	Type:          schema.TypeList,
	MaxItems:      1,
	Description:   "Basic access authentication.",
	Optional:      true,
	ConflictsWith: []string{"http_destination.0.custom_header_auth"},
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"username": {
				Description: "The username of the authentication. This field is not returned by the API.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"password": {
				Description: "The password of the authentication. This field is not returned by the API.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
	},
}

var httpCustomHeaderAuth = &schema.Schema{
	Type:          schema.TypeList,
	MaxItems:      1,
	Description:   "Custom header access authentication.",
	Optional:      true,
	ConflictsWith: []string{"http_destination.0.basic_auth"},
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"header_name": {
				Description: "The header name of the authentication.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"header_value": {
				Description: "The header value of the authentication. This field is not returned by the API.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
	},
}

var splunkDestination = &schema.Schema{
	Type:          schema.TypeList,
	MaxItems:      1,
	Description:   "The Splunk HTTP Event Collector (HEC) destination.",
	Optional:      true,
	ConflictsWith: []string{"http_destination", "elasticsearch_destination"},
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Description: "The destination for which logs will be forwarded to. Must have HTTPS scheme and forwarding back to Datadog is not allowed.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"access_token": {
				Description: "Access token of the Splunk HTTP Event Collector. This field is not returned by the API.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
	},
}

var elasticsearchDestination = &schema.Schema{
	Type:          schema.TypeList,
	MaxItems:      1,
	Description:   "The Elasticsearch destination.",
	Optional:      true,
	ConflictsWith: []string{"http_destination", "splunk_destination"},
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Description: "The destination for which logs will be forwarded to. Must have HTTPS scheme and forwarding back to Datadog is not allowed.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"index_name": {
				Description: "Name of the Elasticsearch index (must follow [Elasticsearch's criteria](https://www.elastic.co/guide/en/elasticsearch/reference/8.11/indices-create-index.html#indices-create-api-path-params)).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"index_rotation": {
				Description: `Date pattern with US locale and UTC timezone to be appended to the index name after adding '-'
				(that is, '${index_name}-${indexPattern}').
				You can customize the index rotation naming pattern by choosing one of these options:
				- Hourly: 'yyyy-MM-dd-HH' (as an example, it would render: '2022-10-19-09')
				- Daily: 'yyyy-MM-dd' (as an example, it would render: '2022-10-19')
				- Weekly: 'yyyy-'W'ww' (as an example, it would render: '2022-W42')
				- Monthly: 'yyyy-MM' (as an example, it would render: '2022-10')

				If this field is missing or is blank, it means that the index name will always be the same
				(that is, no rotation).`,
				Type:     schema.TypeString,
				Optional: true,
			},

			"basic_auth": elasticsearchBasicAuth,
		},
	},
}

var elasticsearchBasicAuth = &schema.Schema{
	Type:        schema.TypeList,
	MinItems:    1,
	MaxItems:    1,
	Description: "Basic access authentication.",
	Required:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"username": {
				Description: "The username of the authentication. This field is not returned by the API.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"password": {
				Description: "The password of the authentication. This field is not returned by the API.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
	},
}

func resourceDatadogLogsCustomDestination() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Logs Custom Destination API resource, which is used to create and manage Datadog log forwarding.",
		CreateContext: resourceDatadogLogsCustomDestinationCreate,
		ReadContext:   resourceDatadogLogsCustomDestinationRead,
		UpdateContext: resourceDatadogLogsCustomDestinationUpdate,
		DeleteContext: resourceDatadogLogsCustomDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return customDestinationSchema
		},
	}
}

func resourceDatadogLogsCustomDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	logCustomDestinationMutex.Lock()
	defer logCustomDestinationMutex.Unlock()

	ddDestination, err := buildCustomDestinationCreateRequest(d)
	if err != nil {
		return diag.FromErr(err)
	}

	createdDestination, httpResponse, err := apiInstances.GetLogsCustomDestinationsApiV2().CreateLogsCustomDestination(auth, *ddDestination)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "failed to create logs custom destination using Datadog API")
	}
	if err := utils.CheckForUnparsed(createdDestination); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*createdDestination.GetData().Id)
	return updateLogsCustomDestinationState(d, createdDestination.GetData().Attributes)
}

func resourceDatadogLogsCustomDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ddDestination, httpResponse, err := apiInstances.GetLogsCustomDestinationsApiV2().GetLogsCustomDestination(auth, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return utils.TranslateClientErrorDiag(err, httpResponse, "failed to get logs custom destination using Datadog API")
	}
	if err := utils.CheckForUnparsed(ddDestination); err != nil {
		return diag.FromErr(err)
	}

	return updateLogsCustomDestinationState(d, ddDestination.GetData().Attributes)
}

func resourceDatadogLogsCustomDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	logCustomDestinationMutex.Lock()
	defer logCustomDestinationMutex.Unlock()

	ddDestination, err := buildCustomDestinationUpdateRequest(d)
	if err != nil {
		return diag.FromErr(err)
	}

	updatedDestination, httpResponse, err := apiInstances.GetLogsCustomDestinationsApiV2().UpdateLogsCustomDestination(auth, d.Id(), *ddDestination)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "failed to update logs custom destination using Datadog API")
	}
	if err := utils.CheckForUnparsed(updatedDestination); err != nil {
		return diag.FromErr(err)
	}

	return updateLogsCustomDestinationState(d, updatedDestination.GetData().Attributes)
}

func resourceDatadogLogsCustomDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	logCustomDestinationMutex.Lock()
	defer logCustomDestinationMutex.Unlock()

	httpResponse, err := apiInstances.GetLogsCustomDestinationsApiV2().DeleteLogsCustomDestination(auth, d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "failed to delete logs custom destination using Datadog APIe")
	}

	return nil
}

func buildCustomDestinationCreateRequest(d *schema.ResourceData) (*datadogV2.CustomDestinationCreateRequest, error) {
	destination := datadogV2.NewCustomDestinationCreateRequest()
	destination.Data = datadogV2.NewCustomDestinationCreateRequestDefinitionWithDefaults()

	forwarderDestination, err := buildForwarderDestination(d)
	if err != nil {
		return nil, err
	}

	attributes := datadogV2.NewCustomDestinationCreateRequestAttributesWithDefaults()
	attributes.SetName(d.Get("name").(string))
	attributes.SetQuery(d.Get("query").(string))
	attributes.SetEnabled(d.Get("enabled").(bool))
	attributes.SetForwarderDestination(*forwarderDestination)
	attributes.SetForwardTags(d.Get("forward_tags").(bool))

	if rawList, ok := d.GetOk("forward_tags_restriction_list"); ok {
		var list []string
		for _, item := range rawList.([]interface{}) {
			list = append(list, item.(string))
		}
		attributes.SetForwardTagsRestrictionList(list)
	}

	if rawListType, ok := d.GetOk("forward_tags_restriction_list_type"); ok {
		listType, err := datadogV2.NewCustomDestinationAttributeTagsRestrictionListTypeFromValue(rawListType.(string))
		if err != nil {
			return nil, err
		}
		attributes.SetForwardTagsRestrictionListType(*listType)
	}

	destination.Data.SetAttributes(*attributes)
	return destination, nil
}

func buildCustomDestinationUpdateRequest(d *schema.ResourceData) (*datadogV2.CustomDestinationUpdateRequest, error) {
	destination := datadogV2.NewCustomDestinationUpdateRequest()
	destination.Data = datadogV2.NewCustomDestinationUpdateRequestDefinitionWithDefaults()
	destination.Data.SetId(d.Id())

	forwarderDestination, err := buildForwarderDestination(d)
	if err != nil {
		return destination, err
	}

	attributes := datadogV2.NewCustomDestinationUpdateRequestAttributesWithDefaults()
	attributes.SetName(d.Get("name").(string))
	attributes.SetQuery(d.Get("query").(string))
	attributes.SetEnabled(d.Get("enabled").(bool))
	attributes.SetForwarderDestination(*forwarderDestination)
	attributes.SetForwardTags(d.Get("forward_tags").(bool))

	if rawList, ok := d.GetOk("forward_tags_restriction_list"); ok {
		var list []string
		for _, item := range rawList.([]interface{}) {
			list = append(list, item.(string))
		}
		attributes.SetForwardTagsRestrictionList(list)
	}

	if rawListType, ok := d.GetOk("forward_tags_restriction_list_type"); ok {
		listType, err := datadogV2.NewCustomDestinationAttributeTagsRestrictionListTypeFromValue(rawListType.(string))
		if err != nil {
			return nil, err
		}
		attributes.SetForwardTagsRestrictionListType(*listType)
	}

	destination.Data.SetAttributes(*attributes)
	return destination, nil
}

func buildForwarderDestination(d *schema.ResourceData) (*datadogV2.CustomDestinationForwardDestination, error) {
	forwarderBuilderMap := map[string]func(map[string]interface{}) (*datadogV2.CustomDestinationForwardDestination, error){
		"http_destination":          buildHttpDestination,
		"splunk_destination":        buildSplunkDestination,
		"elasticsearch_destination": buildElasticsearchDestination,
	}

	for field, fn := range forwarderBuilderMap {
		if entry, ok := d.GetOk(field); ok {
			list := entry.([]interface{})
			if len(list) == 1 {
				return fn(list[0].(map[string]interface{}))
			}
		}
	}

	return nil, fmt.Errorf("no valid forwarder destination was found")
}

func buildHttpDestination(d map[string]interface{}) (*datadogV2.CustomDestinationForwardDestination, error) {
	auth, err := buildHttpDestinationAuth(d)
	if err != nil {
		return nil, err
	}

	http := datadogV2.NewCustomDestinationForwardDestinationHttpWithDefaults()
	http.SetAuth(*auth)
	http.SetEndpoint(d["endpoint"].(string))

	destination := datadogV2.CustomDestinationForwardDestinationHttpAsCustomDestinationForwardDestination(http)
	return &destination, nil
}

func buildHttpDestinationAuth(d map[string]interface{}) (*datadogV2.CustomDestinationHttpDestinationAuth, error) {
	if entries, ok := d["basic_auth"].([]interface{}); ok && len(entries) == 1 {
		entry := entries[0].(map[string]interface{})
		basic := datadogV2.NewCustomDestinationHttpDestinationAuthBasicWithDefaults()
		basic.SetUsername(entry["username"].(string))
		basic.SetPassword(entry["password"].(string))

		auth := datadogV2.CustomDestinationHttpDestinationAuthBasicAsCustomDestinationHttpDestinationAuth(basic)
		return &auth, nil
	}

	if entries, ok := d["custom_header_auth"].([]interface{}); ok && len(entries) == 1 {
		entry := entries[0].(map[string]interface{})
		basic := datadogV2.NewCustomDestinationHttpDestinationAuthCustomHeaderWithDefaults()
		basic.SetHeaderName(entry["header_name"].(string))
		basic.SetHeaderValue(entry["header_value"].(string))

		auth := datadogV2.CustomDestinationHttpDestinationAuthCustomHeaderAsCustomDestinationHttpDestinationAuth(basic)
		return &auth, nil
	}

	return nil, fmt.Errorf("no valid http destination authentication method was found")
}

func buildSplunkDestination(d map[string]interface{}) (*datadogV2.CustomDestinationForwardDestination, error) {
	splunk := datadogV2.NewCustomDestinationForwardDestinationSplunkWithDefaults()
	splunk.SetEndpoint(d["endpoint"].(string))
	splunk.SetAccessToken(d["access_token"].(string))

	destination := datadogV2.CustomDestinationForwardDestinationSplunkAsCustomDestinationForwardDestination(splunk)
	return &destination, nil
}

func buildElasticsearchDestination(d map[string]interface{}) (*datadogV2.CustomDestinationForwardDestination, error) {
	auth, err := buildElasticsearchDestinationAuth(d)
	if err != nil {
		return nil, err
	}

	elasticsearch := datadogV2.NewCustomDestinationForwardDestinationElasticsearchWithDefaults()
	elasticsearch.SetAuth(*auth)
	elasticsearch.SetEndpoint(d["endpoint"].(string))
	elasticsearch.SetIndexName(d["index_name"].(string))
	elasticsearch.SetIndexRotation(d["index_rotation"].(string))

	destination := datadogV2.CustomDestinationForwardDestinationElasticsearchAsCustomDestinationForwardDestination(elasticsearch)
	return &destination, nil
}

func buildElasticsearchDestinationAuth(d map[string]interface{}) (*datadogV2.CustomDestinationElasticsearchDestinationAuth, error) {
	if entries, ok := d["basic_auth"].([]interface{}); ok && len(entries) == 1 {
		entry := entries[0].(map[string]interface{})
		auth := datadogV2.NewCustomDestinationElasticsearchDestinationAuthWithDefaults()
		auth.SetUsername(entry["username"].(string))
		auth.SetPassword(entry["password"].(string))

		return auth, nil
	}

	return nil, fmt.Errorf("no valid elasticsearch destination authentication method was found")
}

func updateLogsCustomDestinationState(d *schema.ResourceData, destinationAttrs *datadogV2.CustomDestinationResponseAttributes) diag.Diagnostics {
	if err := d.Set("name", destinationAttrs.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("query", destinationAttrs.GetQuery()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("enabled", destinationAttrs.GetEnabled()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("forward_tags", destinationAttrs.GetForwardTags()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("forward_tags_restriction_list", destinationAttrs.GetForwardTagsRestrictionList()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("forward_tags_restriction_list_type", destinationAttrs.GetForwardTagsRestrictionListType()); err != nil {
		return diag.FromErr(err)
	}

	httpDestination := destinationAttrs.GetForwarderDestination().CustomDestinationResponseForwardDestinationHttp
	if httpDestination != nil {
		destination := make(map[string]interface{})
		destination["endpoint"] = httpDestination.GetEndpoint()

		basicAuth := httpDestination.GetAuth().CustomDestinationResponseHttpDestinationAuthBasic
		if basicAuth != nil {
			// NOTE: Basic auth values are not returned by the API.
			destination["basic_auth"] = []map[string]interface{}{{
				"username": d.Get("http_destination.0.basic_auth.0.username").(string),
				"password": d.Get("http_destination.0.basic_auth.0.password").(string),
			}}
		}

		customHeaderAuth := httpDestination.GetAuth().CustomDestinationResponseHttpDestinationAuthCustomHeader
		if customHeaderAuth != nil {
			destination["custom_header_auth"] = []map[string]interface{}{{
				"header_name": customHeaderAuth.GetHeaderName(),
				// NOTE: Header value are not returned by the API.
				"header_value": d.Get("http_destination.0.custom_header_auth.0.header_value").(string),
			}}
		}

		if err := d.Set("http_destination", []map[string]interface{}{destination}); err != nil {
			return diag.FromErr(err)
		}
	}

	splunkDestination := destinationAttrs.GetForwarderDestination().CustomDestinationResponseForwardDestinationSplunk
	if splunkDestination != nil {
		destination := make(map[string]interface{})
		destination["endpoint"] = splunkDestination.GetEndpoint()

		// NOTE: Access token is not returned by the API.
		destination["access_token"] = d.Get("splunk_destination.0.access_token").(string)

		if err := d.Set("splunk_destination", []map[string]interface{}{destination}); err != nil {
			return diag.FromErr(err)
		}
	}

	elasticsearchDestination := destinationAttrs.GetForwarderDestination().CustomDestinationResponseForwardDestinationElasticsearch
	if elasticsearchDestination != nil {
		destination := make(map[string]interface{})
		destination["endpoint"] = elasticsearchDestination.GetEndpoint()
		destination["index_name"] = elasticsearchDestination.GetIndexName()
		destination["index_rotation"] = elasticsearchDestination.GetIndexRotation()

		// NOTE: Basic auth values are not returned by the API.
		destination["basic_auth"] = []map[string]interface{}{{
			"username": d.Get("elasticsearch_destination.0.basic_auth.0.username").(string),
			"password": d.Get("elasticsearch_destination.0.basic_auth.0.password").(string),
		}}

		if err := d.Set("elasticsearch_destination", []map[string]interface{}{destination}); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
