package fwprovider

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider/observability_pipeline"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &observabilityPipelineResource{}
	_ resource.ResourceWithImportState = &observabilityPipelineResource{}
)

type observabilityPipelineResource struct {
	Api  *datadogV2.ObservabilityPipelinesApi
	Auth context.Context
}

// Note on nested block design:
// SingleNestedBlocks are not allowed in this resource schema. Instead, we use ListNestedBlock
// with size validation: listvalidator.SizeAtMost(1) and listvalidator.IsRequired()(for required blocks).
// We do this to make the TF schema more robust, future-proof and
// eliminate potential breaking changes related to required/optional blocks and fields.
// See hashicorp/terraform-provider-aws#35813 as an example of the same approach.

type observabilityPipelineModel struct {
	ID     types.String  `tfsdk:"id"`
	Name   types.String  `tfsdk:"name"`
	Config []configModel `tfsdk:"config"`
}

type configModel struct {
	PipelineType          types.String           `tfsdk:"pipeline_type"`
	UseLegacySearchSyntax types.Bool             `tfsdk:"use_legacy_search_syntax"`
	Sources               []*sourceModel         `tfsdk:"source"`
	ProcessorGroups       []*processorGroupModel `tfsdk:"processor_group"`
	Destinations          []*destinationModel    `tfsdk:"destination"`
}

type destinationModel struct {
	Id     types.String `tfsdk:"id"`
	Inputs types.List   `tfsdk:"inputs"`

	DatadogLogsDestination            []*datadogLogsDestinationModel                                   `tfsdk:"datadog_logs"`
	GoogleCloudStorageDestination     []*gcsDestinationModel                                           `tfsdk:"google_cloud_storage"`
	GooglePubSubDestination           []*googlePubSubDestinationModel                                  `tfsdk:"google_pubsub"`
	SplunkHecDestination              []*splunkHecDestinationModel                                     `tfsdk:"splunk_hec"`
	SumoLogicDestination              []*sumoLogicDestinationModel                                     `tfsdk:"sumo_logic"`
	RsyslogDestination                []*rsyslogDestinationModel                                       `tfsdk:"rsyslog"`
	SyslogNgDestination               []*syslogNgDestinationModel                                      `tfsdk:"syslog_ng"`
	ElasticsearchDestination          []*elasticsearchDestinationModel                                 `tfsdk:"elasticsearch"`
	AzureStorageDestination           []*azureStorageDestinationModel                                  `tfsdk:"azure_storage"`
	MicrosoftSentinelDestination      []*microsoftSentinelDestinationModel                             `tfsdk:"microsoft_sentinel"`
	GoogleSecopsDestination           []*googleSecopsDestinationModel                                  `tfsdk:"google_secops"`
	NewRelicDestination               []*newRelicDestinationModel                                      `tfsdk:"new_relic"`
	SentinelOneDestination            []*sentinelOneDestinationModel                                   `tfsdk:"sentinel_one"`
	OpenSearchDestination             []*opensearchDestinationModel                                    `tfsdk:"opensearch"`
	AmazonOpenSearchDestination       []*amazonOpenSearchDestinationModel                              `tfsdk:"amazon_opensearch"`
	SocketDestination                 []*observability_pipeline.SocketDestinationModel                 `tfsdk:"socket"`
	AmazonS3Destination               []*observability_pipeline.AmazonS3DestinationModel               `tfsdk:"amazon_s3"`
	AmazonSecurityLakeDestination     []*observability_pipeline.AmazonSecurityLakeDestinationModel     `tfsdk:"amazon_security_lake"`
	CrowdStrikeNextGenSiemDestination []*observability_pipeline.CrowdStrikeNextGenSiemDestinationModel `tfsdk:"crowdstrike_next_gen_siem"`
	DatadogMetricsDestination         []*datadogMetricsDestinationModel                                `tfsdk:"datadog_metrics"`
	HttpClientDestination             []*httpClientDestinationModel                                    `tfsdk:"http_client"`
	CloudPremDestination              []*observability_pipeline.CloudPremDestinationModel              `tfsdk:"cloud_prem"`
	KafkaDestination                  []*observability_pipeline.KafkaDestinationModel                  `tfsdk:"kafka"`
}

type datadogMetricsDestinationModel struct {
	// No additional fields needed - only id and inputs (defined in destinationModel)
}

type httpClientDestinationModel struct {
	Encoding     types.String                            `tfsdk:"encoding"`
	Compression  []httpClientDestinationCompressionModel `tfsdk:"compression"`
	AuthStrategy types.String                            `tfsdk:"auth_strategy"`
	Tls          []observability_pipeline.TlsModel       `tfsdk:"tls"`
}

type httpClientDestinationCompressionModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
}

type sourceModel struct {
	Id                       types.String                                       `tfsdk:"id"`
	DatadogAgentSource       []*datadogAgentSourceModel                         `tfsdk:"datadog_agent"`
	KafkaSource              []*kafkaSourceModel                                `tfsdk:"kafka"`
	RsyslogSource            []*rsyslogSourceModel                              `tfsdk:"rsyslog"`
	SyslogNgSource           []*syslogNgSourceModel                             `tfsdk:"syslog_ng"`
	SumoLogicSource          []*sumoLogicSourceModel                            `tfsdk:"sumo_logic"`
	FluentdSource            []*fluentdSourceModel                              `tfsdk:"fluentd"`
	FluentBitSource          []*fluentBitSourceModel                            `tfsdk:"fluent_bit"`
	HttpServerSource         []*httpServerSourceModel                           `tfsdk:"http_server"`
	AmazonS3Source           []*amazonS3SourceModel                             `tfsdk:"amazon_s3"`
	SplunkHecSource          []*splunkHecSourceModel                            `tfsdk:"splunk_hec"`
	SplunkTcpSource          []*splunkTcpSourceModel                            `tfsdk:"splunk_tcp"`
	AmazonDataFirehoseSource []*amazonDataFirehoseSourceModel                   `tfsdk:"amazon_data_firehose"`
	HttpClientSource         []*httpClientSourceModel                           `tfsdk:"http_client"`
	GooglePubSubSource       []*googlePubSubSourceModel                         `tfsdk:"google_pubsub"`
	LogstashSource           []*logstashSourceModel                             `tfsdk:"logstash"`
	SocketSource             []*observability_pipeline.SocketSourceModel        `tfsdk:"socket"`
	OpentelemetrySource      []*observability_pipeline.OpentelemetrySourceModel `tfsdk:"opentelemetry"`
}

type logstashSourceModel struct {
	Tls []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type datadogAgentSourceModel struct {
	Tls []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type kafkaSourceModel struct {
	GroupId           types.String                      `tfsdk:"group_id"`
	Topics            []types.String                    `tfsdk:"topics"`
	LibrdkafkaOptions []librdkafkaOptionModel           `tfsdk:"librdkafka_option"`
	Sasl              []kafkaSourceSaslModel            `tfsdk:"sasl"`
	Tls               []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type librdkafkaOptionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type kafkaSourceSaslModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
}

type amazonS3SourceModel struct {
	Region types.String                          `tfsdk:"region"` // AWS region where the S3 bucket resides
	Auth   []observability_pipeline.AwsAuthModel `tfsdk:"auth"`   // AWS authentication credentials
	Tls    []observability_pipeline.TlsModel     `tfsdk:"tls"`    // TLS encryption configuration
}

type processorGroupModel struct {
	Id          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	DisplayName types.String `tfsdk:"display_name"`
	Include     types.String `tfsdk:"include"`
	Inputs      types.List   `tfsdk:"inputs"`

	Processors []*processorModel `tfsdk:"processor"`
}

type processorModel struct {
	Id          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Include     types.String `tfsdk:"include"`
	DisplayName types.String `tfsdk:"display_name"`

	FilterProcessor               []*filterProcessorModel                             `tfsdk:"filter"`
	ParseJsonProcessor            []*parseJsonProcessorModel                          `tfsdk:"parse_json"`
	AddFieldsProcessor            []*addFieldsProcessor                               `tfsdk:"add_fields"`
	RenameFieldsProcessor         []*renameFieldsProcessorModel                       `tfsdk:"rename_fields"`
	RemoveFieldsProcessor         []*removeFieldsProcessorModel                       `tfsdk:"remove_fields"`
	QuotaProcessor                []*quotaProcessorModel                              `tfsdk:"quota"`
	GenerateMetricsProcessor      []*generateMetricsProcessorModel                    `tfsdk:"generate_datadog_metrics"`
	ParseGrokProcessor            []*parseGrokProcessorModel                          `tfsdk:"parse_grok"`
	SampleProcessor               []*sampleProcessorModel                             `tfsdk:"sample"`
	SensitiveDataScannerProcessor []*sensitiveDataScannerProcessorModel               `tfsdk:"sensitive_data_scanner"`
	DedupeProcessor               []*dedupeProcessorModel                             `tfsdk:"dedupe"`
	ReduceProcessor               []*reduceProcessorModel                             `tfsdk:"reduce"`
	ThrottleProcessor             []*throttleProcessorModel                           `tfsdk:"throttle"`
	AddEnvVarsProcessor           []*addEnvVarsProcessorModel                         `tfsdk:"add_env_vars"`
	EnrichmentTableProcessor      []*enrichmentTableProcessorModel                    `tfsdk:"enrichment_table"`
	OcsfMapperProcessor           []*ocsfMapperProcessorModel                         `tfsdk:"ocsf_mapper"`
	DatadogTagsProcessor          []*observability_pipeline.DatadogTagsProcessorModel `tfsdk:"datadog_tags"`
	CustomProcessor               []*observability_pipeline.CustomProcessorModel      `tfsdk:"custom_processor"`
	AddHostnameProcessor          []*addHostnameProcessorModel                        `tfsdk:"add_hostname"`
	ParseXMLProcessor             []*parseXMLProcessorModel                           `tfsdk:"parse_xml"`
	SplitArrayProcessor           []*splitArrayProcessorModel                         `tfsdk:"split_array"`
	MetricTagsProcessor           []*metricTagsProcessorModel                         `tfsdk:"metric_tags"`
}

type metricTagsProcessorModel struct {
	Rules []metricTagsProcessorRuleModel `tfsdk:"rule"`
}

type metricTagsProcessorRuleModel struct {
	Include types.String   `tfsdk:"include"`
	Mode    types.String   `tfsdk:"mode"`
	Action  types.String   `tfsdk:"action"`
	Keys    []types.String `tfsdk:"keys"`
}

type ocsfMapperProcessorModel struct {
	Mapping []ocsfMappingModel `tfsdk:"mapping"`
}

type ocsfMappingModel struct {
	Include        types.String `tfsdk:"include"`
	LibraryMapping types.String `tfsdk:"library_mapping"`
}

type enrichmentTableProcessorModel struct {
	Target         types.String                    `tfsdk:"target"`
	File           []enrichmentFileModel           `tfsdk:"file"`
	GeoIp          []enrichmentGeoIpModel          `tfsdk:"geoip"`
	ReferenceTable []enrichmentReferenceTableModel `tfsdk:"reference_table"`
}

type enrichmentFileModel struct {
	Path     types.String        `tfsdk:"path"`
	Encoding []fileEncodingModel `tfsdk:"encoding"`
	Key      []fileKeyItemModel  `tfsdk:"key"`
}

type fileEncodingModel struct {
	Type            types.String `tfsdk:"type"`
	Delimiter       types.String `tfsdk:"delimiter"`
	IncludesHeaders types.Bool   `tfsdk:"includes_headers"`
}

type fileKeyItemModel struct {
	Column     types.String `tfsdk:"column"`
	Comparison types.String `tfsdk:"comparison"`
	Field      types.String `tfsdk:"field"`
}

type enrichmentGeoIpModel struct {
	KeyField types.String `tfsdk:"key_field"`
	Locale   types.String `tfsdk:"locale"`
	Path     types.String `tfsdk:"path"`
}

type enrichmentReferenceTableModel struct {
	KeyField types.String `tfsdk:"key_field"`
	TableId  types.String `tfsdk:"table_id"`
	Columns  types.List   `tfsdk:"columns"`
}

type addEnvVarsProcessorModel struct {
	Variables []envVarMappingModel `tfsdk:"variable"`
}

type envVarMappingModel struct {
	Field types.String `tfsdk:"field"`
	Name  types.String `tfsdk:"name"`
}

type throttleProcessorModel struct {
	Threshold types.Int64    `tfsdk:"threshold"`
	Window    types.Float64  `tfsdk:"window"`
	GroupBy   []types.String `tfsdk:"group_by"`
}

type reduceProcessorModel struct {
	GroupBy         []types.String       `tfsdk:"group_by"`
	MergeStrategies []mergeStrategyModel `tfsdk:"merge_strategy"`
}

type mergeStrategyModel struct {
	Path     types.String `tfsdk:"path"`
	Strategy types.String `tfsdk:"strategy"`
}

type dedupeProcessorModel struct {
	Fields []types.String `tfsdk:"fields"`
	Mode   types.String   `tfsdk:"mode"`
}

type filterProcessorModel struct {
}

type parseJsonProcessorModel struct {
	Field types.String `tfsdk:"field"`
}

type addFieldsProcessor struct {
	Fields []fieldValue `tfsdk:"field"`
}

type renameFieldsProcessorModel struct {
	Fields []renameFieldItemModel `tfsdk:"field"`
}

type renameFieldItemModel struct {
	Source         types.String `tfsdk:"source"`
	Destination    types.String `tfsdk:"destination"`
	PreserveSource types.Bool   `tfsdk:"preserve_source"`
}

type removeFieldsProcessorModel struct {
	Fields types.List `tfsdk:"fields"`
}

type quotaProcessorModel struct {
	Name                        types.String         `tfsdk:"name"`
	DropEvents                  types.Bool           `tfsdk:"drop_events"`
	Limit                       []quotaLimitModel    `tfsdk:"limit"`
	PartitionFields             []types.String       `tfsdk:"partition_fields"`
	IgnoreWhenMissingPartitions types.Bool           `tfsdk:"ignore_when_missing_partitions"`
	Overrides                   []quotaOverrideModel `tfsdk:"override"`
	OverflowAction              types.String         `tfsdk:"overflow_action"`
	TooManyBucketsAction        types.String         `tfsdk:"too_many_buckets_action"`
}

type quotaLimitModel struct {
	Enforce types.String `tfsdk:"enforce"` // "bytes" or "events"
	Limit   types.Int64  `tfsdk:"limit"`
}

type quotaOverrideModel struct {
	Fields []fieldValue      `tfsdk:"field"`
	Limit  []quotaLimitModel `tfsdk:"limit"`
}

type fieldValue struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type amazonOpenSearchDestinationModel struct {
	BulkIndex types.String                `tfsdk:"bulk_index"`
	Auth      []amazonOpenSearchAuthModel `tfsdk:"auth"`
}

type amazonOpenSearchAuthModel struct {
	Strategy    types.String `tfsdk:"strategy"`
	AwsRegion   types.String `tfsdk:"aws_region"`
	AssumeRole  types.String `tfsdk:"assume_role"`
	ExternalId  types.String `tfsdk:"external_id"`
	SessionName types.String `tfsdk:"session_name"`
}

type opensearchDestinationModel struct {
	BulkIndex  types.String                           `tfsdk:"bulk_index"`
	DataStream []opensearchDestinationDataStreamModel `tfsdk:"data_stream"`
}

type opensearchDestinationDataStreamModel struct {
	Dtype     types.String `tfsdk:"dtype"`
	Dataset   types.String `tfsdk:"dataset"`
	Namespace types.String `tfsdk:"namespace"`
}

type sentinelOneDestinationModel struct {
	Region types.String `tfsdk:"region"`
}

type newRelicDestinationModel struct {
	Region types.String `tfsdk:"region"`
}

type googleSecopsDestinationModel struct {
	Auth       []gcpAuthModel `tfsdk:"auth"`
	CustomerId types.String   `tfsdk:"customer_id"`
	Encoding   types.String   `tfsdk:"encoding"`
	LogType    types.String   `tfsdk:"log_type"`
}

type googlePubSubDestinationModel struct {
	Project  types.String                      `tfsdk:"project"`
	Topic    types.String                      `tfsdk:"topic"`
	Auth     []gcpAuthModel                    `tfsdk:"auth"`
	Encoding types.String                      `tfsdk:"encoding"`
	Tls      []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type datadogLogsDestinationModel struct {
	Routes []datadogLogsDestinationRouteModel `tfsdk:"routes"`
}

type datadogLogsDestinationRouteModel struct {
	RouteId   types.String `tfsdk:"route_id"`
	Include   types.String `tfsdk:"include"`
	Site      types.String `tfsdk:"site"`
	ApiKeyKey types.String `tfsdk:"api_key_key"`
}

type parseGrokProcessorModel struct {
	DisableLibraryRules types.Bool                    `tfsdk:"disable_library_rules"`
	Rules               []parseGrokProcessorRuleModel `tfsdk:"rule"`
}

type parseGrokProcessorRuleModel struct {
	Source       types.String    `tfsdk:"source"`
	MatchRules   []grokRuleModel `tfsdk:"match_rule"`
	SupportRules []grokRuleModel `tfsdk:"support_rule"`
}

type grokRuleModel struct {
	Name types.String `tfsdk:"name"`
	Rule types.String `tfsdk:"rule"`
}

type sampleProcessorModel struct {
	Percentage types.Float64  `tfsdk:"percentage"`
	GroupBy    []types.String `tfsdk:"group_by"`
}

type fluentdSourceModel struct {
	Tls []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type fluentBitSourceModel struct {
	Tls []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type httpServerSourceModel struct {
	AuthStrategy types.String                      `tfsdk:"auth_strategy"`
	Decoding     types.String                      `tfsdk:"decoding"`
	Tls          []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type splunkHecSourceModel struct {
	Tls []observability_pipeline.TlsModel `tfsdk:"tls"` // TLS encryption settings for secure ingestion.
}

type generateMetricsProcessorModel struct {
	Metrics []generatedMetricModel `tfsdk:"metric"`
}

type generatedMetricModel struct {
	Name       types.String           `tfsdk:"name"`
	Include    types.String           `tfsdk:"include"`
	MetricType types.String           `tfsdk:"metric_type"`
	GroupBy    types.List             `tfsdk:"group_by"`
	Value      []generatedMetricValue `tfsdk:"value"`
}

type generatedMetricValue struct {
	Strategy types.String `tfsdk:"strategy"`
	Field    types.String `tfsdk:"field"`
}

type splunkTcpSourceModel struct {
	Tls []observability_pipeline.TlsModel `tfsdk:"tls"` // TLS encryption settings for secure transmission.
}

type splunkHecDestinationModel struct {
	AutoExtractTimestamp types.Bool   `tfsdk:"auto_extract_timestamp"`
	Encoding             types.String `tfsdk:"encoding"`
	Sourcetype           types.String `tfsdk:"sourcetype"`
	Index                types.String `tfsdk:"index"`
}

type gcsDestinationModel struct {
	Bucket       types.String    `tfsdk:"bucket"`
	KeyPrefix    types.String    `tfsdk:"key_prefix"`
	StorageClass types.String    `tfsdk:"storage_class"`
	Acl          types.String    `tfsdk:"acl"`
	Auth         []gcpAuthModel  `tfsdk:"auth"`
	Metadata     []metadataEntry `tfsdk:"metadata"`
}

type metadataEntry struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type sumoLogicDestinationModel struct {
	Encoding             types.String             `tfsdk:"encoding"`
	HeaderHostName       types.String             `tfsdk:"header_host_name"`
	HeaderSourceName     types.String             `tfsdk:"header_source_name"`
	HeaderSourceCategory types.String             `tfsdk:"header_source_category"`
	HeaderCustomFields   []headerCustomFieldModel `tfsdk:"header_custom_field"`
}

type headerCustomFieldModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type rsyslogSourceModel struct {
	Mode types.String                      `tfsdk:"mode"`
	Tls  []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type syslogNgSourceModel struct {
	Mode types.String                      `tfsdk:"mode"`
	Tls  []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type rsyslogDestinationModel struct {
	Keepalive types.Int64                       `tfsdk:"keepalive"`
	Tls       []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type syslogNgDestinationModel struct {
	Keepalive types.Int64                       `tfsdk:"keepalive"`
	Tls       []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type elasticsearchDestinationModel struct {
	ApiVersion types.String                              `tfsdk:"api_version"`
	BulkIndex  types.String                              `tfsdk:"bulk_index"`
	DataStream []elasticsearchDestinationDataStreamModel `tfsdk:"data_stream"`
}

type elasticsearchDestinationDataStreamModel struct {
	Dtype     types.String `tfsdk:"dtype"`
	Dataset   types.String `tfsdk:"dataset"`
	Namespace types.String `tfsdk:"namespace"`
}

type azureStorageDestinationModel struct {
	ContainerName types.String `tfsdk:"container_name"`
	BlobPrefix    types.String `tfsdk:"blob_prefix"`
}

type microsoftSentinelDestinationModel struct {
	ClientId       types.String `tfsdk:"client_id"`
	TenantId       types.String `tfsdk:"tenant_id"`
	DcrImmutableId types.String `tfsdk:"dcr_immutable_id"`
	Table          types.String `tfsdk:"table"`
}

type sensitiveDataScannerProcessorModel struct {
	Rules []sensitiveDataScannerProcessorRule `tfsdk:"rule"`
}

type sensitiveDataScannerProcessorRule struct {
	Name           types.String                                  `tfsdk:"name"`
	Tags           []types.String                                `tfsdk:"tags"`
	KeywordOptions []sensitiveDataScannerProcessorKeywordOptions `tfsdk:"keyword_options"` // it's not a list in the API thus the plural name
	Pattern        []sensitiveDataScannerProcessorPattern        `tfsdk:"pattern"`
	Scope          []sensitiveDataScannerProcessorScope          `tfsdk:"scope"`
	OnMatch        []sensitiveDataScannerProcessorAction         `tfsdk:"on_match"`
}

// Nested structs (extracted per your preference)
type sensitiveDataScannerProcessorKeywordOptions struct {
	Keywords  []types.String `tfsdk:"keywords"`
	Proximity types.Int64    `tfsdk:"proximity"`
}

type sensitiveDataScannerProcessorPattern struct {
	Custom  []sensitiveDataScannerCustomPattern  `tfsdk:"custom"`
	Library []sensitiveDataScannerLibraryPattern `tfsdk:"library"`
}

type sensitiveDataScannerCustomPattern struct {
	Rule        types.String `tfsdk:"rule"`
	Description types.String `tfsdk:"description"`
}

type sensitiveDataScannerLibraryPattern struct {
	Id                     types.String `tfsdk:"id"`
	UseRecommendedKeywords types.Bool   `tfsdk:"use_recommended_keywords"`
	Description            types.String `tfsdk:"description"`
}

type sensitiveDataScannerProcessorScope struct {
	Include []sensitiveDataScannerScopeOptions `tfsdk:"include"`
	Exclude []sensitiveDataScannerScopeOptions `tfsdk:"exclude"`
	All     *bool                              `tfsdk:"all"`
}

type sensitiveDataScannerScopeOptions struct {
	Fields []types.String `tfsdk:"fields"`
}

type sensitiveDataScannerProcessorAction struct {
	Redact        []sensitiveDataScannerRedactAction        `tfsdk:"redact"`
	Hash          []sensitiveDataScannerHashAction          `tfsdk:"hash"`
	PartialRedact []sensitiveDataScannerPartialRedactAction `tfsdk:"partial_redact"`
}

type sensitiveDataScannerRedactAction struct {
	Replace types.String `tfsdk:"replace"`
}

type sensitiveDataScannerHashAction struct {
	// no fields; schema allows empty options
}

type sensitiveDataScannerPartialRedactAction struct {
	Characters types.Int64  `tfsdk:"characters"`
	Direction  types.String `tfsdk:"direction"` // "first" | "last"
}

type sumoLogicSourceModel struct {
}

type addHostnameProcessorModel struct {
	// No additional fields beyond common processor fields
}

type parseXMLProcessorModel struct {
	Field            types.String `tfsdk:"field"`
	IncludeAttr      types.Bool   `tfsdk:"include_attr"`
	AlwaysUseTextKey types.Bool   `tfsdk:"always_use_text_key"`
	ParseNumber      types.Bool   `tfsdk:"parse_number"`
	ParseBool        types.Bool   `tfsdk:"parse_bool"`
	ParseNull        types.Bool   `tfsdk:"parse_null"`
	AttrPrefix       types.String `tfsdk:"attr_prefix"`
	TextKey          types.String `tfsdk:"text_key"`
}

type splitArrayProcessorModel struct {
	Arrays []splitArrayConfigModel `tfsdk:"array"`
}

type splitArrayConfigModel struct {
	Include types.String `tfsdk:"include"`
	Field   types.String `tfsdk:"field"`
}

type amazonDataFirehoseSourceModel struct {
	Auth []observability_pipeline.AwsAuthModel `tfsdk:"auth"`
	Tls  []observability_pipeline.TlsModel     `tfsdk:"tls"`
}

type httpClientSourceModel struct {
	Decoding       types.String                      `tfsdk:"decoding"`
	ScrapeInterval types.Int64                       `tfsdk:"scrape_interval_secs"`
	ScrapeTimeout  types.Int64                       `tfsdk:"scrape_timeout_secs"`
	AuthStrategy   types.String                      `tfsdk:"auth_strategy"`
	Tls            []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type googlePubSubSourceModel struct {
	Project      types.String                      `tfsdk:"project"`
	Subscription types.String                      `tfsdk:"subscription"`
	Decoding     types.String                      `tfsdk:"decoding"`
	Auth         []gcpAuthModel                    `tfsdk:"auth"`
	Tls          []observability_pipeline.TlsModel `tfsdk:"tls"`
}

type gcpAuthModel struct {
	CredentialsFile types.String `tfsdk:"credentials_file"`
}

func NewObservabilitPipelineResource() resource.Resource {
	return &observabilityPipelineResource{}
}

func (r *observabilityPipelineResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetObsPipelinesV2()
	r.Auth = providerData.Auth
}

func (r *observabilityPipelineResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "observability_pipeline"
}

func (r *observabilityPipelineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog Observability Pipeline resource. Observability Pipelines allows you to collect and process logs within your own infrastructure, and then route them to downstream integrations. \n\n" +
			"Datadog recommends using the `-parallelism=1` option to apply this resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The pipeline name.",
			},
		},
		Blocks: map[string]schema.Block{
			"config": schema.ListNestedBlock{
				Description: "Configuration for the pipeline.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"pipeline_type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The type of data being ingested. Defaults to `logs` if not specified.",
							Validators: []validator.String{
								stringvalidator.OneOf("logs", "metrics"),
							},
						},
						"use_legacy_search_syntax": schema.BoolAttribute{
							Optional: true,
							Description: "Set to `true` to continue using the legacy search syntax while migrating filter queries. " +
								"After migrating all queries to the new syntax, set to `false`. " +
								"The legacy syntax is deprecated and will eventually be removed. " +
								"Requires Observability Pipelines Worker 2.11 or later. " +
								"See https://docs.datadoghq.com/observability_pipelines/guide/upgrade_your_filter_queries_to_the_new_search_syntax/ for more information.",
						},
					},
					Blocks: map[string]schema.Block{
						"source": schema.ListNestedBlock{
							Description: "List of sources.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Required:    true,
										Description: "The unique identifier for this source.",
									},
								},
								Blocks: map[string]schema.Block{
									"datadog_agent": schema.ListNestedBlock{
										Description: "The `datadog_agent` source collects logs from the Datadog Agent.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"kafka": schema.ListNestedBlock{
										Description: "The `kafka` source ingests data from Apache Kafka topics.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"group_id": schema.StringAttribute{
													Required:    true,
													Description: "The Kafka consumer group ID.",
												},
												"topics": schema.ListAttribute{
													Required:    true,
													Description: "A list of Kafka topic names to subscribe to. The source ingests messages from each topic specified.",
													ElementType: types.StringType,
												},
											},
											Blocks: map[string]schema.Block{
												"librdkafka_option": schema.ListNestedBlock{
													Description: "Advanced librdkafka client configuration options.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"name": schema.StringAttribute{
																Required:    true,
																Description: "The name of the librdkafka option.",
															},
															"value": schema.StringAttribute{
																Required:    true,
																Description: "The value of the librdkafka option.",
															},
														},
													},
												},
												"sasl": schema.ListNestedBlock{
													Description: "SASL authentication settings.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"mechanism": schema.StringAttribute{
																Required:    true,
																Description: "SASL mechanism to use (e.g., PLAIN, SCRAM-SHA-256, SCRAM-SHA-512).",
																Validators: []validator.String{
																	stringvalidator.OneOf("PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"),
																},
															},
														},
													},
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"fluentd": schema.ListNestedBlock{
										Description: "The `fluentd source ingests logs from a Fluentd-compatible service.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"fluent_bit": schema.ListNestedBlock{
										Description: "The `fluent_bit` source ingests logs from Fluent Bit.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"http_server": schema.ListNestedBlock{
										Description: "The `http_server` source collects logs over HTTP POST from external services.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"auth_strategy": schema.StringAttribute{
													Required:    true,
													Description: "HTTP authentication method.",
													Validators: []validator.String{
														stringvalidator.OneOf("none", "plain"),
													},
												},
												"decoding": decodingSchema(),
											},
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"amazon_s3": schema.ListNestedBlock{
										Description: "The `amazon_s3` source ingests logs from an Amazon S3 bucket. It supports AWS authentication and TLS encryption.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"region": schema.StringAttribute{
													Required:    true,
													Description: "AWS region where the S3 bucket resides.",
												},
											},
											Blocks: map[string]schema.Block{
												"auth": observability_pipeline.AwsAuthSchema(),
												"tls":  observability_pipeline.TlsSchema(),
											},
										},
									},
									"splunk_hec": schema.ListNestedBlock{
										Description: "The `splunk_hec` source implements the Splunk HTTP Event Collector (HEC) API.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"splunk_tcp": schema.ListNestedBlock{
										Description: "The `splunk_tcp` source receives logs from a Splunk Universal Forwarder over TCP. TLS is supported for secure transmission.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"rsyslog": schema.ListNestedBlock{
										Description: "The `rsyslog` source listens for logs over TCP or UDP from an `rsyslog` server using the syslog protocol.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"mode": schema.StringAttribute{
													Optional:    true,
													Description: "Protocol used by the syslog source to receive messages.",
												},
											},
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"syslog_ng": schema.ListNestedBlock{
										Description: "The `syslog_ng` source listens for logs over TCP or UDP from a `syslog-ng` server using the syslog protocol.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"mode": schema.StringAttribute{
													Optional:    true,
													Description: "Protocol used by the syslog source to receive messages.",
												},
											},
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"sumo_logic": schema.ListNestedBlock{
										Description:  "The `sumo_logic` source receives logs from Sumo Logic collectors.",
										NestedObject: schema.NestedBlockObject{},
									},
									"amazon_data_firehose": schema.ListNestedBlock{
										Description: "The `amazon_data_firehose` source ingests logs from AWS Data Firehose.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"auth": observability_pipeline.AwsAuthSchema(),
												"tls":  observability_pipeline.TlsSchema(),
											},
										},
									},
									"http_client": schema.ListNestedBlock{
										Description: "The `http_client` source scrapes logs from HTTP endpoints at regular intervals.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"decoding": schema.StringAttribute{
													Required:    true,
													Description: "The decoding format used to interpret incoming logs.",
												},
												"scrape_interval_secs": schema.Int64Attribute{
													Optional:    true,
													Description: "The interval (in seconds) between HTTP scrape requests.",
												},
												"scrape_timeout_secs": schema.Int64Attribute{
													Optional:    true,
													Description: "The timeout (in seconds) for each scrape request.",
												},
												"auth_strategy": schema.StringAttribute{
													Optional:    true,
													Description: "Optional authentication strategy for HTTP requests.",
													Validators: []validator.String{
														stringvalidator.OneOf("none", "basic", "bearer", "custom"),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"google_pubsub": schema.ListNestedBlock{
										Description: "The `google_pubsub` source ingests logs from a Google Cloud Pub/Sub subscription.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"project": schema.StringAttribute{
													Required:    true,
													Description: "The GCP project ID that owns the Pub/Sub subscription.",
												},
												"subscription": schema.StringAttribute{
													Required:    true,
													Description: "The Pub/Sub subscription name from which messages are consumed.",
												},
												"decoding": schema.StringAttribute{
													Required:    true,
													Description: "The decoding format used to interpret incoming logs.",
												},
											},
											Blocks: map[string]schema.Block{
												"auth": gcpAuthSchema(),
												"tls":  observability_pipeline.TlsSchema(),
											},
										},
									},
									"logstash": schema.ListNestedBlock{
										Description: "The `logstash` source ingests logs from a Logstash forwarder.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"socket":        observability_pipeline.SocketSourceSchema(),
									"opentelemetry": observability_pipeline.OpentelemetrySourceSchema(),
								},
							},
						},
						"processor_group": schema.ListNestedBlock{
							Description: "A processor group containing common configuration and nested processors.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Required:    true,
										Description: "The unique ID of the processor group.",
									},
									"enabled": schema.BoolAttribute{
										Required:    true,
										Description: "Whether this processor group is enabled.",
									},
									"include": schema.StringAttribute{
										Required:    true,
										Description: "A Datadog search query used to determine which logs this processor group targets.",
									},
									"inputs": schema.ListAttribute{
										Required:    true,
										ElementType: types.StringType,
										Description: "A list of component IDs whose output is used as the input for this processor group.",
									},
									"display_name": schema.StringAttribute{
										Optional:    true,
										Description: "A human-friendly name of the processor group.",
									},
								},
								Blocks: map[string]schema.Block{
									"processor": schema.ListNestedBlock{
										Description: "The processor contained in this group.",
										NestedObject: schema.NestedBlockObject{
											Validators: []validator.Object{
												observability_pipeline.ExactlyOneProcessorValidator{},
											},
											Attributes: map[string]schema.Attribute{
												"id": schema.StringAttribute{
													Required:    true,
													Description: "The unique identifier for this processor.",
												},
												"enabled": schema.BoolAttribute{
													Required:    true,
													Description: "Whether this processor is enabled.",
												},
												"include": schema.StringAttribute{
													Required:    true,
													Description: "A Datadog search query used to determine which logs this processor targets.",
												},
												"display_name": schema.StringAttribute{
													Optional:    true,
													Description: "A human-friendly name for this processor.",
												},
											},
											Blocks: map[string]schema.Block{
												"filter": schema.ListNestedBlock{
													Description: "The `filter` processor allows conditional processing of logs based on a Datadog search query. Logs that match the `include` query are passed through; others are discarded.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
													},
												},
												"parse_json": schema.ListNestedBlock{
													Description: "The `parse_json` processor extracts JSON from a specified field and flattens it into the event. This is useful when logs contain embedded JSON as a string.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"field": schema.StringAttribute{
																Required:    true,
																Description: "The field to parse.",
															},
														},
													},
												},
												"parse_xml": schema.ListNestedBlock{
													Description: "The `parse_xml` processor parses XML from a specified field and extracts it into the event.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"field": schema.StringAttribute{
																Required:    true,
																Description: "The path to the log field on which you want to parse XML.",
															},
															"include_attr": schema.BoolAttribute{
																Optional:    true,
																Description: "Whether to include XML attributes in the parsed output.",
															},
															"always_use_text_key": schema.BoolAttribute{
																Optional:    true,
																Description: "Whether to always store text inside an object using the text key even when no attributes exist.",
															},
															"parse_number": schema.BoolAttribute{
																Optional:    true,
																Description: "Whether to parse numeric values from strings.",
															},
															"parse_bool": schema.BoolAttribute{
																Optional:    true,
																Description: "Whether to parse boolean values from strings.",
															},
															"parse_null": schema.BoolAttribute{
																Optional:    true,
																Description: "Whether to parse null values.",
															},
															"attr_prefix": schema.StringAttribute{
																Optional:    true,
																Description: "The prefix to use for XML attributes in the parsed output. If the field is left empty, the original attribute key is used.",
															},
															"text_key": schema.StringAttribute{
																Optional:    true,
																Description: "The key name to use for the text node when XML attributes are appended.",
															},
														},
													},
												},
												"add_fields": schema.ListNestedBlock{
													Description: "The `add_fields` processor adds static key-value fields to logs.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
														Blocks: map[string]schema.Block{
															"field": schema.ListNestedBlock{
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																},
																Description: "A list of static fields (key-value pairs) that is added to each log event processed by this component.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"name": schema.StringAttribute{
																			Required:    true,
																			Description: "The field name to add.",
																		},
																		"value": schema.StringAttribute{
																			Required:    true,
																			Description: "The value to assign to the field.",
																		},
																	},
																},
															},
														},
													},
												},
												"add_hostname": schema.ListNestedBlock{
													Description: "The `add_hostname` processor adds the hostname to log events.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
													},
												},
												"rename_fields": schema.ListNestedBlock{
													Description: "The `rename_fields` processor changes field names.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
														Blocks: map[string]schema.Block{
															"field": schema.ListNestedBlock{
																Validators: []validator.List{
																	// this is the only way to make the list of fields required in Terraform
																	listvalidator.IsRequired(),
																},
																Description: "List of fields to rename.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"source": schema.StringAttribute{
																			Required:    true,
																			Description: "Source field to rename.",
																		},
																		"destination": schema.StringAttribute{
																			Required:    true,
																			Description: "Destination field name.",
																		},
																		"preserve_source": schema.BoolAttribute{
																			Required:    true,
																			Description: "Whether to keep the original field.",
																		},
																	},
																},
															},
														},
													},
												},
												"remove_fields": schema.ListNestedBlock{
													Description: "The `remove_fields` processor deletes specified fields from logs.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"fields": schema.ListAttribute{
																Required:    true,
																Description: "List of fields to remove from the events.",
																ElementType: types.StringType,
															},
														},
													},
												},
												"quota": schema.ListNestedBlock{
													Description: "The `quota` processor measures logging traffic for logs that match a specified filter. When the configured daily quota is met, the processor can drop or alert.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"name": schema.StringAttribute{
																Required:    true,
																Description: "The name of the quota.",
															},
															"drop_events": schema.BoolAttribute{
																Optional:    true,
																Description: "Whether to drop events exceeding the limit.",
															},
															"ignore_when_missing_partitions": schema.BoolAttribute{
																Optional:    true,
																Description: "Whether to ignore when partition fields are missing.",
															},
															"partition_fields": schema.ListAttribute{
																Optional:    true,
																ElementType: types.StringType,
																Description: "List of partition fields.",
															},
															"overflow_action": schema.StringAttribute{
																Optional:    true,
																Description: "The action to take when the quota is exceeded: `drop`, `no_action`, or `overflow_routing`.",
															},
															"too_many_buckets_action": schema.StringAttribute{
																Optional:    true,
																Description: "The action to take when the max number of buckets is exceeded: `drop`, `no_action`, or `overflow_routing`.",
															},
														},
														Blocks: map[string]schema.Block{
															"limit": schema.ListNestedBlock{
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"enforce": schema.StringAttribute{
																			Required:    true,
																			Description: "Whether to enforce by 'bytes' or 'events'.",
																			Validators: []validator.String{
																				stringvalidator.OneOf("bytes", "events"),
																			},
																		},
																		"limit": schema.Int64Attribute{
																			Required:    true,
																			Description: "The daily quota limit.",
																		},
																	},
																},
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeAtMost(1),
																},
															},
															"override": schema.ListNestedBlock{
																Description: "The overrides for field-specific quotas.",
																NestedObject: schema.NestedBlockObject{
																	Blocks: map[string]schema.Block{
																		"limit": schema.ListNestedBlock{
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"enforce": schema.StringAttribute{
																						Required:    true,
																						Description: "Whether to enforce by 'bytes' or 'events'.",
																						Validators: []validator.String{
																							stringvalidator.OneOf("bytes", "events"),
																						},
																					},
																					"limit": schema.Int64Attribute{
																						Required:    true,
																						Description: "The daily quota limit.",
																					},
																				},
																			},
																			Validators: []validator.List{
																				listvalidator.IsRequired(),
																				listvalidator.SizeAtMost(1),
																			},
																		},
																		"field": schema.ListNestedBlock{
																			Description: "Fields that trigger this override.",
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"name": schema.StringAttribute{
																						Description: "The field name.",
																						Required:    true,
																					},
																					"value": schema.StringAttribute{
																						Description: "The field value.",
																						Required:    true,
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
												"sensitive_data_scanner": schema.ListNestedBlock{
													Description: "The `sensitive_data_scanner` processor detects and optionally redacts sensitive data in log events.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
														Blocks: map[string]schema.Block{
															"rule": schema.ListNestedBlock{
																Description: "A list of rules for identifying and acting on sensitive data patterns.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"name": schema.StringAttribute{
																			Required:    true,
																			Description: "A name identifying the rule.",
																		},
																		"tags": schema.ListAttribute{
																			Required:    true,
																			ElementType: types.StringType,
																			Description: "Tags assigned to this rule for filtering and classification.",
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"keyword_options": schema.ListNestedBlock{
																			Description: "Keyword-based proximity matching for sensitive data.",
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"keywords": schema.ListAttribute{
																						Optional:    true,
																						ElementType: types.StringType,
																						Description: "A list of keywords to match near the sensitive pattern.",
																					},
																					"proximity": schema.Int64Attribute{
																						Optional:    true,
																						Description: "Maximum number of tokens between a keyword and a sensitive value match.",
																					},
																				},
																			},
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																		},
																		"pattern": schema.ListNestedBlock{
																			Description: "Pattern detection configuration for identifying sensitive data using either a custom regex or a library reference.",
																			NestedObject: schema.NestedBlockObject{
																				Blocks: map[string]schema.Block{
																					"custom": schema.ListNestedBlock{
																						Description: "Pattern detection using a custom regular expression.",
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								"rule": schema.StringAttribute{
																									Optional:    true,
																									Description: "A regular expression used to detect sensitive values. Must be a valid regex.",
																								},
																								"description": schema.StringAttribute{
																									Optional:    true,
																									Description: "Human-readable description providing context about a sensitive data scanner rule.",
																								},
																							},
																						},
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																					},
																					"library": schema.ListNestedBlock{
																						Description: "Pattern detection using a predefined pattern from the sensitive data scanner pattern library.",
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								"id": schema.StringAttribute{
																									Optional:    true,
																									Description: "Identifier for a predefined pattern from the sensitive data scanner pattern library.",
																								},
																								"use_recommended_keywords": schema.BoolAttribute{
																									Optional:    true,
																									Description: "Whether to augment the pattern with recommended keywords (optional).",
																								},
																								"description": schema.StringAttribute{
																									Optional:    true,
																									Description: "Human-readable description providing context about a sensitive data scanner rule.",
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
																		"scope": schema.ListNestedBlock{
																			Description: "Field-level targeting options that determine where the scanner should operate.",
																			NestedObject: schema.NestedBlockObject{
																				Blocks: map[string]schema.Block{
																					"include": schema.ListNestedBlock{
																						Description: "Explicitly include these fields for scanning.",
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								"fields": schema.ListAttribute{
																									Optional:    true,
																									ElementType: types.StringType,
																									Description: "The fields to include in scanning.",
																								},
																							},
																						},
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																					},
																					"exclude": schema.ListNestedBlock{
																						Description: "Explicitly exclude these fields from scanning.",
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								"fields": schema.ListAttribute{
																									Optional:    true,
																									ElementType: types.StringType,
																									Description: "The fields to exclude from scanning.",
																								},
																							},
																						},
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																					},
																				},
																				Attributes: map[string]schema.Attribute{
																					"all": schema.BoolAttribute{
																						Optional:    true,
																						Description: "Scan all fields.",
																					},
																				},
																			},
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																		},
																		"on_match": schema.ListNestedBlock{
																			Description: "The action to take when a sensitive value is found.",
																			NestedObject: schema.NestedBlockObject{
																				Blocks: map[string]schema.Block{
																					"redact": schema.ListNestedBlock{
																						Description: "Redacts the matched value.",
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								"replace": schema.StringAttribute{
																									Optional:    true,
																									Description: "Replacement string for redacted values (e.g., `***`).",
																								},
																							},
																						},
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																					},
																					"hash": schema.ListNestedBlock{
																						Description: "Hashes the matched value.",
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{}, // empty options
																						},
																						Validators: []validator.List{
																							listvalidator.SizeAtMost(1),
																						},
																					},
																					"partial_redact": schema.ListNestedBlock{
																						Description: "Redacts part of the matched value (e.g., keep last 4 characters).",
																						NestedObject: schema.NestedBlockObject{
																							Attributes: map[string]schema.Attribute{
																								"characters": schema.Int64Attribute{
																									Optional:    true,
																									Description: "Number of characters to keep.",
																								},
																								"direction": schema.StringAttribute{
																									Optional:    true,
																									Description: "Direction from which to keep characters: `first` or `last`.",
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
																	},
																},
															},
														},
													},
												},
												"generate_datadog_metrics": schema.ListNestedBlock{
													Description: "The `generate_datadog_metrics` processor creates custom metrics from logs. Metrics can be counters, gauges, or distributions and optionally grouped by log fields.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
														Blocks: map[string]schema.Block{
															"metric": schema.ListNestedBlock{
																Description: "Configuration for generating individual metrics.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"name": schema.StringAttribute{
																			Required:    true,
																			Description: "Name of the custom metric to be created.",
																		},
																		"include": schema.StringAttribute{
																			Required:    true,
																			Description: "Datadog filter query to match logs for metric generation.",
																		},
																		"metric_type": schema.StringAttribute{
																			Required:    true,
																			Description: "Type of metric to create.",
																		},
																		"group_by": schema.ListAttribute{
																			Optional:    true,
																			ElementType: types.StringType,
																			Description: "Optional fields used to group the metric series.",
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"value": schema.ListNestedBlock{
																			Description: "Specifies how the value of the generated metric is computed.",
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"strategy": schema.StringAttribute{
																						Required:    true,
																						Description: "Metric value strategy: `increment_by_one` or `increment_by_field`.",
																					},
																					"field": schema.StringAttribute{
																						Optional:    true,
																						Description: "Name of the log field containing the numeric value to increment the metric by (used only for `increment_by_field`).",
																					},
																				},
																			},
																			Validators: []validator.List{
																				listvalidator.IsRequired(),
																				listvalidator.SizeAtMost(1),
																			},
																		},
																	},
																},
															},
														},
													},
												},
												"parse_grok": schema.ListNestedBlock{
													Description: "The `parse_grok` processor extracts structured fields from unstructured log messages using Grok patterns.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"disable_library_rules": schema.BoolAttribute{
																Optional:    true,
																Description: "If set to `true`, disables the default Grok rules provided by Datadog.",
															},
														},
														Blocks: map[string]schema.Block{
															"rule": schema.ListNestedBlock{
																Description: "The list of Grok parsing rules. If multiple parsing rules are provided, they are evaluated in order. The first successful match is applied.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"source": schema.StringAttribute{
																			Required:    true,
																			Description: "The name of the field in the log event to apply the Grok rules to.",
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"match_rule": schema.ListNestedBlock{
																			Description: "A list of Grok parsing rules that define how to extract fields from the source field. Each rule must contain a name and a valid Grok pattern.",
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"name": schema.StringAttribute{
																						Required:    true,
																						Description: "The name of the rule.",
																					},
																					"rule": schema.StringAttribute{
																						Required:    true,
																						Description: "The definition of the Grok rule.",
																					},
																				},
																			},
																		},
																		"support_rule": schema.ListNestedBlock{
																			Description: "A list of helper Grok rules that can be referenced by the parsing rules.",
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"name": schema.StringAttribute{
																						Required:    true,
																						Description: "The name of the helper Grok rule.",
																					},
																					"rule": schema.StringAttribute{
																						Required:    true,
																						Description: "The definition of the helper Grok rule.",
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
												"sample": schema.ListNestedBlock{
													Description: "The `sample` processor allows probabilistic sampling of logs at a fixed rate.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"percentage": schema.Float64Attribute{
																Required:    true,
																Description: "The percentage of logs to sample.",
															},
															"group_by": schema.ListAttribute{
																Optional:    true,
																ElementType: types.StringType,
																Description: "Optional list of fields to group events by. Each group is sampled independently.",
															},
														},
													},
												},
												"dedupe": schema.ListNestedBlock{
													Description: "The `dedupe` processor removes duplicate fields in log events.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"fields": schema.ListAttribute{
																Required:    true,
																ElementType: types.StringType,
																Description: "A list of log field paths to check for duplicates.",
															},
															"mode": schema.StringAttribute{
																Required:    true,
																Description: "The deduplication mode to apply to the fields.",
															},
														},
													},
												},
												"reduce": schema.ListNestedBlock{
													Description: "The `reduce` processor aggregates and merges logs based on matching keys and merge strategies.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"group_by": schema.ListAttribute{
																Required:    true,
																ElementType: types.StringType,
																Description: "A list of fields used to group log events for merging.",
															},
														},
														Blocks: map[string]schema.Block{
															"merge_strategy": schema.ListNestedBlock{
																Description: "List of merge strategies defining how values from grouped events should be combined.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"path": schema.StringAttribute{
																			Required:    true,
																			Description: "The field path in the log event.",
																		},
																		"strategy": schema.StringAttribute{
																			Required:    true,
																			Description: "The merge strategy to apply.",
																		},
																	},
																},
															},
														},
													},
												},
												"split_array": schema.ListNestedBlock{
													Description: "The `split_array` processor splits array fields into separate events based on configured rules.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
														Blocks: map[string]schema.Block{
															"array": schema.ListNestedBlock{
																Description: "A list of array split configurations.",
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeAtMost(15),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"include": schema.StringAttribute{
																			Required:    true,
																			Description: "A Datadog search query used to determine which logs this array split operation targets.",
																		},
																		"field": schema.StringAttribute{
																			Required:    true,
																			Description: "The path to the array field to split.",
																		},
																	},
																},
															},
														},
													},
												},
												"throttle": schema.ListNestedBlock{
													Description: "The `throttle` processor limits the number of events that pass through over a given time window.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"threshold": schema.Int64Attribute{
																Required:    true,
																Description: "The number of events to allow before throttling is applied.",
															},
															"window": schema.Float64Attribute{
																Required:    true,
																Description: "The time window in seconds over which the threshold applies.",
															},
															"group_by": schema.ListAttribute{
																Optional:    true,
																ElementType: types.StringType,
																Description: "Optional list of fields used to group events before applying throttling.",
															},
														},
													},
												},
												"add_env_vars": schema.ListNestedBlock{
													Description: "The `add_env_vars` processor adds environment variable values to log events.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
														Blocks: map[string]schema.Block{
															"variable": schema.ListNestedBlock{
																Description: "A list of environment variable mappings to apply to log fields.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"field": schema.StringAttribute{
																			Required:    true,
																			Description: "The target field in the log event.",
																		},
																		"name": schema.StringAttribute{
																			Required:    true,
																			Description: "The name of the environment variable to read.",
																		},
																	},
																},
															},
														},
													},
												},
												"enrichment_table": schema.ListNestedBlock{
													Description: "The `enrichment_table` processor enriches logs using a static CSV file or GeoIP database.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"target": schema.StringAttribute{
																Required:    true,
																Description: "Path where enrichment results should be stored in the log.",
															},
														},
														Blocks: map[string]schema.Block{
															"file": schema.ListNestedBlock{
																Description: "Defines a static enrichment table loaded from a CSV file.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"path": schema.StringAttribute{
																			Optional:    true,
																			Description: "Path to the CSV file.",
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"encoding": schema.ListNestedBlock{
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"type": schema.StringAttribute{
																						Required:    true,
																						Description: "File encoding format.",
																					},
																					"delimiter": schema.StringAttribute{
																						Required:    true,
																						Description: "The `encoding` `delimiter`.",
																					},
																					"includes_headers": schema.BoolAttribute{
																						Optional:    true,
																						Description: "The `encoding` `includes_headers`.",
																					},
																				},
																			},
																			Validators: []validator.List{
																				listvalidator.IsRequired(),
																				listvalidator.SizeAtMost(1),
																			},
																		},
																		"key": schema.ListNestedBlock{
																			Description: "Key fields used to look up enrichment values.",
																			NestedObject: schema.NestedBlockObject{
																				Attributes: map[string]schema.Attribute{
																					"column": schema.StringAttribute{
																						Optional:    true,
																						Description: "The `items` `column`.",
																					},
																					"comparison": schema.StringAttribute{
																						Optional:    true,
																						Description: "The comparison method (e.g. equals).",
																					},
																					"field": schema.StringAttribute{
																						Optional:    true,
																						Description: "The `items` `field`.",
																					},
																				},
																			},
																		},
																	},
																},
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
															},
															"geoip": schema.ListNestedBlock{
																Description: "Uses a GeoIP database to enrich logs based on an IP field.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"key_field": schema.StringAttribute{
																			Optional:    true,
																			Description: "Path to the IP field in the log.",
																		},
																		"locale": schema.StringAttribute{
																			Optional:    true,
																			Description: "Locale used to resolve geographical names.",
																		},
																		"path": schema.StringAttribute{
																			Optional:    true,
																			Description: "Path to the GeoIP database file.",
																		},
																	},
																},
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
															},
															"reference_table": schema.ListNestedBlock{
																Description: "Uses a Datadog reference table to enrich logs.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"key_field": schema.StringAttribute{
																			Required:    true,
																			Description: "Path to the field in the log event to match against the reference table.",
																		},
																		"table_id": schema.StringAttribute{
																			Required:    true,
																			Description: "The unique identifier of the reference table.",
																		},
																		"columns": schema.ListAttribute{
																			Optional:    true,
																			ElementType: types.StringType,
																			Description: "List of column names to include from the reference table. If not provided, all columns are included.",
																		},
																	},
																},
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
															},
														},
													},
												},
												"ocsf_mapper": schema.ListNestedBlock{
													Description: "The `ocsf_mapper` processor transforms logs into the OCSF schema using predefined library mappings.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
														Blocks: map[string]schema.Block{
															"mapping": schema.ListNestedBlock{
																Description: "List of OCSF mapping entries using library mapping.",
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"include": schema.StringAttribute{
																			Required:    true,
																			Description: "Search query for selecting which logs the mapping applies to.",
																		},
																		"library_mapping": schema.StringAttribute{
																			Required:    true,
																			Description: "Predefined library mapping for log transformation.",
																		},
																	},
																},
															},
														},
													},
												},
												"datadog_tags":     observability_pipeline.DatadogTagsProcessorSchema(),
												"custom_processor": observability_pipeline.CustomProcessorSchema(),
												"metric_tags": schema.ListNestedBlock{
													Description: "The `metric_tags` processor filters metrics based on their tags using Datadog tag key patterns.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{},
														Blocks: map[string]schema.Block{
															"rule": schema.ListNestedBlock{
																Description: "A list of rules for filtering metric tags.",
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeAtMost(100),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"include": schema.StringAttribute{
																			Required:    true,
																			Description: "A Datadog search query used to determine which metrics this rule targets.",
																		},
																		"mode": schema.StringAttribute{
																			Required:    true,
																			Description: "The processing mode for tag filtering.",
																			Validators: []validator.String{
																				stringvalidator.OneOf("filter"),
																			},
																		},
																		"action": schema.StringAttribute{
																			Required:    true,
																			Description: "The action to take on tags with matching keys.",
																			Validators: []validator.String{
																				stringvalidator.OneOf("include", "exclude"),
																			},
																		},
																		"keys": schema.ListAttribute{
																			ElementType: types.StringType,
																			Required:    true,
																			Description: "A list of tag keys to include or exclude.",
																			Validators: []validator.List{
																				listvalidator.SizeAtLeast(1),
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"destination": schema.ListNestedBlock{
							Description: "List of destinations.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Required:    true,
										Description: "The unique identifier for this destination.",
									},
									"inputs": schema.ListAttribute{
										Required:    true,
										Description: "A list of component IDs whose output is used as the `input` for this component.",
										ElementType: types.StringType,
									},
								},
								Blocks: map[string]schema.Block{
									"datadog_logs": schema.ListNestedBlock{
										Description: "The `datadog_logs` destination forwards logs to Datadog Log Management.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"routes": schema.ListNestedBlock{
													Description: "A list of routing rules that forward matching logs to Datadog using dedicated API keys.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"route_id": schema.StringAttribute{
																Required:    true,
																Description: "Unique identifier for this route within the destination.",
															},
															"include": schema.StringAttribute{
																Required:    true,
																Description: "A Datadog search query that determines which logs are forwarded using this route.",
															},
															"site": schema.StringAttribute{
																Required:    true,
																Description: "Datadog site where matching logs are sent (for example, `us1`).",
															},
															"api_key_key": schema.StringAttribute{
																Required:    true,
																Description: "Name of the environment variable or secret that stores the Datadog API key used by this route.",
															},
														},
													},
													Validators: []validator.List{
														listvalidator.SizeAtMost(100),
													},
												},
											},
										},
									},
									"datadog_metrics": schema.ListNestedBlock{
										Description:  "The `datadog_metrics` destination forwards metrics to Datadog.",
										NestedObject: schema.NestedBlockObject{},
									},
									"http_client": schema.ListNestedBlock{
										Description: "The `http_client` destination sends data to an HTTP endpoint.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"encoding": schema.StringAttribute{
													Required:    true,
													Description: "Encoding format for events.",
													Validators: []validator.String{
														stringvalidator.OneOf("json"),
													},
												},
												"auth_strategy": schema.StringAttribute{
													Optional:    true,
													Description: "HTTP authentication strategy.",
													Validators: []validator.String{
														stringvalidator.OneOf("none", "basic", "bearer"),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"compression": schema.ListNestedBlock{
													Description: "Compression configuration for HTTP requests.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"algorithm": schema.StringAttribute{
																Required:    true,
																Description: "Compression algorithm.",
																Validators: []validator.String{
																	stringvalidator.OneOf("gzip"),
																},
															},
														},
													},
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
												},
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"google_cloud_storage": schema.ListNestedBlock{
										Description: "The `google_cloud_storage` destination stores logs in a Google Cloud Storage (GCS) bucket.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"bucket": schema.StringAttribute{
													Required:    true,
													Description: "Name of the GCS bucket.",
												},
												"key_prefix": schema.StringAttribute{
													Optional:    true,
													Description: "Optional prefix for object keys within the GCS bucket.",
												},
												"storage_class": schema.StringAttribute{
													Required:    true,
													Description: "Storage class used for objects stored in GCS.",
												},
												"acl": schema.StringAttribute{
													Optional:    true,
													Description: "Access control list setting for objects written to the bucket.",
												},
											},
											Blocks: map[string]schema.Block{
												"auth": gcpAuthSchema(),
												"metadata": schema.ListNestedBlock{
													Description: "Custom metadata key-value pairs added to each object.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"name": schema.StringAttribute{
																Required:    true,
																Description: "The metadata key.",
															},
															"value": schema.StringAttribute{
																Required:    true,
																Description: "The metadata value.",
															},
														},
													},
												},
											},
										},
									},
									"google_pubsub": schema.ListNestedBlock{
										Description: "The `google_pubsub` destination publishes logs to a Google Cloud Pub/Sub topic.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"project": schema.StringAttribute{
													Required:    true,
													Description: "The GCP project ID that owns the Pub/Sub topic.",
												},
												"topic": schema.StringAttribute{
													Required:    true,
													Description: "The Pub/Sub topic name to publish logs to.",
												},
												"encoding": schema.StringAttribute{
													Optional:    true,
													Description: "Encoding format for log events. Valid values: `json`, `raw_message`.",
												},
											},
											Blocks: map[string]schema.Block{
												"auth": gcpAuthSchema(),
												"tls":  observability_pipeline.TlsSchema(),
											},
										},
									},
									"splunk_hec": schema.ListNestedBlock{
										Description: "The `splunk_hec` destination forwards logs to Splunk using the HTTP Event Collector (HEC).",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"auto_extract_timestamp": schema.BoolAttribute{
													Optional:    true,
													Description: "If `true`, Splunk tries to extract timestamps from incoming log events.",
												},
												"encoding": schema.StringAttribute{
													Required:    true,
													Description: "Encoding format for log events. Valid values: `json`, `raw_message`.",
												},
												"sourcetype": schema.StringAttribute{
													Optional:    true,
													Description: "The Splunk sourcetype to assign to log events.",
												},
												"index": schema.StringAttribute{
													Optional:    true,
													Description: "Optional name of the Splunk index where logs are written.",
												},
											},
										},
									},
									"sumo_logic": schema.ListNestedBlock{
										Description: "The `sumo_logic` destination forwards logs to Sumo Logic.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"encoding": schema.StringAttribute{
													Optional:    true,
													Description: "The output encoding format.",
												},
												"header_host_name": schema.StringAttribute{
													Optional:    true,
													Description: "Optional override for the host name header.",
												},
												"header_source_name": schema.StringAttribute{
													Optional:    true,
													Description: "Optional override for the source name header.",
												},
												"header_source_category": schema.StringAttribute{
													Optional:    true,
													Description: "Optional override for the source category header.",
												},
											},
											Blocks: map[string]schema.Block{
												"header_custom_field": schema.ListNestedBlock{
													Description: "A list of custom headers to include in the request to Sumo Logic.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"name": schema.StringAttribute{
																Optional:    true,
																Description: "The header field name.",
															},
															"value": schema.StringAttribute{
																Optional:    true,
																Description: "The header field value.",
															},
														},
													},
												},
											},
										},
									},
									"rsyslog": schema.ListNestedBlock{
										Description: "The `rsyslog` destination forwards logs to an external `rsyslog` server over TCP or UDP using the syslog protocol.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"keepalive": schema.Int64Attribute{
													Optional:    true,
													Description: "Optional socket keepalive duration in milliseconds.",
												},
											},
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"syslog_ng": schema.ListNestedBlock{
										Description: "The `syslog_ng` destination forwards logs to an external `syslog-ng` server over TCP or UDP using the syslog protocol.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"keepalive": schema.Int64Attribute{
													Optional:    true,
													Description: "Optional socket keepalive duration in milliseconds.",
												},
											},
											Blocks: map[string]schema.Block{
												"tls": observability_pipeline.TlsSchema(),
											},
										},
									},
									"elasticsearch": schema.ListNestedBlock{
										Description: "The `elasticsearch` destination writes logs to an Elasticsearch cluster.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"api_version": schema.StringAttribute{
													Optional:    true,
													Description: "The Elasticsearch API version to use. Set to `auto` to auto-detect.",
												},
												"bulk_index": schema.StringAttribute{
													Optional:    true,
													Description: "The index or datastream to write logs to in Elasticsearch.",
												},
											},
											Blocks: map[string]schema.Block{
												"data_stream": schema.ListNestedBlock{
													Description: "Configuration options for writing to Elasticsearch Data Streams instead of a fixed index.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"dtype": schema.StringAttribute{
																Optional:    true,
																Description: "The data stream type for your logs. This determines how logs are categorized within the data stream.",
															},
															"dataset": schema.StringAttribute{
																Optional:    true,
																Description: "The data stream dataset for your logs. This groups logs by their source or application.",
															},
															"namespace": schema.StringAttribute{
																Optional:    true,
																Description: "The data stream namespace for your logs. This separates logs into different environments or domains.",
															},
														},
													},
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ConflictsWith(frameworkPath.MatchRelative().AtParent().AtName("bulk_index")),
													},
												},
											},
										},
									},
									"opensearch": schema.ListNestedBlock{
										Description: "The `opensearch` destination writes logs to an OpenSearch cluster.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"bulk_index": schema.StringAttribute{
													Optional:    true,
													Description: "The index or datastream to write logs to.",
												},
											},
											Blocks: map[string]schema.Block{
												"data_stream": schema.ListNestedBlock{
													Description: "Configuration options for writing to OpenSearch Data Streams instead of a fixed index.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"dtype": schema.StringAttribute{
																Optional:    true,
																Description: "The data stream type for your logs. This determines how logs are categorized within the data stream.",
															},
															"dataset": schema.StringAttribute{
																Optional:    true,
																Description: "The data stream dataset for your logs. This groups logs by their source or application.",
															},
															"namespace": schema.StringAttribute{
																Optional:    true,
																Description: "The data stream namespace for your logs. This separates logs into different environments or domains.",
															},
														},
													},
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ConflictsWith(frameworkPath.MatchRelative().AtParent().AtName("bulk_index")),
													},
												},
											},
										},
									},
									"amazon_opensearch": schema.ListNestedBlock{
										Description: "The `amazon_opensearch` destination writes logs to Amazon OpenSearch.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"bulk_index": schema.StringAttribute{
													Optional:    true,
													Description: "The index or datastream to write logs to.",
												},
											},
											Blocks: map[string]schema.Block{
												"auth": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"strategy": schema.StringAttribute{
																Required:    true,
																Description: "The authentication strategy to use (e.g. aws or basic).",
															},
															"aws_region": schema.StringAttribute{
																Optional:    true,
																Description: "AWS region override (if applicable).",
															},
															"assume_role": schema.StringAttribute{
																Optional:    true,
																Description: "ARN of the role to assume.",
															},
															"external_id": schema.StringAttribute{
																Optional:    true,
																Description: "External ID for assumed role.",
															},
															"session_name": schema.StringAttribute{
																Optional:    true,
																Description: "Session name for assumed role.",
															},
														},
													},
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtMost(1),
													},
												},
											},
										},
									},
									"azure_storage": schema.ListNestedBlock{
										Description: "The `azure_storage` destination forwards logs to an Azure Blob Storage container.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"container_name": schema.StringAttribute{
													Required:    true,
													Description: "The name of the Azure Blob Storage container to store logs in.",
												},
												"blob_prefix": schema.StringAttribute{
													Optional:    true,
													Description: "Optional prefix for blobs written to the container.",
												},
											},
										},
									},
									"microsoft_sentinel": schema.ListNestedBlock{
										Description: "The `microsoft_sentinel` destination forwards logs to Microsoft Sentinel.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"client_id": schema.StringAttribute{
													Required:    true,
													Description: "Azure AD client ID used for authentication.",
												},
												"tenant_id": schema.StringAttribute{
													Required:    true,
													Description: "Azure AD tenant ID.",
												},
												"dcr_immutable_id": schema.StringAttribute{
													Required:    true,
													Description: "The immutable ID of the Data Collection Rule (DCR).",
												},
												"table": schema.StringAttribute{
													Required:    true,
													Description: "The name of the Log Analytics table where logs will be sent.",
												},
											},
										},
									},
									"google_secops": schema.ListNestedBlock{
										Description: "The `google_chronicle` destination sends logs to Google SecOps.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"customer_id": schema.StringAttribute{
													Required:    true,
													Description: "The Google SecOps customer ID.",
												},
												"encoding": schema.StringAttribute{
													Required:    true,
													Description: "The encoding format for the logs sent to Google SecOps.",
													Validators: []validator.String{
														stringvalidator.OneOf("json", "raw_message"),
													},
												},
												"log_type": schema.StringAttribute{
													Required:    true,
													Description: "The log type metadata associated with the Google SecOps destination.",
												},
											},
											Blocks: map[string]schema.Block{
												"auth": gcpAuthSchema(),
											},
										},
									},
									"new_relic": schema.ListNestedBlock{
										Description: "The `new_relic` destination sends logs to the New Relic platform.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"region": schema.StringAttribute{
													Required:    true,
													Description: "The New Relic region.",
												},
											},
										},
									},
									"sentinel_one": schema.ListNestedBlock{
										Description: "The `sentinel_one` destination sends logs to SentinelOne.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"region": schema.StringAttribute{
													Required:    true,
													Description: "The SentinelOne region to send logs to.",
												},
											},
										},
									},
									"socket":                    observability_pipeline.SocketDestinationSchema(),
									"amazon_s3":                 observability_pipeline.AmazonS3DestinationSchema(),
									"amazon_security_lake":      observability_pipeline.AmazonSecurityLakeDestinationSchema(),
									"crowdstrike_next_gen_siem": observability_pipeline.CrowdStrikeNextGenSiemDestinationSchema(),
									"cloud_prem":                observability_pipeline.CloudPremDestinationSchema(),
									"kafka":                     observability_pipeline.KafkaDestinationSchema(),
								},
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func gcpAuthSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "GCP credentials used to authenticate with Google Cloud services.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"credentials_file": schema.StringAttribute{
					Required:    true,
					Description: "Path to the GCP service account key file.",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

func expandGcpAuth(auth []gcpAuthModel) *datadogV2.ObservabilityPipelineGcpAuth {
	if len(auth) == 0 {
		return nil
	}

	return &datadogV2.ObservabilityPipelineGcpAuth{
		CredentialsFile: auth[0].CredentialsFile.ValueString(),
	}
}

func flattenGcpAuth(auth *datadogV2.ObservabilityPipelineGcpAuth) []gcpAuthModel {
	if auth == nil {
		return nil
	}

	return []gcpAuthModel{
		{
			CredentialsFile: types.StringValue(auth.CredentialsFile),
		},
	}
}

func (r *observabilityPipelineResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *observabilityPipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state observabilityPipelineModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...) // Read config from plan
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := expandPipeline(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := datadogV2.NewObservabilityPipelineSpecWithDefaults()
	createReq.Data = *datadogV2.NewObservabilityPipelineSpecDataWithDefaults()
	createReq.Data.Attributes = body.Data.Attributes

	// Used for debugging purposes in the TF tests to display the payload sent to the Public API
	if os.Getenv("TF_LOG") == "DEBUG" {
		reqBytes, _ := json.MarshalIndent(createReq, "", "  ")
		log.Printf("[DEBUG] Creating pipeline with request: %s", string(reqBytes))
	}

	result, _, err := r.Api.CreatePipeline(r.Auth, *createReq)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Pipeline"))
		return
	}
	if err := utils.CheckForUnparsed(result); err != nil {
		resp.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	flattenPipeline(ctx, &state, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // Save to state
}

func (r *observabilityPipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state observabilityPipelineModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Load current state
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	result, httpResp, err := r.Api.GetPipeline(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Pipeline"))
		return
	}
	if err := utils.CheckForUnparsed(result); err != nil {
		resp.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	flattenPipeline(ctx, &state, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // Save to state
}

func (r *observabilityPipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state observabilityPipelineModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...) // Read config from plan
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	body, diags := expandPipeline(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Used for debugging purposes in the TF tests to display the payload sent to the Public API
	if os.Getenv("TF_LOG") == "DEBUG" {
		reqBytes, _ := json.MarshalIndent(body, "", "  ")
		log.Printf("[DEBUG] Updating pipeline %s with request: %s", id, string(reqBytes))
	}

	result, _, err := r.Api.UpdatePipeline(r.Auth, id, *body)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Pipeline"))
		return
	}
	if err := utils.CheckForUnparsed(result); err != nil {
		resp.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	flattenPipeline(ctx, &state, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // Save to state
}

func (r *observabilityPipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state observabilityPipelineModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Load current state
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	httpResp, err := r.Api.DeletePipeline(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting Pipeline"))
		return
	}
}

// --- Expansion - converting TF state to API model ---
func expandPipeline(ctx context.Context, state *observabilityPipelineModel) (*datadogV2.ObservabilityPipeline, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := datadogV2.NewObservabilityPipelineWithDefaults()
	data := datadogV2.NewObservabilityPipelineDataWithDefaults()
	attrs := datadogV2.NewObservabilityPipelineDataAttributesWithDefaults()

	if !state.Name.IsNull() {
		attrs.SetName(state.Name.ValueString())
	}

	config := datadogV2.NewObservabilityPipelineConfigWithDefaults()

	// Always set pipeline_type, defaulting to "logs" if not specified
	pipelineType := "logs"
	if len(state.Config) > 0 && !state.Config[0].PipelineType.IsNull() && !state.Config[0].PipelineType.IsUnknown() {
		pipelineType = state.Config[0].PipelineType.ValueString()
	}
	config.SetPipelineType(datadogV2.ObservabilityPipelineConfigPipelineType(pipelineType))

	if len(state.Config) > 0 && !state.Config[0].UseLegacySearchSyntax.IsNull() && !state.Config[0].UseLegacySearchSyntax.IsUnknown() {
		config.SetUseLegacySearchSyntax(state.Config[0].UseLegacySearchSyntax.ValueBool())
	}

	// Sources
	for _, sourceBlock := range state.Config[0].Sources {
		sourceId := sourceBlock.Id.ValueString()
		for _, s := range sourceBlock.DatadogAgentSource {
			config.Sources = append(config.Sources, expandDatadogAgentSource(s, sourceId))
		}
		for _, k := range sourceBlock.KafkaSource {
			config.Sources = append(config.Sources, expandKafkaSource(k, sourceId))
		}
		for _, f := range sourceBlock.FluentdSource {
			config.Sources = append(config.Sources, expandFluentdSource(f, sourceId))
		}
		for _, f := range sourceBlock.FluentBitSource {
			config.Sources = append(config.Sources, expandFluentBitSource(f, sourceId))
		}
		for _, s := range sourceBlock.HttpServerSource {
			config.Sources = append(config.Sources, expandHttpServerSource(s, sourceId))
		}
		for _, s := range sourceBlock.SplunkHecSource {
			config.Sources = append(config.Sources, expandSplunkHecSource(s, sourceId))
		}
		for _, s := range sourceBlock.SplunkTcpSource {
			config.Sources = append(config.Sources, expandSplunkTcpSource(s, sourceId))
		}
		for _, s := range sourceBlock.AmazonS3Source {
			config.Sources = append(config.Sources, expandAmazonS3Source(s, sourceId))
		}
		for _, s := range sourceBlock.RsyslogSource {
			config.Sources = append(config.Sources, expandRsyslogSource(s, sourceId))
		}
		for _, s := range sourceBlock.SyslogNgSource {
			config.Sources = append(config.Sources, expandSyslogNgSource(s, sourceId))
		}
		for _, s := range sourceBlock.SumoLogicSource {
			config.Sources = append(config.Sources, expandSumoLogicSource(s, sourceId))
		}
		for _, a := range sourceBlock.AmazonDataFirehoseSource {
			config.Sources = append(config.Sources, expandAmazonDataFirehoseSource(a, sourceId))
		}
		for _, h := range sourceBlock.HttpClientSource {
			config.Sources = append(config.Sources, expandHttpClientSource(h, sourceId))
		}
		for _, g := range sourceBlock.GooglePubSubSource {
			config.Sources = append(config.Sources, expandGooglePubSubSource(g, sourceId))
		}
		for _, l := range sourceBlock.LogstashSource {
			config.Sources = append(config.Sources, expandLogstashSource(l, sourceId))
		}
		for _, s := range sourceBlock.SocketSource {
			item, d := observability_pipeline.ExpandSocketSource(s, sourceId)
			diags.Append(d...)
			if d.HasError() {
				return nil, diags
			}
			config.Sources = append(config.Sources, item)
		}
		for _, o := range sourceBlock.OpentelemetrySource {
			config.Sources = append(config.Sources, observability_pipeline.ExpandOpentelemetrySource(o, sourceId))
		}
	}

	// Processors - iterate through processor groups
	for _, group := range state.Config[0].ProcessorGroups {
		processorGroup := expandProcessorGroup(ctx, group)
		config.ProcessorGroups = append(config.ProcessorGroups, processorGroup)
	}

	// Destinations
	for _, dest := range state.Config[0].Destinations {
		for _, d := range dest.DatadogLogsDestination {
			config.Destinations = append(config.Destinations, expandDatadogLogsDestination(ctx, dest, d))
		}
		for _, d := range dest.DatadogMetricsDestination {
			config.Destinations = append(config.Destinations, expandDatadogMetricsDestination(ctx, dest, d))
		}
		for _, d := range dest.HttpClientDestination {
			config.Destinations = append(config.Destinations, expandHttpClientDestination(ctx, dest, d))
		}
		for _, d := range dest.SplunkHecDestination {
			config.Destinations = append(config.Destinations, expandSplunkHecDestination(ctx, dest, d))
		}
		for _, d := range dest.GoogleCloudStorageDestination {
			config.Destinations = append(config.Destinations, expandGoogleCloudStorageDestination(ctx, dest, d))
		}
		for _, d := range dest.GooglePubSubDestination {
			config.Destinations = append(config.Destinations, expandGooglePubSubDestination(ctx, dest, d))
		}
		for _, d := range dest.SumoLogicDestination {
			config.Destinations = append(config.Destinations, expandSumoLogicDestination(ctx, dest, d))
		}
		for _, d := range dest.RsyslogDestination {
			config.Destinations = append(config.Destinations, expandRsyslogDestination(ctx, dest, d))
		}
		for _, d := range dest.SyslogNgDestination {
			config.Destinations = append(config.Destinations, expandSyslogNgDestination(ctx, dest, d))
		}
		for _, d := range dest.ElasticsearchDestination {
			config.Destinations = append(config.Destinations, expandElasticsearchDestination(ctx, dest, d))
		}
		for _, d := range dest.AzureStorageDestination {
			config.Destinations = append(config.Destinations, expandAzureStorageDestination(ctx, dest, d))
		}
		for _, d := range dest.MicrosoftSentinelDestination {
			config.Destinations = append(config.Destinations, expandMicrosoftSentinelDestination(ctx, dest, d))
		}
		for _, d := range dest.GoogleSecopsDestination {
			config.Destinations = append(config.Destinations, expandGoogleSecopsDestination(ctx, dest, d))
		}
		for _, d := range dest.NewRelicDestination {
			config.Destinations = append(config.Destinations, expandNewRelicDestination(ctx, dest, d))
		}
		for _, d := range dest.SentinelOneDestination {
			config.Destinations = append(config.Destinations, expandSentinelOneDestination(ctx, dest, d))
		}
		for _, d := range dest.OpenSearchDestination {
			config.Destinations = append(config.Destinations, expandOpenSearchDestination(ctx, dest, d))
		}
		for _, d := range dest.AmazonOpenSearchDestination {
			config.Destinations = append(config.Destinations, expandAmazonOpenSearchDestination(ctx, dest, d))
		}
		for _, d := range dest.SocketDestination {
			item, socketDiags := observability_pipeline.ExpandSocketDestination(ctx, dest.Id.ValueString(), dest.Inputs, d)
			diags.Append(socketDiags...)
			if socketDiags.HasError() {
				return nil, diags
			}
			config.Destinations = append(config.Destinations, item)
		}
		for _, d := range dest.AmazonS3Destination {
			config.Destinations = append(config.Destinations, observability_pipeline.ExpandAmazonS3Destination(ctx, dest.Id.ValueString(), dest.Inputs, d))
		}
		for _, d := range dest.AmazonSecurityLakeDestination {
			config.Destinations = append(config.Destinations, observability_pipeline.ExpandObservabilityPipelinesAmazonSecurityLakeDestination(ctx, dest.Id.ValueString(), dest.Inputs, d))
		}
		for _, d := range dest.CrowdStrikeNextGenSiemDestination {
			config.Destinations = append(config.Destinations, observability_pipeline.ExpandCrowdStrikeNextGenSiemDestination(ctx, dest.Id.ValueString(), dest.Inputs, d))
		}
		for _, d := range dest.CloudPremDestination {
			config.Destinations = append(config.Destinations, observability_pipeline.ExpandCloudPremDestination(ctx, dest.Id.ValueString(), dest.Inputs, d))
		}
		for _, d := range dest.KafkaDestination {
			config.Destinations = append(config.Destinations, observability_pipeline.ExpandKafkaDestination(ctx, dest.Id.ValueString(), dest.Inputs, d))
		}
	}

	attrs.SetConfig(*config)
	data.SetAttributes(*attrs)
	req.SetData(*data)
	return req, diags
}

// --- Flattening - converting API model to TF state ---
func flattenPipeline(ctx context.Context, state *observabilityPipelineModel, resp *datadogV2.ObservabilityPipeline) {
	state.ID = types.StringValue(resp.Data.GetId())
	attrs := resp.Data.GetAttributes()
	state.Name = types.StringValue(attrs.GetName())

	cfg := attrs.GetConfig()
	outCfg := configModel{}

	if pt, ok := cfg.GetPipelineTypeOk(); ok {
		outCfg.PipelineType = types.StringValue(string(*pt))
	} else {
		// API doesn't return pipeline_type when it's "logs" (the default)
		// Set it explicitly to avoid state drift
		outCfg.PipelineType = types.StringValue("logs")
	}

	useLegacySearchSyntax := types.BoolNull()
	if v, ok := cfg.GetUseLegacySearchSyntaxOk(); ok {
		useLegacySearchSyntax = types.BoolValue(*v)
	}
	outCfg.UseLegacySearchSyntax = useLegacySearchSyntax

	for _, src := range cfg.GetSources() {
		sourceBlock := &sourceModel{}

		if a := flattenDatadogAgentSource(src.ObservabilityPipelineDatadogAgentSource); a != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineDatadogAgentSource.GetId())
			sourceBlock.DatadogAgentSource = append(sourceBlock.DatadogAgentSource, a)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if k := flattenKafkaSource(src.ObservabilityPipelineKafkaSource); k != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineKafkaSource.GetId())
			sourceBlock.KafkaSource = append(sourceBlock.KafkaSource, k)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if f := flattenFluentdSource(src.ObservabilityPipelineFluentdSource); f != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineFluentdSource.GetId())
			sourceBlock.FluentdSource = append(sourceBlock.FluentdSource, f)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if f := flattenFluentBitSource(src.ObservabilityPipelineFluentBitSource); f != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineFluentBitSource.GetId())
			sourceBlock.FluentBitSource = append(sourceBlock.FluentBitSource, f)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if s := flattenHttpServerSource(src.ObservabilityPipelineHttpServerSource); s != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineHttpServerSource.GetId())
			sourceBlock.HttpServerSource = append(sourceBlock.HttpServerSource, s)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if s := flattenSplunkHecSource(src.ObservabilityPipelineSplunkHecSource); s != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineSplunkHecSource.GetId())
			sourceBlock.SplunkHecSource = append(sourceBlock.SplunkHecSource, s)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if s := flattenSplunkTcpSource(src.ObservabilityPipelineSplunkTcpSource); s != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineSplunkTcpSource.GetId())
			sourceBlock.SplunkTcpSource = append(sourceBlock.SplunkTcpSource, s)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if s3 := flattenAmazonS3Source(src.ObservabilityPipelineAmazonS3Source); s3 != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineAmazonS3Source.GetId())
			sourceBlock.AmazonS3Source = append(sourceBlock.AmazonS3Source, s3)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if r := flattenRsyslogSource(src.ObservabilityPipelineRsyslogSource); r != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineRsyslogSource.GetId())
			sourceBlock.RsyslogSource = append(sourceBlock.RsyslogSource, r)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if s := flattenSyslogNgSource(src.ObservabilityPipelineSyslogNgSource); s != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineSyslogNgSource.GetId())
			sourceBlock.SyslogNgSource = append(sourceBlock.SyslogNgSource, s)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if s := flattenSumoLogicSource(src.ObservabilityPipelineSumoLogicSource); s != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineSumoLogicSource.GetId())
			sourceBlock.SumoLogicSource = append(sourceBlock.SumoLogicSource, s)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if f := flattenAmazonDataFirehoseSource(src.ObservabilityPipelineAmazonDataFirehoseSource); f != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineAmazonDataFirehoseSource.GetId())
			sourceBlock.AmazonDataFirehoseSource = append(sourceBlock.AmazonDataFirehoseSource, f)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if h := flattenHttpClientSource(src.ObservabilityPipelineHttpClientSource); h != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineHttpClientSource.GetId())
			sourceBlock.HttpClientSource = append(sourceBlock.HttpClientSource, h)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if g := flattenGooglePubSubSource(src.ObservabilityPipelineGooglePubSubSource); g != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineGooglePubSubSource.GetId())
			sourceBlock.GooglePubSubSource = append(sourceBlock.GooglePubSubSource, g)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if l := flattenLogstashSource(src.ObservabilityPipelineLogstashSource); l != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineLogstashSource.GetId())
			sourceBlock.LogstashSource = append(sourceBlock.LogstashSource, l)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if s := observability_pipeline.FlattenSocketSource(src.ObservabilityPipelineSocketSource); s != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineSocketSource.GetId())
			sourceBlock.SocketSource = append(sourceBlock.SocketSource, s)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		} else if o := observability_pipeline.FlattenOpentelemetrySource(src.ObservabilityPipelineOpentelemetrySource); o != nil {
			sourceBlock.Id = types.StringValue(src.ObservabilityPipelineOpentelemetrySource.GetId())
			sourceBlock.OpentelemetrySource = append(sourceBlock.OpentelemetrySource, o)
			outCfg.Sources = append(outCfg.Sources, sourceBlock)
		}
	}

	// Process processor groups - each group may contain one or more processors
	for _, group := range cfg.GetProcessorGroups() {
		flattenedGroup := flattenProcessorGroup(ctx, &group)
		if flattenedGroup != nil {
			outCfg.ProcessorGroups = append(outCfg.ProcessorGroups, flattenedGroup)
		}
	}

	for _, d := range cfg.GetDestinations() {
		destBlock := &destinationModel{}

		if logs := flattenDatadogLogsDestination(ctx, d.ObservabilityPipelineDatadogLogsDestination); logs != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineDatadogLogsDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineDatadogLogsDestination.GetInputs())
			destBlock.DatadogLogsDestination = append(destBlock.DatadogLogsDestination, logs)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if metrics := flattenDatadogMetricsDestination(ctx, d.ObservabilityPipelineDatadogMetricsDestination); metrics != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineDatadogMetricsDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineDatadogMetricsDestination.GetInputs())
			destBlock.DatadogMetricsDestination = append(destBlock.DatadogMetricsDestination, metrics)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if httpClient := flattenHttpClientDestination(ctx, d.ObservabilityPipelineHttpClientDestination); httpClient != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineHttpClientDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineHttpClientDestination.GetInputs())
			destBlock.HttpClientDestination = append(destBlock.HttpClientDestination, httpClient)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if chronicle := flattenGoogleSecopsDestination(ctx, d.ObservabilityPipelineGoogleChronicleDestination); chronicle != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineGoogleChronicleDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineGoogleChronicleDestination.GetInputs())
			destBlock.GoogleSecopsDestination = append(destBlock.GoogleSecopsDestination, chronicle)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if newrelic := flattenNewRelicDestination(ctx, d.ObservabilityPipelineNewRelicDestination); newrelic != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineNewRelicDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineNewRelicDestination.GetInputs())
			destBlock.NewRelicDestination = append(destBlock.NewRelicDestination, newrelic)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if sentinelone := flattenSentinelOneDestination(ctx, d.ObservabilityPipelineSentinelOneDestination); sentinelone != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineSentinelOneDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineSentinelOneDestination.GetInputs())
			destBlock.SentinelOneDestination = append(destBlock.SentinelOneDestination, sentinelone)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if hec := flattenSplunkHecDestination(ctx, d.ObservabilityPipelineSplunkHecDestination); hec != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineSplunkHecDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineSplunkHecDestination.GetInputs())
			destBlock.SplunkHecDestination = append(destBlock.SplunkHecDestination, hec)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if gcs := flattenGoogleCloudStorageDestination(ctx, d.ObservabilityPipelineGoogleCloudStorageDestination); gcs != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineGoogleCloudStorageDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineGoogleCloudStorageDestination.GetInputs())
			destBlock.GoogleCloudStorageDestination = append(destBlock.GoogleCloudStorageDestination, gcs)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if pubsub := flattenGooglePubSubDestination(ctx, d.ObservabilityPipelineGooglePubSubDestination); pubsub != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineGooglePubSubDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineGooglePubSubDestination.GetInputs())
			destBlock.GooglePubSubDestination = append(destBlock.GooglePubSubDestination, pubsub)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if s := flattenSumoLogicDestination(ctx, d.ObservabilityPipelineSumoLogicDestination); s != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineSumoLogicDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineSumoLogicDestination.GetInputs())
			destBlock.SumoLogicDestination = append(destBlock.SumoLogicDestination, s)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if r := flattenRsyslogDestination(ctx, d.ObservabilityPipelineRsyslogDestination); r != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineRsyslogDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineRsyslogDestination.GetInputs())
			destBlock.RsyslogDestination = append(destBlock.RsyslogDestination, r)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if s := flattenSyslogNgDestination(ctx, d.ObservabilityPipelineSyslogNgDestination); s != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineSyslogNgDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineSyslogNgDestination.GetInputs())
			destBlock.SyslogNgDestination = append(destBlock.SyslogNgDestination, s)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if e := flattenElasticsearchDestination(ctx, d.ObservabilityPipelineElasticsearchDestination); e != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineElasticsearchDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineElasticsearchDestination.GetInputs())
			destBlock.ElasticsearchDestination = append(destBlock.ElasticsearchDestination, e)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if a := flattenAzureStorageDestination(ctx, d.AzureStorageDestination); a != nil {
			destBlock.Id = types.StringValue(d.AzureStorageDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.AzureStorageDestination.GetInputs())
			destBlock.AzureStorageDestination = append(destBlock.AzureStorageDestination, a)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if m := flattenMicrosoftSentinelDestination(ctx, d.MicrosoftSentinelDestination); m != nil {
			destBlock.Id = types.StringValue(d.MicrosoftSentinelDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.MicrosoftSentinelDestination.GetInputs())
			destBlock.MicrosoftSentinelDestination = append(destBlock.MicrosoftSentinelDestination, m)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if opensearch := flattenOpenSearchDestination(ctx, d.ObservabilityPipelineOpenSearchDestination); opensearch != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineOpenSearchDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineOpenSearchDestination.GetInputs())
			destBlock.OpenSearchDestination = append(destBlock.OpenSearchDestination, opensearch)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if amazonopensearch := flattenAmazonOpenSearchDestination(d.ObservabilityPipelineAmazonOpenSearchDestination); amazonopensearch != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineAmazonOpenSearchDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineAmazonOpenSearchDestination.GetInputs())
			destBlock.AmazonOpenSearchDestination = append(destBlock.AmazonOpenSearchDestination, amazonopensearch)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if socket := observability_pipeline.FlattenSocketDestination(ctx, d.ObservabilityPipelineSocketDestination); socket != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineSocketDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineSocketDestination.GetInputs())
			destBlock.SocketDestination = append(destBlock.SocketDestination, socket)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if s3 := observability_pipeline.FlattenAmazonS3Destination(ctx, d.ObservabilityPipelineAmazonS3Destination); s3 != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineAmazonS3Destination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineAmazonS3Destination.GetInputs())
			destBlock.AmazonS3Destination = append(destBlock.AmazonS3Destination, s3)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if securitylake := observability_pipeline.FlattenObservabilityPipelinesAmazonSecurityLakeDestination(ctx, d.ObservabilityPipelineAmazonSecurityLakeDestination); securitylake != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineAmazonSecurityLakeDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineAmazonSecurityLakeDestination.GetInputs())
			destBlock.AmazonSecurityLakeDestination = append(destBlock.AmazonSecurityLakeDestination, securitylake)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if crowdstrike := observability_pipeline.FlattenCrowdStrikeNextGenSiemDestination(ctx, d.ObservabilityPipelineCrowdStrikeNextGenSiemDestination); crowdstrike != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineCrowdStrikeNextGenSiemDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineCrowdStrikeNextGenSiemDestination.GetInputs())
			destBlock.CrowdStrikeNextGenSiemDestination = append(destBlock.CrowdStrikeNextGenSiemDestination, crowdstrike)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if cloudprem := observability_pipeline.FlattenCloudPremDestination(ctx, d.ObservabilityPipelineCloudPremDestination); cloudprem != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineCloudPremDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineCloudPremDestination.GetInputs())
			destBlock.CloudPremDestination = append(destBlock.CloudPremDestination, cloudprem)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		} else if kafka := observability_pipeline.FlattenKafkaDestination(ctx, d.ObservabilityPipelineKafkaDestination); kafka != nil {
			destBlock.Id = types.StringValue(d.ObservabilityPipelineKafkaDestination.GetId())
			destBlock.Inputs, _ = types.ListValueFrom(ctx, types.StringType, d.ObservabilityPipelineKafkaDestination.GetInputs())
			destBlock.KafkaDestination = append(destBlock.KafkaDestination, kafka)
			outCfg.Destinations = append(outCfg.Destinations, destBlock)
		}
	}

	state.Config = []configModel{outCfg}
}

// ---------- Sources ----------

func flattenDatadogAgentSource(src *datadogV2.ObservabilityPipelineDatadogAgentSource) *datadogAgentSourceModel {
	if src == nil {
		return nil
	}
	out := &datadogAgentSourceModel{}
	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}
	return out
}

func expandDatadogAgentSource(src *datadogAgentSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	agent := datadogV2.NewObservabilityPipelineDatadogAgentSourceWithDefaults()
	agent.SetId(id)
	agent.Tls = observability_pipeline.ExpandTls(src.Tls)
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineDatadogAgentSource: agent,
	}
}

func flattenKafkaSource(src *datadogV2.ObservabilityPipelineKafkaSource) *kafkaSourceModel {
	if src == nil {
		return nil
	}
	out := &kafkaSourceModel{
		GroupId: types.StringValue(src.GetGroupId()),
	}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	// Topics is required by the API (always present, even if empty)
	// Initialize as empty slice to preserve [] vs null distinction
	topics := []types.String{}
	for _, topic := range src.GetTopics() {
		topics = append(topics, types.StringValue(topic))
	}
	out.Topics = topics
	if sasl, ok := src.GetSaslOk(); ok {
		out.Sasl = []kafkaSourceSaslModel{
			{
				Mechanism: types.StringValue(string(sasl.GetMechanism())),
			},
		}
	}
	for _, opt := range src.GetLibrdkafkaOptions() {
		out.LibrdkafkaOptions = append(out.LibrdkafkaOptions, librdkafkaOptionModel{
			Name:  types.StringValue(opt.Name),
			Value: types.StringValue(opt.Value),
		})
	}
	return out
}

func expandKafkaSource(src *kafkaSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	source := datadogV2.NewObservabilityPipelineKafkaSourceWithDefaults()
	source.SetId(id)
	source.SetGroupId(src.GroupId.ValueString())
	// Initialize as empty slice, not nil, to ensure it serializes as [] not null
	topics := []string{}
	for _, t := range src.Topics {
		topics = append(topics, t.ValueString())
	}
	source.SetTopics(topics)

	source.Tls = observability_pipeline.ExpandTls(src.Tls)

	if len(src.Sasl) > 0 {
		sasl := src.Sasl[0]
		mechanism, _ := datadogV2.NewObservabilityPipelineKafkaSaslMechanismFromValue(sasl.Mechanism.ValueString())
		if mechanism != nil {
			saslConfig := datadogV2.ObservabilityPipelineKafkaSasl{}
			saslConfig.SetMechanism(*mechanism)
			source.SetSasl(saslConfig)
		}
	}

	if len(src.LibrdkafkaOptions) > 0 {
		opts := []datadogV2.ObservabilityPipelineKafkaLibrdkafkaOption{}
		for _, opt := range src.LibrdkafkaOptions {
			opts = append(opts, datadogV2.ObservabilityPipelineKafkaLibrdkafkaOption{
				Name:  opt.Name.ValueString(),
				Value: opt.Value.ValueString(),
			})
		}
		source.SetLibrdkafkaOptions(opts)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineKafkaSource: source,
	}
}

// ---------- Processors ----------

// createProcessorModel creates a processorModel with common fields populated
// This function could be removed once we move `processorModel` to `processor_common.go`
// and split all processor types into their own files.
func createProcessorModel(proc observability_pipeline.BaseProcessor) *processorModel {
	displayName, _ := proc.GetDisplayNameOk()
	return &processorModel{
		Id:          types.StringValue(proc.GetId()),
		Enabled:     types.BoolValue(proc.GetEnabled()),
		Include:     types.StringValue(proc.GetInclude()),
		DisplayName: types.StringPointerValue(displayName),
	}
}

func flattenFilterProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineFilterProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	// Filter processor has no processor-specific fields, only common fields
	model.FilterProcessor = append(model.FilterProcessor, &filterProcessorModel{})
	return model
}

// flattenProcessorGroup converts a processor group from API model to Terraform model
func flattenProcessorGroup(ctx context.Context, group *datadogV2.ObservabilityPipelineConfigProcessorGroup) *processorGroupModel {
	if group == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, group.GetInputs())

	var processorsList []*processorModel
	processors := group.GetProcessors()
	for _, p := range processors {
		var procModel *processorModel

		if p.ObservabilityPipelineFilterProcessor != nil {
			procModel = flattenFilterProcessor(ctx, p.ObservabilityPipelineFilterProcessor)
		} else if p.ObservabilityPipelineParseJSONProcessor != nil {
			procModel = flattenParseJsonProcessor(ctx, p.ObservabilityPipelineParseJSONProcessor)
		} else if p.ObservabilityPipelineAddFieldsProcessor != nil {
			procModel = flattenAddFieldsProcessor(ctx, p.ObservabilityPipelineAddFieldsProcessor)
		} else if p.ObservabilityPipelineRenameFieldsProcessor != nil {
			procModel = flattenRenameFieldsProcessor(ctx, p.ObservabilityPipelineRenameFieldsProcessor)
		} else if p.ObservabilityPipelineRemoveFieldsProcessor != nil {
			procModel = flattenRemoveFieldsProcessor(ctx, p.ObservabilityPipelineRemoveFieldsProcessor)
		} else if p.ObservabilityPipelineQuotaProcessor != nil {
			procModel = flattenQuotaProcessor(ctx, p.ObservabilityPipelineQuotaProcessor)
		} else if p.ObservabilityPipelineSensitiveDataScannerProcessor != nil {
			procModel = flattenSensitiveDataScannerProcessor(ctx, p.ObservabilityPipelineSensitiveDataScannerProcessor)
		} else if p.ObservabilityPipelineGenerateMetricsProcessor != nil {
			procModel = flattenGenerateDatadogMetricsProcessor(ctx, p.ObservabilityPipelineGenerateMetricsProcessor)
		} else if p.ObservabilityPipelineParseGrokProcessor != nil {
			procModel = flattenParseGrokProcessor(ctx, p.ObservabilityPipelineParseGrokProcessor)
		} else if p.ObservabilityPipelineSampleProcessor != nil {
			procModel = flattenSampleProcessor(ctx, p.ObservabilityPipelineSampleProcessor)
		} else if p.ObservabilityPipelineDedupeProcessor != nil {
			procModel = flattenDedupeProcessor(ctx, p.ObservabilityPipelineDedupeProcessor)
		} else if p.ObservabilityPipelineReduceProcessor != nil {
			procModel = flattenReduceProcessor(ctx, p.ObservabilityPipelineReduceProcessor)
		} else if p.ObservabilityPipelineThrottleProcessor != nil {
			procModel = flattenThrottleProcessor(ctx, p.ObservabilityPipelineThrottleProcessor)
		} else if p.ObservabilityPipelineAddEnvVarsProcessor != nil {
			procModel = flattenAddEnvVarsProcessor(ctx, p.ObservabilityPipelineAddEnvVarsProcessor)
		} else if p.ObservabilityPipelineEnrichmentTableProcessor != nil {
			procModel = flattenEnrichmentTableProcessor(ctx, p.ObservabilityPipelineEnrichmentTableProcessor)
		} else if p.ObservabilityPipelineOcsfMapperProcessor != nil {
			procModel = flattenOcsfMapperProcessor(ctx, p.ObservabilityPipelineOcsfMapperProcessor)
		} else if p.ObservabilityPipelineDatadogTagsProcessor != nil {
			procModel = flattenDatadogTagsProcessor(ctx, p.ObservabilityPipelineDatadogTagsProcessor)
		} else if p.ObservabilityPipelineCustomProcessor != nil {
			procModel = flattenCustomProcessor(ctx, p.ObservabilityPipelineCustomProcessor)
		} else if p.ObservabilityPipelineAddHostnameProcessor != nil {
			procModel = flattenAddHostnameProcessor(ctx, p.ObservabilityPipelineAddHostnameProcessor)
		} else if p.ObservabilityPipelineParseXMLProcessor != nil {
			procModel = flattenParseXMLProcessor(ctx, p.ObservabilityPipelineParseXMLProcessor)
		} else if p.ObservabilityPipelineSplitArrayProcessor != nil {
			procModel = flattenSplitArrayProcessor(ctx, p.ObservabilityPipelineSplitArrayProcessor)
		} else if p.ObservabilityPipelineMetricTagsProcessor != nil {
			procModel = flattenMetricTagsProcessor(ctx, p.ObservabilityPipelineMetricTagsProcessor)
		}

		if procModel != nil {
			processorsList = append(processorsList, procModel)
		}
	}

	out := &processorGroupModel{
		Id:         types.StringValue(group.GetId()),
		Enabled:    types.BoolValue(group.GetEnabled()),
		Include:    types.StringValue(group.GetInclude()),
		Inputs:     inputs,
		Processors: processorsList,
	}
	if group.DisplayName != nil {
		out.DisplayName = types.StringValue(group.GetDisplayName())
	}
	return out
}

// expandProcessorGroup converts a processor group from Terraform model to API model
func expandProcessorGroup(ctx context.Context, group *processorGroupModel) datadogV2.ObservabilityPipelineConfigProcessorGroup {
	apiGroup := datadogV2.NewObservabilityPipelineConfigProcessorGroupWithDefaults()

	// Set group-level fields
	apiGroup.SetId(group.Id.ValueString())
	apiGroup.SetEnabled(group.Enabled.ValueBool())
	apiGroup.SetInclude(group.Include.ValueString())
	if !group.DisplayName.IsNull() {
		apiGroup.SetDisplayName(group.DisplayName.ValueString())
	}
	var inputs []string
	group.Inputs.ElementsAs(ctx, &inputs, false)
	apiGroup.SetInputs(inputs)

	// Process the nested processors and get all items
	var processorItems []datadogV2.ObservabilityPipelineConfigProcessorItem
	for _, processor := range group.Processors {
		items := expandProcessorTypes(ctx, processor)
		processorItems = append(processorItems, items...)
	}
	if len(processorItems) > 0 {
		apiGroup.SetProcessors(processorItems)
	}

	return *apiGroup
}

// expandProcessorTypes converts the processor types model to a list of processor items
// Uses the processor-level id, enabled, include, and display_name for all processors in the group
func expandProcessorTypes(ctx context.Context, processor *processorModel) []datadogV2.ObservabilityPipelineConfigProcessorItem {
	var items []datadogV2.ObservabilityPipelineConfigProcessorItem

	// Create common fields struct to pass to all expand functions
	var displayName *string
	if !processor.DisplayName.IsNull() {
		dn := processor.DisplayName.ValueString()
		displayName = &dn
	}
	common := observability_pipeline.BaseProcessorFields{
		Id:          processor.Id.ValueString(),
		Enabled:     processor.Enabled.ValueBool(),
		Include:     processor.Include.ValueString(),
		DisplayName: displayName,
	}

	// Check each processor type and expand if present
	for _, p := range processor.FilterProcessor {
		items = append(items, expandFilterProcessorItem(ctx, common, p))
	}
	for _, p := range processor.ParseJsonProcessor {
		items = append(items, expandParseJsonProcessorItem(ctx, common, p))
	}
	for _, p := range processor.AddFieldsProcessor {
		items = append(items, expandAddFieldsProcessorItem(ctx, common, p))
	}
	for _, p := range processor.RenameFieldsProcessor {
		items = append(items, expandRenameFieldsProcessorItem(ctx, common, p))
	}
	for _, p := range processor.RemoveFieldsProcessor {
		items = append(items, expandRemoveFieldsProcessorItem(ctx, common, p))
	}
	for _, p := range processor.QuotaProcessor {
		items = append(items, expandQuotaProcessorItem(ctx, common, p))
	}
	for _, p := range processor.DedupeProcessor {
		items = append(items, expandDedupeProcessorItem(ctx, common, p))
	}
	for _, p := range processor.ReduceProcessor {
		items = append(items, expandReduceProcessorItem(ctx, common, p))
	}
	for _, p := range processor.ThrottleProcessor {
		items = append(items, expandThrottleProcessorItem(ctx, common, p))
	}
	for _, p := range processor.AddEnvVarsProcessor {
		items = append(items, expandAddEnvVarsProcessorItem(ctx, common, p))
	}
	for _, p := range processor.EnrichmentTableProcessor {
		items = append(items, expandEnrichmentTableProcessorItem(ctx, common, p))
	}
	for _, p := range processor.OcsfMapperProcessor {
		items = append(items, expandOcsfMapperProcessorItem(ctx, common, p))
	}
	for _, p := range processor.ParseGrokProcessor {
		items = append(items, expandParseGrokProcessorItem(ctx, common, p))
	}
	for _, p := range processor.SampleProcessor {
		items = append(items, expandSampleProcessorItem(ctx, common, p))
	}
	for _, p := range processor.GenerateMetricsProcessor {
		items = append(items, expandGenerateMetricsProcessorItem(ctx, common, p))
	}
	for _, p := range processor.SensitiveDataScannerProcessor {
		items = append(items, expandSensitiveDataScannerProcessorItem(ctx, common, p))
	}
	for _, p := range processor.CustomProcessor {
		items = append(items, observability_pipeline.ExpandCustomProcessor(common, p))
	}
	for _, p := range processor.DatadogTagsProcessor {
		items = append(items, observability_pipeline.ExpandDatadogTagsProcessor(common, p))
	}
	for _, p := range processor.AddHostnameProcessor {
		items = append(items, expandAddHostnameProcessorItem(ctx, common, p))
	}
	for _, p := range processor.ParseXMLProcessor {
		items = append(items, expandParseXMLProcessorItem(ctx, common, p))
	}
	for _, p := range processor.SplitArrayProcessor {
		items = append(items, expandSplitArrayProcessorItem(ctx, common, p))
	}
	for _, p := range processor.MetricTagsProcessor {
		items = append(items, expandMetricTagsProcessorItem(ctx, common, p))
	}

	return items
}

func expandFilterProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *filterProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineFilterProcessorWithDefaults()
	common.ApplyTo(proc)

	return datadogV2.ObservabilityPipelineFilterProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func flattenParseJsonProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineParseJSONProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	model.ParseJsonProcessor = append(model.ParseJsonProcessor, &parseJsonProcessorModel{
		Field: types.StringValue(src.Field),
	})
	return model
}

func expandParseJsonProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *parseJsonProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineParseJSONProcessorWithDefaults()
	common.ApplyTo(proc)
	proc.SetField(src.Field.ValueString())

	return datadogV2.ObservabilityPipelineParseJSONProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func flattenAddFieldsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineAddFieldsProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	addFields := &addFieldsProcessor{}
	for _, f := range src.Fields {
		addFields.Fields = append(addFields.Fields, fieldValue{
			Name:  types.StringValue(f.Name),
			Value: types.StringValue(f.Value),
		})
	}
	model.AddFieldsProcessor = append(model.AddFieldsProcessor, addFields)
	return model
}

func flattenRenameFieldsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineRenameFieldsProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	renameFields := &renameFieldsProcessorModel{}
	for _, f := range src.Fields {
		renameFields.Fields = append(renameFields.Fields, renameFieldItemModel{
			Source:         types.StringValue(f.Source),
			Destination:    types.StringValue(f.Destination),
			PreserveSource: types.BoolValue(f.PreserveSource),
		})
	}
	model.RenameFieldsProcessor = append(model.RenameFieldsProcessor, renameFields)
	return model
}

func flattenRemoveFieldsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineRemoveFieldsProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	// Use nil slice for optional fields - only populate if non-empty to preserve null in state
	var fields []types.String
	for _, f := range src.Fields {
		fields = append(fields, types.StringValue(f))
	}
	fieldList, _ := types.ListValueFrom(ctx, types.StringType, fields)
	model.RemoveFieldsProcessor = append(model.RemoveFieldsProcessor, &removeFieldsProcessorModel{
		Fields: fieldList,
	})
	return model
}

func flattenQuotaProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineQuotaProcessor) *processorModel {
	if src == nil {
		return nil
	}

	model := createProcessorModel(src)

	limit := src.GetLimit()
	// PartitionFields is optional - only populate if present to distinguish null from []
	var partitionFields []types.String
	if pf, ok := src.GetPartitionFieldsOk(); ok {
		partitionFields = []types.String{}
		for _, p := range *pf {
			partitionFields = append(partitionFields, types.StringValue(p))
		}
	}

	quota := &quotaProcessorModel{
		Name: types.StringValue(src.GetName()),
		Limit: []quotaLimitModel{
			{
				Enforce: types.StringValue(string(limit.GetEnforce())),
				Limit:   types.Int64Value(limit.GetLimit()),
			},
		},
		PartitionFields: partitionFields,
	}

	if dropEvents, ok := src.GetDropEventsOk(); ok && dropEvents != nil {
		quota.DropEvents = types.BoolPointerValue(dropEvents)
	}

	if ignoreMissing, ok := src.GetIgnoreWhenMissingPartitionsOk(); ok {
		quota.IgnoreWhenMissingPartitions = types.BoolPointerValue(ignoreMissing)
	}

	if overflowAction, ok := src.GetOverflowActionOk(); ok {
		quota.OverflowAction = types.StringValue(string(*overflowAction))
	}

	if tooManyBucketsAction, ok := src.GetTooManyBucketsActionOk(); ok {
		quota.TooManyBucketsAction = types.StringValue(string(*tooManyBucketsAction))
	}

	for _, o := range src.GetOverrides() {
		override := quotaOverrideModel{
			Limit: []quotaLimitModel{
				{
					Enforce: types.StringValue(string(o.Limit.GetEnforce())),
					Limit:   types.Int64Value(o.Limit.GetLimit()),
				},
			},
		}
		for _, f := range o.GetFields() {
			override.Fields = append(override.Fields, fieldValue{
				Name:  types.StringValue(f.Name),
				Value: types.StringValue(f.Value),
			})
		}
		quota.Overrides = append(quota.Overrides, override)
	}

	model.QuotaProcessor = append(model.QuotaProcessor, quota)
	return model
}

func flattenSensitiveDataScannerProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineSensitiveDataScannerProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	scanner := &sensitiveDataScannerProcessorModel{}
	for _, rule := range src.GetRules() {
		r := sensitiveDataScannerProcessorRule{
			Name: types.StringValue(rule.GetName()),
		}
		// Tags is optional - only populate if present to distinguish null from []
		if tags, ok := rule.GetTagsOk(); ok {
			tagsList := []types.String{}
			for _, t := range *tags {
				tagsList = append(tagsList, types.StringValue(t))
			}
			r.Tags = tagsList
		}

		if ko := rule.KeywordOptions; ko != nil {
			// Keywords is required by the API (always present, even if empty)
			// Initialize as empty slice to preserve [] vs null distinction
			keywords := []types.String{}
			for _, k := range ko.GetKeywords() {
				keywords = append(keywords, types.StringValue(k))
			}
			r.KeywordOptions = []sensitiveDataScannerProcessorKeywordOptions{
				{
					Keywords:  keywords,
					Proximity: types.Int64Value(ko.GetProximity()),
				},
			}
		}

		// Flatten Pattern
		if pattern, ok := rule.GetPatternOk(); ok {
			outPattern := sensitiveDataScannerProcessorPattern{}
			if pattern.ObservabilityPipelineSensitiveDataScannerProcessorCustomPattern != nil {
				options := pattern.ObservabilityPipelineSensitiveDataScannerProcessorCustomPattern.GetOptions()
				outPattern.Custom = []sensitiveDataScannerCustomPattern{
					{
						Rule: types.StringValue(options.GetRule()),
					},
				}
				if desc, ok := options.GetDescriptionOk(); ok {
					outPattern.Custom[0].Description = types.StringPointerValue(desc)
				}
			}
			if pattern.ObservabilityPipelineSensitiveDataScannerProcessorLibraryPattern != nil {
				options := pattern.ObservabilityPipelineSensitiveDataScannerProcessorLibraryPattern.GetOptions()
				outPattern.Library = []sensitiveDataScannerLibraryPattern{
					{
						Id: types.StringValue(options.GetId()),
					},
				}
				if desc, ok := options.GetDescriptionOk(); ok {
					outPattern.Library[0].Description = types.StringPointerValue(desc)
				}
				if useKw, ok := options.GetUseRecommendedKeywordsOk(); ok {
					outPattern.Library[0].UseRecommendedKeywords = types.BoolPointerValue(useKw)
				}
			}
			r.Pattern = append(r.Pattern, outPattern)
		}
		// Flatten Scope
		scope := rule.GetScope()
		outScope := sensitiveDataScannerProcessorScope{}
		if scope.ObservabilityPipelineSensitiveDataScannerProcessorScopeInclude != nil {
			options := scope.ObservabilityPipelineSensitiveDataScannerProcessorScopeInclude.GetOptions()
			// Fields is required by the API (always present, even if empty)
			// Initialize as empty slice to preserve [] vs null distinction
			fields := []types.String{}
			for _, f := range options.GetFields() {
				fields = append(fields, types.StringValue(f))
			}
			outScope.Include = []sensitiveDataScannerScopeOptions{
				{
					Fields: fields,
				},
			}
		}
		if scope.ObservabilityPipelineSensitiveDataScannerProcessorScopeExclude != nil {
			options := scope.ObservabilityPipelineSensitiveDataScannerProcessorScopeExclude.GetOptions()
			// Fields is required by the API (always present, even if empty)
			// Initialize as empty slice to preserve [] vs null distinction
			fields := []types.String{}
			for _, f := range options.GetFields() {
				fields = append(fields, types.StringValue(f))
			}
			outScope.Exclude = []sensitiveDataScannerScopeOptions{
				{
					Fields: fields,
				},
			}
		}
		if scope.ObservabilityPipelineSensitiveDataScannerProcessorScopeAll != nil {
			all := true
			outScope.All = &all
		}
		r.Scope = append(r.Scope, outScope)

		// Flatten OnMatch
		onMatch := rule.GetOnMatch()
		outOnMatch := sensitiveDataScannerProcessorAction{}
		if onMatch.ObservabilityPipelineSensitiveDataScannerProcessorActionRedact != nil {
			options := onMatch.ObservabilityPipelineSensitiveDataScannerProcessorActionRedact.GetOptions()
			outOnMatch.Redact = []sensitiveDataScannerRedactAction{
				{
					Replace: types.StringValue(options.GetReplace()),
				},
			}
		}
		if onMatch.ObservabilityPipelineSensitiveDataScannerProcessorActionHash != nil {
			outOnMatch.Hash = []sensitiveDataScannerHashAction{
				{},
			}
		}
		if onMatch.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedact != nil {
			options := onMatch.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedact.GetOptions()
			outOnMatch.PartialRedact = []sensitiveDataScannerPartialRedactAction{
				{
					Characters: types.Int64Value(options.GetCharacters()),
					Direction:  types.StringValue(string(options.GetDirection())),
				},
			}
		}
		r.OnMatch = append(r.OnMatch, outOnMatch)

		scanner.Rules = append(scanner.Rules, r)
	}
	model.SensitiveDataScannerProcessor = append(model.SensitiveDataScannerProcessor, scanner)
	return model
}

func flattenGenerateDatadogMetricsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineGenerateMetricsProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	genMetrics := &generateMetricsProcessorModel{}
	for _, metric := range src.GetMetrics() {
		groupByList, _ := types.ListValueFrom(ctx, types.StringType, metric.GetGroupBy())
		m := generatedMetricModel{
			Name:       types.StringValue(metric.GetName()),
			Include:    types.StringValue(metric.GetInclude()),
			MetricType: types.StringValue(string(metric.GetMetricType())),
			GroupBy:    groupByList,
		}
		// Handle value
		if metric.Value.ObservabilityPipelineGeneratedMetricIncrementByOne != nil {
			m.Value = []generatedMetricValue{
				{
					Strategy: types.StringValue("increment_by_one"),
				},
			}
		} else if metric.Value.ObservabilityPipelineGeneratedMetricIncrementByField != nil {
			m.Value = []generatedMetricValue{
				{
					Strategy: types.StringValue("increment_by_field"),
					Field:    types.StringValue(metric.Value.ObservabilityPipelineGeneratedMetricIncrementByField.GetField()),
				},
			}
		}
		genMetrics.Metrics = append(genMetrics.Metrics, m)
	}
	model.GenerateMetricsProcessor = append(model.GenerateMetricsProcessor, genMetrics)
	return model
}

func flattenParseGrokProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineParseGrokProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	grok := &parseGrokProcessorModel{
		DisableLibraryRules: types.BoolValue(src.GetDisableLibraryRules()),
	}
	for _, rule := range src.GetRules() {
		r := parseGrokProcessorRuleModel{
			Source: types.StringValue(rule.GetSource()),
		}
		for _, m := range rule.GetMatchRules() {
			r.MatchRules = append(r.MatchRules, grokRuleModel{
				Name: types.StringValue(m.GetName()),
				Rule: types.StringValue(m.GetRule()),
			})
		}
		for _, s := range rule.GetSupportRules() {
			r.SupportRules = append(r.SupportRules, grokRuleModel{
				Name: types.StringValue(s.GetName()),
				Rule: types.StringValue(s.GetRule()),
			})
		}
		grok.Rules = append(grok.Rules, r)
	}
	model.ParseGrokProcessor = append(model.ParseGrokProcessor, grok)
	return model
}

func flattenSampleProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineSampleProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	sample := &sampleProcessorModel{}
	if percentage, ok := src.GetPercentageOk(); ok {
		sample.Percentage = types.Float64PointerValue(percentage)
	}
	// Use nil slice for optional fields - only populate if non-empty to preserve null in state
	var groupBy []types.String
	for _, g := range src.GetGroupBy() {
		groupBy = append(groupBy, types.StringValue(g))
	}
	sample.GroupBy = groupBy
	model.SampleProcessor = append(model.SampleProcessor, sample)
	return model
}

func flattenDedupeProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineDedupeProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	// Fields is required by the API (always present, even if empty)
	// Initialize as empty slice to preserve [] vs null distinction
	fields := []types.String{}
	for _, f := range src.GetFields() {
		fields = append(fields, types.StringValue(f))
	}
	model.DedupeProcessor = append(model.DedupeProcessor, &dedupeProcessorModel{
		Fields: fields,
		Mode:   types.StringValue(string(src.GetMode())),
	})
	return model
}

func flattenReduceProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineReduceProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	// GroupBy is required by the API (always present, even if empty)
	// Initialize as empty slice to preserve [] vs null distinction
	groupBy := []types.String{}
	for _, g := range src.GetGroupBy() {
		groupBy = append(groupBy, types.StringValue(g))
	}

	reduce := &reduceProcessorModel{
		GroupBy: groupBy,
	}
	for _, strategy := range src.GetMergeStrategies() {
		reduce.MergeStrategies = append(reduce.MergeStrategies, mergeStrategyModel{
			Path:     types.StringValue(strategy.GetPath()),
			Strategy: types.StringValue(string(strategy.GetStrategy())),
		})
	}
	model.ReduceProcessor = append(model.ReduceProcessor, reduce)
	return model
}

func flattenThrottleProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineThrottleProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	// Use nil slice for optional fields - only populate if non-empty to preserve null in state
	var groupBy []types.String
	for _, g := range src.GetGroupBy() {
		groupBy = append(groupBy, types.StringValue(g))
	}
	model.ThrottleProcessor = append(model.ThrottleProcessor, &throttleProcessorModel{
		Threshold: types.Int64Value(src.GetThreshold()),
		Window:    types.Float64Value(src.GetWindow()),
		GroupBy:   groupBy,
	})
	return model
}

func flattenAddEnvVarsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineAddEnvVarsProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	envVars := &addEnvVarsProcessorModel{}
	for _, v := range src.GetVariables() {
		envVars.Variables = append(envVars.Variables, envVarMappingModel{
			Field: types.StringValue(v.GetField()),
			Name:  types.StringValue(v.GetName()),
		})
	}
	model.AddEnvVarsProcessor = append(model.AddEnvVarsProcessor, envVars)
	return model
}

func flattenEnrichmentTableProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineEnrichmentTableProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	enrichment := &enrichmentTableProcessorModel{
		Target: types.StringValue(src.GetTarget()),
	}
	if src.File != nil {
		enrichment.File = []enrichmentFileModel{
			{
				Path: types.StringValue(src.File.GetPath()),
				Encoding: []fileEncodingModel{
					{
						Type:            types.StringValue(string(src.File.Encoding.GetType())),
						Delimiter:       types.StringValue(src.File.Encoding.GetDelimiter()),
						IncludesHeaders: types.BoolValue(src.File.Encoding.GetIncludesHeaders()),
					},
				},
			},
		}
		for _, k := range src.File.GetKey() {
			enrichment.File[0].Key = append(enrichment.File[0].Key, fileKeyItemModel{
				Column:     types.StringValue(k.GetColumn()),
				Comparison: types.StringValue(string(k.GetComparison())),
				Field:      types.StringValue(k.GetField()),
			})
		}
	}
	if src.Geoip != nil {
		enrichment.GeoIp = []enrichmentGeoIpModel{
			{
				KeyField: types.StringValue(src.Geoip.GetKeyField()),
				Locale:   types.StringValue(src.Geoip.GetLocale()),
				Path:     types.StringValue(src.Geoip.GetPath()),
			},
		}
	}
	if src.ReferenceTable != nil {
		refTableModel := enrichmentReferenceTableModel{
			KeyField: types.StringValue(src.ReferenceTable.GetKeyField()),
			TableId:  types.StringValue(src.ReferenceTable.GetTableId()),
		}
		if len(src.ReferenceTable.GetColumns()) > 0 {
			columnsList, _ := types.ListValueFrom(ctx, types.StringType, src.ReferenceTable.GetColumns())
			refTableModel.Columns = columnsList
		} else {
			refTableModel.Columns = types.ListNull(types.StringType)
		}
		enrichment.ReferenceTable = []enrichmentReferenceTableModel{refTableModel}
	}
	model.EnrichmentTableProcessor = append(model.EnrichmentTableProcessor, enrichment)
	return model
}

func flattenOcsfMapperProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineOcsfMapperProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	ocsf := &ocsfMapperProcessorModel{}
	for _, mapping := range src.GetMappings() {
		m := ocsfMappingModel{
			Include: types.StringValue(mapping.GetInclude()),
		}
		if mapping.Mapping.ObservabilityPipelineOcsfMappingLibrary != nil {
			m.LibraryMapping = types.StringValue(string(*mapping.Mapping.ObservabilityPipelineOcsfMappingLibrary))
		}
		ocsf.Mapping = append(ocsf.Mapping, m)
	}
	model.OcsfMapperProcessor = append(model.OcsfMapperProcessor, ocsf)
	return model
}

func flattenDatadogTagsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineDatadogTagsProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	if f := observability_pipeline.FlattenDatadogTagsProcessor(src); f != nil {
		model.DatadogTagsProcessor = append(model.DatadogTagsProcessor, f)
	}
	return model
}

func flattenCustomProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineCustomProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	if f := observability_pipeline.FlattenCustomProcessor(src); f != nil {
		model.CustomProcessor = append(model.CustomProcessor, f)
	}
	return model
}

func expandAddFieldsProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *addFieldsProcessor) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineAddFieldsProcessorWithDefaults()
	common.ApplyTo(proc)

	var fields []datadogV2.ObservabilityPipelineFieldValue
	for _, f := range src.Fields {
		fields = append(fields, datadogV2.ObservabilityPipelineFieldValue{
			Name:  f.Name.ValueString(),
			Value: f.Value.ValueString(),
		})
	}
	proc.SetFields(fields)

	return datadogV2.ObservabilityPipelineAddFieldsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandRenameFieldsProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *renameFieldsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineRenameFieldsProcessorWithDefaults()
	common.ApplyTo(proc)

	var fields []datadogV2.ObservabilityPipelineRenameFieldsProcessorField
	for _, f := range src.Fields {
		fields = append(fields, datadogV2.ObservabilityPipelineRenameFieldsProcessorField{
			Source:         f.Source.ValueString(),
			Destination:    f.Destination.ValueString(),
			PreserveSource: f.PreserveSource.ValueBool(),
		})
	}
	proc.SetFields(fields)

	return datadogV2.ObservabilityPipelineRenameFieldsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandRemoveFieldsProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *removeFieldsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineRemoveFieldsProcessorWithDefaults()
	common.ApplyTo(proc)

	var fields []string
	src.Fields.ElementsAs(ctx, &fields, false)
	proc.SetFields(fields)

	return datadogV2.ObservabilityPipelineRemoveFieldsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandQuotaProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *quotaProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineQuotaProcessorWithDefaults()
	common.ApplyTo(proc)
	proc.SetName(src.Name.ValueString())

	if !src.DropEvents.IsNull() {
		proc.SetDropEvents(src.DropEvents.ValueBool())
	}

	if !src.IgnoreWhenMissingPartitions.IsNull() {
		proc.SetIgnoreWhenMissingPartitions(src.IgnoreWhenMissingPartitions.ValueBool())
	}

	// Only set partition_fields if user specified them in config (to distinguish null from [])
	if src.PartitionFields != nil {
		partitions := []string{}
		for _, p := range src.PartitionFields {
			partitions = append(partitions, p.ValueString())
		}
		proc.SetPartitionFields(partitions)
	}

	if len(src.Limit) > 0 {
		proc.SetLimit(datadogV2.ObservabilityPipelineQuotaProcessorLimit{
			Enforce: datadogV2.ObservabilityPipelineQuotaProcessorLimitEnforceType(src.Limit[0].Enforce.ValueString()),
			Limit:   src.Limit[0].Limit.ValueInt64(),
		})
	}

	if !src.OverflowAction.IsNull() {
		proc.SetOverflowAction(datadogV2.ObservabilityPipelineQuotaProcessorOverflowAction(src.OverflowAction.ValueString()))
	}

	if !src.TooManyBucketsAction.IsNull() {
		proc.SetTooManyBucketsAction(datadogV2.ObservabilityPipelineQuotaProcessorOverflowAction(src.TooManyBucketsAction.ValueString()))
	}

	var overrides []datadogV2.ObservabilityPipelineQuotaProcessorOverride
	for _, o := range src.Overrides {
		override := datadogV2.ObservabilityPipelineQuotaProcessorOverride{
			Limit: datadogV2.ObservabilityPipelineQuotaProcessorLimit{
				Enforce: datadogV2.ObservabilityPipelineQuotaProcessorLimitEnforceType(o.Limit[0].Enforce.ValueString()),
				Limit:   o.Limit[0].Limit.ValueInt64(),
			},
		}
		var fields []datadogV2.ObservabilityPipelineFieldValue
		for _, f := range o.Fields {
			fields = append(fields, datadogV2.ObservabilityPipelineFieldValue{
				Name:  f.Name.ValueString(),
				Value: f.Value.ValueString(),
			})
		}
		override.SetFields(fields)
		overrides = append(overrides, override)
	}
	proc.SetOverrides(overrides)

	return datadogV2.ObservabilityPipelineQuotaProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandDedupeProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *dedupeProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineDedupeProcessorWithDefaults()
	common.ApplyTo(proc)

	// Initialize as empty slice, not nil, to ensure it serializes as [] not null
	fields := []string{}
	for _, f := range src.Fields {
		fields = append(fields, f.ValueString())
	}
	proc.SetFields(fields)
	proc.SetMode(datadogV2.ObservabilityPipelineDedupeProcessorMode(src.Mode.ValueString()))

	return datadogV2.ObservabilityPipelineDedupeProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandReduceProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *reduceProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineReduceProcessorWithDefaults()
	common.ApplyTo(proc)

	// Initialize as empty slice, not nil, to ensure it serializes as [] not null
	groupBy := []string{}
	for _, g := range src.GroupBy {
		groupBy = append(groupBy, g.ValueString())
	}
	proc.SetGroupBy(groupBy)

	var strategies []datadogV2.ObservabilityPipelineReduceProcessorMergeStrategy
	for _, s := range src.MergeStrategies {
		strategies = append(strategies, datadogV2.ObservabilityPipelineReduceProcessorMergeStrategy{
			Path:     s.Path.ValueString(),
			Strategy: datadogV2.ObservabilityPipelineReduceProcessorMergeStrategyStrategy(s.Strategy.ValueString()),
		})
	}
	proc.SetMergeStrategies(strategies)

	return datadogV2.ObservabilityPipelineReduceProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandThrottleProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *throttleProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineThrottleProcessorWithDefaults()
	common.ApplyTo(proc)
	proc.SetThreshold(src.Threshold.ValueInt64())
	proc.SetWindow(src.Window.ValueFloat64())

	// Initialize as empty slice, not nil, to ensure it serializes as [] not null
	groupBy := []string{}
	for _, g := range src.GroupBy {
		groupBy = append(groupBy, g.ValueString())
	}
	if len(groupBy) > 0 {
		proc.SetGroupBy(groupBy)
	}

	return datadogV2.ObservabilityPipelineThrottleProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandAddEnvVarsProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *addEnvVarsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineAddEnvVarsProcessorWithDefaults()
	common.ApplyTo(proc)

	var vars []datadogV2.ObservabilityPipelineAddEnvVarsProcessorVariable
	for _, v := range src.Variables {
		vars = append(vars, datadogV2.ObservabilityPipelineAddEnvVarsProcessorVariable{
			Field: v.Field.ValueString(),
			Name:  v.Name.ValueString(),
		})
	}
	proc.SetVariables(vars)

	return datadogV2.ObservabilityPipelineAddEnvVarsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandEnrichmentTableProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *enrichmentTableProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineEnrichmentTableProcessorWithDefaults()
	common.ApplyTo(proc)
	proc.SetTarget(src.Target.ValueString())

	if len(src.File) > 0 {
		file := datadogV2.ObservabilityPipelineEnrichmentTableFile{
			Path: src.File[0].Path.ValueString(),
		}

		file.Encoding = datadogV2.ObservabilityPipelineEnrichmentTableFileEncoding{
			Type:            datadogV2.ObservabilityPipelineEnrichmentTableFileEncodingType(src.File[0].Encoding[0].Type.ValueString()),
			Delimiter:       src.File[0].Encoding[0].Delimiter.ValueString(),
			IncludesHeaders: src.File[0].Encoding[0].IncludesHeaders.ValueBool(),
		}

		// Set empty schema list - required by API
		file.Schema = []datadogV2.ObservabilityPipelineEnrichmentTableFileSchemaItems{}

		for _, k := range src.File[0].Key {
			file.Key = append(file.Key, datadogV2.ObservabilityPipelineEnrichmentTableFileKeyItems{
				Column:     k.Column.ValueString(),
				Comparison: datadogV2.ObservabilityPipelineEnrichmentTableFileKeyItemsComparison(k.Comparison.ValueString()),
				Field:      k.Field.ValueString(),
			})
		}

		proc.SetFile(file)
	}

	if len(src.GeoIp) > 0 {
		geoip := datadogV2.ObservabilityPipelineEnrichmentTableGeoIp{
			KeyField: src.GeoIp[0].KeyField.ValueString(),
			Locale:   src.GeoIp[0].Locale.ValueString(),
			Path:     src.GeoIp[0].Path.ValueString(),
		}
		proc.SetGeoip(geoip)
	}

	if len(src.ReferenceTable) > 0 {
		refTable := datadogV2.ObservabilityPipelineEnrichmentTableReferenceTable{
			KeyField: src.ReferenceTable[0].KeyField.ValueString(),
			TableId:  src.ReferenceTable[0].TableId.ValueString(),
		}
		if !src.ReferenceTable[0].Columns.IsNull() && !src.ReferenceTable[0].Columns.IsUnknown() {
			var columns []string
			src.ReferenceTable[0].Columns.ElementsAs(ctx, &columns, false)
			refTable.Columns = columns
		}
		proc.ReferenceTable = &refTable
	}

	return datadogV2.ObservabilityPipelineEnrichmentTableProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandOcsfMapperProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *ocsfMapperProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineOcsfMapperProcessorWithDefaults()
	common.ApplyTo(proc)

	var mappings []datadogV2.ObservabilityPipelineOcsfMapperProcessorMapping
	for _, m := range src.Mapping {
		libMapping := datadogV2.ObservabilityPipelineOcsfMappingLibrary(m.LibraryMapping.ValueString())
		mapping := datadogV2.ObservabilityPipelineOcsfMappingLibraryAsObservabilityPipelineOcsfMapperProcessorMappingMapping(&libMapping)
		mappings = append(mappings, datadogV2.ObservabilityPipelineOcsfMapperProcessorMapping{
			Include: m.Include.ValueString(),
			Mapping: mapping,
		})
	}
	proc.SetMappings(mappings)

	return datadogV2.ObservabilityPipelineOcsfMapperProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandParseGrokProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *parseGrokProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineParseGrokProcessorWithDefaults()
	common.ApplyTo(proc)

	if !src.DisableLibraryRules.IsNull() {
		proc.SetDisableLibraryRules(src.DisableLibraryRules.ValueBool())
	}

	var rules []datadogV2.ObservabilityPipelineParseGrokProcessorRule
	for _, r := range src.Rules {
		rule := datadogV2.ObservabilityPipelineParseGrokProcessorRule{
			Source: r.Source.ValueString(),
		}

		var matchRules []datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule
		for _, m := range r.MatchRules {
			matchRules = append(matchRules, datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule{
				Name: m.Name.ValueString(),
				Rule: m.Rule.ValueString(),
			})
		}
		rule.SetMatchRules(matchRules)

		var supportRules []datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule
		for _, s := range r.SupportRules {
			supportRules = append(supportRules, datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule{
				Name: s.Name.ValueString(),
				Rule: s.Rule.ValueString(),
			})
		}
		rule.SetSupportRules(supportRules)

		rules = append(rules, rule)
	}
	proc.SetRules(rules)

	return datadogV2.ObservabilityPipelineParseGrokProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandSampleProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *sampleProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineSampleProcessorWithDefaults()
	common.ApplyTo(proc)

	if !src.Percentage.IsNull() {
		proc.SetPercentage(src.Percentage.ValueFloat64())
	}

	// Only set group_by if there are values
	var groupBy []string
	for _, g := range src.GroupBy {
		groupBy = append(groupBy, g.ValueString())
	}
	if len(groupBy) > 0 {
		proc.SetGroupBy(groupBy)
	}

	return datadogV2.ObservabilityPipelineSampleProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandGenerateMetricsProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *generateMetricsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineGenerateMetricsProcessorWithDefaults()
	common.ApplyTo(proc)

	var metrics []datadogV2.ObservabilityPipelineGeneratedMetric
	for _, m := range src.Metrics {
		// Initialize as empty slice, not nil, to ensure it serializes as [] not null
		groupBy := []string{}
		m.GroupBy.ElementsAs(ctx, &groupBy, false)

		val := datadogV2.ObservabilityPipelineMetricValue{}
		if len(m.Value) > 0 {
			switch m.Value[0].Strategy.ValueString() {
			case "increment_by_one":
				val.ObservabilityPipelineGeneratedMetricIncrementByOne = &datadogV2.ObservabilityPipelineGeneratedMetricIncrementByOne{
					Strategy: "increment_by_one",
				}
			case "increment_by_field":
				val.ObservabilityPipelineGeneratedMetricIncrementByField = &datadogV2.ObservabilityPipelineGeneratedMetricIncrementByField{
					Strategy: "increment_by_field",
					Field:    m.Value[0].Field.ValueString(),
				}
			}
		}

		metrics = append(metrics, datadogV2.ObservabilityPipelineGeneratedMetric{
			Name:       m.Name.ValueString(),
			Include:    m.Include.ValueString(),
			MetricType: datadogV2.ObservabilityPipelineGeneratedMetricMetricType(m.MetricType.ValueString()),
			Value:      val,
			GroupBy:    groupBy,
		})
	}
	proc.SetMetrics(metrics)

	return datadogV2.ObservabilityPipelineGenerateMetricsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandSensitiveDataScannerProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *sensitiveDataScannerProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorWithDefaults()
	common.ApplyTo(proc)

	var rules []datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorRule
	for _, r := range src.Rules {
		rule := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorRuleWithDefaults()

		if !r.Name.IsNull() {
			rule.SetName(r.Name.ValueString())
		}

		// Only set tags if user specified them in config (to distinguish null from [])
		if r.Tags != nil {
			tags := []string{}
			for _, t := range r.Tags {
				tags = append(tags, t.ValueString())
			}
			rule.SetTags(tags)
		}

		if r.KeywordOptions != nil {
			ko := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorKeywordOptionsWithDefaults()
			// Initialize as empty slice, not nil, to ensure it serializes as [] not null
			keywords := []string{}
			for _, k := range r.KeywordOptions[0].Keywords {
				keywords = append(keywords, k.ValueString())
			}
			ko.SetKeywords(keywords)
			if !r.KeywordOptions[0].Proximity.IsNull() {
				ko.SetProximity(r.KeywordOptions[0].Proximity.ValueInt64())
			}
			rule.SetKeywordOptions(*ko)
		}

		// Expand Pattern
		if len(r.Pattern) > 0 {
			tfPattern := r.Pattern[0]
			if len(tfPattern.Custom) > 0 {
				options := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorCustomPatternOptionsWithDefaults()
				options.SetRule(tfPattern.Custom[0].Rule.ValueString())
				if !tfPattern.Custom[0].Description.IsNull() {
					options.SetDescription(tfPattern.Custom[0].Description.ValueString())
				}
				customPattern := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorCustomPattern(
					*options,
					datadogV2.OBSERVABILITYPIPELINESENSITIVEDATASCANNERPROCESSORCUSTOMPATTERNTYPE_CUSTOM,
				)
				pattern := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorCustomPatternAsObservabilityPipelineSensitiveDataScannerProcessorPattern(customPattern)
				rule.SetPattern(pattern)
			} else if len(tfPattern.Library) > 0 {
				options := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorLibraryPatternOptionsWithDefaults()
				options.SetId(tfPattern.Library[0].Id.ValueString())
				if !tfPattern.Library[0].Description.IsNull() {
					options.SetDescription(tfPattern.Library[0].Description.ValueString())
				}
				if !tfPattern.Library[0].UseRecommendedKeywords.IsNull() {
					options.SetUseRecommendedKeywords(tfPattern.Library[0].UseRecommendedKeywords.ValueBool())
				}
				libraryPattern := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorLibraryPattern(
					*options,
					datadogV2.OBSERVABILITYPIPELINESENSITIVEDATASCANNERPROCESSORLIBRARYPATTERNTYPE_LIBRARY,
				)
				pattern := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorLibraryPatternAsObservabilityPipelineSensitiveDataScannerProcessorPattern(libraryPattern)
				rule.SetPattern(pattern)
			}
		}

		// Expand Scope
		if len(r.Scope) > 0 {
			tfScope := r.Scope[0]
			if tfScope.Include != nil {
				// Initialize as empty slice, not nil, to ensure it serializes as [] not null
				fields := []string{}
				for _, f := range tfScope.Include[0].Fields {
					fields = append(fields, f.ValueString())
				}
				options := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorScopeOptionsWithDefaults()
				options.SetFields(fields)
				scopeInclude := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorScopeInclude(
					*options,
					datadogV2.OBSERVABILITYPIPELINESENSITIVEDATASCANNERPROCESSORSCOPEINCLUDETARGET_INCLUDE,
				)
				scope := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScopeIncludeAsObservabilityPipelineSensitiveDataScannerProcessorScope(scopeInclude)
				rule.SetScope(scope)
			} else if len(tfScope.Exclude) > 0 {
				// Initialize as empty slice, not nil, to ensure it serializes as [] not null
				fields := []string{}
				for _, f := range tfScope.Exclude[0].Fields {
					fields = append(fields, f.ValueString())
				}
				options := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorScopeOptionsWithDefaults()
				options.SetFields(fields)
				scopeExclude := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorScopeExclude(
					*options,
					datadogV2.OBSERVABILITYPIPELINESENSITIVEDATASCANNERPROCESSORSCOPEEXCLUDETARGET_EXCLUDE,
				)
				scope := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScopeExcludeAsObservabilityPipelineSensitiveDataScannerProcessorScope(scopeExclude)
				rule.SetScope(scope)
			} else if tfScope.All != nil && *tfScope.All {
				scopeAll := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorScopeAll(
					datadogV2.OBSERVABILITYPIPELINESENSITIVEDATASCANNERPROCESSORSCOPEALLTARGET_ALL,
				)
				scope := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScopeAllAsObservabilityPipelineSensitiveDataScannerProcessorScope(scopeAll)
				rule.SetScope(scope)
			}
		}

		// Expand OnMatch
		if len(r.OnMatch) > 0 {
			tfOnMatch := r.OnMatch[0]
			if len(tfOnMatch.Redact) > 0 {
				options := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorActionRedactOptionsWithDefaults()
				options.SetReplace(tfOnMatch.Redact[0].Replace.ValueString())
				actionRedact := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorActionRedact(
					datadogV2.OBSERVABILITYPIPELINESENSITIVEDATASCANNERPROCESSORACTIONREDACTACTION_REDACT,
					*options,
				)
				action := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionRedactAsObservabilityPipelineSensitiveDataScannerProcessorAction(actionRedact)
				rule.SetOnMatch(action)
			} else if tfOnMatch.Hash != nil {
				actionHash := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorActionHash(
					datadogV2.OBSERVABILITYPIPELINESENSITIVEDATASCANNERPROCESSORACTIONHASHACTION_HASH,
				)
				action := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionHashAsObservabilityPipelineSensitiveDataScannerProcessorAction(actionHash)
				rule.SetOnMatch(action)
			} else if len(tfOnMatch.PartialRedact) > 0 {
				options := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedactOptionsWithDefaults()
				options.SetCharacters(tfOnMatch.PartialRedact[0].Characters.ValueInt64())
				options.SetDirection(datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedactOptionsDirection(tfOnMatch.PartialRedact[0].Direction.ValueString()))
				actionPartialRedact := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedact(
					datadogV2.OBSERVABILITYPIPELINESENSITIVEDATASCANNERPROCESSORACTIONPARTIALREDACTACTION_PARTIAL_REDACT,
					*options,
				)
				action := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedactAsObservabilityPipelineSensitiveDataScannerProcessorAction(actionPartialRedact)
				rule.SetOnMatch(action)
			}
		}

		rules = append(rules, *rule)
	}
	proc.SetRules(rules)

	return datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandAddHostnameProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *addHostnameProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineAddHostnameProcessorWithDefaults()
	common.ApplyTo(proc)

	return datadogV2.ObservabilityPipelineAddHostnameProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandParseXMLProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *parseXMLProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineParseXMLProcessorWithDefaults()
	common.ApplyTo(proc)

	proc.SetField(src.Field.ValueString())

	if !src.IncludeAttr.IsNull() {
		proc.SetIncludeAttr(src.IncludeAttr.ValueBool())
	}
	if !src.AlwaysUseTextKey.IsNull() {
		proc.SetAlwaysUseTextKey(src.AlwaysUseTextKey.ValueBool())
	}
	if !src.ParseNumber.IsNull() {
		proc.SetParseNumber(src.ParseNumber.ValueBool())
	}
	if !src.ParseBool.IsNull() {
		proc.SetParseBool(src.ParseBool.ValueBool())
	}
	if !src.ParseNull.IsNull() {
		proc.SetParseNull(src.ParseNull.ValueBool())
	}
	if !src.AttrPrefix.IsNull() {
		proc.SetAttrPrefix(src.AttrPrefix.ValueString())
	}
	if !src.TextKey.IsNull() {
		proc.SetTextKey(src.TextKey.ValueString())
	}

	return datadogV2.ObservabilityPipelineParseXMLProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func expandSplitArrayProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *splitArrayProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineSplitArrayProcessorWithDefaults()
	common.ApplyTo(proc)

	var arrays []datadogV2.ObservabilityPipelineSplitArrayProcessorArrayConfig
	for _, arr := range src.Arrays {
		arrays = append(arrays, datadogV2.ObservabilityPipelineSplitArrayProcessorArrayConfig{
			Include: arr.Include.ValueString(),
			Field:   arr.Field.ValueString(),
		})
	}
	proc.SetArrays(arrays)

	return datadogV2.ObservabilityPipelineSplitArrayProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func flattenAddHostnameProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineAddHostnameProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	model.AddHostnameProcessor = append(model.AddHostnameProcessor, &addHostnameProcessorModel{})
	return model
}

func flattenParseXMLProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineParseXMLProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	parseXML := &parseXMLProcessorModel{
		Field: types.StringValue(src.GetField()),
	}

	if val, ok := src.GetIncludeAttrOk(); ok {
		parseXML.IncludeAttr = types.BoolValue(*val)
	}
	if val, ok := src.GetAlwaysUseTextKeyOk(); ok {
		parseXML.AlwaysUseTextKey = types.BoolValue(*val)
	}
	if val, ok := src.GetParseNumberOk(); ok {
		parseXML.ParseNumber = types.BoolValue(*val)
	}
	if val, ok := src.GetParseBoolOk(); ok {
		parseXML.ParseBool = types.BoolValue(*val)
	}
	if val, ok := src.GetParseNullOk(); ok {
		parseXML.ParseNull = types.BoolValue(*val)
	}
	if val, ok := src.GetAttrPrefixOk(); ok {
		parseXML.AttrPrefix = types.StringValue(*val)
	}
	if val, ok := src.GetTextKeyOk(); ok {
		parseXML.TextKey = types.StringValue(*val)
	}

	model.ParseXMLProcessor = append(model.ParseXMLProcessor, parseXML)
	return model
}

func flattenSplitArrayProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineSplitArrayProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	splitArray := &splitArrayProcessorModel{}

	for _, arr := range src.GetArrays() {
		splitArray.Arrays = append(splitArray.Arrays, splitArrayConfigModel{
			Include: types.StringValue(arr.GetInclude()),
			Field:   types.StringValue(arr.GetField()),
		})
	}

	model.SplitArrayProcessor = append(model.SplitArrayProcessor, splitArray)
	return model
}
func flattenMetricTagsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineMetricTagsProcessor) *processorModel {
	if src == nil {
		return nil
	}
	model := createProcessorModel(src)
	metricTags := &metricTagsProcessorModel{}
	for _, rule := range src.GetRules() {
		var keys []types.String
		for _, k := range rule.GetKeys() {
			keys = append(keys, types.StringValue(k))
		}
		metricTags.Rules = append(metricTags.Rules, metricTagsProcessorRuleModel{
			Include: types.StringValue(rule.GetInclude()),
			Mode:    types.StringValue(string(rule.GetMode())),
			Action:  types.StringValue(string(rule.GetAction())),
			Keys:    keys,
		})
	}
	model.MetricTagsProcessor = append(model.MetricTagsProcessor, metricTags)
	return model
}

func expandMetricTagsProcessorItem(ctx context.Context, common observability_pipeline.BaseProcessorFields, src *metricTagsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineMetricTagsProcessorWithDefaults()
	common.ApplyTo(proc)

	var rules []datadogV2.ObservabilityPipelineMetricTagsProcessorRule
	for _, r := range src.Rules {
		rule := datadogV2.ObservabilityPipelineMetricTagsProcessorRule{
			Include: r.Include.ValueString(),
			Mode:    datadogV2.ObservabilityPipelineMetricTagsProcessorRuleMode(r.Mode.ValueString()),
			Action:  datadogV2.ObservabilityPipelineMetricTagsProcessorRuleAction(r.Action.ValueString()),
		}
		var keys []string
		for _, k := range r.Keys {
			keys = append(keys, k.ValueString())
		}
		rule.SetKeys(keys)
		rules = append(rules, rule)
	}
	proc.SetRules(rules)

	return datadogV2.ObservabilityPipelineMetricTagsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

// ---------- Destinations ----------

func flattenDatadogLogsDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineDatadogLogsDestination) *datadogLogsDestinationModel {
	if src == nil {
		return nil
	}
	out := &datadogLogsDestinationModel{}

	if routes := src.GetRoutes(); len(routes) > 0 {
		out.Routes = make([]datadogLogsDestinationRouteModel, 0, len(routes))
		for _, route := range routes {
			routeModel := datadogLogsDestinationRouteModel{}

			routeModel.RouteId = types.StringValue(route.GetRouteId())
			routeModel.Include = types.StringValue(route.GetInclude())
			routeModel.Site = types.StringValue(route.GetSite())
			routeModel.ApiKeyKey = types.StringValue(route.GetApiKeyKey())
			out.Routes = append(out.Routes, routeModel)
		}
	}

	return out
}

func expandDatadogLogsDestination(ctx context.Context, dest *destinationModel, src *datadogLogsDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	d := datadogV2.NewObservabilityPipelineDatadogLogsDestinationWithDefaults()
	d.SetId(dest.Id.ValueString())
	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	d.SetInputs(inputs)

	if len(src.Routes) > 0 {
		routes := make([]datadogV2.ObservabilityPipelineDatadogLogsDestinationRoute, 0, len(src.Routes))
		for _, route := range src.Routes {
			apiRoute := datadogV2.ObservabilityPipelineDatadogLogsDestinationRoute{}
			apiRoute.SetRouteId(route.RouteId.ValueString())
			apiRoute.SetInclude(route.Include.ValueString())
			apiRoute.SetSite(route.Site.ValueString())
			apiRoute.SetApiKeyKey(route.ApiKeyKey.ValueString())
			routes = append(routes, apiRoute)
		}

		d.SetRoutes(routes)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineDatadogLogsDestination: d,
	}
}

func flattenDatadogMetricsDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineDatadogMetricsDestination) *datadogMetricsDestinationModel {
	if src == nil {
		return nil
	}
	return &datadogMetricsDestinationModel{}
}

func expandDatadogMetricsDestination(ctx context.Context, dest *destinationModel, src *datadogMetricsDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	d := datadogV2.NewObservabilityPipelineDatadogMetricsDestinationWithDefaults()
	d.SetId(dest.Id.ValueString())
	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	d.SetInputs(inputs)
	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineDatadogMetricsDestination: d,
	}
}

func expandHttpClientDestination(ctx context.Context, dest *destinationModel, src *httpClientDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	d := datadogV2.NewObservabilityPipelineHttpClientDestinationWithDefaults()
	d.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	d.SetInputs(inputs)

	d.SetEncoding(datadogV2.ObservabilityPipelineHttpClientDestinationEncoding(src.Encoding.ValueString()))

	if !src.AuthStrategy.IsNull() {
		d.SetAuthStrategy(datadogV2.ObservabilityPipelineHttpClientDestinationAuthStrategy(src.AuthStrategy.ValueString()))
	}

	if len(src.Compression) > 0 {
		comp := datadogV2.ObservabilityPipelineHttpClientDestinationCompression{
			Algorithm: datadogV2.ObservabilityPipelineHttpClientDestinationCompressionAlgorithm(src.Compression[0].Algorithm.ValueString()),
		}
		d.SetCompression(comp)
	}

	d.Tls = observability_pipeline.ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineHttpClientDestination: d,
	}
}

func flattenHttpClientDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineHttpClientDestination) *httpClientDestinationModel {
	if src == nil {
		return nil
	}

	out := &httpClientDestinationModel{
		Encoding: types.StringValue(string(src.GetEncoding())),
	}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	if auth, ok := src.GetAuthStrategyOk(); ok {
		out.AuthStrategy = types.StringValue(string(*auth))
	}

	if comp, ok := src.GetCompressionOk(); ok {
		out.Compression = []httpClientDestinationCompressionModel{
			{
				Algorithm: types.StringValue(string(comp.GetAlgorithm())),
			},
		}
	}

	return out
}

func expandFluentdSource(src *fluentdSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	source := datadogV2.NewObservabilityPipelineFluentdSourceWithDefaults()
	source.SetId(id)
	source.Tls = observability_pipeline.ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineFluentdSource: source,
	}
}

func expandFluentBitSource(src *fluentBitSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	source := datadogV2.NewObservabilityPipelineFluentBitSourceWithDefaults()
	source.SetId(id)

	if src.Tls != nil {
		source.Tls = observability_pipeline.ExpandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineFluentBitSource: source,
	}
}

func flattenFluentdSource(src *datadogV2.ObservabilityPipelineFluentdSource) *fluentdSourceModel {
	if src == nil {
		return nil
	}

	out := &fluentdSourceModel{}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	return out
}

func flattenFluentBitSource(src *datadogV2.ObservabilityPipelineFluentBitSource) *fluentBitSourceModel {
	if src == nil {
		return nil
	}

	out := &fluentBitSourceModel{}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	return out
}

func decodingSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Required:    true,
		Description: "The decoding format used to interpret incoming logs.",
		Validators: []validator.String{
			stringvalidator.OneOf("json", "gelf", "syslog", "bytes"),
		},
	}
}

func expandHttpServerSource(src *httpServerSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineHttpServerSourceWithDefaults()
	s.SetId(id)

	s.SetAuthStrategy(datadogV2.ObservabilityPipelineHttpServerSourceAuthStrategy(src.AuthStrategy.ValueString()))
	s.SetDecoding(datadogV2.ObservabilityPipelineDecoding(src.Decoding.ValueString()))

	s.Tls = observability_pipeline.ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineHttpServerSource: s,
	}
}

func flattenHttpServerSource(src *datadogV2.ObservabilityPipelineHttpServerSource) *httpServerSourceModel {
	if src == nil {
		return nil
	}

	out := &httpServerSourceModel{
		AuthStrategy: types.StringValue(string(src.GetAuthStrategy())),
		Decoding:     types.StringValue(string(src.GetDecoding())),
	}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	return out
}

func expandSplunkHecSource(src *splunkHecSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineSplunkHecSourceWithDefaults()
	s.SetId(id)

	if src.Tls != nil {
		s.Tls = observability_pipeline.ExpandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSplunkHecSource: s,
	}
}

func flattenSplunkHecSource(src *datadogV2.ObservabilityPipelineSplunkHecSource) *splunkHecSourceModel {
	if src == nil {
		return nil
	}

	out := &splunkHecSourceModel{}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	return out
}

func expandGoogleCloudStorageDestination(ctx context.Context, destModel *destinationModel, d *gcsDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineGoogleCloudStorageDestinationWithDefaults()

	dest.SetId(destModel.Id.ValueString())
	dest.SetBucket(d.Bucket.ValueString())
	dest.SetStorageClass(datadogV2.ObservabilityPipelineGoogleCloudStorageDestinationStorageClass(d.StorageClass.ValueString()))

	if !d.Acl.IsNull() {
		dest.SetAcl(datadogV2.ObservabilityPipelineGoogleCloudStorageDestinationAcl(d.Acl.ValueString()))
	}

	if !d.KeyPrefix.IsNull() {
		dest.SetKeyPrefix(d.KeyPrefix.ValueString())
	}

	if auth := expandGcpAuth(d.Auth); auth != nil {
		dest.SetAuth(*auth)
	}

	var metadata []datadogV2.ObservabilityPipelineMetadataEntry
	for _, m := range d.Metadata {
		metadata = append(metadata, datadogV2.ObservabilityPipelineMetadataEntry{
			Name:  m.Name.ValueString(),
			Value: m.Value.ValueString(),
		})
	}
	dest.SetMetadata(metadata)

	var inputs []string
	destModel.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineGoogleCloudStorageDestination: dest,
	}
}

func flattenGoogleCloudStorageDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineGoogleCloudStorageDestination) *gcsDestinationModel {
	if src == nil {
		return nil
	}

	var metadata []metadataEntry
	for _, m := range src.GetMetadata() {
		metadata = append(metadata, metadataEntry{
			Name:  types.StringValue(m.Name),
			Value: types.StringValue(m.Value),
		})
	}

	out := &gcsDestinationModel{
		Bucket:       types.StringValue(src.GetBucket()),
		KeyPrefix:    types.StringPointerValue(src.KeyPrefix),
		StorageClass: types.StringValue(string(src.GetStorageClass())),
		Metadata:     metadata,
	}

	if acl, ok := src.GetAclOk(); ok {
		out.Acl = types.StringValue(string(*acl))
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = flattenGcpAuth(auth)
	}

	return out
}

func expandGooglePubSubDestination(ctx context.Context, dest *destinationModel, d *googlePubSubDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	pubsub := datadogV2.NewObservabilityPipelineGooglePubSubDestinationWithDefaults()
	pubsub.SetId(dest.Id.ValueString())
	pubsub.SetProject(d.Project.ValueString())
	pubsub.SetTopic(d.Topic.ValueString())

	if !d.Encoding.IsNull() {
		pubsub.SetEncoding(datadogV2.ObservabilityPipelineGooglePubSubDestinationEncoding(d.Encoding.ValueString()))
	}

	if len(d.Auth) > 0 {
		auth := datadogV2.ObservabilityPipelineGcpAuth{}
		auth.SetCredentialsFile(d.Auth[0].CredentialsFile.ValueString())
		pubsub.SetAuth(auth)
	}

	pubsub.Tls = observability_pipeline.ExpandTls(d.Tls)

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	pubsub.SetInputs(inputs)

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineGooglePubSubDestination: pubsub,
	}
}

func flattenGooglePubSubDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineGooglePubSubDestination) *googlePubSubDestinationModel {
	if src == nil {
		return nil
	}

	out := &googlePubSubDestinationModel{
		Project: types.StringValue(src.GetProject()),
		Topic:   types.StringValue(src.GetTopic()),
	}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	if encoding, ok := src.GetEncodingOk(); ok {
		out.Encoding = types.StringValue(string(*encoding))
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = flattenGcpAuth(auth)
	}

	return out
}

func expandSplunkTcpSource(src *splunkTcpSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineSplunkTcpSourceWithDefaults()
	s.SetId(id)

	s.Tls = observability_pipeline.ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSplunkTcpSource: s,
	}
}

func flattenSplunkTcpSource(src *datadogV2.ObservabilityPipelineSplunkTcpSource) *splunkTcpSourceModel {
	if src == nil {
		return nil
	}

	out := &splunkTcpSourceModel{}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	return out
}

func expandSplunkHecDestination(ctx context.Context, dest *destinationModel, d *splunkHecDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	splunk := datadogV2.NewObservabilityPipelineSplunkHecDestinationWithDefaults()

	splunk.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	splunk.SetInputs(inputs)

	if !d.AutoExtractTimestamp.IsNull() {
		splunk.SetAutoExtractTimestamp(d.AutoExtractTimestamp.ValueBool())
	}
	if !d.Encoding.IsNull() {
		splunk.SetEncoding(datadogV2.ObservabilityPipelineSplunkHecDestinationEncoding(d.Encoding.ValueString()))
	}
	if !d.Sourcetype.IsNull() {
		splunk.SetSourcetype(d.Sourcetype.ValueString())
	}
	if !d.Index.IsNull() {
		splunk.SetIndex(d.Index.ValueString())
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSplunkHecDestination: splunk,
	}
}

func flattenSplunkHecDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSplunkHecDestination) *splunkHecDestinationModel {
	if src == nil {
		return nil
	}

	autoExtractTimestamp := types.BoolNull()
	if src.HasAutoExtractTimestamp() {
		autoExtractTimestamp = types.BoolValue(src.GetAutoExtractTimestamp())
	}

	return &splunkHecDestinationModel{
		AutoExtractTimestamp: autoExtractTimestamp,
		Encoding:             types.StringValue(string(*src.Encoding)),
		Sourcetype:           types.StringPointerValue(src.Sourcetype),
		Index:                types.StringPointerValue(src.Index),
	}
}

func expandAmazonS3Source(src *amazonS3SourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineAmazonS3SourceWithDefaults()
	s.SetId(id)

	s.SetRegion(src.Region.ValueString())

	if len(src.Auth) > 0 {
		s.SetAuth(observability_pipeline.ExpandAwsAuth(src.Auth[0]))
	}

	s.Tls = observability_pipeline.ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineAmazonS3Source: s,
	}
}

func flattenAmazonS3Source(src *datadogV2.ObservabilityPipelineAmazonS3Source) *amazonS3SourceModel {
	if src == nil {
		return nil
	}

	out := &amazonS3SourceModel{
		Region: types.StringValue(src.GetRegion()),
	}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = observability_pipeline.FlattenAwsAuth(auth)
	}

	return out
}

func expandSumoLogicDestination(ctx context.Context, dest *destinationModel, src *sumoLogicDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	sumo := datadogV2.NewObservabilityPipelineSumoLogicDestinationWithDefaults()
	sumo.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	sumo.SetInputs(inputs)

	if !src.Encoding.IsNull() {
		sumo.SetEncoding(datadogV2.ObservabilityPipelineSumoLogicDestinationEncoding(src.Encoding.ValueString()))
	}
	if !src.HeaderHostName.IsNull() {
		sumo.SetHeaderHostName(src.HeaderHostName.ValueString())
	}
	if !src.HeaderSourceName.IsNull() {
		sumo.SetHeaderSourceName(src.HeaderSourceName.ValueString())
	}
	if !src.HeaderSourceCategory.IsNull() {
		sumo.SetHeaderSourceCategory(src.HeaderSourceCategory.ValueString())
	}

	if len(src.HeaderCustomFields) > 0 {
		var fields []datadogV2.ObservabilityPipelineSumoLogicDestinationHeaderCustomFieldsItem
		for _, f := range src.HeaderCustomFields {
			fields = append(fields, datadogV2.ObservabilityPipelineSumoLogicDestinationHeaderCustomFieldsItem{
				Name:  f.Name.ValueString(),
				Value: f.Value.ValueString(),
			})
		}
		sumo.SetHeaderCustomFields(fields)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSumoLogicDestination: sumo,
	}
}

func flattenSumoLogicDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSumoLogicDestination) *sumoLogicDestinationModel {
	if src == nil {
		return nil
	}

	out := &sumoLogicDestinationModel{}

	if v, ok := src.GetEncodingOk(); ok {
		out.Encoding = types.StringValue(string(*v))
	}
	if v, ok := src.GetHeaderHostNameOk(); ok {
		out.HeaderHostName = types.StringValue(*v)
	}
	if v, ok := src.GetHeaderSourceNameOk(); ok {
		out.HeaderSourceName = types.StringValue(*v)
	}
	if v, ok := src.GetHeaderSourceCategoryOk(); ok {
		out.HeaderSourceCategory = types.StringValue(*v)
	}
	if v, ok := src.GetHeaderCustomFieldsOk(); ok {
		for _, f := range *v {
			out.HeaderCustomFields = append(out.HeaderCustomFields, headerCustomFieldModel{
				Name:  types.StringValue(f.Name),
				Value: types.StringValue(f.Value),
			})
		}
	}

	return out
}

func expandRsyslogSource(src *rsyslogSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	obj := datadogV2.NewObservabilityPipelineRsyslogSourceWithDefaults()
	obj.SetId(id)
	if !src.Mode.IsNull() {
		obj.SetMode(datadogV2.ObservabilityPipelineSyslogSourceMode(src.Mode.ValueString()))
	}
	obj.Tls = observability_pipeline.ExpandTls(src.Tls)
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineRsyslogSource: obj,
	}
}

func flattenRsyslogSource(src *datadogV2.ObservabilityPipelineRsyslogSource) *rsyslogSourceModel {
	if src == nil {
		return nil
	}
	out := &rsyslogSourceModel{}
	if v, ok := src.GetModeOk(); ok {
		out.Mode = types.StringValue(string(*v))
	}
	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	return out
}

func expandSyslogNgSource(src *syslogNgSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	obj := datadogV2.NewObservabilityPipelineSyslogNgSourceWithDefaults()
	obj.SetId(id)
	if !src.Mode.IsNull() {
		obj.SetMode(datadogV2.ObservabilityPipelineSyslogSourceMode(src.Mode.ValueString()))
	}
	obj.Tls = observability_pipeline.ExpandTls(src.Tls)
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSyslogNgSource: obj,
	}
}

func flattenSyslogNgSource(src *datadogV2.ObservabilityPipelineSyslogNgSource) *syslogNgSourceModel {
	if src == nil {
		return nil
	}
	out := &syslogNgSourceModel{}
	if v, ok := src.GetModeOk(); ok {
		out.Mode = types.StringValue(string(*v))
	}
	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}
	return out
}

func expandRsyslogDestination(ctx context.Context, dest *destinationModel, src *rsyslogDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewObservabilityPipelineRsyslogDestinationWithDefaults()
	obj.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	if !src.Keepalive.IsNull() {
		obj.SetKeepalive(src.Keepalive.ValueInt64())
	}
	obj.Tls = observability_pipeline.ExpandTls(src.Tls)
	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineRsyslogDestination: obj,
	}
}

func flattenRsyslogDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineRsyslogDestination) *rsyslogDestinationModel {
	if src == nil {
		return nil
	}
	out := &rsyslogDestinationModel{}
	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}
	if v, ok := src.GetKeepaliveOk(); ok {
		out.Keepalive = types.Int64Value(*v)
	}
	return out
}

func expandSyslogNgDestination(ctx context.Context, dest *destinationModel, src *syslogNgDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewObservabilityPipelineSyslogNgDestinationWithDefaults()
	obj.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	if !src.Keepalive.IsNull() {
		obj.SetKeepalive(src.Keepalive.ValueInt64())
	}
	obj.Tls = observability_pipeline.ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSyslogNgDestination: obj,
	}
}

func flattenSyslogNgDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSyslogNgDestination) *syslogNgDestinationModel {
	if src == nil {
		return nil
	}
	out := &syslogNgDestinationModel{}
	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}
	if v, ok := src.GetKeepaliveOk(); ok {
		out.Keepalive = types.Int64Value(*v)
	}
	return out
}

func expandElasticsearchDestination(ctx context.Context, dest *destinationModel, src *elasticsearchDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewObservabilityPipelineElasticsearchDestinationWithDefaults()
	obj.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	if !src.ApiVersion.IsNull() {
		obj.SetApiVersion(datadogV2.ObservabilityPipelineElasticsearchDestinationApiVersion(src.ApiVersion.ValueString()))
	}
	if !src.BulkIndex.IsNull() {
		obj.SetBulkIndex(src.BulkIndex.ValueString())
	}
	if len(src.DataStream) > 0 {
		ds := datadogV2.NewObservabilityPipelineElasticsearchDestinationDataStream()
		if !src.DataStream[0].Dtype.IsNull() {
			ds.SetDtype(src.DataStream[0].Dtype.ValueString())
		}
		if !src.DataStream[0].Dataset.IsNull() {
			ds.SetDataset(src.DataStream[0].Dataset.ValueString())
		}
		if !src.DataStream[0].Namespace.IsNull() {
			ds.SetNamespace(src.DataStream[0].Namespace.ValueString())
		}
		obj.DataStream = ds
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineElasticsearchDestination: obj,
	}
}

func flattenElasticsearchDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineElasticsearchDestination) *elasticsearchDestinationModel {
	if src == nil {
		return nil
	}
	out := &elasticsearchDestinationModel{}
	if v, ok := src.GetApiVersionOk(); ok {
		out.ApiVersion = types.StringValue(string(*v))
	}
	if v, ok := src.GetBulkIndexOk(); ok {
		out.BulkIndex = types.StringValue(*v)
	}
	if ds, ok := src.GetDataStreamOk(); ok && ds != nil {
		dsModel := elasticsearchDestinationDataStreamModel{}
		if v, ok := ds.GetDtypeOk(); ok {
			dsModel.Dtype = types.StringValue(*v)
		}
		if v, ok := ds.GetDatasetOk(); ok {
			dsModel.Dataset = types.StringValue(*v)
		}
		if v, ok := ds.GetNamespaceOk(); ok {
			dsModel.Namespace = types.StringValue(*v)
		}
		out.DataStream = []elasticsearchDestinationDataStreamModel{dsModel}
	}
	return out
}

func expandAzureStorageDestination(ctx context.Context, dest *destinationModel, src *azureStorageDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewAzureStorageDestinationWithDefaults()
	obj.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	obj.SetContainerName(src.ContainerName.ValueString())

	if !src.BlobPrefix.IsNull() {
		obj.SetBlobPrefix(src.BlobPrefix.ValueString())
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		AzureStorageDestination: obj,
	}
}

func flattenAzureStorageDestination(ctx context.Context, src *datadogV2.AzureStorageDestination) *azureStorageDestinationModel {
	if src == nil {
		return nil
	}
	out := &azureStorageDestinationModel{
		ContainerName: types.StringValue(src.GetContainerName()),
	}
	if v, ok := src.GetBlobPrefixOk(); ok {
		out.BlobPrefix = types.StringValue(*v)
	}
	return out
}

func expandMicrosoftSentinelDestination(ctx context.Context, dest *destinationModel, src *microsoftSentinelDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewMicrosoftSentinelDestinationWithDefaults()
	obj.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	obj.SetClientId(src.ClientId.ValueString())
	obj.SetTenantId(src.TenantId.ValueString())
	obj.SetDcrImmutableId(src.DcrImmutableId.ValueString())
	obj.SetTable(src.Table.ValueString())

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		MicrosoftSentinelDestination: obj,
	}
}

func flattenMicrosoftSentinelDestination(ctx context.Context, src *datadogV2.MicrosoftSentinelDestination) *microsoftSentinelDestinationModel {
	if src == nil {
		return nil
	}
	return &microsoftSentinelDestinationModel{
		ClientId:       types.StringValue(src.GetClientId()),
		TenantId:       types.StringValue(src.GetTenantId()),
		DcrImmutableId: types.StringValue(src.GetDcrImmutableId()),
		Table:          types.StringValue(src.GetTable()),
	}
}

func expandSumoLogicSource(src *sumoLogicSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	obj := datadogV2.NewObservabilityPipelineSumoLogicSourceWithDefaults()
	obj.SetId(id)

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSumoLogicSource: obj,
	}
}

func flattenSumoLogicSource(src *datadogV2.ObservabilityPipelineSumoLogicSource) *sumoLogicSourceModel {
	if src == nil {
		return nil
	}
	return &sumoLogicSourceModel{}
}

func expandAmazonDataFirehoseSource(src *amazonDataFirehoseSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	firehose := datadogV2.NewObservabilityPipelineAmazonDataFirehoseSourceWithDefaults()
	firehose.SetId(id)

	if len(src.Auth) > 0 {
		firehose.SetAuth(observability_pipeline.ExpandAwsAuth(src.Auth[0]))
	}

	if src.Tls != nil {
		firehose.Tls = observability_pipeline.ExpandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineAmazonDataFirehoseSource: firehose,
	}
}

func flattenAmazonDataFirehoseSource(src *datadogV2.ObservabilityPipelineAmazonDataFirehoseSource) *amazonDataFirehoseSourceModel {
	if src == nil {
		return nil
	}

	out := &amazonDataFirehoseSourceModel{}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = observability_pipeline.FlattenAwsAuth(auth)
	}

	return out
}

func expandHttpClientSource(src *httpClientSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	httpSrc := datadogV2.NewObservabilityPipelineHttpClientSourceWithDefaults()
	httpSrc.SetId(id)
	httpSrc.SetDecoding(datadogV2.ObservabilityPipelineDecoding(src.Decoding.ValueString()))

	if !src.ScrapeInterval.IsNull() {
		httpSrc.SetScrapeIntervalSecs(src.ScrapeInterval.ValueInt64())
	}
	if !src.ScrapeTimeout.IsNull() {
		httpSrc.SetScrapeTimeoutSecs(src.ScrapeTimeout.ValueInt64())
	}
	if !src.AuthStrategy.IsNull() {
		auth := datadogV2.ObservabilityPipelineHttpClientSourceAuthStrategy(src.AuthStrategy.ValueString())
		httpSrc.SetAuthStrategy(auth)
	}
	httpSrc.Tls = observability_pipeline.ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineHttpClientSource: httpSrc,
	}
}

func flattenHttpClientSource(src *datadogV2.ObservabilityPipelineHttpClientSource) *httpClientSourceModel {
	if src == nil {
		return nil
	}

	out := &httpClientSourceModel{
		Decoding: types.StringValue(string(src.GetDecoding())),
	}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	if v, ok := src.GetScrapeIntervalSecsOk(); ok {
		out.ScrapeInterval = types.Int64Value(*v)
	}
	if v, ok := src.GetScrapeTimeoutSecsOk(); ok {
		out.ScrapeTimeout = types.Int64Value(*v)
	}
	if v, ok := src.GetAuthStrategyOk(); ok && v != nil {
		out.AuthStrategy = types.StringValue(string(*v))
	}

	return out
}

func expandGooglePubSubSource(src *googlePubSubSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	pubsub := datadogV2.NewObservabilityPipelineGooglePubSubSourceWithDefaults()
	pubsub.SetId(id)
	pubsub.SetProject(src.Project.ValueString())
	pubsub.SetSubscription(src.Subscription.ValueString())
	pubsub.SetDecoding(datadogV2.ObservabilityPipelineDecoding(src.Decoding.ValueString()))

	if auth := expandGcpAuth(src.Auth); auth != nil {
		pubsub.SetAuth(*auth)
	}

	pubsub.Tls = observability_pipeline.ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineGooglePubSubSource: pubsub,
	}
}

func flattenGooglePubSubSource(src *datadogV2.ObservabilityPipelineGooglePubSubSource) *googlePubSubSourceModel {
	if src == nil {
		return nil
	}
	out := &googlePubSubSourceModel{
		Project:      types.StringValue(src.GetProject()),
		Subscription: types.StringValue(src.GetSubscription()),
		Decoding:     types.StringValue(string(src.GetDecoding())),
	}

	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = flattenGcpAuth(auth)
	}

	return out
}

func expandLogstashSource(src *logstashSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	logstash := datadogV2.NewObservabilityPipelineLogstashSourceWithDefaults()
	logstash.SetId(id)
	logstash.Tls = observability_pipeline.ExpandTls(src.Tls)
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineLogstashSource: logstash,
	}
}

func flattenLogstashSource(src *datadogV2.ObservabilityPipelineLogstashSource) *logstashSourceModel {
	if src == nil {
		return nil
	}
	out := &logstashSourceModel{}
	if src.Tls != nil {
		out.Tls = observability_pipeline.FlattenTls(src.Tls)
	}

	return out
}

func expandGoogleSecopsDestination(ctx context.Context, dest *destinationModel, src *googleSecopsDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	chronicle := datadogV2.NewObservabilityPipelineGoogleChronicleDestinationWithDefaults()
	chronicle.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	chronicle.SetInputs(inputs)

	if len(src.Auth) > 0 {
		auth := datadogV2.ObservabilityPipelineGcpAuth{}
		if !src.Auth[0].CredentialsFile.IsNull() {
			auth.SetCredentialsFile(src.Auth[0].CredentialsFile.ValueString())
		}
		chronicle.SetAuth(auth)
	}

	if !src.CustomerId.IsNull() {
		chronicle.SetCustomerId(src.CustomerId.ValueString())
	}
	if !src.Encoding.IsNull() {
		chronicle.SetEncoding(datadogV2.ObservabilityPipelineGoogleChronicleDestinationEncoding(src.Encoding.ValueString()))
	}
	if !src.LogType.IsNull() {
		chronicle.SetLogType(src.LogType.ValueString())
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineGoogleChronicleDestination: chronicle,
	}
}

func flattenGoogleSecopsDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineGoogleChronicleDestination) *googleSecopsDestinationModel {
	if src == nil {
		return nil
	}

	out := &googleSecopsDestinationModel{}

	if v, ok := src.GetCustomerIdOk(); ok && v != nil && *v != "" {
		out.CustomerId = types.StringValue(*v)
	}
	if v, ok := src.GetEncodingOk(); ok && v != nil && string(*v) != "" {
		out.Encoding = types.StringValue(string(*v))
	}
	if v, ok := src.GetLogTypeOk(); ok && v != nil && *v != "" {
		out.LogType = types.StringValue(*v)
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = flattenGcpAuth(auth)
	}

	return out
}

func expandNewRelicDestination(ctx context.Context, dest *destinationModel, src *newRelicDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	newrelic := datadogV2.NewObservabilityPipelineNewRelicDestinationWithDefaults()
	newrelic.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	newrelic.SetInputs(inputs)

	newrelic.SetRegion(datadogV2.ObservabilityPipelineNewRelicDestinationRegion(src.Region.ValueString()))

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineNewRelicDestination: newrelic,
	}
}

func flattenNewRelicDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineNewRelicDestination) *newRelicDestinationModel {
	if src == nil {
		return nil
	}

	return &newRelicDestinationModel{
		Region: types.StringValue(string(src.GetRegion())),
	}
}

func expandSentinelOneDestination(ctx context.Context, dest *destinationModel, src *sentinelOneDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	sentinel := datadogV2.NewObservabilityPipelineSentinelOneDestinationWithDefaults()
	sentinel.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	sentinel.SetInputs(inputs)

	sentinel.SetRegion(datadogV2.ObservabilityPipelineSentinelOneDestinationRegion(src.Region.ValueString()))

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSentinelOneDestination: sentinel,
	}
}

func flattenSentinelOneDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSentinelOneDestination) *sentinelOneDestinationModel {
	if src == nil {
		return nil
	}

	return &sentinelOneDestinationModel{
		Region: types.StringValue(string(src.GetRegion())),
	}
}

func expandOpenSearchDestination(ctx context.Context, dest *destinationModel, src *opensearchDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	opensearch := datadogV2.NewObservabilityPipelineOpenSearchDestinationWithDefaults()
	opensearch.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	opensearch.SetInputs(inputs)

	if !src.BulkIndex.IsNull() {
		opensearch.SetBulkIndex(src.BulkIndex.ValueString())
	}

	if len(src.DataStream) > 0 {
		ds := datadogV2.NewObservabilityPipelineOpenSearchDestinationDataStream()
		if !src.DataStream[0].Dtype.IsNull() {
			ds.SetDtype(src.DataStream[0].Dtype.ValueString())
		}
		if !src.DataStream[0].Dataset.IsNull() {
			ds.SetDataset(src.DataStream[0].Dataset.ValueString())
		}
		if !src.DataStream[0].Namespace.IsNull() {
			ds.SetNamespace(src.DataStream[0].Namespace.ValueString())
		}
		opensearch.DataStream = ds
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineOpenSearchDestination: opensearch,
	}
}

func flattenOpenSearchDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineOpenSearchDestination) *opensearchDestinationModel {
	if src == nil {
		return nil
	}

	out := &opensearchDestinationModel{
		BulkIndex: types.StringPointerValue(src.BulkIndex),
	}

	if ds, ok := src.GetDataStreamOk(); ok && ds != nil {
		out.DataStream = []opensearchDestinationDataStreamModel{{
			Dtype:     types.StringPointerValue(ds.Dtype),
			Dataset:   types.StringPointerValue(ds.Dataset),
			Namespace: types.StringPointerValue(ds.Namespace),
		}}
	}

	return out
}

func expandAmazonOpenSearchDestination(ctx context.Context, dest *destinationModel, src *amazonOpenSearchDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	amazonopensearch := datadogV2.NewObservabilityPipelineAmazonOpenSearchDestinationWithDefaults()
	amazonopensearch.SetId(dest.Id.ValueString())

	var inputs []string
	dest.Inputs.ElementsAs(ctx, &inputs, false)
	amazonopensearch.SetInputs(inputs)

	if !src.BulkIndex.IsNull() {
		amazonopensearch.SetBulkIndex(src.BulkIndex.ValueString())
	}

	if len(src.Auth) > 0 {
		authSrc := src.Auth[0]
		auth := datadogV2.ObservabilityPipelineAmazonOpenSearchDestinationAuth{
			Strategy: datadogV2.ObservabilityPipelineAmazonOpenSearchDestinationAuthStrategy(authSrc.Strategy.ValueString()),
		}
		if !authSrc.AwsRegion.IsNull() {
			auth.AwsRegion = authSrc.AwsRegion.ValueStringPointer()
		}
		if !authSrc.AssumeRole.IsNull() {
			auth.AssumeRole = authSrc.AssumeRole.ValueStringPointer()
		}
		if !authSrc.ExternalId.IsNull() {
			auth.ExternalId = authSrc.ExternalId.ValueStringPointer()
		}
		if !authSrc.SessionName.IsNull() {
			auth.SessionName = authSrc.SessionName.ValueStringPointer()
		}
		amazonopensearch.SetAuth(auth)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineAmazonOpenSearchDestination: amazonopensearch,
	}
}

func flattenAmazonOpenSearchDestination(src *datadogV2.ObservabilityPipelineAmazonOpenSearchDestination) *amazonOpenSearchDestinationModel {
	if src == nil {
		return nil
	}

	model := &amazonOpenSearchDestinationModel{}

	if v, ok := src.GetBulkIndexOk(); ok {
		model.BulkIndex = types.StringValue(*v)
	}

	model.Auth = []amazonOpenSearchAuthModel{
		{
			Strategy:    types.StringValue(string(src.Auth.Strategy)),
			AwsRegion:   types.StringPointerValue(src.Auth.AwsRegion),
			AssumeRole:  types.StringPointerValue(src.Auth.AssumeRole),
			ExternalId:  types.StringPointerValue(src.Auth.ExternalId),
			SessionName: types.StringPointerValue(src.Auth.SessionName),
		},
	}

	return model
}
