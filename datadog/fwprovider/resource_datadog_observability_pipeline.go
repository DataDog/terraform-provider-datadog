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
	Processors   processorsModel   `tfsdk:"processors"`
	Destinations destinationsModel `tfsdk:"destinations"`
}
type sourcesModel struct {
	DatadogAgentSource []*datadogAgentSourceModel `tfsdk:"datadog_agent"`
	KafkaSource        []*kafkaSourceModel        `tfsdk:"kafka"`
	RsyslogSource      []*rsyslogSourceModel      `tfsdk:"rsyslog"`
	SyslogNgSource     []*syslogNgSourceModel     `tfsdk:"syslog_ng"`
}

// / Source models
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
	FilterProcessor               []*filterProcessorModel               `tfsdk:"filter"`
	ParseJsonProcessor            []*parseJsonProcessorModel            `tfsdk:"parse_json"`
	AddFieldsProcessor            []*addFieldsProcessor                 `tfsdk:"add_fields"`
	RenameFieldsProcessor         []*renameFieldsProcessorModel         `tfsdk:"rename_fields"`
	RemoveFieldsProcessor         []*removeFieldsProcessorModel         `tfsdk:"remove_fields"`
	QuotaProcessor                []*quotaProcessorModel                `tfsdk:"quota"`
	SensitiveDataScannerProcessor []*sensitiveDataScannerProcessorModel `tfsdk:"sensitive_data_scanner"`
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
	DatadogLogsDestination       []*datadogLogsDestinationModel       `tfsdk:"datadog_logs"`
	SumoLogicDestination         []*sumoLogicDestinationModel         `tfsdk:"sumo_logic"`
	RsyslogDestination           []*rsyslogDestinationModel           `tfsdk:"rsyslog"`
	SyslogNgDestination          []*syslogNgDestinationModel          `tfsdk:"syslog_ng"`
	ElasticsearchDestination     []*elasticsearchDestinationModel     `tfsdk:"elasticsearch"`
	AzureStorageDestination      []*azureStorageDestinationModel      `tfsdk:"azure_storage"`
	MicrosoftSentinelDestination []*microsoftSentinelDestinationModel `tfsdk:"microsoft_sentinel"`
}

type datadogLogsDestinationModel struct {
	Id     types.String `tfsdk:"id"`
	Inputs types.List   `tfsdk:"inputs"`
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
	for _, s := range state.Config.Sources.RsyslogSource {
		config.Sources = append(config.Sources, expandRsyslogSource(s))
	}
	for _, s := range state.Config.Sources.SyslogNgSource {
		config.Sources = append(config.Sources, expandSyslogNgSource(s))
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
	for _, p := range state.Config.Processors.SensitiveDataScannerProcessor {
		config.Processors = append(config.Processors, expandSensitiveDataScannerProcessor(ctx, p))
	}

	// Destinations
	for _, d := range state.Config.Destinations.DatadogLogsDestination {
		config.Destinations = append(config.Destinations, expandDatadogLogsDestination(ctx, d))
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
		if r := flattenRsyslogSource(src.ObservabilityPipelineRsyslogSource); r != nil {
			outCfg.Sources.RsyslogSource = append(outCfg.Sources.RsyslogSource, r)
		}
		if s := flattenSyslogNgSource(src.ObservabilityPipelineSyslogNgSource); s != nil {
			outCfg.Sources.SyslogNgSource = append(outCfg.Sources.SyslogNgSource, s)
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
		if s := flattenSensitiveDataScannerProcessor(ctx, p.ObservabilityPipelineSensitiveDataScannerProcessor); s != nil {
			outCfg.Processors.SensitiveDataScannerProcessor = append(outCfg.Processors.SensitiveDataScannerProcessor, s)
		}
	}
	for _, d := range cfg.GetDestinations() {
		if logs := flattenDatadogLogsDestination(ctx, d.ObservabilityPipelineDatadogLogsDestination); logs != nil {
			outCfg.Destinations.DatadogLogsDestination = append(outCfg.Destinations.DatadogLogsDestination, logs)
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
