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

type observabilityPipelineModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Config *configModel `tfsdk:"config"` // config must be a pointer to allow terraform import
}

type configModel struct {
	Sources      sourcesModel      `tfsdk:"sources"`
	Processors   processorsModel   `tfsdk:"processors"`
	Destinations destinationsModel `tfsdk:"destinations"`
}
type sourcesModel struct {
	DatadogAgentSource       []*datadogAgentSourceModel                  `tfsdk:"datadog_agent"`
	KafkaSource              []*kafkaSourceModel                         `tfsdk:"kafka"`
	RsyslogSource            []*rsyslogSourceModel                       `tfsdk:"rsyslog"`
	SyslogNgSource           []*syslogNgSourceModel                      `tfsdk:"syslog_ng"`
	SumoLogicSource          []*sumoLogicSourceModel                     `tfsdk:"sumo_logic"`
	FluentdSource            []*fluentdSourceModel                       `tfsdk:"fluentd"`
	FluentBitSource          []*fluentBitSourceModel                     `tfsdk:"fluent_bit"`
	HttpServerSource         []*httpServerSourceModel                    `tfsdk:"http_server"`
	AmazonS3Source           []*amazonS3SourceModel                      `tfsdk:"amazon_s3"`
	SplunkHecSource          []*splunkHecSourceModel                     `tfsdk:"splunk_hec"`
	SplunkTcpSource          []*splunkTcpSourceModel                     `tfsdk:"splunk_tcp"`
	AmazonDataFirehoseSource []*amazonDataFirehoseSourceModel            `tfsdk:"amazon_data_firehose"`
	HttpClientSource         []*httpClientSourceModel                    `tfsdk:"http_client"`
	GooglePubSubSource       []*googlePubSubSourceModel                  `tfsdk:"google_pubsub"`
	LogstashSource           []*logstashSourceModel                      `tfsdk:"logstash"`
	SocketSource             []*observability_pipeline.SocketSourceModel `tfsdk:"socket"`
}

type logstashSourceModel struct {
	Id  types.String `tfsdk:"id"`
	Tls *tlsModel    `tfsdk:"tls"`
}

type datadogAgentSourceModel struct {
	Id  types.String `tfsdk:"id"`
	Tls *tlsModel    `tfsdk:"tls"`
}

type kafkaSourceModel struct {
	Id                types.String            `tfsdk:"id"`
	GroupId           types.String            `tfsdk:"group_id"`
	Topics            []types.String          `tfsdk:"topics"`
	LibrdkafkaOptions []librdkafkaOptionModel `tfsdk:"librdkafka_option"`
	Sasl              *kafkaSourceSaslModel   `tfsdk:"sasl"`
	Tls               *tlsModel               `tfsdk:"tls"`
}

type librdkafkaOptionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type kafkaSourceSaslModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
}

type amazonS3SourceModel struct {
	Id     types.String                         `tfsdk:"id"`     // Unique identifier for the component
	Region types.String                         `tfsdk:"region"` // AWS region where the S3 bucket resides
	Auth   *observability_pipeline.AwsAuthModel `tfsdk:"auth"`   // AWS authentication credentials
	Tls    *tlsModel                            `tfsdk:"tls"`    // TLS encryption configuration
}

type tlsModel struct {
	CrtFile types.String `tfsdk:"crt_file"`
	CaFile  types.String `tfsdk:"ca_file"`
	KeyFile types.String `tfsdk:"key_file"`
}

// Processor models

type processorsModel struct {
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
}

type ocsfMapperProcessorModel struct {
	Id      types.String       `tfsdk:"id"`
	Include types.String       `tfsdk:"include"`
	Inputs  types.List         `tfsdk:"inputs"`
	Mapping []ocsfMappingModel `tfsdk:"mapping"`
}

type ocsfMappingModel struct {
	Include        types.String `tfsdk:"include"`
	LibraryMapping types.String `tfsdk:"library_mapping"`
}

type enrichmentTableProcessorModel struct {
	Id      types.String          `tfsdk:"id"`
	Include types.String          `tfsdk:"include"`
	Inputs  types.List            `tfsdk:"inputs"`
	Target  types.String          `tfsdk:"target"`
	File    *enrichmentFileModel  `tfsdk:"file"`
	GeoIp   *enrichmentGeoIpModel `tfsdk:"geoip"`
}

type enrichmentFileModel struct {
	Path     types.String          `tfsdk:"path"`
	Encoding fileEncodingModel     `tfsdk:"encoding"`
	Schema   []fileSchemaItemModel `tfsdk:"schema"`
	Key      []fileKeyItemModel    `tfsdk:"key"`
}

type fileEncodingModel struct {
	Type            types.String `tfsdk:"type"`
	Delimiter       types.String `tfsdk:"delimiter"`
	IncludesHeaders types.Bool   `tfsdk:"includes_headers"`
}

type fileSchemaItemModel struct {
	Column types.String `tfsdk:"column"`
	Type   types.String `tfsdk:"type"`
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

type addEnvVarsProcessorModel struct {
	Id        types.String         `tfsdk:"id"`
	Include   types.String         `tfsdk:"include"`
	Inputs    types.List           `tfsdk:"inputs"`
	Variables []envVarMappingModel `tfsdk:"variables"`
}

type envVarMappingModel struct {
	Field types.String `tfsdk:"field"`
	Name  types.String `tfsdk:"name"`
}

type throttleProcessorModel struct {
	Id        types.String   `tfsdk:"id"`
	Include   types.String   `tfsdk:"include"`
	Inputs    types.List     `tfsdk:"inputs"`
	Threshold types.Int64    `tfsdk:"threshold"`
	Window    types.Float64  `tfsdk:"window"`
	GroupBy   []types.String `tfsdk:"group_by"`
}

type reduceProcessorModel struct {
	Id              types.String         `tfsdk:"id"`
	Include         types.String         `tfsdk:"include"`
	Inputs          types.List           `tfsdk:"inputs"`
	GroupBy         []types.String       `tfsdk:"group_by"`
	MergeStrategies []mergeStrategyModel `tfsdk:"merge_strategies"`
}

type mergeStrategyModel struct {
	Path     types.String `tfsdk:"path"`
	Strategy types.String `tfsdk:"strategy"`
}

type dedupeProcessorModel struct {
	Id      types.String   `tfsdk:"id"`
	Include types.String   `tfsdk:"include"`
	Inputs  types.List     `tfsdk:"inputs"`
	Fields  []types.String `tfsdk:"fields"`
	Mode    types.String   `tfsdk:"mode"`
}

type filterProcessorModel struct {
	Id      types.String `tfsdk:"id"`
	Include types.String `tfsdk:"include"`
	Inputs  types.List   `tfsdk:"inputs"`
}

type parseJsonProcessorModel struct {
	Id      types.String `tfsdk:"id"`
	Inputs  types.List   `tfsdk:"inputs"`
	Include types.String `tfsdk:"include"`
	Field   types.String `tfsdk:"field"`
}

type addFieldsProcessor struct {
	Id      types.String `tfsdk:"id"`
	Include types.String `tfsdk:"include"`
	Inputs  types.List   `tfsdk:"inputs"`
	Fields  []fieldValue `tfsdk:"field"`
}

type renameFieldsProcessorModel struct {
	Id      types.String           `tfsdk:"id"`
	Include types.String           `tfsdk:"include"`
	Inputs  types.List             `tfsdk:"inputs"`
	Fields  []renameFieldItemModel `tfsdk:"field"`
}

type renameFieldItemModel struct {
	Source         types.String `tfsdk:"source"`
	Destination    types.String `tfsdk:"destination"`
	PreserveSource types.Bool   `tfsdk:"preserve_source"`
}

type removeFieldsProcessorModel struct {
	Id      types.String `tfsdk:"id"`
	Include types.String `tfsdk:"include"`
	Inputs  types.List   `tfsdk:"inputs"`
	Fields  types.List   `tfsdk:"fields"`
}

type quotaProcessorModel struct {
	Id                          types.String         `tfsdk:"id"`
	Include                     types.String         `tfsdk:"include"`
	Inputs                      types.List           `tfsdk:"inputs"`
	Name                        types.String         `tfsdk:"name"`
	DropEvents                  types.Bool           `tfsdk:"drop_events"`
	Limit                       quotaLimitModel      `tfsdk:"limit"`
	PartitionFields             []types.String       `tfsdk:"partition_fields"`
	IgnoreWhenMissingPartitions types.Bool           `tfsdk:"ignore_when_missing_partitions"`
	Overrides                   []quotaOverrideModel `tfsdk:"overrides"`
	OverflowAction              types.String         `tfsdk:"overflow_action"`
}

type quotaLimitModel struct {
	Enforce types.String `tfsdk:"enforce"` // "bytes" or "events"
	Limit   types.Int64  `tfsdk:"limit"`
}

type quotaOverrideModel struct {
	Fields []fieldValue    `tfsdk:"field"`
	Limit  quotaLimitModel `tfsdk:"limit"`
}

type fieldValue struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// Destination models

type destinationsModel struct {
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
	GoogleChronicleDestination        []*googleChronicleDestinationModel                               `tfsdk:"google_chronicle"`
	NewRelicDestination               []*newRelicDestinationModel                                      `tfsdk:"new_relic"`
	SentinelOneDestination            []*sentinelOneDestinationModel                                   `tfsdk:"sentinel_one"`
	OpenSearchDestination             []*opensearchDestinationModel                                    `tfsdk:"opensearch"`
	AmazonOpenSearchDestination       []*amazonOpenSearchDestinationModel                              `tfsdk:"amazon_opensearch"`
	SocketDestination                 []*observability_pipeline.SocketDestinationModel                 `tfsdk:"socket"`
	AmazonS3Destination               []*observability_pipeline.AmazonS3DestinationModel               `tfsdk:"amazon_s3"`
	AmazonSecurityLakeDestination     []*observability_pipeline.AmazonSecurityLakeDestinationModel     `tfsdk:"amazon_security_lake"`
	CrowdStrikeNextGenSiemDestination []*observability_pipeline.CrowdStrikeNextGenSiemDestinationModel `tfsdk:"crowdstrike_next_gen_siem"`
}

type amazonOpenSearchDestinationModel struct {
	Id        types.String               `tfsdk:"id"`
	Inputs    types.List                 `tfsdk:"inputs"`
	BulkIndex types.String               `tfsdk:"bulk_index"`
	Auth      *amazonOpenSearchAuthModel `tfsdk:"auth"`
}

type amazonOpenSearchAuthModel struct {
	Strategy    types.String `tfsdk:"strategy"`
	AwsRegion   types.String `tfsdk:"aws_region"`
	AssumeRole  types.String `tfsdk:"assume_role"`
	ExternalId  types.String `tfsdk:"external_id"`
	SessionName types.String `tfsdk:"session_name"`
}

type opensearchDestinationModel struct {
	Id        types.String `tfsdk:"id"`
	Inputs    types.List   `tfsdk:"inputs"`
	BulkIndex types.String `tfsdk:"bulk_index"`
}

type sentinelOneDestinationModel struct {
	Id     types.String `tfsdk:"id"`
	Inputs types.List   `tfsdk:"inputs"`
	Region types.String `tfsdk:"region"`
}

type newRelicDestinationModel struct {
	Id     types.String `tfsdk:"id"`
	Inputs types.List   `tfsdk:"inputs"`
	Region types.String `tfsdk:"region"`
}

type googleChronicleDestinationModel struct {
	Id         types.String  `tfsdk:"id"`
	Inputs     types.List    `tfsdk:"inputs"`
	Auth       *gcpAuthModel `tfsdk:"auth"`
	CustomerId types.String  `tfsdk:"customer_id"`
	Encoding   types.String  `tfsdk:"encoding"`
	LogType    types.String  `tfsdk:"log_type"`
}

type googlePubSubDestinationModel struct {
	Id       types.String  `tfsdk:"id"`
	Inputs   types.List    `tfsdk:"inputs"`
	Project  types.String  `tfsdk:"project"`
	Topic    types.String  `tfsdk:"topic"`
	Auth     *gcpAuthModel `tfsdk:"auth"`
	Encoding types.String  `tfsdk:"encoding"`
	Tls      *tlsModel     `tfsdk:"tls"`
}

type datadogLogsDestinationModel struct {
	Id     types.String `tfsdk:"id"`
	Inputs types.List   `tfsdk:"inputs"`
}

type parseGrokProcessorModel struct {
	Id                  types.String                  `tfsdk:"id"`
	Include             types.String                  `tfsdk:"include"`
	Inputs              types.List                    `tfsdk:"inputs"`
	DisableLibraryRules types.Bool                    `tfsdk:"disable_library_rules"`
	Rules               []parseGrokProcessorRuleModel `tfsdk:"rules"`
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
	Id         types.String  `tfsdk:"id"`
	Include    types.String  `tfsdk:"include"`
	Inputs     types.List    `tfsdk:"inputs"`
	Rate       types.Int64   `tfsdk:"rate"`
	Percentage types.Float64 `tfsdk:"percentage"`
}

type fluentdSourceModel struct {
	Id  types.String `tfsdk:"id"`
	Tls *tlsModel    `tfsdk:"tls"`
}

type fluentBitSourceModel struct {
	Id  types.String `tfsdk:"id"`
	Tls *tlsModel    `tfsdk:"tls"`
}

type httpServerSourceModel struct {
	Id           types.String `tfsdk:"id"`
	AuthStrategy types.String `tfsdk:"auth_strategy"`
	Decoding     types.String `tfsdk:"decoding"`
	Tls          *tlsModel    `tfsdk:"tls"`
}

type splunkHecSourceModel struct {
	Id  types.String `tfsdk:"id"`  // The unique identifier for this component.
	Tls *tlsModel    `tfsdk:"tls"` // TLS encryption settings for secure ingestion.
}

type generateMetricsProcessorModel struct {
	Id      types.String           `tfsdk:"id"`
	Include types.String           `tfsdk:"include"`
	Inputs  types.List             `tfsdk:"inputs"`
	Metrics []generatedMetricModel `tfsdk:"metrics"`
}

type generatedMetricModel struct {
	Name       types.String          `tfsdk:"name"`
	Include    types.String          `tfsdk:"include"`
	MetricType types.String          `tfsdk:"metric_type"`
	GroupBy    types.List            `tfsdk:"group_by"`
	Value      *generatedMetricValue `tfsdk:"value"`
}

type generatedMetricValue struct {
	Strategy types.String `tfsdk:"strategy"`
	Field    types.String `tfsdk:"field"`
}

type splunkTcpSourceModel struct {
	Id  types.String `tfsdk:"id"`  // The unique identifier for this component.
	Tls *tlsModel    `tfsdk:"tls"` // TLS encryption settings for secure transmission.
}

type splunkHecDestinationModel struct {
	Id                   types.String `tfsdk:"id"`
	Inputs               types.List   `tfsdk:"inputs"`
	AutoExtractTimestamp types.Bool   `tfsdk:"auto_extract_timestamp"`
	Encoding             types.String `tfsdk:"encoding"`
	Sourcetype           types.String `tfsdk:"sourcetype"`
	Index                types.String `tfsdk:"index"`
}

type gcsDestinationModel struct {
	Id           types.String    `tfsdk:"id"`
	Inputs       types.List      `tfsdk:"inputs"`
	Bucket       types.String    `tfsdk:"bucket"`
	KeyPrefix    types.String    `tfsdk:"key_prefix"`
	StorageClass types.String    `tfsdk:"storage_class"`
	Acl          types.String    `tfsdk:"acl"`
	Auth         gcpAuthModel    `tfsdk:"auth"`
	Metadata     []metadataEntry `tfsdk:"metadata"`
}

type metadataEntry struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type sumoLogicDestinationModel struct {
	Id                   types.String             `tfsdk:"id"`
	Inputs               types.List               `tfsdk:"inputs"`
	Encoding             types.String             `tfsdk:"encoding"`
	HeaderHostName       types.String             `tfsdk:"header_host_name"`
	HeaderSourceName     types.String             `tfsdk:"header_source_name"`
	HeaderSourceCategory types.String             `tfsdk:"header_source_category"`
	HeaderCustomFields   []headerCustomFieldModel `tfsdk:"header_custom_fields"`
}

type headerCustomFieldModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type rsyslogSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Mode types.String `tfsdk:"mode"`
	Tls  *tlsModel    `tfsdk:"tls"`
}

type syslogNgSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Mode types.String `tfsdk:"mode"`
	Tls  *tlsModel    `tfsdk:"tls"`
}

type rsyslogDestinationModel struct {
	Id        types.String `tfsdk:"id"`
	Inputs    types.List   `tfsdk:"inputs"`
	Keepalive types.Int64  `tfsdk:"keepalive"`
	Tls       *tlsModel    `tfsdk:"tls"`
}

type syslogNgDestinationModel struct {
	Id        types.String `tfsdk:"id"`
	Inputs    types.List   `tfsdk:"inputs"`
	Keepalive types.Int64  `tfsdk:"keepalive"`
	Tls       *tlsModel    `tfsdk:"tls"`
}

type elasticsearchDestinationModel struct {
	Id         types.String `tfsdk:"id"`
	Inputs     types.List   `tfsdk:"inputs"`
	ApiVersion types.String `tfsdk:"api_version"`
	BulkIndex  types.String `tfsdk:"bulk_index"`
}

type azureStorageDestinationModel struct {
	Id            types.String `tfsdk:"id"`
	Inputs        types.List   `tfsdk:"inputs"`
	ContainerName types.String `tfsdk:"container_name"`
	BlobPrefix    types.String `tfsdk:"blob_prefix"`
}

type microsoftSentinelDestinationModel struct {
	Id             types.String `tfsdk:"id"`
	Inputs         types.List   `tfsdk:"inputs"`
	ClientId       types.String `tfsdk:"client_id"`
	TenantId       types.String `tfsdk:"tenant_id"`
	DcrImmutableId types.String `tfsdk:"dcr_immutable_id"`
	Table          types.String `tfsdk:"table"`
}

type sensitiveDataScannerProcessorModel struct {
	Id      types.String                        `tfsdk:"id"`
	Include types.String                        `tfsdk:"include"`
	Inputs  types.List                          `tfsdk:"inputs"`
	Rules   []sensitiveDataScannerProcessorRule `tfsdk:"rules"`
}

type sensitiveDataScannerProcessorRule struct {
	Name           types.String                                 `tfsdk:"name"`
	Tags           []types.String                               `tfsdk:"tags"`
	KeywordOptions *sensitiveDataScannerProcessorKeywordOptions `tfsdk:"keyword_options"`
	Pattern        *sensitiveDataScannerProcessorPattern        `tfsdk:"pattern"`
	Scope          *sensitiveDataScannerProcessorScope          `tfsdk:"scope"`
	OnMatch        *sensitiveDataScannerProcessorAction         `tfsdk:"on_match"`
}

// Nested structs (extracted per your preference)
type sensitiveDataScannerProcessorKeywordOptions struct {
	Keywords  []types.String `tfsdk:"keywords"`
	Proximity types.Int64    `tfsdk:"proximity"`
}

type sensitiveDataScannerProcessorPattern struct {
	Custom  *sensitiveDataScannerCustomPattern  `tfsdk:"custom"`
	Library *sensitiveDataScannerLibraryPattern `tfsdk:"library"`
}

type sensitiveDataScannerCustomPattern struct {
	Rule types.String `tfsdk:"rule"`
}

type sensitiveDataScannerLibraryPattern struct {
	Id                     types.String `tfsdk:"id"`
	UseRecommendedKeywords types.Bool   `tfsdk:"use_recommended_keywords"`
}

type sensitiveDataScannerProcessorScope struct {
	Include *sensitiveDataScannerScopeOptions `tfsdk:"include"`
	Exclude *sensitiveDataScannerScopeOptions `tfsdk:"exclude"`
	All     *bool                             `tfsdk:"all"`
}

type sensitiveDataScannerScopeOptions struct {
	Fields []types.String `tfsdk:"fields"`
}

type sensitiveDataScannerProcessorAction struct {
	Redact        *sensitiveDataScannerRedactAction        `tfsdk:"redact"`
	Hash          *sensitiveDataScannerHashAction          `tfsdk:"hash"`
	PartialRedact *sensitiveDataScannerPartialRedactAction `tfsdk:"partial_redact"`
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
	Id types.String `tfsdk:"id"`
}

type amazonDataFirehoseSourceModel struct {
	Id   types.String                         `tfsdk:"id"`
	Auth *observability_pipeline.AwsAuthModel `tfsdk:"auth"`
	Tls  *tlsModel                            `tfsdk:"tls"`
}

type httpClientSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Decoding       types.String `tfsdk:"decoding"`
	ScrapeInterval types.Int64  `tfsdk:"scrape_interval_secs"`
	ScrapeTimeout  types.Int64  `tfsdk:"scrape_timeout_secs"`
	AuthStrategy   types.String `tfsdk:"auth_strategy"`
	Tls            *tlsModel    `tfsdk:"tls"`
}

type googlePubSubSourceModel struct {
	Id           types.String  `tfsdk:"id"`
	Project      types.String  `tfsdk:"project"`
	Subscription types.String  `tfsdk:"subscription"`
	Decoding     types.String  `tfsdk:"decoding"`
	Auth         *gcpAuthModel `tfsdk:"auth"`
	Tls          *tlsModel     `tfsdk:"tls"`
}

type gcpAuthModel struct {
	CredentialsFile types.String `tfsdk:"credentials_file"`
}

type amazonSecurityLakeDestinationModel struct {
	Id               types.String                         `tfsdk:"id"`
	Inputs           types.List                           `tfsdk:"inputs"`
	Bucket           types.String                         `tfsdk:"bucket"`
	Region           types.String                         `tfsdk:"region"`
	CustomSourceName types.String                         `tfsdk:"custom_source_name"`
	Tls              *tlsModel                            `tfsdk:"tls"`
	Auth             *observability_pipeline.AwsAuthModel `tfsdk:"auth"`
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
		Description: "Provides a Datadog Observability Pipeline resource. Observability Pipelines allows you to collect and process logs within your own infrastructure, and then route them to downstream integrations. " +
			"This resource is in **Preview**. Reach out to Datadog support to enable it for your account.   \n\n" +
			"Datadog recommends using the `-parallelism=1` option to apply this resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The pipeline name.",
			},
		},
		Blocks: map[string]schema.Block{
			"config": schema.SingleNestedBlock{
				Description: "Configuration for the pipeline.",
				Blocks: map[string]schema.Block{
					"sources": schema.SingleNestedBlock{
						Description: "List of sources.",
						Blocks: map[string]schema.Block{
							"datadog_agent": schema.ListNestedBlock{
								Description: "The `datadog_agent` source collects logs from the Datadog Agent.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the source.",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"kafka": schema.ListNestedBlock{
								Description: "The `kafka` source ingests data from Apache Kafka topics.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the source.",
										},
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
										"sasl": schema.SingleNestedBlock{
											Description: "SASL authentication settings.",
											Attributes: map[string]schema.Attribute{
												"mechanism": schema.StringAttribute{
													Optional:    true, // must be optional to make the block optional
													Description: "SASL mechanism to use (e.g., PLAIN, SCRAM-SHA-256, SCRAM-SHA-512).",
													Validators: []validator.String{
														stringvalidator.OneOf("PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"),
													},
												},
											},
										},
										"tls": tlsSchema(),
									},
								},
							},
							"fluentd": schema.ListNestedBlock{
								Description: "The `fluent` source ingests logs from a Fluentd-compatible service.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (for example, as the `input` to downstream components).",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"fluent_bit": schema.ListNestedBlock{
								Description: "The `fluent` source ingests logs from Fluent Bit.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (for example, as the `input` to downstream components).",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"http_server": schema.ListNestedBlock{
								Description: "The `http_server` source collects logs over HTTP POST from external services.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "Unique ID for the HTTP server source.",
										},
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
										"tls": tlsSchema(),
									},
								},
							},
							"amazon_s3": schema.ListNestedBlock{
								Description: "The `amazon_s3` source ingests logs from an Amazon S3 bucket. It supports AWS authentication and TLS encryption.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
										"region": schema.StringAttribute{
											Required:    true,
											Description: "AWS region where the S3 bucket resides.",
										},
									},
									Blocks: map[string]schema.Block{
										"auth": schema.SingleNestedBlock{
											Description: "AWS authentication credentials used for accessing AWS services such as S3. If omitted, the system's default credentials are used (for example, the IAM role and environment variables).",
											Attributes: map[string]schema.Attribute{
												"assume_role": schema.StringAttribute{
													Optional:    true,
													Description: "The Amazon Resource Name (ARN) of the role to assume.",
												},
												"external_id": schema.StringAttribute{
													Optional:    true,
													Description: "A unique identifier for cross-account role assumption.",
												},
												"session_name": schema.StringAttribute{
													Optional:    true,
													Description: "A session identifier used for logging and tracing the assumed role session.",
												},
											},
										},
										"tls": tlsSchema(),
									},
								},
							},
							"splunk_hec": schema.ListNestedBlock{
								Description: "The `splunk_hec` source implements the Splunk HTTP Event Collector (HEC) API.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"splunk_tcp": schema.ListNestedBlock{
								Description: "The `splunk_tcp` source receives logs from a Splunk Universal Forwarder over TCP. TLS is supported for secure transmission.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"rsyslog": schema.ListNestedBlock{
								Description: "The `rsyslog` source listens for logs over TCP or UDP from an `rsyslog` server using the syslog protocol.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
										"mode": schema.StringAttribute{
											Optional:    true,
											Description: "Protocol used by the syslog source to receive messages.",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"syslog_ng": schema.ListNestedBlock{
								Description: "The `syslog_ng` source listens for logs over TCP or UDP from a `syslog-ng` server using the syslog protocol.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
										"mode": schema.StringAttribute{
											Optional:    true,
											Description: "Protocol used by the syslog source to receive messages.",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"sumo_logic": schema.ListNestedBlock{
								Description: "The `sumo_logic` source receives logs from Sumo Logic collectors.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
									},
								},
							},
							"amazon_data_firehose": schema.ListNestedBlock{
								Description: "The `amazon_data_firehose` source ingests logs from AWS Data Firehose.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
									},
									Blocks: map[string]schema.Block{
										"auth": schema.SingleNestedBlock{
											Description: "AWS authentication credentials used for accessing AWS services such as S3. If omitted, the system's default credentials are used (for example, the IAM role and environment variables).",
											Attributes: map[string]schema.Attribute{
												"assume_role": schema.StringAttribute{
													Optional:    true,
													Description: "The Amazon Resource Name (ARN) of the role to assume.",
												},
												"external_id": schema.StringAttribute{
													Optional:    true,
													Description: "A unique identifier for cross-account role assumption.",
												},
												"session_name": schema.StringAttribute{
													Optional:    true,
													Description: "A session identifier used for logging and tracing the assumed role session.",
												},
											},
										},
										"tls": tlsSchema(),
									},
								},
							},
							"http_client": schema.ListNestedBlock{
								Description: "The `http_client` source scrapes logs from HTTP endpoints at regular intervals.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
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
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"google_pubsub": schema.ListNestedBlock{
								Description: "The `google_pubsub` source ingests logs from a Google Cloud Pub/Sub subscription.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
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
										"auth": schema.SingleNestedBlock{
											Description: "GCP credentials used to authenticate with Google Cloud Storage.",
											Attributes: map[string]schema.Attribute{
												"credentials_file": schema.StringAttribute{
													Required:    true,
													Description: "Path to the GCP service account key file.",
												},
											},
										},
										"tls": tlsSchema(),
									},
								},
							},
							"logstash": schema.ListNestedBlock{
								Description: "The `logstash` source ingests logs from a Logstash forwarder.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"socket": observability_pipeline.SocketSourceSchema(),
						},
					},
					"processors": schema.SingleNestedBlock{
						Description: "List of processors.",
						Blocks: map[string]schema.Block{
							"filter": schema.ListNestedBlock{
								Description: "The `filter` processor allows conditional processing of logs based on a Datadog search query. Logs that match the `include` query are passed through; others are discarded.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs should pass through the filter. Logs that match this query continue to downstream components; others are dropped.",
										},
										"inputs": schema.ListAttribute{
											Description: "The inputs for the processor.",
											ElementType: types.StringType,
											Required:    true,
										},
									},
								},
							},
							"parse_json": schema.ListNestedBlock{
								Description: "The `parse_json` processor extracts JSON from a specified field and flattens it into the event. This is useful when logs contain embedded JSON as a string.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "The inputs for the processor.",
											ElementType: types.StringType,
										},
										"field": schema.StringAttribute{
											Required:    true,
											Description: "The field to parse.",
										},
									},
								},
							},
							"add_fields": schema.ListNestedBlock{
								Description: "The `add_fields` processor adds static key-value fields to logs.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Description: "The inputs for the processor.",
											ElementType: types.StringType,
											Required:    true,
										},
									},
									Blocks: map[string]schema.Block{
										"field": schema.ListNestedBlock{
											Validators: []validator.List{
												// this is the only way to make the list of fields required in Terraform
												listvalidator.SizeAtLeast(1),
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
							"rename_fields": schema.ListNestedBlock{
								Description: "The `rename_fields` processor changes field names.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "The inputs for the processor.",
											ElementType: types.StringType,
										},
									},
									Blocks: map[string]schema.Block{
										"field": schema.ListNestedBlock{
											Validators: []validator.List{
												// this is the only way to make the list of fields required in Terraform
												listvalidator.SizeAtLeast(1),
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
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "The inputs for the processor.",
											ElementType: types.StringType,
										},
										"fields": schema.ListAttribute{
											Required:    true,
											Description: "List of fields to remove from the events.",
											ElementType: types.StringType,
										},
									},
								},
							},
							"quota": schema.ListNestedBlock{
								Description: "The `quota` measures logging traffic for logs that match a specified filter. When the configured daily quota is met, the processor can drop or alert.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "The inputs for the processor.",
										},
										"name": schema.StringAttribute{
											Required:    true,
											Description: "The name of the quota.",
										},
										"drop_events": schema.BoolAttribute{
											Required:    true,
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
									},
									Blocks: map[string]schema.Block{
										"limit": schema.SingleNestedBlock{
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
										"overrides": schema.ListNestedBlock{
											Description: "The overrides for field-specific quotas.",
											NestedObject: schema.NestedBlockObject{
												Blocks: map[string]schema.Block{
													"limit": schema.SingleNestedBlock{
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
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
											ElementType: types.StringType,
										},
									},
									Blocks: map[string]schema.Block{
										"rules": schema.ListNestedBlock{
											Description: "A list of rules for identifying and acting on sensitive data patterns.",
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Optional:    true,
														Description: "A name identifying the rule.",
													},
													"tags": schema.ListAttribute{
														Optional:    true,
														ElementType: types.StringType,
														Description: "Tags assigned to this rule for filtering and classification.",
													},
												},
												Blocks: map[string]schema.Block{
													"keyword_options": schema.SingleNestedBlock{
														Description: "Keyword-based proximity matching for sensitive data.",
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
													"pattern": schema.SingleNestedBlock{
														Description: "Pattern detection configuration for identifying sensitive data using either a custom regex or a library reference.",
														Blocks: map[string]schema.Block{
															"custom": schema.SingleNestedBlock{
																Description: "Pattern detection using a custom regular expression.",
																Attributes: map[string]schema.Attribute{
																	"rule": schema.StringAttribute{
																		Optional:    true,
																		Description: "A regular expression used to detect sensitive values. Must be a valid regex.",
																	},
																},
															},
															"library": schema.SingleNestedBlock{
																Description: "Pattern detection using a predefined pattern from the sensitive data scanner pattern library.",
																Attributes: map[string]schema.Attribute{
																	"id": schema.StringAttribute{
																		Optional:    true,
																		Description: "Identifier for a predefined pattern from the sensitive data scanner pattern library.",
																	},
																	"use_recommended_keywords": schema.BoolAttribute{
																		Optional:    true,
																		Description: "Whether to augment the pattern with recommended keywords (optional).",
																	},
																},
															},
														},
													},
													"scope": schema.SingleNestedBlock{
														Description: "Field-level targeting options that determine where the scanner should operate.",
														Blocks: map[string]schema.Block{
															"include": schema.SingleNestedBlock{
																Description: "Explicitly include these fields for scanning.",
																Attributes: map[string]schema.Attribute{
																	"fields": schema.ListAttribute{
																		Optional:    true,
																		ElementType: types.StringType,
																		Description: "The fields to include in scanning.",
																	},
																},
															},
															"exclude": schema.SingleNestedBlock{
																Description: "Explicitly exclude these fields from scanning.",
																Attributes: map[string]schema.Attribute{
																	"fields": schema.ListAttribute{
																		Optional:    true,
																		ElementType: types.StringType,
																		Description: "The fields to exclude from scanning.",
																	},
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
													"on_match": schema.SingleNestedBlock{
														Description: "The action to take when a sensitive value is found.",
														Blocks: map[string]schema.Block{
															"redact": schema.SingleNestedBlock{
																Description: "Redacts the matched value.",
																Attributes: map[string]schema.Attribute{
																	"replace": schema.StringAttribute{
																		Optional:    true,
																		Description: "Replacement string for redacted values (e.g., `***`).",
																	},
																},
															},
															"hash": schema.SingleNestedBlock{
																Description: "Hashes the matched value.",
																Attributes:  map[string]schema.Attribute{}, // empty options
															},
															"partial_redact": schema.SingleNestedBlock{
																Description: "Redacts part of the matched value (e.g., keep last 4 characters).",
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
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this processor.",
										},
									},
									Blocks: map[string]schema.Block{
										"metrics": schema.ListNestedBlock{
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
													"value": schema.SingleNestedBlock{
														Description: "Specifies how the value of the generated metric is computed.",
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
												},
											},
										},
									},
								},
							},
							"parse_grok": schema.ListNestedBlock{
								Description: "The `parse_grok` processor extracts structured fields from unstructured log messages using Grok patterns.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "A unique identifier for this processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
										},
										"disable_library_rules": schema.BoolAttribute{
											Optional:    true,
											Description: "If set to `true`, disables the default Grok rules provided by Datadog.",
										},
									},
									Blocks: map[string]schema.Block{
										"rules": schema.ListNestedBlock{
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
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (for example, as the `input` to downstream components).",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
										},
										"rate": schema.Int64Attribute{
											Optional:    true,
											Description: "Number of events to sample (1 in N).",
										},
										"percentage": schema.Float64Attribute{
											Optional:    true,
											Description: "The percentage of logs to sample.",
										},
									},
								},
							},
							"dedupe": schema.ListNestedBlock{
								Description: "The `dedupe` processor removes duplicate fields in log events.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the input for this processor.",
										},
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
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the input for this processor.",
										},
										"group_by": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of fields used to group log events for merging.",
										},
									},
									Blocks: map[string]schema.Block{
										"merge_strategies": schema.ListNestedBlock{
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
							"throttle": schema.ListNestedBlock{
								Description: "The `throttle` processor limits the number of events that pass through over a given time window.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the input for this processor.",
										},
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
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this processor in the pipeline.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the input for this processor.",
										},
									},
									Blocks: map[string]schema.Block{
										"variables": schema.ListNestedBlock{
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
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this processor.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "A Datadog search query used to determine which logs this processor targets.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the input for this processor.",
										},
										"target": schema.StringAttribute{
											Required:    true,
											Description: "Path where enrichment results should be stored in the log.",
										},
									},
									Blocks: map[string]schema.Block{
										"file": schema.SingleNestedBlock{
											Description: "Defines a static enrichment table loaded from a CSV file.",
											Attributes: map[string]schema.Attribute{
												"path": schema.StringAttribute{
													Optional:    true,
													Description: "Path to the CSV file.",
												},
											},
											Blocks: map[string]schema.Block{
												"encoding": schema.SingleNestedBlock{
													Attributes: map[string]schema.Attribute{
														"type": schema.StringAttribute{
															Optional:    true,
															Description: "File encoding format.",
														},
														"delimiter": schema.StringAttribute{
															Optional:    true,
															Description: "The `encoding` `delimiter`.",
														},
														"includes_headers": schema.BoolAttribute{
															Optional:    true,
															Description: "The `encoding` `includes_headers`.",
														},
													},
												},
												"schema": schema.ListNestedBlock{
													Description: "Schema defining column names and their types.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"column": schema.StringAttribute{
																Optional:    true,
																Description: "The `items` `column`.",
															},
															"type": schema.StringAttribute{
																Optional:    true,
																Description: "The type of the column (e.g. string, boolean, integer, etc.).",
															},
														},
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
										"geoip": schema.SingleNestedBlock{
											Description: "Uses a GeoIP database to enrich logs based on an IP field.",
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
									},
								},
							},
							"ocsf_mapper": schema.ListNestedBlock{
								Description: "The `ocsf_mapper` processor transforms logs into the OCSF schema using predefined library mappings.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"include": schema.StringAttribute{
											Required:    true,
											Description: "Search query to select logs.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "List of component IDs whose output is used as input.",
										},
									},
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
						},
					},
					"destinations": schema.SingleNestedBlock{
						Description: "List of destinations.",
						Blocks: map[string]schema.Block{
							"datadog_logs": schema.ListNestedBlock{
								Description: "The `datadog_logs` destination forwards logs to Datadog Log Management.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique ID of the destination.",
										},
										"inputs": schema.ListAttribute{
											Description: "The inputs for the destination.",
											ElementType: types.StringType,
											Required:    true,
										},
									},
								},
							},
							"google_cloud_storage": schema.ListNestedBlock{
								Description: "The `google_cloud_storage` destination stores logs in a Google Cloud Storage (GCS) bucket.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "Unique identifier for the destination component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
										},
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
											Required:    true,
											Description: "Access control list setting for objects written to the bucket.",
										},
									},
									Blocks: map[string]schema.Block{
										"auth": schema.SingleNestedBlock{
											Description: "GCP credentials used to authenticate with Google Cloud Storage.",
											Attributes: map[string]schema.Attribute{
												"credentials_file": schema.StringAttribute{
													Required:    true,
													Description: "Path to the GCP service account key file.",
												},
											},
										},
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
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
										},
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
										"auth": schema.SingleNestedBlock{
											Description: "GCP credentials used to authenticate with Google Cloud Pub/Sub.",
											Attributes: map[string]schema.Attribute{
												"credentials_file": schema.StringAttribute{
													Optional:    true,
													Description: "Path to the GCP service account key file.",
												},
											},
										},
										"tls": tlsSchema(),
									},
								},
							},
							"splunk_hec": schema.ListNestedBlock{
								Description: "The `splunk_hec` destination forwards logs to Splunk using the HTTP Event Collector (HEC).",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
										},
										"auto_extract_timestamp": schema.BoolAttribute{
											Optional:    true,
											Description: "If `true`, Splunk tries to extract timestamps from incoming log events.",
										},
										"encoding": schema.StringAttribute{
											Optional:    true,
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
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
											ElementType: types.StringType,
										},
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
										"header_custom_fields": schema.ListNestedBlock{
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
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
											ElementType: types.StringType,
										},
										"keepalive": schema.Int64Attribute{
											Optional:    true,
											Description: "Optional socket keepalive duration in milliseconds.",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"syslog_ng": schema.ListNestedBlock{
								Description: "The `syslog_ng` destination forwards logs to an external `syslog-ng` server over TCP or UDP using the syslog protocol.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
											ElementType: types.StringType,
										},
										"keepalive": schema.Int64Attribute{
											Optional:    true,
											Description: "Optional socket keepalive duration in milliseconds.",
										},
									},
									Blocks: map[string]schema.Block{
										"tls": tlsSchema(),
									},
								},
							},
							"elasticsearch": schema.ListNestedBlock{
								Description: "The `elasticsearch` destination writes logs to an Elasticsearch cluster.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
											ElementType: types.StringType,
										},
										"api_version": schema.StringAttribute{
											Optional:    true,
											Description: "The Elasticsearch API version to use. Set to `auto` to auto-detect.",
										},
										"bulk_index": schema.StringAttribute{
											Optional:    true,
											Description: "The index or datastream to write logs to in Elasticsearch.",
										},
									},
								},
							},
							"opensearch": schema.ListNestedBlock{
								Description: "The `opensearch` destination writes logs to an OpenSearch cluster.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as input.",
										},
										"bulk_index": schema.StringAttribute{
											Optional:    true,
											Description: "The index or datastream to write logs to.",
										},
									},
								},
							},
							"amazon_opensearch": schema.ListNestedBlock{
								Description: "The `amazon_opensearch` destination writes logs to Amazon OpenSearch.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the input for this component.",
										},
										"bulk_index": schema.StringAttribute{
											Optional:    true,
											Description: "The index or datastream to write logs to.",
										},
									},
									Blocks: map[string]schema.Block{
										"auth": schema.SingleNestedBlock{
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
									},
								},
							},
							"azure_storage": schema.ListNestedBlock{
								Description: "The `azure_storage` destination forwards logs to an Azure Blob Storage container.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
											ElementType: types.StringType,
										},
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
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
											ElementType: types.StringType,
										},
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
							"google_chronicle": schema.ListNestedBlock{
								Description: "The `google_chronicle` destination sends logs to Google Chronicle.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
										},
										"customer_id": schema.StringAttribute{
											Optional:    true,
											Description: "The Google Chronicle customer ID.",
										},
										"encoding": schema.StringAttribute{
											Optional:    true,
											Description: "The encoding format for the logs sent to Chronicle.",
										},
										"log_type": schema.StringAttribute{
											Optional:    true,
											Description: "The log type metadata associated with the Chronicle destination.",
										},
									},
									Blocks: map[string]schema.Block{
										"auth": schema.SingleNestedBlock{
											Description: "GCP credentials used to authenticate with Google Cloud Storage.",
											Attributes: map[string]schema.Attribute{
												"credentials_file": schema.StringAttribute{
													Optional:    true,
													Description: "Path to the GCP service account key file.",
												},
											},
										},
									},
								},
							},
							"new_relic": schema.ListNestedBlock{
								Description: "The `new_relic` destination sends logs to the New Relic platform.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
										},
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
										"id": schema.StringAttribute{
											Required:    true,
											Description: "The unique identifier for this component.",
										},
										"inputs": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "A list of component IDs whose output is used as the `input` for this component.",
										},
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
						},
					},
				},
			},
		},
	}
}

func tlsSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Configuration for enabling TLS encryption between the pipeline component and external services.",
		Attributes: map[string]schema.Attribute{
			"crt_file": schema.StringAttribute{
				Optional:    true, // must be optional to make the block optional
				Description: "Path to the TLS client certificate file used to authenticate the pipeline component with upstream or downstream services.",
			},
			"ca_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to the Certificate Authority (CA) file used to validate the server's TLS certificate.",
			},
			"key_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to the private key file associated with the TLS client certificate. Used for mutual TLS authentication.",
			},
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

	// Sources
	for _, s := range state.Config.Sources.DatadogAgentSource {
		config.Sources = append(config.Sources, expandDatadogAgentSource(s))
	}
	for _, k := range state.Config.Sources.KafkaSource {
		config.Sources = append(config.Sources, expandKafkaSource(k))
	}
	for _, f := range state.Config.Sources.FluentdSource {
		config.Sources = append(config.Sources, expandFluentdSource(f))
	}
	for _, f := range state.Config.Sources.FluentBitSource {
		config.Sources = append(config.Sources, expandFluentBitSource(f))
	}
	for _, s := range state.Config.Sources.HttpServerSource {
		config.Sources = append(config.Sources, expandHttpServerSource(s))
	}
	for _, s := range state.Config.Sources.SplunkHecSource {
		config.Sources = append(config.Sources, expandSplunkHecSource(s))
	}
	for _, s := range state.Config.Sources.SplunkTcpSource {
		config.Sources = append(config.Sources, expandSplunkTcpSource(s))
	}
	for _, s := range state.Config.Sources.AmazonS3Source {
		config.Sources = append(config.Sources, expandAmazonS3Source(s))
	}
	for _, s := range state.Config.Sources.RsyslogSource {
		config.Sources = append(config.Sources, expandRsyslogSource(s))
	}
	for _, s := range state.Config.Sources.SyslogNgSource {
		config.Sources = append(config.Sources, expandSyslogNgSource(s))
	}
	for _, s := range state.Config.Sources.SumoLogicSource {
		config.Sources = append(config.Sources, expandSumoLogicSource(s))
	}
	for _, a := range state.Config.Sources.AmazonDataFirehoseSource {
		config.Sources = append(config.Sources, expandAmazonDataFirehoseSource(a))
	}
	for _, h := range state.Config.Sources.HttpClientSource {
		config.Sources = append(config.Sources, expandHttpClientSource(h))
	}
	for _, g := range state.Config.Sources.GooglePubSubSource {
		config.Sources = append(config.Sources, expandGooglePubSubSource(g))
	}
	for _, l := range state.Config.Sources.LogstashSource {
		config.Sources = append(config.Sources, expandLogstashSource(l))
	}
	for _, s := range state.Config.Sources.SocketSource {
		config.Sources = append(config.Sources, observability_pipeline.ExpandSocketSource(s))
	}

	// Processors
	for _, p := range state.Config.Processors.FilterProcessor {
		config.Processors = append(config.Processors, expandFilterProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.ParseJsonProcessor {
		config.Processors = append(config.Processors, expandParseJsonProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.AddFieldsProcessor {
		config.Processors = append(config.Processors, expandAddFieldsProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.RenameFieldsProcessor {
		config.Processors = append(config.Processors, expandRenameFieldsProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.RemoveFieldsProcessor {
		config.Processors = append(config.Processors, expandRemoveFieldsProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.QuotaProcessor {
		config.Processors = append(config.Processors, expandQuotaProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.DedupeProcessor {
		config.Processors = append(config.Processors, expandDedupeProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.ReduceProcessor {
		config.Processors = append(config.Processors, expandReduceProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.ThrottleProcessor {
		config.Processors = append(config.Processors, expandThrottleProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.AddEnvVarsProcessor {
		config.Processors = append(config.Processors, expandAddEnvVarsProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.EnrichmentTableProcessor {
		config.Processors = append(config.Processors, expandEnrichmentTableProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.OcsfMapperProcessor {
		config.Processors = append(config.Processors, expandOcsfMapperProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.CustomProcessor {
		config.Processors = append(config.Processors, observability_pipeline.ExpandCustomProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.ParseGrokProcessor {
		config.Processors = append(config.Processors, expandParseGrokProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.SampleProcessor {
		config.Processors = append(config.Processors, expandSampleProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.GenerateMetricsProcessor {
		config.Processors = append(config.Processors, expandGenerateMetricsProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.SensitiveDataScannerProcessor {
		config.Processors = append(config.Processors, expandSensitiveDataScannerProcessor(ctx, p))
	}
	for _, p := range state.Config.Processors.DatadogTagsProcessor {
		config.Processors = append(config.Processors, expandDatadogTagsProcessor(ctx, p))
	}

	// Destinations
	for _, d := range state.Config.Destinations.DatadogLogsDestination {
		config.Destinations = append(config.Destinations, expandDatadogLogsDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.SplunkHecDestination {
		config.Destinations = append(config.Destinations, expandSplunkHecDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.GoogleCloudStorageDestination {
		config.Destinations = append(config.Destinations, expandGoogleCloudStorageDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.GooglePubSubDestination {
		config.Destinations = append(config.Destinations, expandGooglePubSubDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.SumoLogicDestination {
		config.Destinations = append(config.Destinations, expandSumoLogicDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.RsyslogDestination {
		config.Destinations = append(config.Destinations, expandRsyslogDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.SyslogNgDestination {
		config.Destinations = append(config.Destinations, expandSyslogNgDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.ElasticsearchDestination {
		config.Destinations = append(config.Destinations, expandElasticsearchDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.AzureStorageDestination {
		config.Destinations = append(config.Destinations, expandAzureStorageDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.MicrosoftSentinelDestination {
		config.Destinations = append(config.Destinations, expandMicrosoftSentinelDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.GoogleChronicleDestination {
		config.Destinations = append(config.Destinations, expandGoogleChronicleDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.NewRelicDestination {
		config.Destinations = append(config.Destinations, expandNewRelicDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.SentinelOneDestination {
		config.Destinations = append(config.Destinations, expandSentinelOneDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.OpenSearchDestination {
		config.Destinations = append(config.Destinations, expandOpenSearchDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.AmazonOpenSearchDestination {
		config.Destinations = append(config.Destinations, expandAmazonOpenSearchDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.SocketDestination {
		config.Destinations = append(config.Destinations, observability_pipeline.ExpandSocketDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.AmazonS3Destination {
		config.Destinations = append(config.Destinations, observability_pipeline.ExpandAmazonS3Destination(ctx, d))
	}
	for _, d := range state.Config.Destinations.AmazonSecurityLakeDestination {
		config.Destinations = append(config.Destinations, observability_pipeline.ExpandObservabilityPipelinesAmazonSecurityLakeDestination(ctx, d))
	}
	for _, d := range state.Config.Destinations.CrowdStrikeNextGenSiemDestination {
		config.Destinations = append(config.Destinations, observability_pipeline.ExpandCrowdStrikeNextGenSiemDestination(ctx, d))
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

	for _, src := range cfg.GetSources() {

		if a := flattenDatadogAgentSource(src.ObservabilityPipelineDatadogAgentSource); a != nil {
			outCfg.Sources.DatadogAgentSource = append(outCfg.Sources.DatadogAgentSource, a)
		}
		if k := flattenKafkaSource(src.ObservabilityPipelineKafkaSource); k != nil {
			outCfg.Sources.KafkaSource = append(outCfg.Sources.KafkaSource, k)
		}
		if f := flattenFluentdSource(src.ObservabilityPipelineFluentdSource); f != nil {
			outCfg.Sources.FluentdSource = append(outCfg.Sources.FluentdSource, f)
		}
		if f := flattenFluentBitSource(src.ObservabilityPipelineFluentBitSource); f != nil {
			outCfg.Sources.FluentBitSource = append(outCfg.Sources.FluentBitSource, f)
		}
		if s := flattenHttpServerSource(src.ObservabilityPipelineHttpServerSource); s != nil {
			outCfg.Sources.HttpServerSource = append(outCfg.Sources.HttpServerSource, s)
		}

		if s := flattenSplunkHecSource(src.ObservabilityPipelineSplunkHecSource); s != nil {
			outCfg.Sources.SplunkHecSource = append(outCfg.Sources.SplunkHecSource, s)
		}

		if s := flattenSplunkTcpSource(src.ObservabilityPipelineSplunkTcpSource); s != nil {
			outCfg.Sources.SplunkTcpSource = append(outCfg.Sources.SplunkTcpSource, s)
		}

		if s3 := flattenAmazonS3Source(src.ObservabilityPipelineAmazonS3Source); s3 != nil {
			outCfg.Sources.AmazonS3Source = append(outCfg.Sources.AmazonS3Source, s3)
		}
		if r := flattenRsyslogSource(src.ObservabilityPipelineRsyslogSource); r != nil {
			outCfg.Sources.RsyslogSource = append(outCfg.Sources.RsyslogSource, r)
		}
		if s := flattenSyslogNgSource(src.ObservabilityPipelineSyslogNgSource); s != nil {
			outCfg.Sources.SyslogNgSource = append(outCfg.Sources.SyslogNgSource, s)
		}
		if s := flattenSumoLogicSource(src.ObservabilityPipelineSumoLogicSource); s != nil {
			outCfg.Sources.SumoLogicSource = append(outCfg.Sources.SumoLogicSource, s)
		}
		if f := flattenAmazonDataFirehoseSource(src.ObservabilityPipelineAmazonDataFirehoseSource); f != nil {
			outCfg.Sources.AmazonDataFirehoseSource = append(outCfg.Sources.AmazonDataFirehoseSource, f)
		}
		if h := flattenHttpClientSource(src.ObservabilityPipelineHttpClientSource); h != nil {
			outCfg.Sources.HttpClientSource = append(outCfg.Sources.HttpClientSource, h)
		}
		if g := flattenGooglePubSubSource(src.ObservabilityPipelineGooglePubSubSource); g != nil {
			outCfg.Sources.GooglePubSubSource = append(outCfg.Sources.GooglePubSubSource, g)
		}
		if l := flattenLogstashSource(src.ObservabilityPipelineLogstashSource); l != nil {
			outCfg.Sources.LogstashSource = append(outCfg.Sources.LogstashSource, l)
		}
		if s := observability_pipeline.FlattenSocketSource(src.ObservabilityPipelineSocketSource); s != nil {
			outCfg.Sources.SocketSource = append(outCfg.Sources.SocketSource, s)
		}
	}

	for _, p := range cfg.GetProcessors() {

		if f := flattenFilterProcessor(ctx, p.ObservabilityPipelineFilterProcessor); f != nil {
			outCfg.Processors.FilterProcessor = append(outCfg.Processors.FilterProcessor, f)
		}

		if f := flattenParseJsonProcessor(ctx, p.ObservabilityPipelineParseJSONProcessor); f != nil {
			outCfg.Processors.ParseJsonProcessor = append(outCfg.Processors.ParseJsonProcessor, f)
		}

		if f := flattenAddFieldsProcessor(ctx, p.ObservabilityPipelineAddFieldsProcessor); f != nil {
			outCfg.Processors.AddFieldsProcessor = append(outCfg.Processors.AddFieldsProcessor, f)
		}

		if f := flattenRenameFieldsProcessor(ctx, p.ObservabilityPipelineRenameFieldsProcessor); f != nil {
			outCfg.Processors.RenameFieldsProcessor = append(outCfg.Processors.RenameFieldsProcessor, f)
		}

		if f := flattenRemoveFieldsProcessor(ctx, p.ObservabilityPipelineRemoveFieldsProcessor); f != nil {
			outCfg.Processors.RemoveFieldsProcessor = append(outCfg.Processors.RemoveFieldsProcessor, f)
		}

		if f := flattenQuotaProcessor(ctx, p.ObservabilityPipelineQuotaProcessor); f != nil {
			outCfg.Processors.QuotaProcessor = append(outCfg.Processors.QuotaProcessor, f)
		}
		if f := flattenParseGrokProcessor(ctx, p.ObservabilityPipelineParseGrokProcessor); f != nil {
			outCfg.Processors.ParseGrokProcessor = append(outCfg.Processors.ParseGrokProcessor, f)
		}
		if s := flattenSampleProcessor(ctx, p.ObservabilityPipelineSampleProcessor); s != nil {
			outCfg.Processors.SampleProcessor = append(outCfg.Processors.SampleProcessor, s)
		}

		if f := flattenGenerateMetricsProcessor(ctx, p.ObservabilityPipelineGenerateMetricsProcessor); f != nil {
			outCfg.Processors.GenerateMetricsProcessor = append(outCfg.Processors.GenerateMetricsProcessor, f)
		}

		if s := flattenSensitiveDataScannerProcessor(ctx, p.ObservabilityPipelineSensitiveDataScannerProcessor); s != nil {
			outCfg.Processors.SensitiveDataScannerProcessor = append(outCfg.Processors.SensitiveDataScannerProcessor, s)
		}
		if f := flattenDedupeProcessor(ctx, p.ObservabilityPipelineDedupeProcessor); f != nil {
			outCfg.Processors.DedupeProcessor = append(outCfg.Processors.DedupeProcessor, f)
		}
		if f := flattenReduceProcessor(ctx, p.ObservabilityPipelineReduceProcessor); f != nil {
			outCfg.Processors.ReduceProcessor = append(outCfg.Processors.ReduceProcessor, f)
		}
		if f := flattenThrottleProcessor(ctx, p.ObservabilityPipelineThrottleProcessor); f != nil {
			outCfg.Processors.ThrottleProcessor = append(outCfg.Processors.ThrottleProcessor, f)
		}
		if f := flattenAddEnvVarsProcessor(ctx, p.ObservabilityPipelineAddEnvVarsProcessor); f != nil {
			outCfg.Processors.AddEnvVarsProcessor = append(outCfg.Processors.AddEnvVarsProcessor, f)
		}
		if f := flattenEnrichmentTableProcessor(ctx, p.ObservabilityPipelineEnrichmentTableProcessor); f != nil {
			outCfg.Processors.EnrichmentTableProcessor = append(outCfg.Processors.EnrichmentTableProcessor, f)
		}
		if f := flattenOcsfMapperProcessor(ctx, p.ObservabilityPipelineOcsfMapperProcessor); f != nil {
			outCfg.Processors.OcsfMapperProcessor = append(outCfg.Processors.OcsfMapperProcessor, f)
		}
		if f := flattenDatadogTagsProcessor(ctx, p.ObservabilityPipelineDatadogTagsProcessor); f != nil {
			outCfg.Processors.DatadogTagsProcessor = append(outCfg.Processors.DatadogTagsProcessor, f)
		}
		if f := observability_pipeline.FlattenCustomProcessor(ctx, p.ObservabilityPipelineCustomProcessor); f != nil {
			outCfg.Processors.CustomProcessor = append(outCfg.Processors.CustomProcessor, f)
		}
	}

	for _, d := range cfg.GetDestinations() {

		if logs := flattenDatadogLogsDestination(ctx, d.ObservabilityPipelineDatadogLogsDestination); logs != nil {
			outCfg.Destinations.DatadogLogsDestination = append(outCfg.Destinations.DatadogLogsDestination, logs)
		}
		if d := flattenGoogleChronicleDestination(ctx, d.ObservabilityPipelineGoogleChronicleDestination); d != nil {
			outCfg.Destinations.GoogleChronicleDestination = append(outCfg.Destinations.GoogleChronicleDestination, d)
		}
		if d := flattenNewRelicDestination(ctx, d.ObservabilityPipelineNewRelicDestination); d != nil {
			outCfg.Destinations.NewRelicDestination = append(outCfg.Destinations.NewRelicDestination, d)
		}
		if d := flattenSentinelOneDestination(ctx, d.ObservabilityPipelineSentinelOneDestination); d != nil {
			outCfg.Destinations.SentinelOneDestination = append(outCfg.Destinations.SentinelOneDestination, d)
		}
		if hec := flattenSplunkHecDestination(ctx, d.ObservabilityPipelineSplunkHecDestination); hec != nil {
			outCfg.Destinations.SplunkHecDestination = append(outCfg.Destinations.SplunkHecDestination, hec)
		}

		if gcs := flattenGoogleCloudStorageDestination(ctx, d.ObservabilityPipelineGoogleCloudStorageDestination); gcs != nil {
			outCfg.Destinations.GoogleCloudStorageDestination = append(outCfg.Destinations.GoogleCloudStorageDestination, gcs)
		}

		if pubsub := flattenGooglePubSubDestination(ctx, d.ObservabilityPipelineGooglePubSubDestination); pubsub != nil {
			outCfg.Destinations.GooglePubSubDestination = append(outCfg.Destinations.GooglePubSubDestination, pubsub)
		}

		if s := flattenSumoLogicDestination(ctx, d.ObservabilityPipelineSumoLogicDestination); s != nil {
			outCfg.Destinations.SumoLogicDestination = append(outCfg.Destinations.SumoLogicDestination, s)
		}
		if r := flattenRsyslogDestination(ctx, d.ObservabilityPipelineRsyslogDestination); r != nil {
			outCfg.Destinations.RsyslogDestination = append(outCfg.Destinations.RsyslogDestination, r)
		}
		if s := flattenSyslogNgDestination(ctx, d.ObservabilityPipelineSyslogNgDestination); s != nil {
			outCfg.Destinations.SyslogNgDestination = append(outCfg.Destinations.SyslogNgDestination, s)
		}
		if e := flattenElasticsearchDestination(ctx, d.ObservabilityPipelineElasticsearchDestination); e != nil {
			outCfg.Destinations.ElasticsearchDestination = append(outCfg.Destinations.ElasticsearchDestination, e)
		}
		if a := flattenAzureStorageDestination(ctx, d.AzureStorageDestination); a != nil {
			outCfg.Destinations.AzureStorageDestination = append(outCfg.Destinations.AzureStorageDestination, a)
		}
		if m := flattenMicrosoftSentinelDestination(ctx, d.MicrosoftSentinelDestination); m != nil {
			outCfg.Destinations.MicrosoftSentinelDestination = append(outCfg.Destinations.MicrosoftSentinelDestination, m)
		}
		if d := flattenOpenSearchDestination(ctx, d.ObservabilityPipelineOpenSearchDestination); d != nil {
			outCfg.Destinations.OpenSearchDestination = append(outCfg.Destinations.OpenSearchDestination, d)
		}
		if d := flattenAmazonOpenSearchDestination(ctx, d.ObservabilityPipelineAmazonOpenSearchDestination); d != nil {
			outCfg.Destinations.AmazonOpenSearchDestination = append(outCfg.Destinations.AmazonOpenSearchDestination, d)
		}
		if d := observability_pipeline.FlattenSocketDestination(ctx, d.ObservabilityPipelineSocketDestination); d != nil {
			outCfg.Destinations.SocketDestination = append(outCfg.Destinations.SocketDestination, d)
		}
		if d := observability_pipeline.FlattenAmazonS3Destination(ctx, d.ObservabilityPipelineAmazonS3Destination); d != nil {
			outCfg.Destinations.AmazonS3Destination = append(outCfg.Destinations.AmazonS3Destination, d)
		}
		if d := observability_pipeline.FlattenObservabilityPipelinesAmazonSecurityLakeDestination(ctx, d.ObservabilityPipelineAmazonSecurityLakeDestination); d != nil {
			outCfg.Destinations.AmazonSecurityLakeDestination = append(outCfg.Destinations.AmazonSecurityLakeDestination, d)
		}
		if d := observability_pipeline.FlattenCrowdStrikeNextGenSiemDestination(ctx, d.ObservabilityPipelineCrowdStrikeNextGenSiemDestination); d != nil {
			outCfg.Destinations.CrowdStrikeNextGenSiemDestination = append(outCfg.Destinations.CrowdStrikeNextGenSiemDestination, d)
		}

	}

	state.Config = &outCfg
}

// ---------- Sources ----------

func flattenDatadogAgentSource(src *datadogV2.ObservabilityPipelineDatadogAgentSource) *datadogAgentSourceModel {
	if src == nil {
		return nil
	}
	out := &datadogAgentSourceModel{
		Id: types.StringValue(src.Id),
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}
	return out
}

func expandDatadogAgentSource(src *datadogAgentSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	agent := datadogV2.NewObservabilityPipelineDatadogAgentSourceWithDefaults()
	agent.SetId(src.Id.ValueString())
	if src.Tls != nil {
		agent.Tls = expandTls(src.Tls)
	}
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineDatadogAgentSource: agent,
	}
}

func flattenKafkaSource(src *datadogV2.ObservabilityPipelineKafkaSource) *kafkaSourceModel {
	if src == nil {
		return nil
	}
	out := &kafkaSourceModel{
		Id:      types.StringValue(src.GetId()),
		GroupId: types.StringValue(src.GetGroupId()),
	}
	for _, topic := range src.GetTopics() {
		out.Topics = append(out.Topics, types.StringValue(topic))
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}
	if sasl, ok := src.GetSaslOk(); ok {
		out.Sasl = &kafkaSourceSaslModel{
			Mechanism: types.StringValue(string(sasl.GetMechanism())),
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

func expandKafkaSource(src *kafkaSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	source := datadogV2.NewObservabilityPipelineKafkaSourceWithDefaults()
	source.SetId(src.Id.ValueString())
	source.SetGroupId(src.GroupId.ValueString())
	var topics []string
	for _, t := range src.Topics {
		topics = append(topics, t.ValueString())
	}
	source.SetTopics(topics)

	if src.Tls != nil {
		source.Tls = expandTls(src.Tls)
	}

	if src.Sasl != nil {
		mechanism, _ := datadogV2.NewObservabilityPipelinePipelineKafkaSourceSaslMechanismFromValue(src.Sasl.Mechanism.ValueString())
		if mechanism != nil {
			sasl := datadogV2.ObservabilityPipelineKafkaSourceSasl{}
			sasl.SetMechanism(*mechanism)
			source.SetSasl(sasl)
		}
	}

	if len(src.LibrdkafkaOptions) > 0 {
		opts := []datadogV2.ObservabilityPipelineKafkaSourceLibrdkafkaOption{}
		for _, opt := range src.LibrdkafkaOptions {
			opts = append(opts, datadogV2.ObservabilityPipelineKafkaSourceLibrdkafkaOption{
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

func flattenFilterProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineFilterProcessor) *filterProcessorModel {
	if src == nil {
		return nil
	}
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)
	return &filterProcessorModel{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
	}
}

func expandFilterProcessor(ctx context.Context, src *filterProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineFilterProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())
	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)
	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineFilterProcessor: proc,
	}
}

func flattenParseJsonProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineParseJSONProcessor) *parseJsonProcessorModel {
	if src == nil {
		return nil
	}
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)
	return &parseJsonProcessorModel{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
		Field:   types.StringValue(src.Field),
	}
}

func expandParseJsonProcessor(ctx context.Context, src *parseJsonProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineParseJSONProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())
	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)
	proc.SetField(src.Field.ValueString())
	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineParseJSONProcessor: proc,
	}
}

func flattenAddFieldsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineAddFieldsProcessor) *addFieldsProcessor {
	if src == nil {
		return nil
	}
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)
	out := &addFieldsProcessor{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
	}
	for _, f := range src.Fields {
		out.Fields = append(out.Fields, fieldValue{
			Name:  types.StringValue(f.Name),
			Value: types.StringValue(f.Value),
		})
	}
	return out
}

func expandAddFieldsProcessor(ctx context.Context, src *addFieldsProcessor) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineAddFieldsProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())
	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)
	var fields []datadogV2.ObservabilityPipelineFieldValue
	for _, f := range src.Fields {
		fields = append(fields, datadogV2.ObservabilityPipelineFieldValue{
			Name:  f.Name.ValueString(),
			Value: f.Value.ValueString(),
		})
	}
	proc.SetFields(fields)
	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineAddFieldsProcessor: proc,
	}
}

func flattenRenameFieldsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineRenameFieldsProcessor) *renameFieldsProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	out := &renameFieldsProcessorModel{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
	}

	for _, f := range src.Fields {
		out.Fields = append(out.Fields, renameFieldItemModel{
			Source:         types.StringValue(f.Source),
			Destination:    types.StringValue(f.Destination),
			PreserveSource: types.BoolValue(f.PreserveSource),
		})
	}

	return out
}

func expandRenameFieldsProcessor(ctx context.Context, src *renameFieldsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineRenameFieldsProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	var fields []datadogV2.ObservabilityPipelineRenameFieldsProcessorField
	for _, f := range src.Fields {
		fields = append(fields, datadogV2.ObservabilityPipelineRenameFieldsProcessorField{
			Source:         f.Source.ValueString(),
			Destination:    f.Destination.ValueString(),
			PreserveSource: f.PreserveSource.ValueBool(),
		})
	}
	proc.SetFields(fields)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineRenameFieldsProcessor: proc,
	}
}

func flattenRemoveFieldsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineRemoveFieldsProcessor) *removeFieldsProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)
	fields, _ := types.ListValueFrom(ctx, types.StringType, src.Fields)

	return &removeFieldsProcessorModel{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
		Fields:  fields,
	}
}

func expandRemoveFieldsProcessor(ctx context.Context, src *removeFieldsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineRemoveFieldsProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	var fields []string
	src.Fields.ElementsAs(ctx, &fields, false)
	proc.SetFields(fields)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineRemoveFieldsProcessor: proc,
	}
}

func flattenQuotaProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineQuotaProcessor) *quotaProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)
	partitionFields, _ := types.ListValueFrom(ctx, types.StringType, src.PartitionFields)

	var partitions []types.String
	for _, p := range partitionFields.Elements() {
		if strVal, ok := p.(types.String); ok {
			partitions = append(partitions, strVal)
		}
	}

	out := &quotaProcessorModel{
		Id:              types.StringValue(src.Id),
		Include:         types.StringValue(src.Include),
		Name:            types.StringValue(src.Name),
		DropEvents:      types.BoolValue(src.DropEvents),
		Inputs:          inputs,
		PartitionFields: partitions,
		Limit: quotaLimitModel{
			Enforce: types.StringValue(string(src.Limit.Enforce)),
			Limit:   types.Int64Value(src.Limit.Limit),
		},
	}

	if src.IgnoreWhenMissingPartitions != nil {
		out.IgnoreWhenMissingPartitions = types.BoolPointerValue(src.IgnoreWhenMissingPartitions)
	}

	if src.OverflowAction != nil {
		out.OverflowAction = types.StringValue(string(*src.OverflowAction))
	}

	for _, o := range src.Overrides {
		override := quotaOverrideModel{
			Limit: quotaLimitModel{
				Enforce: types.StringValue(string(o.Limit.Enforce)),
				Limit:   types.Int64Value(o.Limit.Limit),
			},
		}
		for _, f := range o.Fields {
			override.Fields = append(override.Fields, fieldValue{
				Name:  types.StringValue(f.Name),
				Value: types.StringValue(f.Value),
			})
		}
		out.Overrides = append(out.Overrides, override)
	}

	return out
}

func expandQuotaProcessor(ctx context.Context, src *quotaProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineQuotaProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())
	proc.SetName(src.Name.ValueString())
	proc.SetDropEvents(src.DropEvents.ValueBool())
	if !src.IgnoreWhenMissingPartitions.IsNull() {
		proc.SetIgnoreWhenMissingPartitions(src.IgnoreWhenMissingPartitions.ValueBool())
	}

	var inputs, partitions []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	for _, p := range src.PartitionFields {
		partitions = append(partitions, p.ValueString())
	}
	proc.SetInputs(inputs)
	proc.SetPartitionFields(partitions)

	proc.SetLimit(datadogV2.ObservabilityPipelineQuotaProcessorLimit{
		Enforce: datadogV2.ObservabilityPipelineQuotaProcessorLimitEnforceType(src.Limit.Enforce.ValueString()),
		Limit:   src.Limit.Limit.ValueInt64(),
	})

	if !src.OverflowAction.IsNull() {
		proc.SetOverflowAction(datadogV2.ObservabilityPipelineQuotaProcessorOverflowAction(src.OverflowAction.ValueString()))
	}

	var overrides []datadogV2.ObservabilityPipelineQuotaProcessorOverride
	for _, o := range src.Overrides {
		var fields []datadogV2.ObservabilityPipelineFieldValue
		for _, f := range o.Fields {
			fields = append(fields, datadogV2.ObservabilityPipelineFieldValue{
				Name:  f.Name.ValueString(),
				Value: f.Value.ValueString(),
			})
		}
		overrides = append(overrides, datadogV2.ObservabilityPipelineQuotaProcessorOverride{
			Fields: fields,
			Limit: datadogV2.ObservabilityPipelineQuotaProcessorLimit{
				Enforce: datadogV2.ObservabilityPipelineQuotaProcessorLimitEnforceType(o.Limit.Enforce.ValueString()),
				Limit:   o.Limit.Limit.ValueInt64(),
			},
		})
	}
	proc.SetOverrides(overrides)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineQuotaProcessor: proc,
	}
}

// ---------- Destinations ----------

func flattenDatadogLogsDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineDatadogLogsDestination) *datadogLogsDestinationModel {
	if src == nil {
		return nil
	}
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)
	return &datadogLogsDestinationModel{
		Id:     types.StringValue(src.Id),
		Inputs: inputs,
	}
}

func expandDatadogLogsDestination(ctx context.Context, src *datadogLogsDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineDatadogLogsDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())
	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)
	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineDatadogLogsDestination: dest,
	}
}

func flattenTls(src *datadogV2.ObservabilityPipelineTls) tlsModel {
	return tlsModel{
		CrtFile: types.StringValue(src.CrtFile),
		CaFile:  types.StringPointerValue(src.CaFile),
		KeyFile: types.StringPointerValue(src.KeyFile),
	}
}

func expandTls(tlsTF *tlsModel) *datadogV2.ObservabilityPipelineTls {
	tls := &datadogV2.ObservabilityPipelineTls{}
	tls.SetCrtFile(tlsTF.CrtFile.ValueString())
	if !tlsTF.CaFile.IsNull() {
		tls.SetCaFile(tlsTF.CaFile.ValueString())
	}
	if !tlsTF.KeyFile.IsNull() {
		tls.SetKeyFile(tlsTF.KeyFile.ValueString())
	}
	return tls
}

func expandParseGrokProcessor(ctx context.Context, p *parseGrokProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineParseGrokProcessorWithDefaults()
	proc.SetId(p.Id.ValueString())
	proc.SetInclude(p.Include.ValueString())

	var inputs []string
	p.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	if !p.DisableLibraryRules.IsNull() {
		proc.SetDisableLibraryRules(p.DisableLibraryRules.ValueBool())
	}

	var rules []datadogV2.ObservabilityPipelineParseGrokProcessorRule
	for _, r := range p.Rules {
		var matchRules []datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule
		for _, m := range r.MatchRules {
			matchRules = append(matchRules, datadogV2.ObservabilityPipelineParseGrokProcessorRuleMatchRule{
				Name: m.Name.ValueString(),
				Rule: m.Rule.ValueString(),
			})
		}

		var supportRules []datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule
		for _, s := range r.SupportRules {
			supportRules = append(supportRules, datadogV2.ObservabilityPipelineParseGrokProcessorRuleSupportRule{
				Name: s.Name.ValueString(),
				Rule: s.Rule.ValueString(),
			})
		}

		rules = append(rules, datadogV2.ObservabilityPipelineParseGrokProcessorRule{
			Source:       r.Source.ValueString(),
			MatchRules:   matchRules,
			SupportRules: supportRules,
		})
	}
	proc.SetRules(rules)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineParseGrokProcessor: proc,
	}
}

func flattenParseGrokProcessor(ctx context.Context, proc *datadogV2.ObservabilityPipelineParseGrokProcessor) *parseGrokProcessorModel {
	if proc == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, proc.GetInputs())

	out := &parseGrokProcessorModel{
		Id:                  types.StringValue(proc.GetId()),
		Include:             types.StringValue(proc.GetInclude()),
		Inputs:              inputs,
		DisableLibraryRules: types.BoolValue(proc.GetDisableLibraryRules()),
	}

	for _, r := range proc.GetRules() {
		var matchRules []grokRuleModel
		for _, m := range r.MatchRules {
			matchRules = append(matchRules, grokRuleModel{
				Name: types.StringValue(m.Name),
				Rule: types.StringValue(m.Rule),
			})
		}

		var supportRules []grokRuleModel
		for _, s := range r.SupportRules {
			supportRules = append(supportRules, grokRuleModel{
				Name: types.StringValue(s.Name),
				Rule: types.StringValue(s.Rule),
			})
		}

		out.Rules = append(out.Rules, parseGrokProcessorRuleModel{
			Source:       types.StringValue(r.Source),
			MatchRules:   matchRules,
			SupportRules: supportRules,
		})
	}

	return out
}

func expandSampleProcessor(ctx context.Context, p *sampleProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineSampleProcessorWithDefaults()
	proc.SetId(p.Id.ValueString())
	proc.SetInclude(p.Include.ValueString())

	var inputs []string
	p.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	if !p.Rate.IsNull() {
		proc.SetRate(p.Rate.ValueInt64())
	}
	if !p.Percentage.IsNull() {
		proc.SetPercentage(p.Percentage.ValueFloat64())
	}

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineSampleProcessor: proc,
	}
}

func flattenSampleProcessor(ctx context.Context, proc *datadogV2.ObservabilityPipelineSampleProcessor) *sampleProcessorModel {
	if proc == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, proc.GetInputs())

	out := &sampleProcessorModel{
		Id:      types.StringValue(proc.GetId()),
		Include: types.StringValue(proc.GetInclude()),
		Inputs:  inputs,
	}

	if rate, ok := proc.GetRateOk(); ok {
		out.Rate = types.Int64Value(*rate)
	} else {
		out.Rate = types.Int64Null()
	}

	if pct, ok := proc.GetPercentageOk(); ok {
		out.Percentage = types.Float64Value(*pct)
	} else {
		out.Percentage = types.Float64Null()
	}

	return out
}

func expandFluentdSource(src *fluentdSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	source := datadogV2.NewObservabilityPipelineFluentdSourceWithDefaults()
	source.SetId(src.Id.ValueString())

	if src.Tls != nil {
		source.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineFluentdSource: source,
	}
}

func expandFluentBitSource(src *fluentBitSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	source := datadogV2.NewObservabilityPipelineFluentBitSourceWithDefaults()
	source.SetId(src.Id.ValueString())

	if src.Tls != nil {
		source.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineFluentBitSource: source,
	}
}

func flattenFluentdSource(src *datadogV2.ObservabilityPipelineFluentdSource) *fluentdSourceModel {
	if src == nil {
		return nil
	}

	out := &fluentdSourceModel{
		Id: types.StringValue(src.GetId()),
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}
	return out
}

func flattenFluentBitSource(src *datadogV2.ObservabilityPipelineFluentBitSource) *fluentBitSourceModel {
	if src == nil {
		return nil
	}

	out := &fluentBitSourceModel{
		Id: types.StringValue(src.GetId()),
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
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

func expandHttpServerSource(src *httpServerSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineHttpServerSourceWithDefaults()
	s.SetId(src.Id.ValueString())

	s.SetAuthStrategy(datadogV2.ObservabilityPipelineHttpServerSourceAuthStrategy(src.AuthStrategy.ValueString()))
	s.SetDecoding(datadogV2.ObservabilityPipelineDecoding(src.Decoding.ValueString()))

	if src.Tls != nil {
		s.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineHttpServerSource: s,
	}
}

func flattenHttpServerSource(src *datadogV2.ObservabilityPipelineHttpServerSource) *httpServerSourceModel {
	if src == nil {
		return nil
	}

	out := &httpServerSourceModel{
		Id:           types.StringValue(src.GetId()),
		AuthStrategy: types.StringValue(string(src.GetAuthStrategy())),
		Decoding:     types.StringValue(string(src.GetDecoding())),
	}

	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

func expandSplunkHecSource(src *splunkHecSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineSplunkHecSourceWithDefaults()

	s.SetId(src.Id.ValueString())

	if src.Tls != nil {
		s.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSplunkHecSource: s,
	}
}

func flattenSplunkHecSource(src *datadogV2.ObservabilityPipelineSplunkHecSource) *splunkHecSourceModel {
	if src == nil {
		return nil
	}

	out := &splunkHecSourceModel{
		Id: types.StringValue(src.GetId()),
	}

	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

func expandGoogleCloudStorageDestination(ctx context.Context, d *gcsDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineGoogleCloudStorageDestinationWithDefaults()

	dest.SetId(d.Id.ValueString())
	dest.SetBucket(d.Bucket.ValueString())
	dest.SetStorageClass(datadogV2.ObservabilityPipelineGoogleCloudStorageDestinationStorageClass(d.StorageClass.ValueString()))
	dest.SetAcl(datadogV2.ObservabilityPipelineGoogleCloudStorageDestinationAcl(d.Acl.ValueString()))

	if !d.KeyPrefix.IsNull() {
		dest.SetKeyPrefix(d.KeyPrefix.ValueString())
	}

	dest.SetAuth(datadogV2.ObservabilityPipelineGcpAuth{
		CredentialsFile: d.Auth.CredentialsFile.ValueString(),
	})

	var metadata []datadogV2.ObservabilityPipelineMetadataEntry
	for _, m := range d.Metadata {
		metadata = append(metadata, datadogV2.ObservabilityPipelineMetadataEntry{
			Name:  m.Name.ValueString(),
			Value: m.Value.ValueString(),
		})
	}
	dest.SetMetadata(metadata)

	var inputs []string
	d.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineGoogleCloudStorageDestination: dest,
	}
}

func flattenGoogleCloudStorageDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineGoogleCloudStorageDestination) *gcsDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())

	var metadata []metadataEntry
	for _, m := range src.GetMetadata() {
		metadata = append(metadata, metadataEntry{
			Name:  types.StringValue(m.Name),
			Value: types.StringValue(m.Value),
		})
	}

	return &gcsDestinationModel{
		Id:           types.StringValue(src.GetId()),
		Bucket:       types.StringValue(src.GetBucket()),
		KeyPrefix:    types.StringPointerValue(src.KeyPrefix),
		StorageClass: types.StringValue(string(src.GetStorageClass())),
		Acl:          types.StringValue(string(src.GetAcl())),
		Auth: gcpAuthModel{
			CredentialsFile: types.StringValue(src.Auth.CredentialsFile),
		},
		Metadata: metadata,
		Inputs:   inputs,
	}
}

func expandGooglePubSubDestination(ctx context.Context, d *googlePubSubDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineGooglePubSubDestinationWithDefaults()
	dest.SetId(d.Id.ValueString())
	dest.SetProject(d.Project.ValueString())
	dest.SetTopic(d.Topic.ValueString())

	if !d.Encoding.IsNull() {
		dest.SetEncoding(datadogV2.ObservabilityPipelineGooglePubSubDestinationEncoding(d.Encoding.ValueString()))
	}

	if d.Auth != nil {
		auth := datadogV2.ObservabilityPipelineGcpAuth{}
		auth.SetCredentialsFile(d.Auth.CredentialsFile.ValueString())
		dest.SetAuth(auth)
	}

	if d.Tls != nil {
		dest.Tls = expandTls(d.Tls)
	}

	var inputs []string
	d.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineGooglePubSubDestination: dest,
	}
}

func flattenGooglePubSubDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineGooglePubSubDestination) *googlePubSubDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())

	out := &googlePubSubDestinationModel{
		Id:      types.StringValue(src.GetId()),
		Project: types.StringValue(src.GetProject()),
		Topic:   types.StringValue(src.GetTopic()),
		Inputs:  inputs,
	}

	if encoding, ok := src.GetEncodingOk(); ok {
		out.Encoding = types.StringValue(string(*encoding))
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = &gcpAuthModel{
			CredentialsFile: types.StringValue(auth.CredentialsFile),
		}
	}

	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

func expandSplunkTcpSource(src *splunkTcpSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineSplunkTcpSourceWithDefaults()
	s.SetId(src.Id.ValueString())

	if src.Tls != nil {
		s.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSplunkTcpSource: s,
	}
}

func flattenSplunkTcpSource(src *datadogV2.ObservabilityPipelineSplunkTcpSource) *splunkTcpSourceModel {
	if src == nil {
		return nil
	}

	out := &splunkTcpSourceModel{
		Id: types.StringValue(src.GetId()),
	}

	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

func expandGenerateMetricsProcessor(ctx context.Context, p *generateMetricsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineGenerateMetricsProcessorWithDefaults()

	proc.SetId(p.Id.ValueString())
	proc.SetInclude(p.Include.ValueString())

	var inputs []string
	p.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	var metrics []datadogV2.ObservabilityPipelineGeneratedMetric
	for _, m := range p.Metrics {
		val := datadogV2.ObservabilityPipelineMetricValue{}

		switch m.Value.Strategy.ValueString() {
		case "increment_by_one":
			val.ObservabilityPipelineGeneratedMetricIncrementByOne = &datadogV2.ObservabilityPipelineGeneratedMetricIncrementByOne{
				Strategy: "increment_by_one",
			}
		case "increment_by_field":
			val.ObservabilityPipelineGeneratedMetricIncrementByField = &datadogV2.ObservabilityPipelineGeneratedMetricIncrementByField{
				Strategy: "increment_by_field",
				Field:    m.Value.Field.ValueString(),
			}
		}

		groupBy := []string{}
		m.GroupBy.ElementsAs(ctx, &groupBy, false)

		metrics = append(metrics, datadogV2.ObservabilityPipelineGeneratedMetric{
			Name:       m.Name.ValueString(),
			Include:    m.Include.ValueString(),
			MetricType: datadogV2.ObservabilityPipelineGeneratedMetricMetricType(m.MetricType.ValueString()),
			Value:      val,
			GroupBy:    groupBy,
		})
	}

	proc.SetMetrics(metrics)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineGenerateMetricsProcessor: proc,
	}
}

func flattenGenerateMetricsProcessor(ctx context.Context, proc *datadogV2.ObservabilityPipelineGenerateMetricsProcessor) *generateMetricsProcessorModel {
	if proc == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, proc.GetInputs())

	var metrics []generatedMetricModel
	for _, m := range proc.GetMetrics() {
		groupBy, _ := types.ListValueFrom(ctx, types.StringType, m.GroupBy)

		val := &generatedMetricValue{}
		if m.Value.ObservabilityPipelineGeneratedMetricIncrementByOne != nil {
			val.Strategy = types.StringValue("increment_by_one")
		}
		if m.Value.ObservabilityPipelineGeneratedMetricIncrementByField != nil {
			val.Strategy = types.StringValue("increment_by_field")
			val.Field = types.StringValue(m.Value.ObservabilityPipelineGeneratedMetricIncrementByField.Field)
		}

		metrics = append(metrics, generatedMetricModel{
			Name:       types.StringValue(m.Name),
			Include:    types.StringValue(m.Include),
			MetricType: types.StringValue(string(m.MetricType)),
			GroupBy:    groupBy,
			Value:      val,
		})
	}

	return &generateMetricsProcessorModel{
		Id:      types.StringValue(proc.GetId()),
		Include: types.StringValue(proc.GetInclude()),
		Inputs:  inputs,
		Metrics: metrics,
	}
}

func expandSplunkHecDestination(ctx context.Context, d *splunkHecDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineSplunkHecDestinationWithDefaults()

	dest.SetId(d.Id.ValueString())

	var inputs []string
	d.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	if !d.AutoExtractTimestamp.IsNull() {
		dest.SetAutoExtractTimestamp(d.AutoExtractTimestamp.ValueBool())
	}
	if !d.Encoding.IsNull() {
		dest.SetEncoding(datadogV2.ObservabilityPipelineSplunkHecDestinationEncoding(d.Encoding.ValueString()))
	}
	if !d.Sourcetype.IsNull() {
		dest.SetSourcetype(d.Sourcetype.ValueString())
	}
	if !d.Index.IsNull() {
		dest.SetIndex(d.Index.ValueString())
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSplunkHecDestination: dest,
	}
}

func flattenSplunkHecDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSplunkHecDestination) *splunkHecDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())

	return &splunkHecDestinationModel{
		Id:                   types.StringValue(src.GetId()),
		Inputs:               inputs,
		AutoExtractTimestamp: types.BoolValue(src.GetAutoExtractTimestamp()),
		Encoding:             types.StringValue(string(*src.Encoding)),
		Sourcetype:           types.StringPointerValue(src.Sourcetype),
		Index:                types.StringPointerValue(src.Index),
	}
}

func expandAmazonS3Source(src *amazonS3SourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineAmazonS3SourceWithDefaults()

	s.SetId(src.Id.ValueString())
	s.SetRegion(src.Region.ValueString())

	if src.Auth != nil {
		auth := observability_pipeline.ExpandAwsAuth(src.Auth)
		if auth != nil {
			s.SetAuth(*auth)
		}
	}

	if src.Tls != nil {
		s.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineAmazonS3Source: s,
	}
}

func flattenAmazonS3Source(src *datadogV2.ObservabilityPipelineAmazonS3Source) *amazonS3SourceModel {
	if src == nil {
		return nil
	}

	out := &amazonS3SourceModel{
		Id:     types.StringValue(src.GetId()),
		Region: types.StringValue(src.GetRegion()),
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = observability_pipeline.FlattenAwsAuth(auth)
	}

	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

func expandSumoLogicDestination(ctx context.Context, src *sumoLogicDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineSumoLogicDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	if !src.Encoding.IsNull() {
		dest.SetEncoding(datadogV2.ObservabilityPipelineSumoLogicDestinationEncoding(src.Encoding.ValueString()))
	}
	if !src.HeaderHostName.IsNull() {
		dest.SetHeaderHostName(src.HeaderHostName.ValueString())
	}
	if !src.HeaderSourceName.IsNull() {
		dest.SetHeaderSourceName(src.HeaderSourceName.ValueString())
	}
	if !src.HeaderSourceCategory.IsNull() {
		dest.SetHeaderSourceCategory(src.HeaderSourceCategory.ValueString())
	}

	if len(src.HeaderCustomFields) > 0 {
		var fields []datadogV2.ObservabilityPipelineSumoLogicDestinationHeaderCustomFieldsItem
		for _, f := range src.HeaderCustomFields {
			fields = append(fields, datadogV2.ObservabilityPipelineSumoLogicDestinationHeaderCustomFieldsItem{
				Name:  f.Name.ValueString(),
				Value: f.Value.ValueString(),
			})
		}
		dest.SetHeaderCustomFields(fields)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSumoLogicDestination: dest,
	}
}

func flattenSumoLogicDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSumoLogicDestination) *sumoLogicDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())

	out := &sumoLogicDestinationModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
	}

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

func expandRsyslogSource(src *rsyslogSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	obj := datadogV2.NewObservabilityPipelineRsyslogSourceWithDefaults()
	obj.SetId(src.Id.ValueString())
	if !src.Mode.IsNull() {
		obj.SetMode(datadogV2.ObservabilityPipelineSyslogSourceMode(src.Mode.ValueString()))
	}
	if src.Tls != nil {
		obj.Tls = expandTls(src.Tls)
	}
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineRsyslogSource: obj,
	}
}

func flattenRsyslogSource(src *datadogV2.ObservabilityPipelineRsyslogSource) *rsyslogSourceModel {
	if src == nil {
		return nil
	}
	out := &rsyslogSourceModel{
		Id: types.StringValue(src.GetId()),
	}
	if v, ok := src.GetModeOk(); ok {
		out.Mode = types.StringValue(string(*v))
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}
	return out
}

func expandSyslogNgSource(src *syslogNgSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	obj := datadogV2.NewObservabilityPipelineSyslogNgSourceWithDefaults()
	obj.SetId(src.Id.ValueString())
	if !src.Mode.IsNull() {
		obj.SetMode(datadogV2.ObservabilityPipelineSyslogSourceMode(src.Mode.ValueString()))
	}
	if src.Tls != nil {
		obj.Tls = expandTls(src.Tls)
	}
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSyslogNgSource: obj,
	}
}

func flattenSyslogNgSource(src *datadogV2.ObservabilityPipelineSyslogNgSource) *syslogNgSourceModel {
	if src == nil {
		return nil
	}
	out := &syslogNgSourceModel{
		Id: types.StringValue(src.GetId()),
	}
	if v, ok := src.GetModeOk(); ok {
		out.Mode = types.StringValue(string(*v))
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}
	return out
}

func expandRsyslogDestination(ctx context.Context, src *rsyslogDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewObservabilityPipelineRsyslogDestinationWithDefaults()
	obj.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	if !src.Keepalive.IsNull() {
		obj.SetKeepalive(src.Keepalive.ValueInt64())
	}
	if src.Tls != nil {
		obj.Tls = expandTls(src.Tls)
	}
	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineRsyslogDestination: obj,
	}
}

func flattenRsyslogDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineRsyslogDestination) *rsyslogDestinationModel {
	if src == nil {
		return nil
	}
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())
	out := &rsyslogDestinationModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
	}
	if v, ok := src.GetKeepaliveOk(); ok {
		out.Keepalive = types.Int64Value(*v)
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}
	return out
}

func expandSyslogNgDestination(ctx context.Context, src *syslogNgDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewObservabilityPipelineSyslogNgDestinationWithDefaults()
	obj.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	if !src.Keepalive.IsNull() {
		obj.SetKeepalive(src.Keepalive.ValueInt64())
	}
	if src.Tls != nil {
		obj.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSyslogNgDestination: obj,
	}
}

func flattenSyslogNgDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSyslogNgDestination) *syslogNgDestinationModel {
	if src == nil {
		return nil
	}
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())
	out := &syslogNgDestinationModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
	}
	if v, ok := src.GetKeepaliveOk(); ok {
		out.Keepalive = types.Int64Value(*v)
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}
	return out
}

func expandElasticsearchDestination(ctx context.Context, src *elasticsearchDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewObservabilityPipelineElasticsearchDestinationWithDefaults()
	obj.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	if !src.ApiVersion.IsNull() {
		obj.SetApiVersion(datadogV2.ObservabilityPipelineElasticsearchDestinationApiVersion(src.ApiVersion.ValueString()))
	}
	if !src.BulkIndex.IsNull() {
		obj.SetBulkIndex(src.BulkIndex.ValueString())
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineElasticsearchDestination: obj,
	}
}

func flattenElasticsearchDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineElasticsearchDestination) *elasticsearchDestinationModel {
	if src == nil {
		return nil
	}
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())
	out := &elasticsearchDestinationModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
	}
	if v, ok := src.GetApiVersionOk(); ok {
		out.ApiVersion = types.StringValue(string(*v))
	}
	if v, ok := src.GetBulkIndexOk(); ok {
		out.BulkIndex = types.StringValue(*v)
	}
	return out
}

func expandAzureStorageDestination(ctx context.Context, src *azureStorageDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewAzureStorageDestinationWithDefaults()
	obj.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
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
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())
	out := &azureStorageDestinationModel{
		Id:            types.StringValue(src.GetId()),
		Inputs:        inputs,
		ContainerName: types.StringValue(src.GetContainerName()),
	}
	if v, ok := src.GetBlobPrefixOk(); ok {
		out.BlobPrefix = types.StringValue(*v)
	}
	return out
}

func expandMicrosoftSentinelDestination(ctx context.Context, src *microsoftSentinelDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	obj := datadogV2.NewMicrosoftSentinelDestinationWithDefaults()
	obj.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
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
	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())
	return &microsoftSentinelDestinationModel{
		Id:             types.StringValue(src.GetId()),
		Inputs:         inputs,
		ClientId:       types.StringValue(src.GetClientId()),
		TenantId:       types.StringValue(src.GetTenantId()),
		DcrImmutableId: types.StringValue(src.GetDcrImmutableId()),
		Table:          types.StringValue(src.GetTable()),
	}
}

func expandSensitiveDataScannerProcessor(ctx context.Context, src *sensitiveDataScannerProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	obj := datadogV2.NewObservabilityPipelineSensitiveDataScannerProcessorWithDefaults()

	obj.SetId(src.Id.ValueString())
	obj.SetInclude(src.Include.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	obj.SetInputs(inputs)

	var rules []datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorRule
	for _, rule := range src.Rules {
		r := datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorRule{
			Name: rule.Name.ValueString(),
		}

		for _, tag := range rule.Tags {
			r.Tags = append(r.Tags, tag.ValueString())
		}

		if rule.KeywordOptions != nil {
			r.KeywordOptions = &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorKeywordOptions{}

			for _, k := range rule.KeywordOptions.Keywords {
				r.KeywordOptions.Keywords = append(r.KeywordOptions.Keywords, k.ValueString())
			}

			r.KeywordOptions.Proximity = rule.KeywordOptions.Proximity.ValueInt64()
		}

		if rule.Pattern != nil {
			if rule.Pattern.Custom != nil {
				r.Pattern = datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorPattern{
					ObservabilityPipelineSensitiveDataScannerProcessorCustomPattern: &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorCustomPattern{
						Type: "custom",
						Options: datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorCustomPatternOptions{
							Rule: rule.Pattern.Custom.Rule.ValueString(),
						},
					},
				}
			} else if rule.Pattern.Library != nil {
				r.Pattern = datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorPattern{
					ObservabilityPipelineSensitiveDataScannerProcessorLibraryPattern: &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorLibraryPattern{
						Type: "library",
						Options: datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorLibraryPatternOptions{
							Id: rule.Pattern.Library.Id.ValueString(),
						},
					},
				}
				if !rule.Pattern.Library.UseRecommendedKeywords.IsNull() {
					r.Pattern.ObservabilityPipelineSensitiveDataScannerProcessorLibraryPattern.Options.
						SetUseRecommendedKeywords(rule.Pattern.Library.UseRecommendedKeywords.ValueBool())
				}
			}
		}

		if rule.Scope != nil {
			if rule.Scope.Include != nil {
				r.Scope = datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScope{
					ObservabilityPipelineSensitiveDataScannerProcessorScopeInclude: &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScopeInclude{
						Target: "include",
						Options: datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScopeOptions{
							Fields: extractStringList(rule.Scope.Include.Fields),
						},
					},
				}
			} else if rule.Scope.Exclude != nil {
				r.Scope = datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScope{
					ObservabilityPipelineSensitiveDataScannerProcessorScopeExclude: &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScopeExclude{
						Target: "exclude",
						Options: datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScopeOptions{
							Fields: extractStringList(rule.Scope.Exclude.Fields),
						},
					},
				}
			} else if rule.Scope.All != nil && *rule.Scope.All {
				r.Scope = datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScope{
					ObservabilityPipelineSensitiveDataScannerProcessorScopeAll: &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorScopeAll{
						Target: "all",
					},
				}
			}
		}

		if rule.OnMatch != nil {
			if rule.OnMatch.Redact != nil {
				r.OnMatch = datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorAction{
					ObservabilityPipelineSensitiveDataScannerProcessorActionRedact: &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionRedact{
						Action: "redact",
						Options: datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionRedactOptions{
							Replace: rule.OnMatch.Redact.Replace.ValueString(),
						},
					},
				}
			} else if rule.OnMatch.Hash != nil {
				r.OnMatch = datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorAction{
					ObservabilityPipelineSensitiveDataScannerProcessorActionHash: &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionHash{
						Action: "hash",
					},
				}
			} else if rule.OnMatch.PartialRedact != nil {
				r.OnMatch = datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorAction{
					ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedact: &datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedact{
						Action: "partial_redact",
						Options: datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedactOptions{
							Characters: rule.OnMatch.PartialRedact.Characters.ValueInt64(),
							Direction:  datadogV2.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedactOptionsDirection(rule.OnMatch.PartialRedact.Direction.ValueString()),
						},
					},
				}
			}
		}

		rules = append(rules, r)
	}
	obj.SetRules(rules)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineSensitiveDataScannerProcessor: obj,
	}
}

func extractStringList(list []types.String) []string {
	var out []string
	for _, s := range list {
		out = append(out, s.ValueString())
	}
	return out
}

func wrapStringList(vals []string) []types.String {
	out := make([]types.String, len(vals))
	for i, v := range vals {
		out[i] = types.StringValue(v)
	}
	return out
}

func flattenSensitiveDataScannerProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineSensitiveDataScannerProcessor) *sensitiveDataScannerProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())
	out := &sensitiveDataScannerProcessorModel{
		Id:      types.StringValue(src.GetId()),
		Include: types.StringValue(src.GetInclude()),
		Inputs:  inputs,
	}

	for _, r := range src.GetRules() {
		rule := sensitiveDataScannerProcessorRule{
			Name: types.StringValue(r.GetName()),
		}

		for _, tag := range r.GetTags() {
			rule.Tags = append(rule.Tags, types.StringValue(tag))
		}

		if ko, ok := r.GetKeywordOptionsOk(); ok {
			rule.KeywordOptions = &sensitiveDataScannerProcessorKeywordOptions{
				Proximity: types.Int64Value(ko.Proximity),
			}
			for _, k := range ko.Keywords {
				rule.KeywordOptions.Keywords = append(rule.KeywordOptions.Keywords, types.StringValue(k))
			}
		}

		switch p := r.Pattern; {
		case p.ObservabilityPipelineSensitiveDataScannerProcessorCustomPattern != nil:
			rule.Pattern = &sensitiveDataScannerProcessorPattern{
				Custom: &sensitiveDataScannerCustomPattern{
					Rule: types.StringValue(p.ObservabilityPipelineSensitiveDataScannerProcessorCustomPattern.Options.Rule),
				},
			}
		case p.ObservabilityPipelineSensitiveDataScannerProcessorLibraryPattern != nil:
			opts := p.ObservabilityPipelineSensitiveDataScannerProcessorLibraryPattern.Options
			rule.Pattern = &sensitiveDataScannerProcessorPattern{
				Library: &sensitiveDataScannerLibraryPattern{
					Id: types.StringValue(opts.Id),
				},
			}
			if v, ok := opts.GetUseRecommendedKeywordsOk(); ok {
				rule.Pattern.Library.UseRecommendedKeywords = types.BoolValue(*v)
			}
		}

		switch s := r.Scope; {
		case s.ObservabilityPipelineSensitiveDataScannerProcessorScopeInclude != nil:
			rule.Scope = &sensitiveDataScannerProcessorScope{
				Include: &sensitiveDataScannerScopeOptions{
					Fields: wrapStringList(s.ObservabilityPipelineSensitiveDataScannerProcessorScopeInclude.Options.Fields),
				},
			}
		case s.ObservabilityPipelineSensitiveDataScannerProcessorScopeExclude != nil:
			rule.Scope = &sensitiveDataScannerProcessorScope{
				Exclude: &sensitiveDataScannerScopeOptions{
					Fields: wrapStringList(s.ObservabilityPipelineSensitiveDataScannerProcessorScopeExclude.Options.Fields),
				},
			}
		case s.ObservabilityPipelineSensitiveDataScannerProcessorScopeAll != nil:
			all := true
			rule.Scope = &sensitiveDataScannerProcessorScope{
				All: &all,
			}
		}

		switch a := r.OnMatch; {
		case a.ObservabilityPipelineSensitiveDataScannerProcessorActionRedact != nil:
			rule.OnMatch = &sensitiveDataScannerProcessorAction{
				Redact: &sensitiveDataScannerRedactAction{
					Replace: types.StringValue(a.ObservabilityPipelineSensitiveDataScannerProcessorActionRedact.Options.Replace),
				},
			}
		case a.ObservabilityPipelineSensitiveDataScannerProcessorActionHash != nil:
			rule.OnMatch = &sensitiveDataScannerProcessorAction{
				Hash: &sensitiveDataScannerHashAction{},
			}
		case a.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedact != nil:
			rule.OnMatch = &sensitiveDataScannerProcessorAction{
				PartialRedact: &sensitiveDataScannerPartialRedactAction{
					Characters: types.Int64Value(a.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedact.Options.Characters),
					Direction:  types.StringValue(string(a.ObservabilityPipelineSensitiveDataScannerProcessorActionPartialRedact.Options.Direction)),
				},
			}
		}

		out.Rules = append(out.Rules, rule)
	}

	return out
}

func expandSumoLogicSource(src *sumoLogicSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	obj := datadogV2.NewObservabilityPipelineSumoLogicSourceWithDefaults()
	obj.SetId(src.Id.ValueString())

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSumoLogicSource: obj,
	}
}

func flattenSumoLogicSource(src *datadogV2.ObservabilityPipelineSumoLogicSource) *sumoLogicSourceModel {
	if src == nil {
		return nil
	}
	return &sumoLogicSourceModel{
		Id: types.StringValue(src.GetId()),
	}
}

func expandAmazonDataFirehoseSource(src *amazonDataFirehoseSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	firehose := datadogV2.NewObservabilityPipelineAmazonDataFirehoseSourceWithDefaults()
	firehose.SetId(src.Id.ValueString())

	if src.Auth != nil {
		auth := observability_pipeline.ExpandAwsAuth(src.Auth)
		if auth != nil {
			firehose.SetAuth(*auth)
		}
	}

	if src.Tls != nil {
		firehose.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineAmazonDataFirehoseSource: firehose,
	}
}

func flattenAmazonDataFirehoseSource(src *datadogV2.ObservabilityPipelineAmazonDataFirehoseSource) *amazonDataFirehoseSourceModel {
	if src == nil {
		return nil
	}

	out := &amazonDataFirehoseSourceModel{
		Id: types.StringValue(src.GetId()),
	}

	if auth, ok := src.GetAuthOk(); ok {
		out.Auth = observability_pipeline.FlattenAwsAuth(auth)
	}

	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

func expandHttpClientSource(src *httpClientSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	httpSrc := datadogV2.NewObservabilityPipelineHttpClientSourceWithDefaults()
	httpSrc.SetId(src.Id.ValueString())
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
	if src.Tls != nil {
		httpSrc.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineHttpClientSource: httpSrc,
	}
}

func flattenHttpClientSource(src *datadogV2.ObservabilityPipelineHttpClientSource) *httpClientSourceModel {
	if src == nil {
		return nil
	}

	out := &httpClientSourceModel{
		Id:       types.StringValue(src.GetId()),
		Decoding: types.StringValue(string(src.GetDecoding())),
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
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

func expandGooglePubSubSource(src *googlePubSubSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	pubsub := datadogV2.NewObservabilityPipelineGooglePubSubSourceWithDefaults()
	pubsub.SetId(src.Id.ValueString())
	pubsub.SetProject(src.Project.ValueString())
	pubsub.SetSubscription(src.Subscription.ValueString())
	pubsub.SetDecoding(datadogV2.ObservabilityPipelineDecoding(src.Decoding.ValueString()))

	if src.Auth != nil {
		auth := datadogV2.ObservabilityPipelineGcpAuth{}
		auth.SetCredentialsFile(src.Auth.CredentialsFile.ValueString())
		pubsub.SetAuth(auth)
	}

	if src.Tls != nil {
		pubsub.Tls = expandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineGooglePubSubSource: pubsub,
	}
}

func flattenGooglePubSubSource(src *datadogV2.ObservabilityPipelineGooglePubSubSource) *googlePubSubSourceModel {
	if src == nil {
		return nil
	}
	out := &googlePubSubSourceModel{
		Id:           types.StringValue(src.GetId()),
		Project:      types.StringValue(src.GetProject()),
		Subscription: types.StringValue(src.GetSubscription()),
		Decoding:     types.StringValue(string(src.GetDecoding())),
	}

	out.Auth = &gcpAuthModel{
		CredentialsFile: types.StringValue(src.Auth.CredentialsFile),
	}

	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

func expandLogstashSource(src *logstashSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	logstash := datadogV2.NewObservabilityPipelineLogstashSourceWithDefaults()
	logstash.SetId(src.Id.ValueString())
	if src.Tls != nil {
		logstash.Tls = expandTls(src.Tls)
	}
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineLogstashSource: logstash,
	}
}

func flattenLogstashSource(src *datadogV2.ObservabilityPipelineLogstashSource) *logstashSourceModel {
	if src == nil {
		return nil
	}
	out := &logstashSourceModel{
		Id: types.StringValue(src.GetId()),
	}
	if src.Tls != nil {
		tls := flattenTls(src.Tls)
		out.Tls = &tls
	}
	return out
}

func expandDedupeProcessor(ctx context.Context, src *dedupeProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineDedupeProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	var fields []string
	for _, f := range src.Fields {
		fields = append(fields, f.ValueString())
	}
	proc.SetFields(fields)

	proc.SetMode(datadogV2.ObservabilityPipelineDedupeProcessorMode(src.Mode.ValueString()))

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineDedupeProcessor: proc,
	}
}

func flattenDedupeProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineDedupeProcessor) *dedupeProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	var fields []types.String
	for _, f := range src.Fields {
		fields = append(fields, types.StringValue(f))
	}

	return &dedupeProcessorModel{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
		Fields:  fields,
		Mode:    types.StringValue(string(src.Mode)),
	}
}

func expandReduceProcessor(ctx context.Context, src *reduceProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineReduceProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())

	var inputs, groupBy []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	for _, g := range src.GroupBy {
		groupBy = append(groupBy, g.ValueString())
	}
	proc.SetInputs(inputs)
	proc.SetGroupBy(groupBy)

	var strategies []datadogV2.ObservabilityPipelineReduceProcessorMergeStrategy
	for _, s := range src.MergeStrategies {
		strategies = append(strategies, datadogV2.ObservabilityPipelineReduceProcessorMergeStrategy{
			Path:     s.Path.ValueString(),
			Strategy: datadogV2.ObservabilityPipelineReduceProcessorMergeStrategyStrategy(s.Strategy.ValueString()),
		})
	}
	proc.SetMergeStrategies(strategies)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineReduceProcessor: proc,
	}
}

func flattenReduceProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineReduceProcessor) *reduceProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	var groupBy []types.String
	for _, g := range src.GroupBy {
		groupBy = append(groupBy, types.StringValue(g))
	}

	var strategies []mergeStrategyModel
	for _, s := range src.MergeStrategies {
		strategies = append(strategies, mergeStrategyModel{
			Path:     types.StringValue(s.Path),
			Strategy: types.StringValue(string(s.Strategy)),
		})
	}

	return &reduceProcessorModel{
		Id:              types.StringValue(src.Id),
		Include:         types.StringValue(src.Include),
		Inputs:          inputs,
		GroupBy:         groupBy,
		MergeStrategies: strategies,
	}
}

func expandThrottleProcessor(ctx context.Context, src *throttleProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineThrottleProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	proc.SetThreshold(src.Threshold.ValueInt64())
	proc.SetWindow(src.Window.ValueFloat64())

	var groupBy []string
	for _, g := range src.GroupBy {
		groupBy = append(groupBy, g.ValueString())
	}
	if len(groupBy) > 0 {
		proc.SetGroupBy(groupBy)
	}

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineThrottleProcessor: proc,
	}
}

func flattenThrottleProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineThrottleProcessor) *throttleProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	var groupBy []types.String
	for _, g := range src.GroupBy {
		groupBy = append(groupBy, types.StringValue(g))
	}

	return &throttleProcessorModel{
		Id:        types.StringValue(src.Id),
		Include:   types.StringValue(src.Include),
		Inputs:    inputs,
		Threshold: types.Int64Value(src.Threshold),
		Window:    types.Float64Value(src.Window),
		GroupBy:   groupBy,
	}
}

func expandAddEnvVarsProcessor(ctx context.Context, src *addEnvVarsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineAddEnvVarsProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	var vars []datadogV2.ObservabilityPipelineAddEnvVarsProcessorVariable
	for _, v := range src.Variables {
		vars = append(vars, datadogV2.ObservabilityPipelineAddEnvVarsProcessorVariable{
			Field: v.Field.ValueString(),
			Name:  v.Name.ValueString(),
		})
	}
	proc.SetVariables(vars)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineAddEnvVarsProcessor: proc,
	}
}

func flattenAddEnvVarsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineAddEnvVarsProcessor) *addEnvVarsProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	var vars []envVarMappingModel
	for _, v := range src.Variables {
		vars = append(vars, envVarMappingModel{
			Field: types.StringValue(v.Field),
			Name:  types.StringValue(v.Name),
		})
	}

	return &addEnvVarsProcessorModel{
		Id:        types.StringValue(src.Id),
		Include:   types.StringValue(src.Include),
		Inputs:    inputs,
		Variables: vars,
	}
}

func expandEnrichmentTableProcessor(ctx context.Context, src *enrichmentTableProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineEnrichmentTableProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())
	proc.SetTarget(src.Target.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	if src.File != nil {
		file := datadogV2.ObservabilityPipelineEnrichmentTableFile{
			Path: src.File.Path.ValueString(),
		}
		file.Encoding = datadogV2.ObservabilityPipelineEnrichmentTableFileEncoding{
			Type:            datadogV2.ObservabilityPipelineEnrichmentTableFileEncodingType(src.File.Encoding.Type.ValueString()),
			Delimiter:       src.File.Encoding.Delimiter.ValueString(),
			IncludesHeaders: src.File.Encoding.IncludesHeaders.ValueBool(),
		}
		for _, s := range src.File.Schema {
			file.Schema = append(file.Schema, datadogV2.ObservabilityPipelineEnrichmentTableFileSchemaItems{
				Column: s.Column.ValueString(),
				Type:   datadogV2.ObservabilityPipelineEnrichmentTableFileSchemaItemsType(s.Type.ValueString()),
			})
		}
		for _, k := range src.File.Key {
			file.Key = append(file.Key, datadogV2.ObservabilityPipelineEnrichmentTableFileKeyItems{
				Column:     k.Column.ValueString(),
				Comparison: datadogV2.ObservabilityPipelineEnrichmentTableFileKeyItemsComparison(k.Comparison.ValueString()),
				Field:      k.Field.ValueString(),
			})
		}
		proc.File = &file
	}

	if src.GeoIp != nil {
		proc.Geoip = &datadogV2.ObservabilityPipelineEnrichmentTableGeoIp{
			KeyField: src.GeoIp.KeyField.ValueString(),
			Locale:   src.GeoIp.Locale.ValueString(),
			Path:     src.GeoIp.Path.ValueString(),
		}
	}

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineEnrichmentTableProcessor: proc,
	}
}

func flattenEnrichmentTableProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineEnrichmentTableProcessor) *enrichmentTableProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	out := &enrichmentTableProcessorModel{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
		Target:  types.StringValue(src.Target),
	}

	if src.File != nil {
		file := enrichmentFileModel{
			Path: types.StringValue(src.File.Path),
		}
		file.Encoding = fileEncodingModel{
			Type:            types.StringValue(string(src.File.Encoding.Type)),
			Delimiter:       types.StringValue(src.File.Encoding.Delimiter),
			IncludesHeaders: types.BoolValue(src.File.Encoding.IncludesHeaders),
		}
		for _, s := range src.File.Schema {
			file.Schema = append(file.Schema, fileSchemaItemModel{
				Column: types.StringValue(s.Column),
				Type:   types.StringValue(string(s.Type)),
			})
		}
		for _, k := range src.File.Key {
			file.Key = append(file.Key, fileKeyItemModel{
				Column:     types.StringValue(k.Column),
				Comparison: types.StringValue(string(k.Comparison)),
				Field:      types.StringValue(k.Field),
			})
		}
		out.File = &file
	}

	if src.Geoip != nil {
		out.GeoIp = &enrichmentGeoIpModel{
			KeyField: types.StringValue(src.Geoip.KeyField),
			Locale:   types.StringValue(src.Geoip.Locale),
			Path:     types.StringValue(src.Geoip.Path),
		}
	}

	return out
}

func expandGoogleChronicleDestination(ctx context.Context, src *googleChronicleDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineGoogleChronicleDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	if src.Auth != nil {
		auth := datadogV2.ObservabilityPipelineGcpAuth{}
		if !src.Auth.CredentialsFile.IsNull() {
			auth.SetCredentialsFile(src.Auth.CredentialsFile.ValueString())
		}
		dest.Auth = auth
	}

	if !src.CustomerId.IsNull() {
		dest.SetCustomerId(src.CustomerId.ValueString())
	}
	if !src.Encoding.IsNull() {
		dest.SetEncoding(datadogV2.ObservabilityPipelineGoogleChronicleDestinationEncoding(src.Encoding.ValueString()))
	}
	if !src.LogType.IsNull() {
		dest.SetLogType(src.LogType.ValueString())
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineGoogleChronicleDestination: dest,
	}
}

func flattenGoogleChronicleDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineGoogleChronicleDestination) *googleChronicleDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	out := &googleChronicleDestinationModel{
		Id:         types.StringValue(src.GetId()),
		Inputs:     inputs,
		CustomerId: types.StringValue(src.GetCustomerId()),
		Encoding:   types.StringValue(string(src.GetEncoding())),
		LogType:    types.StringValue(src.GetLogType()),
	}

	out.Auth = &gcpAuthModel{
		CredentialsFile: types.StringValue(src.Auth.CredentialsFile),
	}

	return out
}

func expandNewRelicDestination(ctx context.Context, src *newRelicDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineNewRelicDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	dest.SetRegion(datadogV2.ObservabilityPipelineNewRelicDestinationRegion(src.Region.ValueString()))

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineNewRelicDestination: dest,
	}
}

func flattenNewRelicDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineNewRelicDestination) *newRelicDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	return &newRelicDestinationModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
		Region: types.StringValue(string(src.GetRegion())),
	}
}

func expandSentinelOneDestination(ctx context.Context, src *sentinelOneDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineSentinelOneDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	dest.SetRegion(datadogV2.ObservabilityPipelineSentinelOneDestinationRegion(src.Region.ValueString()))

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSentinelOneDestination: dest,
	}
}

func flattenSentinelOneDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSentinelOneDestination) *sentinelOneDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	return &sentinelOneDestinationModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
		Region: types.StringValue(string(src.GetRegion())),
	}
}

func expandOcsfMapperProcessor(ctx context.Context, src *ocsfMapperProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineOcsfMapperProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	var mappings []datadogV2.ObservabilityPipelineOcsfMapperProcessorMapping
	for _, m := range src.Mapping {
		library := datadogV2.ObservabilityPipelineOcsfMappingLibrary(m.LibraryMapping.ValueString())
		mappings = append(mappings, datadogV2.ObservabilityPipelineOcsfMapperProcessorMapping{
			Include: m.Include.ValueString(),
			Mapping: datadogV2.ObservabilityPipelineOcsfMapperProcessorMappingMapping{
				ObservabilityPipelineOcsfMappingLibrary: &library,
			},
		})
	}

	proc.SetMappings(mappings)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineOcsfMapperProcessor: proc,
	}
}

func flattenOcsfMapperProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineOcsfMapperProcessor) *ocsfMapperProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	var mappings []ocsfMappingModel
	for _, m := range src.Mappings {
		mapping := ocsfMappingModel{
			Include:        types.StringValue(m.Include),
			LibraryMapping: types.StringValue(string(*m.Mapping.ObservabilityPipelineOcsfMappingLibrary)),
		}
		mappings = append(mappings, mapping)
	}

	return &ocsfMapperProcessorModel{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
		Mapping: mappings,
	}
}

func expandOpenSearchDestination(ctx context.Context, src *opensearchDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineOpenSearchDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	if !src.BulkIndex.IsNull() {
		dest.SetBulkIndex(src.BulkIndex.ValueString())
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineOpenSearchDestination: dest,
	}
}

func flattenOpenSearchDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineOpenSearchDestination) *opensearchDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())
	out := &opensearchDestinationModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
	}
	if v, ok := src.GetBulkIndexOk(); ok {
		out.BulkIndex = types.StringValue(*v)
	}

	return out
}

func expandAmazonOpenSearchDestination(ctx context.Context, src *amazonOpenSearchDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineAmazonOpenSearchDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	if !src.BulkIndex.IsNull() {
		dest.SetBulkIndex(src.BulkIndex.ValueString())
	}

	if src.Auth != nil {
		auth := datadogV2.ObservabilityPipelineAmazonOpenSearchDestinationAuth{
			Strategy: datadogV2.ObservabilityPipelineAmazonOpenSearchDestinationAuthStrategy(src.Auth.Strategy.ValueString()),
		}
		if !src.Auth.AwsRegion.IsNull() {
			auth.AwsRegion = src.Auth.AwsRegion.ValueStringPointer()
		}
		if !src.Auth.AssumeRole.IsNull() {
			auth.AssumeRole = src.Auth.AssumeRole.ValueStringPointer()
		}
		if !src.Auth.ExternalId.IsNull() {
			auth.ExternalId = src.Auth.ExternalId.ValueStringPointer()
		}
		if !src.Auth.SessionName.IsNull() {
			auth.SessionName = src.Auth.SessionName.ValueStringPointer()
		}
		dest.SetAuth(auth)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineAmazonOpenSearchDestination: dest,
	}
}

func flattenAmazonOpenSearchDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineAmazonOpenSearchDestination) *amazonOpenSearchDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	model := &amazonOpenSearchDestinationModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
	}

	if v, ok := src.GetBulkIndexOk(); ok {
		model.BulkIndex = types.StringValue(*v)
	}

	model.Auth = &amazonOpenSearchAuthModel{
		Strategy:    types.StringValue(string(src.Auth.Strategy)),
		AwsRegion:   types.StringPointerValue(src.Auth.AwsRegion),
		AssumeRole:  types.StringPointerValue(src.Auth.AssumeRole),
		ExternalId:  types.StringPointerValue(src.Auth.ExternalId),
		SessionName: types.StringPointerValue(src.Auth.SessionName),
	}

	return model
}

func expandDatadogTagsProcessor(ctx context.Context, src *observability_pipeline.DatadogTagsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	return observability_pipeline.ExpandDatadogTagsProcessor(ctx, src)
}

func flattenDatadogTagsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineDatadogTagsProcessor) *observability_pipeline.DatadogTagsProcessorModel {
	return observability_pipeline.FlattenDatadogTagsProcessor(ctx, src)
}

func expandObservabilityPipelinesAmazonSecurityLakeDestination(ctx context.Context, src *amazonSecurityLakeDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineAmazonSecurityLakeDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	if !src.Bucket.IsNull() {
		dest.SetBucket(src.Bucket.ValueString())
	}
	if !src.Region.IsNull() {
		dest.SetRegion(src.Region.ValueString())
	}
	if !src.CustomSourceName.IsNull() {
		dest.SetCustomSourceName(src.CustomSourceName.ValueString())
	}
	if src.Tls != nil {
		dest.Tls = expandTls(src.Tls)
	}
	if src.Auth != nil {
		dest.Auth = observability_pipeline.ExpandAwsAuth(src.Auth)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineAmazonSecurityLakeDestination: dest,
	}
}
