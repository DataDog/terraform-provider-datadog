package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure        = &logsCustomDestinationResource{}
	_ resource.ResourceWithImportState      = &logsCustomDestinationResource{}
	_ resource.ResourceWithConfigValidators = &logsCustomDestinationResource{}
)

type logsCustomDestinationResource struct {
	Api  *datadogV2.LogsCustomDestinationsApi
	Auth context.Context
}

type logsCustomDestinationModel struct {
	ID                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	Query                          types.String `tfsdk:"query"`
	Enabled                        types.Bool   `tfsdk:"enabled"`
	ForwardTags                    types.Bool   `tfsdk:"forward_tags"`
	ForwardTagsRestrictionList     types.List   `tfsdk:"forward_tags_restriction_list"`
	ForwardTagsRestrictionListType types.String `tfsdk:"forward_tags_restriction_list_type"`

	HttpDestination          []HttpDestination          `tfsdk:"http_destination"`
	SplunkDestination        []SplunkDestination        `tfsdk:"splunk_destination"`
	ElasticsearchDestination []ElasticsearchDestination `tfsdk:"elasticsearch_destination"`
}

type HttpDestination struct {
	Endpoint types.String `tfsdk:"endpoint"`

	BasicAuth        []HttpDestinationBasicAuth        `tfsdk:"basic_auth"`
	CustomHeaderAuth []HttpDestinationCustomHeaderAuth `tfsdk:"custom_header_auth"`
}

type HttpDestinationBasicAuth struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type HttpDestinationCustomHeaderAuth struct {
	HeaderName  types.String `tfsdk:"header_name"`
	HeaderValue types.String `tfsdk:"header_value"`
}

type SplunkDestination struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	AccessToken types.String `tfsdk:"access_token"`
}

type ElasticsearchDestination struct {
	Endpoint      types.String `tfsdk:"endpoint"`
	IndexName     types.String `tfsdk:"index_name"`
	IndexRotation types.String `tfsdk:"index_rotation"`

	BasicAuth []ElasticsearchDestinationBasicAuth `tfsdk:"basic_auth"`
}

type ElasticsearchDestinationBasicAuth struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func NewLogsCustomDestinationResource() resource.Resource {
	return &logsCustomDestinationResource{}
}

func (r *logsCustomDestinationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetLogsCustomDestinationsApiV2()
	r.Auth = providerData.Auth
}

func (r *logsCustomDestinationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "logs_custom_destination"
}

func (d *logsCustomDestinationResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		// Require providing exactly one valid destination with auth.
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("http_destination").AtListIndex(0).AtName("basic_auth"),
			path.MatchRoot("http_destination").AtListIndex(0).AtName("custom_header_auth"),
			path.MatchRoot("splunk_destination"),
			path.MatchRoot("elasticsearch_destination").AtListIndex(0).AtName("basic_auth"),
		),
	}
}

func (r *logsCustomDestinationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Logs Custom Destination API resource, which is used to create and manage Datadog log forwarding.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The custom destination name.",
				Required:    true,
			},
			"query": schema.StringAttribute{
				Description: "The custom destination query filter. Logs matching this query are forwarded to the destination.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether logs matching this custom destination should be forwarded or not.",
				Optional:    true,
				Computed:    true,
			},
			"forward_tags": schema.BoolAttribute{
				Description: "Whether tags from the forwarded logs should be forwarded or not.",
				Optional:    true,
				Computed:    true,
			},
			"forward_tags_restriction_list": schema.ListAttribute{
				Description: `List of [tag keys](https://docs.datadoghq.com/getting_started/tagging/#define-tags) to be filtered.
				An empty list represents no restriction is in place and either all or no tags will be
				forwarded depending on ` + "`forward_tags_restriction_list_type`" + ` parameter.`,
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
			},
			"forward_tags_restriction_list_type": schema.StringAttribute{
				Description: `How the ` + "`forward_tags_restriction_list`" + ` parameter should be interpreted.
				If ` + "`ALLOW_LIST`" + `, then only tags whose keys on the forwarded logs match the ones on the restriction list
				are forwarded.
				` + "`BLOCK_LIST`" + ` works the opposite way. It does not forward the tags matching the ones on the list.`,
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewCustomDestinationAttributeTagsRestrictionListTypeFromValue),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"http_destination": schema.ListNestedBlock{
				Description: "The HTTP destination.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"endpoint": schema.StringAttribute{
							Description: "The destination for which logs will be forwarded to. Must have HTTPS scheme. Forwarding back to Datadog is not allowed.",
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"basic_auth": schema.ListNestedBlock{
							Description: "Basic access authentication.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"username": schema.StringAttribute{
										Description: "The username of the authentication. This field is not returned by the API.",
										Required:    true,
										Sensitive:   true,
									},
									"password": schema.StringAttribute{
										Description: "The password of the authentication. This field is not returned by the API.",
										Required:    true,
										Sensitive:   true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},

						"custom_header_auth": schema.ListNestedBlock{
							Description: "Custom header access authentication.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"header_name": schema.StringAttribute{
										Description: "The header name of the authentication.",
										Required:    true,
									},
									"header_value": schema.StringAttribute{
										Description: "The header value of the authentication. This field is not returned by the API.",
										Required:    true,
										Sensitive:   true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},

			"splunk_destination": schema.ListNestedBlock{
				Description: "The Splunk HTTP Event Collector (HEC) destination.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"endpoint": schema.StringAttribute{
							Description: "The destination for which logs will be forwarded to. Must have HTTPS scheme. Forwarding back to Datadog is not allowed.",
							Required:    true,
						},
						"access_token": schema.StringAttribute{
							Description: "Access token of the Splunk HTTP Event Collector. This field is not returned by the API.",
							Required:    true,
							Sensitive:   true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},

			"elasticsearch_destination": schema.ListNestedBlock{
				Description: "The Elasticsearch destination.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"endpoint": schema.StringAttribute{
							Description: "The destination for which logs will be forwarded to. Must have HTTPS scheme. Forwarding back to Datadog is not allowed.",
							Required:    true,
						},
						"index_name": schema.StringAttribute{
							Description: "Name of the Elasticsearch index (must follow [Elasticsearch's criteria](https://www.elastic.co/guide/en/elasticsearch/reference/8.11/indices-create-index.html#indices-create-api-path-params)).",
							Required:    true,
						},
						"index_rotation": schema.StringAttribute{
							Description: `Date pattern with US locale and UTC timezone to be appended to the index name after adding '-'
							(that is, '${index_name}-${indexPattern}').
							You can customize the index rotation naming pattern by choosing one of these options:
							- Hourly: 'yyyy-MM-dd-HH' (as an example, it would render: '2022-10-19-09')
							- Daily: 'yyyy-MM-dd' (as an example, it would render: '2022-10-19')
							- Weekly: 'yyyy-'W'ww' (as an example, it would render: '2022-W42')
							- Monthly: 'yyyy-MM' (as an example, it would render: '2022-10')
							If this field is missing or is blank, it means that the index name will always be the same
							(that is, no rotation).`,
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"basic_auth": schema.ListNestedBlock{
							Description: "Basic access authentication.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"username": schema.StringAttribute{
										Description: "The username of the authentication. This field is not returned by the API.",
										Required:    true,
										Sensitive:   true,
									},
									"password": schema.StringAttribute{
										Description: "The password of the authentication. This field is not returned by the API.",
										Required:    true,
										Sensitive:   true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func (r *logsCustomDestinationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *logsCustomDestinationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state logsCustomDestinationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetLogsCustomDestination(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving logs custom destination"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *logsCustomDestinationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state logsCustomDestinationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildLogsCustomDestinationCreateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateLogsCustomDestination(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating logs custom destination"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *logsCustomDestinationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state logsCustomDestinationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildLogsCustomDestinationUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateLogsCustomDestination(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating logs custom destination"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *logsCustomDestinationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state logsCustomDestinationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteLogsCustomDestination(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting custom destination"))
		return
	}
}

func (r *logsCustomDestinationResource) updateState(ctx context.Context, state *logsCustomDestinationModel, resp *datadogV2.CustomDestinationResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if query, ok := attributes.GetQueryOk(); ok {
		state.Query = types.StringValue(*query)
	}

	if enabled, ok := attributes.GetEnabledOk(); ok {
		state.Enabled = types.BoolValue(*enabled)
	}

	if forwardTags, ok := attributes.GetForwardTagsOk(); ok {
		state.ForwardTags = types.BoolValue(*forwardTags)
	}

	if forwardTagsRestrictionList, ok := attributes.GetForwardTagsRestrictionListOk(); ok {
		tfForwardTagsRestrictionList, _ := types.ListValueFrom(ctx, types.StringType, forwardTagsRestrictionList)
		state.ForwardTagsRestrictionList = tfForwardTagsRestrictionList
	}

	if forwardTagsRestrictionListType, ok := attributes.GetForwardTagsRestrictionListTypeOk(); ok {
		state.ForwardTagsRestrictionListType = types.StringValue(string(*forwardTagsRestrictionListType))
	}

	forwarderDestination, ok := attributes.GetForwarderDestinationOk()
	if !ok {
		return
	}

	if httpDestination := forwarderDestination.CustomDestinationResponseForwardDestinationHttp; httpDestination != nil {
		if len(state.HttpDestination) != 1 {
			state.HttpDestination = []HttpDestination{{}}
		}

		if endpoint, ok := httpDestination.GetEndpointOk(); ok {
			state.HttpDestination[0].Endpoint = types.StringValue(*endpoint)
		}

		// NOTE: Basic auth values are not returned by the API, keep user state.

		if customHeaderAuth := httpDestination.GetAuth().CustomDestinationResponseHttpDestinationAuthCustomHeader; customHeaderAuth != nil {
			if headerName, ok := customHeaderAuth.GetHeaderNameOk(); ok {
				state.HttpDestination[0].CustomHeaderAuth[0].HeaderName = types.StringValue(*headerName)
			}

			// NOTE: Header value are not returned by the API, keep user state.
		}
	}

	if splunkDestination := forwarderDestination.CustomDestinationResponseForwardDestinationSplunk; splunkDestination != nil {
		if len(state.SplunkDestination) != 1 {
			state.SplunkDestination = []SplunkDestination{{}}
		}

		if endpoint, ok := splunkDestination.GetEndpointOk(); ok {
			state.SplunkDestination[0].Endpoint = types.StringValue(*endpoint)
		}

		// NOTE: Access token is not returned by the API, keep user state.
	}

	if elasticsearchDestination := forwarderDestination.CustomDestinationResponseForwardDestinationElasticsearch; elasticsearchDestination != nil {
		if len(state.ElasticsearchDestination) != 1 {
			state.ElasticsearchDestination = []ElasticsearchDestination{{}}
		}

		if endpoint, ok := elasticsearchDestination.GetEndpointOk(); ok {
			state.ElasticsearchDestination[0].Endpoint = types.StringValue(*endpoint)
		}

		if indexName, ok := elasticsearchDestination.GetIndexNameOk(); ok {
			state.ElasticsearchDestination[0].IndexName = types.StringValue(*indexName)
		}

		if indexRotation, ok := elasticsearchDestination.GetIndexRotationOk(); ok {
			state.ElasticsearchDestination[0].IndexRotation = types.StringValue(*indexRotation)
		}

		// NOTE: Basic auth values are not returned by the API, keep user state.
	}
}

func (r *logsCustomDestinationResource) buildLogsCustomDestinationCreateRequestBody(ctx context.Context, state *logsCustomDestinationModel) (*datadogV2.CustomDestinationCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := datadogV2.NewCustomDestinationCreateRequestWithDefaults()
	req.Data = datadogV2.NewCustomDestinationCreateRequestDefinitionWithDefaults()

	attributes := datadogV2.NewCustomDestinationCreateRequestAttributesWithDefaults()

	if forwarderDestination := r.buildLogsCustomDestinationForwarderDestination(state); forwarderDestination != nil {
		attributes.SetForwarderDestination(*forwarderDestination)
	}

	attributes.SetName(state.Name.ValueString())

	if !state.Query.IsUnknown() {
		attributes.SetQuery(state.Query.ValueString())
	}

	if !state.Enabled.IsUnknown() {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}

	if !state.ForwardTags.IsUnknown() {
		attributes.SetForwardTags(state.ForwardTags.ValueBool())
	}

	if !state.ForwardTagsRestrictionList.IsUnknown() {
		var forwardTagsRestrictionList []string
		diags.Append(state.ForwardTagsRestrictionList.ElementsAs(ctx, &forwardTagsRestrictionList, false)...)
		attributes.SetForwardTagsRestrictionList(forwardTagsRestrictionList)
	}

	if !state.ForwardTagsRestrictionListType.IsUnknown() {
		forwardTagsRestrictionListType, err := datadogV2.NewCustomDestinationAttributeTagsRestrictionListTypeFromValue(state.ForwardTagsRestrictionListType.ValueString())
		if err == nil {
			attributes.SetForwardTagsRestrictionListType(*forwardTagsRestrictionListType)
		}
	}

	req.Data.SetAttributes(*attributes)
	return req, diags
}

func (r *logsCustomDestinationResource) buildLogsCustomDestinationUpdateRequestBody(ctx context.Context, state *logsCustomDestinationModel) (*datadogV2.CustomDestinationUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := datadogV2.NewCustomDestinationUpdateRequestWithDefaults()
	req.Data = datadogV2.NewCustomDestinationUpdateRequestDefinitionWithDefaults()
	req.Data.SetId(state.ID.ValueString())

	attributes := datadogV2.NewCustomDestinationUpdateRequestAttributesWithDefaults()

	if forwarderDestination := r.buildLogsCustomDestinationForwarderDestination(state); forwarderDestination != nil {
		attributes.SetForwarderDestination(*forwarderDestination)
	}

	attributes.SetName(state.Name.ValueString())

	if !state.Query.IsUnknown() {
		attributes.SetQuery(state.Query.ValueString())
	}

	if !state.Enabled.IsUnknown() {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}

	if !state.ForwardTags.IsUnknown() {
		attributes.SetForwardTags(state.ForwardTags.ValueBool())
	}

	if !state.ForwardTagsRestrictionList.IsUnknown() {
		var forwardTagsRestrictionList []string
		diags.Append(state.ForwardTagsRestrictionList.ElementsAs(ctx, &forwardTagsRestrictionList, false)...)
		attributes.SetForwardTagsRestrictionList(forwardTagsRestrictionList)
	}

	if !state.ForwardTagsRestrictionListType.IsUnknown() {
		forwardTagsRestrictionListType, err := datadogV2.NewCustomDestinationAttributeTagsRestrictionListTypeFromValue(state.ForwardTagsRestrictionListType.ValueString())
		if err == nil {
			attributes.SetForwardTagsRestrictionListType(*forwardTagsRestrictionListType)
		}
	}

	req.Data.SetAttributes(*attributes)
	return req, diags
}

func (r *logsCustomDestinationResource) buildLogsCustomDestinationForwarderDestination(state *logsCustomDestinationModel) *datadogV2.CustomDestinationForwardDestination {
	if httpDestination := state.HttpDestination; len(httpDestination) == 1 {
		http := datadogV2.NewCustomDestinationForwardDestinationHttpWithDefaults()
		http.SetEndpoint(httpDestination[0].Endpoint.ValueString())

		if basicAuth := httpDestination[0].BasicAuth; len(basicAuth) == 1 {
			auth := datadogV2.NewCustomDestinationHttpDestinationAuthBasicWithDefaults()
			auth.SetUsername(basicAuth[0].Username.ValueString())
			auth.SetPassword(basicAuth[0].Password.ValueString())
			http.SetAuth(datadogV2.CustomDestinationHttpDestinationAuthBasicAsCustomDestinationHttpDestinationAuth(auth))
		}

		if customHeaderAuth := httpDestination[0].CustomHeaderAuth; len(customHeaderAuth) == 1 {
			auth := datadogV2.NewCustomDestinationHttpDestinationAuthCustomHeaderWithDefaults()
			auth.SetHeaderName(customHeaderAuth[0].HeaderName.ValueString())
			auth.SetHeaderValue(customHeaderAuth[0].HeaderValue.ValueString())
			http.SetAuth(datadogV2.CustomDestinationHttpDestinationAuthCustomHeaderAsCustomDestinationHttpDestinationAuth(auth))
		}

		httpOut := datadogV2.CustomDestinationForwardDestinationHttpAsCustomDestinationForwardDestination(http)
		return &httpOut
	}

	if splunkDestination := state.SplunkDestination; len(splunkDestination) == 1 {
		splunk := datadogV2.NewCustomDestinationForwardDestinationSplunkWithDefaults()
		splunk.SetEndpoint(splunkDestination[0].Endpoint.ValueString())
		splunk.SetAccessToken(splunkDestination[0].AccessToken.ValueString())

		splunkOut := datadogV2.CustomDestinationForwardDestinationSplunkAsCustomDestinationForwardDestination(splunk)
		return &splunkOut
	}

	if elasticsearchDestination := state.ElasticsearchDestination; len(elasticsearchDestination) == 1 {
		elasticsearch := datadogV2.NewCustomDestinationForwardDestinationElasticsearchWithDefaults()
		elasticsearch.SetEndpoint(elasticsearchDestination[0].Endpoint.ValueString())
		elasticsearch.SetIndexName(elasticsearchDestination[0].IndexName.ValueString())

		if !elasticsearchDestination[0].IndexRotation.IsNull() {
			elasticsearch.SetIndexRotation(elasticsearchDestination[0].IndexRotation.ValueString())
		}

		if basicAuth := elasticsearchDestination[0].BasicAuth; len(basicAuth) == 1 {
			auth := datadogV2.NewCustomDestinationElasticsearchDestinationAuthWithDefaults()
			auth.SetUsername(basicAuth[0].Username.ValueString())
			auth.SetPassword(basicAuth[0].Password.ValueString())

			elasticsearch.SetAuth(*auth)
		}

		elasticsearchOut := datadogV2.CustomDestinationForwardDestinationElasticsearchAsCustomDestinationForwardDestination(elasticsearch)
		return &elasticsearchOut
	}

	return nil
}
