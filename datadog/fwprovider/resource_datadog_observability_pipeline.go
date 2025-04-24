package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	Config configModel  `tfsdk:"config"`
}

type configModel struct {
	Sources      sourcesModel      `tfsdk:"sources"`
	Processors   *processorsModel  `tfsdk:"processors"`
	Destinations destinationsModel `tfsdk:"destinations"`
}
type sourcesModel struct {
	DatadogAgentSource       []*datadogAgentSourceModel       `tfsdk:"datadog_agent"`
	KafkaSource              []*kafkaSourceModel              `tfsdk:"kafka"`
	AmazonDataFirehoseSource []*amazonDataFirehoseSourceModel `tfsdk:"amazon_data_firehose"`
	HttpClientSource         []*httpClientSourceModel         `tfsdk:"http_client"`
	GooglePubSubSource       []*googlePubSubSourceModel       `tfsdk:"google_pubsub"`
	LogstashSource           []*logstashSourceModel           `tfsdk:"logstash"`
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

type tlsModel struct {
	CrtFile types.String `tfsdk:"crt_file"`
	CaFile  types.String `tfsdk:"ca_file"`
	KeyFile types.String `tfsdk:"key_file"`
}

// Processor models

type processorsModel struct {
	FilterProcessor          []*filterProcessorModel          `tfsdk:"filter"`
	ParseJsonProcessor       []*parseJsonProcessorModel       `tfsdk:"parse_json"`
	AddFieldsProcessor       []*addFieldsProcessor            `tfsdk:"add_fields"`
	RenameFieldsProcessor    []*renameFieldsProcessorModel    `tfsdk:"rename_fields"`
	RemoveFieldsProcessor    []*removeFieldsProcessorModel    `tfsdk:"remove_fields"`
	QuotaProcessor           []*quotaProcessorModel           `tfsdk:"quota"`
	DedupeProcessor          []*dedupeProcessorModel          `tfsdk:"dedupe"`
	ReduceProcessor          []*reduceProcessorModel          `tfsdk:"reduce"`
	ThrottleProcessor        []*throttleProcessorModel        `tfsdk:"throttle"`
	AddEnvVarsProcessor      []*addEnvVarsProcessorModel      `tfsdk:"add_env_vars"`
	EnrichmentTableProcessor []*enrichmentTableProcessorModel `tfsdk:"enrichment_table"`
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
	DatadogLogsDestination     []*datadogLogsDestinationModel     `tfsdk:"datadog_logs"`
	GoogleChronicleDestination []*googleChronicleDestinationModel `tfsdk:"google_chronicle"`
	NewRelicDestination        []*newRelicDestinationModel        `tfsdk:"new_relic"`
	SentinelOneDestination     []*sentinelOneDestinationModel     `tfsdk:"sentinel_one"`
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

type datadogLogsDestinationModel struct {
	Id     types.String `tfsdk:"id"`
	Inputs types.List   `tfsdk:"inputs"`
}

type amazonDataFirehoseSourceModel struct {
	Id   types.String  `tfsdk:"id"`
	Auth *awsAuthModel `tfsdk:"auth"`
	Tls  *tlsModel     `tfsdk:"tls"`
}

type awsAuthModel struct {
	AssumeRole  types.String `tfsdk:"assume_role"`
	ExternalId  types.String `tfsdk:"external_id"`
	SessionName types.String `tfsdk:"session_name"`
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
		Description: "Provides a Datadog Observability Pipeline resource. Observability Pipelines allows you to collect and process logs within your own infrastructure, and then route them to downstream integrations.",
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
													Required:    true,
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
											Description: "AWS authentication credentials used for accessing AWS services such as S3. If omitted, the system’s default credentials are used (for example, the IAM role and environment variables).",
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
											Description: "he inputs for the processor.",
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
				Description: "Path to the Certificate Authority (CA) file used to validate the server’s TLS certificate.",
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

	body, diags := expandPipelineRequest(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := datadogV2.NewObservabilityPipelineCreateRequestWithDefaults()
	createReq.Data = *datadogV2.NewObservabilityPipelineCreateRequestDataWithDefaults()
	createReq.Data.Attributes = body.Data.Attributes

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
	body, diags := expandPipelineRequest(ctx, &state)
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
func expandPipelineRequest(ctx context.Context, state *observabilityPipelineModel) (*datadogV2.ObservabilityPipeline, diag.Diagnostics) {
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

	// Processors
	if state.Config.Processors != nil {
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
	}

	// Destinations
	for _, d := range state.Config.Destinations.DatadogLogsDestination {
		config.Destinations = append(config.Destinations, expandDatadogLogsDestination(ctx, d))
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
	outCfg := configModel{
		Processors: &processorsModel{},
	}

	for _, src := range cfg.GetSources() {
		if a := flattenDatadogAgentSource(src.ObservabilityPipelineDatadogAgentSource); a != nil {
			outCfg.Sources.DatadogAgentSource = append(outCfg.Sources.DatadogAgentSource, a)
		}
		if k := flattenKafkaSource(src.ObservabilityPipelineKafkaSource); k != nil {
			outCfg.Sources.KafkaSource = append(outCfg.Sources.KafkaSource, k)
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
	}

	state.Config = outCfg
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
		Id:                          types.StringValue(src.Id),
		Include:                     types.StringValue(src.Include),
		Name:                        types.StringValue(src.Name),
		DropEvents:                  types.BoolValue(src.DropEvents),
		IgnoreWhenMissingPartitions: types.BoolValue(src.GetIgnoreWhenMissingPartitions()),
		Inputs:                      inputs,
		PartitionFields:             partitions,
		Limit: quotaLimitModel{
			Enforce: types.StringValue(string(src.Limit.Enforce)),
			Limit:   types.Int64Value(src.Limit.Limit),
		},
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
	proc.SetIgnoreWhenMissingPartitions(src.IgnoreWhenMissingPartitions.ValueBool())

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

func expandAmazonDataFirehoseSource(src *amazonDataFirehoseSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	firehose := datadogV2.NewObservabilityPipelineAmazonDataFirehoseSourceWithDefaults()
	firehose.SetId(src.Id.ValueString())

	if src.Auth != nil {
		auth := datadogV2.ObservabilityPipelineAwsAuth{}
		if !src.Auth.AssumeRole.IsNull() {
			auth.SetAssumeRole(src.Auth.AssumeRole.ValueString())
		}
		if !src.Auth.ExternalId.IsNull() {
			auth.SetExternalId(src.Auth.ExternalId.ValueString())
		}
		if !src.Auth.SessionName.IsNull() {
			auth.SetSessionName(src.Auth.SessionName.ValueString())
		}
		firehose.SetAuth(auth)
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

	if src.Auth != nil {
		auth := awsAuthModel{}
		if v, ok := src.GetAuthOk(); ok {
			auth = awsAuthModel{
				AssumeRole:  types.StringPointerValue(v.AssumeRole),
				ExternalId:  types.StringPointerValue(v.ExternalId),
				SessionName: types.StringPointerValue(v.SessionName),
			}
		}
		out.Auth = &auth
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
