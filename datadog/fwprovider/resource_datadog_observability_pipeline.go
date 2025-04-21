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
	AmazonS3Source     []*amazonS3SourceModel     `tfsdk:"amazon_s3"`
	SplunkHecSource    []*splunkHecSourceModel    `tfsdk:"splunk_hec"`
	SplunkTcpSource    []*splunkTcpSourceModel    `tfsdk:"splunk_tcp"`
}

// / Source models
type datadogAgentSourceModel struct {
	Id  types.String `tfsdk:"id"`
	Tls []tlsModel   `tfsdk:"tls"`
}

type kafkaSourceModel struct {
	Id                types.String            `tfsdk:"id"`
	GroupId           types.String            `tfsdk:"group_id"`
	Topics            []types.String          `tfsdk:"topics"`
	LibrdkafkaOptions []librdkafkaOptionModel `tfsdk:"librdkafka_option"`
	Sasl              *kafkaSourceSaslModel   `tfsdk:"sasl"`
	Tls               []tlsModel              `tfsdk:"tls"`
}

type librdkafkaOptionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type kafkaSourceSaslModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
}

type amazonS3SourceModel struct {
	Id     types.String  `tfsdk:"id"`     // Unique identifier for the component
	Region types.String  `tfsdk:"region"` // AWS region where the S3 bucket resides
	Auth   *awsAuthModel `tfsdk:"auth"`   // AWS authentication credentials
	Tls    []tlsModel    `tfsdk:"tls"`    // TLS encryption configuration
}

type awsAuthModel struct {
	AssumeRole  types.String `tfsdk:"assume_role"`  // ARN of the role to assume
	ExternalId  types.String `tfsdk:"external_id"`  // Unique identifier for cross-account assumption
	SessionName types.String `tfsdk:"session_name"` // Session identifier for logging/tracing
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
	GenerateMetricsProcessor []*generateMetricsProcessorModel `tfsdk:"generate_datadog_metrics"`
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
	DatadogLogsDestination        []*datadogLogsDestinationModel `tfsdk:"datadog_logs"`
	GoogleCloudStorageDestination []*gcsDestinationModel         `tfsdk:"google_cloud_storage"`
	SplunkHecDestination          []*splunkHecDestinationModel   `tfsdk:"splunk_hec"`
}
type datadogLogsDestinationModel struct {
	Id     types.String `tfsdk:"id"`
	Inputs types.List   `tfsdk:"inputs"`
}

type splunkHecSourceModel struct {
	Id  types.String `tfsdk:"id"`  // The unique identifier for this component.
	Tls []tlsModel   `tfsdk:"tls"` // TLS encryption settings for secure ingestion.
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
	Tls []tlsModel   `tfsdk:"tls"` // TLS encryption settings for secure transmission.
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

type gcpAuthModel struct {
	CredentialsFile types.String `tfsdk:"credentials_file"`
}

type metadataEntry struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
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
						},
					},
				},
			},
		},
	}
}

func tlsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			// this is the only way to make the block optional
			listvalidator.SizeAtMost(1),
		},
		Description: "Configuration for enabling TLS encryption between the pipeline component and external services.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"crt_file": schema.StringAttribute{
					Required:    true,
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
	for _, s := range state.Config.Sources.SplunkHecSource {
		config.Sources = append(config.Sources, expandSplunkHecSource(s))
	}
	for _, s := range state.Config.Sources.SplunkTcpSource {
		config.Sources = append(config.Sources, expandSplunkTcpSource(s))
	}
	for _, s := range state.Config.Sources.AmazonS3Source {
		config.Sources = append(config.Sources, expandAmazonS3Source(s))
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
	for _, p := range state.Config.Processors.GenerateMetricsProcessor {
		config.Processors = append(config.Processors, expandGenerateMetricsProcessor(ctx, p))
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

		if s := flattenSplunkHecSource(src.ObservabilityPipelineSplunkHecSource); s != nil {
			outCfg.Sources.SplunkHecSource = append(outCfg.Sources.SplunkHecSource, s)
		}

		if s := flattenSplunkTcpSource(src.ObservabilityPipelineSplunkTcpSource); s != nil {
			outCfg.Sources.SplunkTcpSource = append(outCfg.Sources.SplunkTcpSource, s)
		}

		if s3 := flattenAmazonS3Source(src.ObservabilityPipelineAmazonS3Source); s3 != nil {
			outCfg.Sources.AmazonS3Source = append(outCfg.Sources.AmazonS3Source, s3)
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

		if f := flattenGenerateMetricsProcessor(ctx, p.ObservabilityPipelineGenerateMetricsProcessor); f != nil {
			outCfg.Processors.GenerateMetricsProcessor = append(outCfg.Processors.GenerateMetricsProcessor, f)
		}

	}

	for _, d := range cfg.GetDestinations() {

		if logs := flattenDatadogLogsDestination(ctx, d.ObservabilityPipelineDatadogLogsDestination); logs != nil {
			outCfg.Destinations.DatadogLogsDestination = append(outCfg.Destinations.DatadogLogsDestination, logs)
		}
		if hec := flattenSplunkHecDestination(ctx, d.ObservabilityPipelineSplunkHecDestination); hec != nil {
			outCfg.Destinations.SplunkHecDestination = append(outCfg.Destinations.SplunkHecDestination, hec)
		}

		if gcs := flattenGoogleCloudStorageDestination(ctx, d.ObservabilityPipelineGoogleCloudStorageDestination); gcs != nil {
			outCfg.Destinations.GoogleCloudStorageDestination = append(outCfg.Destinations.GoogleCloudStorageDestination, gcs)
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
		out.Tls = []tlsModel{flattenTls(src.Tls)}
	}
	return out
}

func expandDatadogAgentSource(src *datadogAgentSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	agent := datadogV2.NewObservabilityPipelineDatadogAgentSourceWithDefaults()
	agent.SetId(src.Id.ValueString())
	if len(src.Tls) > 0 {
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
		out.Tls = []tlsModel{flattenTls(src.Tls)}
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

	if len(src.Tls) > 0 {
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

func expandTls(src []tlsModel) *datadogV2.ObservabilityPipelineTls {
	tls := &datadogV2.ObservabilityPipelineTls{}
	// there must be no more than one TLS block
	tlsTF := src[0]
	tls.SetCrtFile(tlsTF.CrtFile.ValueString())
	if !tlsTF.CaFile.IsNull() {
		tls.SetCaFile(tlsTF.CaFile.ValueString())
	}
	if !tlsTF.KeyFile.IsNull() {
		tls.SetKeyFile(tlsTF.KeyFile.ValueString())
	}
	return tls
}

func expandSplunkHecSource(src *splunkHecSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineSplunkHecSourceWithDefaults()

	s.SetId(src.Id.ValueString())

	if len(src.Tls) > 0 {
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
		out.Tls = []tlsModel{flattenTls(src.Tls)}
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
		KeyPrefix:    types.StringPointerValue(src.KeyPrefix.Get()),
		StorageClass: types.StringValue(string(src.GetStorageClass())),
		Acl:          types.StringValue(string(src.GetAcl())),
		Auth: gcpAuthModel{
			CredentialsFile: types.StringValue(src.Auth.CredentialsFile),
		},
		Metadata: metadata,
		Inputs:   inputs,
	}
}

func expandSplunkTcpSource(src *splunkTcpSourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineSplunkTcpSourceWithDefaults()
	s.SetId(src.Id.ValueString())

	if len(src.Tls) > 0 {
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
		out.Tls = []tlsModel{flattenTls(src.Tls)}
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
		s.SetAuth(auth)
	}

	if len(src.Tls) > 0 {
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

	if src.Auth != nil {
		out.Auth = &awsAuthModel{
			AssumeRole:  types.StringPointerValue(src.Auth.AssumeRole.Get()),
			ExternalId:  types.StringPointerValue(src.Auth.ExternalId.Get()),
			SessionName: types.StringPointerValue(src.Auth.SessionName.Get()),
		}
	}

	if src.Tls != nil {
		out.Tls = []tlsModel{flattenTls(src.Tls)}
	}

	return out
}
