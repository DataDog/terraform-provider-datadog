package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KafkaDestinationModel represents the Terraform model for kafka destination configuration
type KafkaDestinationModel struct {
	Encoding              types.String            `tfsdk:"encoding"`
	Topic                 types.String            `tfsdk:"topic"`
	Compression           types.String            `tfsdk:"compression"`
	HeadersKey            types.String            `tfsdk:"headers_key"`
	KeyField              types.String            `tfsdk:"key_field"`
	MessageTimeoutMs      types.Int64             `tfsdk:"message_timeout_ms"`
	RateLimitDurationSecs types.Int64             `tfsdk:"rate_limit_duration_secs"`
	RateLimitNum          types.Int64             `tfsdk:"rate_limit_num"`
	SocketTimeoutMs       types.Int64             `tfsdk:"socket_timeout_ms"`
	Sasl                  []KafkaSaslModel        `tfsdk:"sasl"`
	LibrdkafkaOptions     []LibrdkafkaOptionModel `tfsdk:"librdkafka_option"`
	Tls                   []TlsModel              `tfsdk:"tls"`
}

// KafkaSaslModel represents SASL configuration
type KafkaSaslModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
}

// LibrdkafkaOptionModel represents a librdkafka configuration option
type LibrdkafkaOptionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// ExpandKafkaDestination converts the Terraform model to the Datadog API model
func ExpandKafkaDestination(ctx context.Context, id string, inputs types.List, src *KafkaDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	d := datadogV2.NewObservabilityPipelineKafkaDestinationWithDefaults()
	d.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	d.SetInputs(inputsList)

	// Required fields
	d.SetEncoding(datadogV2.ObservabilityPipelineKafkaDestinationEncoding(src.Encoding.ValueString()))
	d.SetTopic(src.Topic.ValueString())

	// Optional string fields
	if !src.HeadersKey.IsNull() {
		d.SetHeadersKey(src.HeadersKey.ValueString())
	}
	if !src.KeyField.IsNull() {
		d.SetKeyField(src.KeyField.ValueString())
	}

	// Optional compression
	if !src.Compression.IsNull() {
		d.SetCompression(datadogV2.ObservabilityPipelineKafkaDestinationCompression(src.Compression.ValueString()))
	}

	// Optional int64 fields
	if !src.MessageTimeoutMs.IsNull() {
		d.SetMessageTimeoutMs(src.MessageTimeoutMs.ValueInt64())
	}
	if !src.RateLimitDurationSecs.IsNull() {
		d.SetRateLimitDurationSecs(src.RateLimitDurationSecs.ValueInt64())
	}
	if !src.RateLimitNum.IsNull() {
		d.SetRateLimitNum(src.RateLimitNum.ValueInt64())
	}
	if !src.SocketTimeoutMs.IsNull() {
		d.SetSocketTimeoutMs(src.SocketTimeoutMs.ValueInt64())
	}

	// SASL configuration
	if len(src.Sasl) > 0 {
		sasl := src.Sasl[0]
		mechanism, _ := datadogV2.NewObservabilityPipelineKafkaSaslMechanismFromValue(sasl.Mechanism.ValueString())
		if mechanism != nil {
			saslConfig := datadogV2.ObservabilityPipelineKafkaSasl{}
			saslConfig.SetMechanism(*mechanism)
			d.SetSasl(saslConfig)
		}
	}

	// Librdkafka options
	if len(src.LibrdkafkaOptions) > 0 {
		opts := []datadogV2.ObservabilityPipelineKafkaLibrdkafkaOption{}
		for _, opt := range src.LibrdkafkaOptions {
			opts = append(opts, datadogV2.ObservabilityPipelineKafkaLibrdkafkaOption{
				Name:  opt.Name.ValueString(),
				Value: opt.Value.ValueString(),
			})
		}
		d.SetLibrdkafkaOptions(opts)
	}

	// TLS configuration
	if len(src.Tls) > 0 {
		d.Tls = ExpandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineKafkaDestination: d,
	}
}

// FlattenKafkaDestination converts the Datadog API model to the Terraform model
func FlattenKafkaDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineKafkaDestination) *KafkaDestinationModel {
	if src == nil {
		return nil
	}

	out := &KafkaDestinationModel{
		Encoding: types.StringValue(string(src.GetEncoding())),
		Topic:    types.StringValue(src.GetTopic()),
	}

	// Optional string fields
	if v, ok := src.GetHeadersKeyOk(); ok {
		out.HeadersKey = types.StringValue(*v)
	}
	if v, ok := src.GetKeyFieldOk(); ok {
		out.KeyField = types.StringValue(*v)
	}

	// Optional compression
	if v, ok := src.GetCompressionOk(); ok {
		out.Compression = types.StringValue(string(*v))
	}

	// Optional int64 fields
	if v, ok := src.GetMessageTimeoutMsOk(); ok {
		out.MessageTimeoutMs = types.Int64Value(*v)
	}
	if v, ok := src.GetRateLimitDurationSecsOk(); ok {
		out.RateLimitDurationSecs = types.Int64Value(*v)
	}
	if v, ok := src.GetRateLimitNumOk(); ok {
		out.RateLimitNum = types.Int64Value(*v)
	}
	if v, ok := src.GetSocketTimeoutMsOk(); ok {
		out.SocketTimeoutMs = types.Int64Value(*v)
	}

	// SASL configuration
	if sasl, ok := src.GetSaslOk(); ok {
		out.Sasl = []KafkaSaslModel{
			{
				Mechanism: types.StringValue(string(sasl.GetMechanism())),
			},
		}
	}

	// Librdkafka options
	for _, opt := range src.GetLibrdkafkaOptions() {
		out.LibrdkafkaOptions = append(out.LibrdkafkaOptions, LibrdkafkaOptionModel{
			Name:  types.StringValue(opt.Name),
			Value: types.StringValue(opt.Value),
		})
	}

	// TLS configuration
	if src.Tls != nil {
		out.Tls = FlattenTls(src.Tls)
	}

	return out
}

// KafkaDestinationSchema returns the schema for kafka destination
func KafkaDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `kafka` destination sends logs to Apache Kafka topics.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"encoding": schema.StringAttribute{
					Required:    true,
					Description: "Encoding format for log events.",
					Validators: []validator.String{
						stringvalidator.OneOf("json", "raw_message"),
					},
				},
				"topic": schema.StringAttribute{
					Required:    true,
					Description: "The Kafka topic name to publish logs to.",
				},
				"compression": schema.StringAttribute{
					Optional:    true,
					Description: "Compression codec for Kafka messages.",
					Validators: []validator.String{
						stringvalidator.OneOf("none", "gzip", "snappy", "lz4", "zstd"),
					},
				},
				"headers_key": schema.StringAttribute{
					Optional:    true,
					Description: "The field name to use for Kafka message headers.",
				},
				"key_field": schema.StringAttribute{
					Optional:    true,
					Description: "The field name to use as the Kafka message key.",
				},
				"message_timeout_ms": schema.Int64Attribute{
					Optional:    true,
					Description: "Maximum time in milliseconds to wait for message delivery confirmation.",
				},
				"rate_limit_duration_secs": schema.Int64Attribute{
					Optional:    true,
					Description: "Duration in seconds for the rate limit window.",
				},
				"rate_limit_num": schema.Int64Attribute{
					Optional:    true,
					Description: "Maximum number of messages allowed per rate limit duration.",
				},
				"socket_timeout_ms": schema.Int64Attribute{
					Optional:    true,
					Description: "Socket timeout in milliseconds for network requests.",
				},
			},
			Blocks: map[string]schema.Block{
				"sasl": schema.ListNestedBlock{
					Description: "Specifies the SASL mechanism for authenticating with a Kafka cluster.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"mechanism": schema.StringAttribute{
								Required:    true,
								Description: "SASL authentication mechanism.",
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
				"librdkafka_option": schema.ListNestedBlock{
					Description: "Optional list of advanced Kafka producer configuration options, defined as key-value pairs.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Required:    true,
								Description: "The name of the librdkafka configuration option.",
							},
							"value": schema.StringAttribute{
								Required:    true,
								Description: "The value of the librdkafka configuration option.",
							},
						},
					},
				},
				"tls": TlsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
	}
}
