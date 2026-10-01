package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MetricEnrichmentTableProcessorModel struct {
	File           []MetricEnrichmentTableFileModel           `tfsdk:"file"`
	ReferenceTable []MetricEnrichmentTableReferenceTableModel `tfsdk:"reference_table"`
}

type MetricEnrichmentTableFileModel struct {
	Path     types.String                             `tfsdk:"path"`
	Encoding []MetricEnrichmentTableFileEncodingModel `tfsdk:"encoding"`
	Key      []MetricEnrichmentTableFileKeyModel      `tfsdk:"key"`
}

type MetricEnrichmentTableFileEncodingModel struct {
	Type            types.String `tfsdk:"type"`
	Delimiter       types.String `tfsdk:"delimiter"`
	IncludesHeaders types.Bool   `tfsdk:"includes_headers"`
}

type MetricEnrichmentTableFileKeyModel struct {
	Column types.String                             `tfsdk:"column"`
	Source []MetricEnrichmentTableLookupSourceModel `tfsdk:"source"`
}

type MetricEnrichmentTableReferenceTableModel struct {
	TableId   types.String                             `tfsdk:"table_id"`
	AppKeyKey types.String                             `tfsdk:"app_key_key"`
	Columns   []types.String                           `tfsdk:"columns"`
	Key       []MetricEnrichmentTableReferenceKeyModel `tfsdk:"key"`
}

type MetricEnrichmentTableReferenceKeyModel struct {
	Source []MetricEnrichmentTableLookupSourceModel `tfsdk:"source"`
}

type MetricEnrichmentTableLookupSourceModel struct {
	Type types.String `tfsdk:"type"`
	Name types.String `tfsdk:"name"`
}

const (
	metricEnrichmentTableLookupSourceMetricName = "metric_name"
	metricEnrichmentTableLookupSourceTag        = "tag"
)

func MetricEnrichmentTableProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Configures the `enrichment_table` processor for `metrics` pipelines. The processor enriches metrics with tags from a static CSV file or a Datadog reference table. It looks up a row using the metric name or a metric tag value. It then adds each column of the matching row as a metric tag, overwriting any existing tag with the same key. Exactly one of `file` or `reference_table` must be configured.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"file": schema.ListNestedBlock{
					Description: "Defines a static enrichment table loaded from a CSV file for metric enrichment.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.ExactlyOneOf(frameworkPath.MatchRelative().AtParent().AtName("reference_table")),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"path": schema.StringAttribute{
								Required:    true,
								Description: "Path to the CSV file.",
							},
						},
						Blocks: map[string]schema.Block{
							"encoding": schema.ListNestedBlock{
								Description: "File encoding format.",
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Required:    true,
											Description: "The encoding format of the file. The value should always be `csv`.",
										},
										"delimiter": schema.StringAttribute{
											Required:    true,
											Description: "The single character that separates columns in the file.",
										},
										"includes_headers": schema.BoolAttribute{
											Required:    true,
											Description: "Whether the first row of the file contains column headers.",
										},
									},
								},
							},
							"key": schema.ListNestedBlock{
								Description: "Defines how to map a metric lookup value to a CSV column during enrichment table lookups.",
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"column": schema.StringAttribute{
											Required:    true,
											Description: "The CSV column name or index to match against the lookup value.",
										},
									},
									Blocks: map[string]schema.Block{
										"source": metricEnrichmentTableLookupSourceSchema(),
									},
								},
							},
						},
					},
				},
				"reference_table": schema.ListNestedBlock{
					Description: "Uses a Datadog reference table to enrich metrics.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"table_id": schema.StringAttribute{
								Required:    true,
								Description: "The unique identifier of the reference table.",
							},
							"app_key_key": schema.StringAttribute{
								Optional:    true,
								Description: "The name of the environment variable or secret that holds the Datadog application key used to access the reference table.",
							},
							"columns": schema.ListAttribute{
								Optional:    true,
								ElementType: types.StringType,
								Description: "A list of column names to include from the reference table. If not provided, all columns are included.",
							},
						},
						Blocks: map[string]schema.Block{
							"key": schema.ListNestedBlock{
								Description: "Defines the metric lookup value used as the reference-table row ID.",
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"source": metricEnrichmentTableLookupSourceSchema(),
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

func metricEnrichmentTableLookupSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Specifies the source of the key value used for metric enrichment table lookups. The lookup key can be either the metric name or a metric tag.",
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Required:    true,
					Description: "The lookup source type. Use `metric_name` to look up the metric name, or `tag` to look up the value of the metric tag set in `name`.",
					Validators: []validator.String{
						stringvalidator.OneOf(metricEnrichmentTableLookupSourceMetricName, metricEnrichmentTableLookupSourceTag),
					},
				},
				"name": schema.StringAttribute{
					Optional:    true,
					Description: "The Datadog tag key used as the lookup key. Required when `type` is `tag`.",
				},
			},
		},
	}
}

func ExpandMetricEnrichmentTableProcessor(common BaseProcessorFields, src *MetricEnrichmentTableProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	var proc datadogV2.ObservabilityPipelineMetricEnrichmentTableProcessor

	if len(src.File) > 0 {
		fileProc := datadogV2.NewObservabilityPipelineMetricEnrichmentTableFileProcessorWithDefaults()
		common.ApplyTo(fileProc)
		fileProc.SetFile(expandMetricEnrichmentTableFile(src.File[0]))
		proc = datadogV2.ObservabilityPipelineMetricEnrichmentTableFileProcessorAsObservabilityPipelineMetricEnrichmentTableProcessor(fileProc)
	} else if len(src.ReferenceTable) > 0 {
		refProc := datadogV2.NewObservabilityPipelineMetricEnrichmentTableReferenceTableProcessorWithDefaults()
		common.ApplyTo(refProc)
		refProc.SetReferenceTable(expandMetricEnrichmentTableReferenceTable(src.ReferenceTable[0]))
		proc = datadogV2.ObservabilityPipelineMetricEnrichmentTableReferenceTableProcessorAsObservabilityPipelineMetricEnrichmentTableProcessor(refProc)
	}

	return datadogV2.ObservabilityPipelineMetricEnrichmentTableProcessorAsObservabilityPipelineConfigProcessorItem(&proc)
}

func expandMetricEnrichmentTableFile(src MetricEnrichmentTableFileModel) datadogV2.ObservabilityPipelineMetricEnrichmentTableFile {
	encoding := src.Encoding[0]
	key := src.Key[0]
	return datadogV2.ObservabilityPipelineMetricEnrichmentTableFile{
		Path: src.Path.ValueString(),
		Encoding: datadogV2.ObservabilityPipelineEnrichmentTableFileEncoding{
			Type:            datadogV2.ObservabilityPipelineEnrichmentTableFileEncodingType(encoding.Type.ValueString()),
			Delimiter:       encoding.Delimiter.ValueString(),
			IncludesHeaders: encoding.IncludesHeaders.ValueBool(),
		},
		Key: datadogV2.ObservabilityPipelineMetricEnrichmentTableFileKey{
			Column: key.Column.ValueString(),
			Source: expandMetricEnrichmentTableLookupSource(key.Source[0]),
		},
	}
}

func expandMetricEnrichmentTableReferenceTable(src MetricEnrichmentTableReferenceTableModel) datadogV2.ObservabilityPipelineMetricEnrichmentTableReferenceTable {
	refTable := datadogV2.ObservabilityPipelineMetricEnrichmentTableReferenceTable{
		TableId: src.TableId.ValueString(),
		Key: datadogV2.ObservabilityPipelineMetricEnrichmentTableReferenceKey{
			Source: expandMetricEnrichmentTableLookupSource(src.Key[0].Source[0]),
		},
	}
	if !src.AppKeyKey.IsNull() {
		refTable.SetAppKeyKey(src.AppKeyKey.ValueString())
	}
	if src.Columns != nil {
		columns := make([]string, 0, len(src.Columns))
		for _, c := range src.Columns {
			columns = append(columns, c.ValueString())
		}
		refTable.SetColumns(columns)
	}
	return refTable
}

func expandMetricEnrichmentTableLookupSource(src MetricEnrichmentTableLookupSourceModel) datadogV2.ObservabilityPipelineMetricEnrichmentTableLookupSource {
	if src.Type.ValueString() == metricEnrichmentTableLookupSourceTag {
		return datadogV2.ObservabilityPipelineMetricEnrichmentTableTagLookupAsObservabilityPipelineMetricEnrichmentTableLookupSource(
			datadogV2.NewObservabilityPipelineMetricEnrichmentTableTagLookup(src.Name.ValueString(), datadogV2.OBSERVABILITYPIPELINEMETRICENRICHMENTTABLETAGLOOKUPTYPE_TAG),
		)
	}
	return datadogV2.ObservabilityPipelineMetricEnrichmentTableMetricNameLookupAsObservabilityPipelineMetricEnrichmentTableLookupSource(
		datadogV2.NewObservabilityPipelineMetricEnrichmentTableMetricNameLookupWithDefaults(),
	)
}

// FlattenMetricEnrichmentTableProcessor returns the populated variant (for the common processor
// fields) and the Terraform model. It returns nil values if no variant is set.
func FlattenMetricEnrichmentTableProcessor(src *datadogV2.ObservabilityPipelineMetricEnrichmentTableProcessor) (BaseProcessor, *MetricEnrichmentTableProcessorModel) {
	if src == nil {
		return nil, nil
	}
	if p := src.ObservabilityPipelineMetricEnrichmentTableFileProcessor; p != nil {
		return p, &MetricEnrichmentTableProcessorModel{
			File: []MetricEnrichmentTableFileModel{flattenMetricEnrichmentTableFile(p.GetFile())},
		}
	}
	if p := src.ObservabilityPipelineMetricEnrichmentTableReferenceTableProcessor; p != nil {
		return p, &MetricEnrichmentTableProcessorModel{
			ReferenceTable: []MetricEnrichmentTableReferenceTableModel{flattenMetricEnrichmentTableReferenceTable(p.GetReferenceTable())},
		}
	}
	return nil, nil
}

func flattenMetricEnrichmentTableFile(src datadogV2.ObservabilityPipelineMetricEnrichmentTableFile) MetricEnrichmentTableFileModel {
	encoding := src.GetEncoding()
	key := src.GetKey()
	return MetricEnrichmentTableFileModel{
		Path: types.StringValue(src.GetPath()),
		Encoding: []MetricEnrichmentTableFileEncodingModel{{
			Type:            types.StringValue(string(encoding.GetType())),
			Delimiter:       types.StringValue(encoding.GetDelimiter()),
			IncludesHeaders: types.BoolValue(encoding.GetIncludesHeaders()),
		}},
		Key: []MetricEnrichmentTableFileKeyModel{{
			Column: types.StringValue(key.GetColumn()),
			Source: flattenMetricEnrichmentTableLookupSource(key.GetSource()),
		}},
	}
}

func flattenMetricEnrichmentTableReferenceTable(src datadogV2.ObservabilityPipelineMetricEnrichmentTableReferenceTable) MetricEnrichmentTableReferenceTableModel {
	key := src.GetKey()
	out := MetricEnrichmentTableReferenceTableModel{
		TableId:   types.StringValue(src.GetTableId()),
		AppKeyKey: types.StringPointerValue(src.AppKeyKey),
		Key: []MetricEnrichmentTableReferenceKeyModel{{
			Source: flattenMetricEnrichmentTableLookupSource(key.GetSource()),
		}},
	}
	if columns, ok := src.GetColumnsOk(); ok {
		for _, c := range *columns {
			out.Columns = append(out.Columns, types.StringValue(c))
		}
	}
	return out
}

func flattenMetricEnrichmentTableLookupSource(src datadogV2.ObservabilityPipelineMetricEnrichmentTableLookupSource) []MetricEnrichmentTableLookupSourceModel {
	if tag := src.ObservabilityPipelineMetricEnrichmentTableTagLookup; tag != nil {
		return []MetricEnrichmentTableLookupSourceModel{{
			Type: types.StringValue(string(tag.GetType())),
			Name: types.StringValue(tag.GetName()),
		}}
	}
	if metricName := src.ObservabilityPipelineMetricEnrichmentTableMetricNameLookup; metricName != nil {
		return []MetricEnrichmentTableLookupSourceModel{{
			Type: types.StringValue(string(metricName.GetType())),
			Name: types.StringNull(),
		}}
	}
	return nil
}
