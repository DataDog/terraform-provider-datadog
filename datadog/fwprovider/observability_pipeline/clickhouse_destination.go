package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ClickhouseDestinationModel represents the Terraform model for the ClickHouse destination.
type ClickhouseDestinationModel struct {
	EndpointUrlKey     types.String                   `tfsdk:"endpoint_url_key"`
	Database           types.String                   `tfsdk:"database"`
	Table              types.String                   `tfsdk:"table"`
	Format             types.String                   `tfsdk:"format"`
	SkipUnknownFields  types.Bool                     `tfsdk:"skip_unknown_fields"`
	DateTimeBestEffort types.Bool                     `tfsdk:"date_time_best_effort"`
	Compression        []ClickhouseCompressionModel   `tfsdk:"compression"`
	Auth               []ClickhouseAuthModel          `tfsdk:"auth"`
	Batch              []ClickhouseBatchModel         `tfsdk:"batch"`
	BatchEncoding      []ClickhouseBatchEncodingModel `tfsdk:"batch_encoding"`
	Tls                []TlsModel                     `tfsdk:"tls"`
	Buffer             []BufferOptionsModel           `tfsdk:"buffer"`
}

// ClickhouseCompressionModel represents the compression configuration for the ClickHouse destination.
type ClickhouseCompressionModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
	Level     types.Int64  `tfsdk:"level"`
}

// ClickhouseAuthModel represents the authentication configuration for the ClickHouse destination.
type ClickhouseAuthModel struct {
	Strategy    types.String `tfsdk:"strategy"`
	UsernameKey types.String `tfsdk:"username_key"`
	PasswordKey types.String `tfsdk:"password_key"`
}

// ClickhouseBatchModel represents the batch configuration for the ClickHouse destination.
type ClickhouseBatchModel struct {
	MaxEvents   types.Int64 `tfsdk:"max_events"`
	TimeoutSecs types.Int64 `tfsdk:"timeout_secs"`
}

// ClickhouseBatchEncodingModel represents the batch encoding configuration for the ClickHouse destination.
type ClickhouseBatchEncodingModel struct {
	Codec               types.String `tfsdk:"codec"`
	AllowNullableFields types.Bool   `tfsdk:"allow_nullable_fields"`
}

// ExpandClickhouseDestination converts the Terraform model to the Datadog API model.
func ExpandClickhouseDestination(ctx context.Context, id string, inputs types.List, src *ClickhouseDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineClickhouseDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	dest.SetTable(src.Table.ValueString())

	if !src.EndpointUrlKey.IsNull() && !src.EndpointUrlKey.IsUnknown() {
		dest.SetEndpointUrlKey(src.EndpointUrlKey.ValueString())
	}
	if !src.Database.IsNull() && !src.Database.IsUnknown() {
		dest.SetDatabase(src.Database.ValueString())
	}
	if !src.Format.IsNull() && !src.Format.IsUnknown() {
		dest.SetFormat(datadogV2.ObservabilityPipelineClickhouseDestinationFormat(src.Format.ValueString()))
	}
	if !src.SkipUnknownFields.IsNull() && !src.SkipUnknownFields.IsUnknown() {
		dest.SetSkipUnknownFields(src.SkipUnknownFields.ValueBool())
	}
	if !src.DateTimeBestEffort.IsNull() && !src.DateTimeBestEffort.IsUnknown() {
		dest.SetDateTimeBestEffort(src.DateTimeBestEffort.ValueBool())
	}

	if len(src.Compression) > 0 {
		c := src.Compression[0]
		obj := datadogV2.NewObservabilityPipelineClickhouseDestinationCompressionObjectWithDefaults()
		if !c.Algorithm.IsNull() && !c.Algorithm.IsUnknown() {
			obj.SetAlgorithm(datadogV2.ObservabilityPipelineClickhouseDestinationCompressionAlgorithm(c.Algorithm.ValueString()))
		}
		if !c.Level.IsNull() && !c.Level.IsUnknown() {
			obj.SetLevel(c.Level.ValueInt64())
		}
		dest.SetCompression(datadogV2.ObservabilityPipelineClickhouseDestinationCompressionObjectAsObservabilityPipelineClickhouseDestinationCompression(obj))
	}

	if len(src.Auth) > 0 {
		a := src.Auth[0]
		auth := datadogV2.NewObservabilityPipelineClickhouseDestinationAuthWithDefaults()
		strategy := "basic"
		if !a.Strategy.IsNull() && !a.Strategy.IsUnknown() {
			strategy = a.Strategy.ValueString()
		}
		auth.SetStrategy(datadogV2.ObservabilityPipelineClickhouseDestinationAuthStrategy(strategy))
		if !a.UsernameKey.IsNull() && !a.UsernameKey.IsUnknown() {
			auth.SetUsernameKey(a.UsernameKey.ValueString())
		}
		if !a.PasswordKey.IsNull() && !a.PasswordKey.IsUnknown() {
			auth.SetPasswordKey(a.PasswordKey.ValueString())
		}
		dest.SetAuth(*auth)
	}

	if len(src.Batch) > 0 {
		b := src.Batch[0]
		batch := datadogV2.NewObservabilityPipelineClickhouseDestinationBatchWithDefaults()
		if !b.MaxEvents.IsNull() && !b.MaxEvents.IsUnknown() {
			batch.SetMaxEvents(b.MaxEvents.ValueInt64())
		}
		if !b.TimeoutSecs.IsNull() && !b.TimeoutSecs.IsUnknown() {
			batch.SetTimeoutSecs(b.TimeoutSecs.ValueInt64())
		}
		dest.SetBatch(*batch)
	}

	if len(src.BatchEncoding) > 0 {
		be := src.BatchEncoding[0]
		batchEncoding := datadogV2.NewObservabilityPipelineClickhouseDestinationBatchEncodingWithDefaults()
		if !be.Codec.IsNull() && !be.Codec.IsUnknown() {
			batchEncoding.SetCodec(datadogV2.ObservabilityPipelineClickhouseDestinationBatchEncodingCodec(be.Codec.ValueString()))
		}
		if !be.AllowNullableFields.IsNull() && !be.AllowNullableFields.IsUnknown() {
			batchEncoding.SetAllowNullableFields(be.AllowNullableFields.ValueBool())
		}
		dest.SetBatchEncoding(*batchEncoding)
	}

	if len(src.Tls) > 0 {
		dest.Tls = ExpandTls(src.Tls)
	}

	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			dest.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineClickhouseDestination: dest,
	}
}

// FlattenClickhouseDestination converts the Datadog API model to the Terraform model.
func FlattenClickhouseDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineClickhouseDestination) *ClickhouseDestinationModel {
	if src == nil {
		return nil
	}

	model := &ClickhouseDestinationModel{
		Table: types.StringValue(src.GetTable()),
	}

	if v, ok := src.GetEndpointUrlKeyOk(); ok {
		model.EndpointUrlKey = types.StringValue(*v)
	} else {
		model.EndpointUrlKey = types.StringNull()
	}

	if v, ok := src.GetDatabaseOk(); ok {
		model.Database = types.StringValue(*v)
	} else {
		model.Database = types.StringNull()
	}

	if v, ok := src.GetFormatOk(); ok {
		model.Format = types.StringValue(string(*v))
	} else {
		model.Format = types.StringNull()
	}

	if v, ok := src.GetSkipUnknownFieldsOk(); ok {
		model.SkipUnknownFields = types.BoolValue(*v)
	} else {
		model.SkipUnknownFields = types.BoolNull()
	}

	if v, ok := src.GetDateTimeBestEffortOk(); ok {
		model.DateTimeBestEffort = types.BoolValue(*v)
	} else {
		model.DateTimeBestEffort = types.BoolNull()
	}

	if comp, ok := src.GetCompressionOk(); ok {
		cm := ClickhouseCompressionModel{Level: types.Int64Null()}
		if obj := comp.ObservabilityPipelineClickhouseDestinationCompressionObject; obj != nil {
			if v, ok2 := obj.GetAlgorithmOk(); ok2 {
				cm.Algorithm = types.StringValue(string(*v))
			}
			if v, ok2 := obj.GetLevelOk(); ok2 {
				cm.Level = types.Int64Value(*v)
			}
		} else if alg := comp.ObservabilityPipelineClickhouseDestinationCompressionAlgorithm; alg != nil {
			cm.Algorithm = types.StringValue(string(*alg))
		}
		model.Compression = []ClickhouseCompressionModel{cm}
	}

	if auth, ok := src.GetAuthOk(); ok {
		am := ClickhouseAuthModel{
			Strategy: types.StringValue("basic"),
		}
		if v, ok2 := auth.GetUsernameKeyOk(); ok2 {
			am.UsernameKey = types.StringValue(*v)
		} else {
			am.UsernameKey = types.StringNull()
		}
		if v, ok2 := auth.GetPasswordKeyOk(); ok2 {
			am.PasswordKey = types.StringValue(*v)
		} else {
			am.PasswordKey = types.StringNull()
		}
		model.Auth = []ClickhouseAuthModel{am}
	}

	if batch, ok := src.GetBatchOk(); ok {
		bm := ClickhouseBatchModel{}
		if v, ok2 := batch.GetMaxEventsOk(); ok2 {
			bm.MaxEvents = types.Int64Value(*v)
		} else {
			bm.MaxEvents = types.Int64Null()
		}
		if v, ok2 := batch.GetTimeoutSecsOk(); ok2 {
			bm.TimeoutSecs = types.Int64Value(*v)
		} else {
			bm.TimeoutSecs = types.Int64Null()
		}
		model.Batch = []ClickhouseBatchModel{bm}
	}

	if be, ok := src.GetBatchEncodingOk(); ok {
		bem := ClickhouseBatchEncodingModel{}
		if v, ok2 := be.GetCodecOk(); ok2 {
			bem.Codec = types.StringValue(string(*v))
		}
		if v, ok2 := be.GetAllowNullableFieldsOk(); ok2 {
			bem.AllowNullableFields = types.BoolValue(*v)
		} else {
			bem.AllowNullableFields = types.BoolNull()
		}
		model.BatchEncoding = []ClickhouseBatchEncodingModel{bem}
	}

	if src.Tls != nil {
		model.Tls = FlattenTls(src.Tls)
	}

	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			model.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	return model
}

// ClickhouseDestinationSchema returns the schema for the ClickHouse destination.
func ClickhouseDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `clickhouse` destination forwards logs to a ClickHouse server via HTTP.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"endpoint_url_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the ClickHouse HTTP endpoint URL. Defaults to `DESTINATION_CLICKHOUSE_ENDPOINT_URL`.",
				},
				"database": schema.StringAttribute{
					Optional:    true,
					Description: "Optional name of the ClickHouse database to write to. When omitted, the user's default database is used.",
				},
				"table": schema.StringAttribute{
					Required:    true,
					Description: "Target ClickHouse table name.",
				},
				"format": schema.StringAttribute{
					Optional:    true,
					Description: "Insert format for events. `json_each_row` maps event fields to columns by name. `json_as_object` and `json_as_string` insert each event into a single JSON or String column. `arrow_stream` batches events with Apache Arrow IPC streaming and requires `batch_encoding`.",
					Validators: []validator.String{
						stringvalidator.OneOf("json_each_row", "json_as_object", "json_as_string", "arrow_stream"),
					},
				},
				"skip_unknown_fields": schema.BoolAttribute{
					Optional:    true,
					Description: "If `true`, fields not present in the target table schema are dropped instead of causing insert errors. When unset, the ClickHouse server's own `input_format_skip_unknown_fields` setting applies.",
				},
				"date_time_best_effort": schema.BoolAttribute{
					Optional:    true,
					Description: "If `true`, enables flexible DateTime parsing on the server side.",
				},
			},
			Blocks: map[string]schema.Block{
				"compression": schema.ListNestedBlock{
					Description: "Compression for outbound HTTP requests. Use `algorithm = \"gzip\"` or `algorithm = \"none\"`.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"algorithm": schema.StringAttribute{
								Required:    true,
								Description: "Compression algorithm. Valid values are `gzip` and `none`.",
								Validators: []validator.String{
									stringvalidator.OneOf("gzip", "none"),
								},
							},
							"level": schema.Int64Attribute{
								Optional:    true,
								Description: "Compression level (1–9). Only valid when `algorithm` is `gzip`.",
								Validators: []validator.Int64{
									int64validator.Between(1, 9),
								},
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"auth": schema.ListNestedBlock{
					Description: "Authentication strategy for ClickHouse HTTP requests. Only `basic` strategy is supported.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"strategy": schema.StringAttribute{
								Required:    true,
								Description: "Authentication strategy. Must be `basic`.",
								Validators: []validator.String{
									stringvalidator.OneOf("basic"),
								},
							},
							"username_key": schema.StringAttribute{
								Optional:    true,
								Description: "Name of the environment variable or secret that holds the ClickHouse username. Defaults to `DESTINATION_CLICKHOUSE_USERNAME`.",
							},
							"password_key": schema.StringAttribute{
								Optional:    true,
								Description: "Name of the environment variable or secret that holds the ClickHouse password. Defaults to `DESTINATION_CLICKHOUSE_PASSWORD`.",
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"batch": schema.ListNestedBlock{
					Description: "Batching configuration for ClickHouse inserts.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"max_events": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum number of events per batch.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"timeout_secs": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum time in seconds before a partial batch is flushed.",
								Validators: []validator.Int64{
									int64validator.Between(1, 65535),
								},
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"batch_encoding": schema.ListNestedBlock{
					Description: "Batch encoding configuration. Required when `format` is `arrow_stream`.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"codec": schema.StringAttribute{
								Required:    true,
								Description: "Batch encoding codec. Must be `arrow_stream`.",
								Validators: []validator.String{
									stringvalidator.OneOf("arrow_stream"),
								},
							},
							"allow_nullable_fields": schema.BoolAttribute{
								Optional:    true,
								Description: "If `true`, allows null values for non-nullable fields in the ClickHouse schema. Defaults to `false`.",
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"tls":    TlsSchema(),
				"buffer": BufferOptionsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
