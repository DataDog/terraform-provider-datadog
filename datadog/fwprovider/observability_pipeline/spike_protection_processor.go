package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SpikeProtectionProcessorModel struct {
	Name                        types.String                                     `tfsdk:"name"`
	SpikeAction                 types.String                                     `tfsdk:"spike_action"`
	SpikeThreshold              types.Float64                                    `tfsdk:"spike_threshold"`
	MinimalThreshold            types.Int64                                      `tfsdk:"minimal_threshold"`
	PartitionFields             []types.String                                   `tfsdk:"partition_fields"`
	ActionWhenMissingPartitions types.String                                     `tfsdk:"action_when_missing_partitions"`
	MinimalThresholdOverrides   []SpikeProtectionProcessorThresholdOverrideModel `tfsdk:"minimal_threshold_override"`
}

type SpikeProtectionProcessorThresholdOverrideModel struct {
	Fields           []SpikeProtectionProcessorFieldValueModel       `tfsdk:"field"`
	MinimalThreshold []SpikeProtectionProcessorMinimalThresholdModel `tfsdk:"minimal_threshold"`
}

type SpikeProtectionProcessorFieldValueModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type SpikeProtectionProcessorMinimalThresholdModel struct {
	Type  types.String `tfsdk:"type"`
	Bytes types.Int64  `tfsdk:"bytes"`
}

func SpikeProtectionProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `spike_protection` processor learns the normal log volume and detects spikes, optionally dropping or rerouting the excess events.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "Name of the spike protection rule.",
				},
				"spike_action": schema.StringAttribute{
					Required:    true,
					Description: "The action to take when a spike is detected. One of `drop`, `no_action`, `overflow_routing`. When `overflow_routing` is used, there must be a destination whose `inputs` reference this processor with the `<processor-id>.overflow_events` suffix to route the overflowing events.",
					Validators: []validator.String{
						stringvalidator.OneOf("drop", "no_action", "overflow_routing"),
					},
				},
				"spike_threshold": schema.Float64Attribute{
					Optional:    true,
					Description: "The number of standard deviations above the learned mean volume that must be exceeded before a spike is detected. Defaults to `5`.",
				},
				"minimal_threshold": schema.Int64Attribute{
					Optional:    true,
					Description: "The minimum volume, in bytes, that must be observed within 24 hours before spike protection can activate. Defaults to 100 GiB.",
				},
				"partition_fields": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
					Description: "A list of fields used to segment log traffic. Volume is learned and evaluated independently for each unique combination of these field values. A maximum of five fields is supported.",
				},
				"action_when_missing_partitions": schema.StringAttribute{
					Optional:    true,
					Description: "The action to take when an event is missing one of the `partition_fields`. One of `ignore`, `default_bucket`.",
					Validators: []validator.String{
						stringvalidator.OneOf("ignore", "default_bucket"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"minimal_threshold_override": schema.ListNestedBlock{
					Description: "A list of per-partition overrides for the minimal threshold, identified by matching field values. Requires `partition_fields` to be set.",
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"field": schema.ListNestedBlock{
								Description: "The set of field matchers that identifies the partition this override applies to. Every field name listed must also appear in `partition_fields`.",
								Validators: []validator.List{
									listvalidator.IsRequired(),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required:    true,
											Description: "The field name.",
										},
										"value": schema.StringAttribute{
											Required:    true,
											Description: "The field value.",
										},
									},
								},
							},
							"minimal_threshold": schema.ListNestedBlock{
								Description: "The minimal threshold to apply to the matching partition.",
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Required:    true,
											Description: "The minimal threshold type. One of `threshold`, `always_keep`. `always_keep` exempts the matching partition from spike protection.",
											Validators: []validator.String{
												stringvalidator.OneOf("threshold", "always_keep"),
											},
										},
										"bytes": schema.Int64Attribute{
											Optional:    true,
											Description: "The minimum volume, in bytes, that must be observed within 24 hours before spike protection can activate for the matching partition. Used only when `type` is `threshold`.",
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

func ExpandSpikeProtectionProcessor(common BaseProcessorFields, src *SpikeProtectionProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineSpikeProtectionProcessorWithDefaults()
	common.ApplyTo(proc)

	proc.SetName(src.Name.ValueString())
	proc.SetSpikeAction(datadogV2.ObservabilityPipelineSpikeProtectionProcessorSpikeAction(src.SpikeAction.ValueString()))
	if !src.SpikeThreshold.IsNull() && !src.SpikeThreshold.IsUnknown() {
		proc.SetSpikeThreshold(src.SpikeThreshold.ValueFloat64())
	}
	if !src.MinimalThreshold.IsNull() && !src.MinimalThreshold.IsUnknown() {
		proc.SetMinimalThreshold(src.MinimalThreshold.ValueInt64())
	}
	if len(src.PartitionFields) > 0 {
		partitionFields := make([]string, 0, len(src.PartitionFields))
		for _, f := range src.PartitionFields {
			partitionFields = append(partitionFields, f.ValueString())
		}
		proc.SetPartitionFields(partitionFields)
	}
	if !src.ActionWhenMissingPartitions.IsNull() && !src.ActionWhenMissingPartitions.IsUnknown() {
		proc.SetActionWhenMissingPartitions(datadogV2.ObservabilityPipelineSpikeProtectionProcessorActionWhenMissingPartitions(src.ActionWhenMissingPartitions.ValueString()))
	}

	if len(src.MinimalThresholdOverrides) > 0 {
		overrides := make([]datadogV2.ObservabilityPipelineSpikeProtectionProcessorThresholdOverride, 0, len(src.MinimalThresholdOverrides))
		for _, o := range src.MinimalThresholdOverrides {
			fields := make([]datadogV2.ObservabilityPipelineFieldValue, 0, len(o.Fields))
			for _, f := range o.Fields {
				fields = append(fields, datadogV2.ObservabilityPipelineFieldValue{
					Name:  f.Name.ValueString(),
					Value: f.Value.ValueString(),
				})
			}

			threshold := datadogV2.ObservabilityPipelineSpikeProtectionProcessorMinimalThreshold{}
			if len(o.MinimalThreshold) > 0 {
				mt := o.MinimalThreshold[0]
				switch mt.Type.ValueString() {
				case "threshold":
					threshold.ObservabilityPipelineSpikeProtectionProcessorMinimalThresholdBytes = datadogV2.NewObservabilityPipelineSpikeProtectionProcessorMinimalThresholdBytes(
						mt.Bytes.ValueInt64(),
						datadogV2.OBSERVABILITYPIPELINESPIKEPROTECTIONPROCESSORMINIMALTHRESHOLDBYTESTYPE_THRESHOLD,
					)
				case "always_keep":
					threshold.ObservabilityPipelineSpikeProtectionProcessorMinimalThresholdAlwaysKeep = datadogV2.NewObservabilityPipelineSpikeProtectionProcessorMinimalThresholdAlwaysKeep(
						datadogV2.OBSERVABILITYPIPELINESPIKEPROTECTIONPROCESSORMINIMALTHRESHOLDALWAYSKEEPTYPE_ALWAYS_KEEP,
					)
				}
			}

			overrides = append(overrides, *datadogV2.NewObservabilityPipelineSpikeProtectionProcessorThresholdOverride(fields, threshold))
		}
		proc.SetMinimalThresholdOverrides(overrides)
	}

	return datadogV2.ObservabilityPipelineSpikeProtectionProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func FlattenSpikeProtectionProcessor(src *datadogV2.ObservabilityPipelineSpikeProtectionProcessor) *SpikeProtectionProcessorModel {
	if src == nil {
		return nil
	}
	model := &SpikeProtectionProcessorModel{
		Name:        types.StringValue(src.GetName()),
		SpikeAction: types.StringValue(string(src.GetSpikeAction())),
	}
	if v, ok := src.GetSpikeThresholdOk(); ok && v != nil {
		model.SpikeThreshold = types.Float64Value(*v)
	} else {
		model.SpikeThreshold = types.Float64Null()
	}
	if v, ok := src.GetMinimalThresholdOk(); ok && v != nil {
		model.MinimalThreshold = types.Int64Value(*v)
	} else {
		model.MinimalThreshold = types.Int64Null()
	}
	for _, f := range src.GetPartitionFields() {
		model.PartitionFields = append(model.PartitionFields, types.StringValue(f))
	}
	if v, ok := src.GetActionWhenMissingPartitionsOk(); ok && v != nil {
		model.ActionWhenMissingPartitions = types.StringValue(string(*v))
	} else {
		model.ActionWhenMissingPartitions = types.StringNull()
	}

	for _, o := range src.GetMinimalThresholdOverrides() {
		oModel := SpikeProtectionProcessorThresholdOverrideModel{}
		for _, f := range o.GetFields() {
			oModel.Fields = append(oModel.Fields, SpikeProtectionProcessorFieldValueModel{
				Name:  types.StringValue(f.GetName()),
				Value: types.StringValue(f.GetValue()),
			})
		}
		mt := o.GetMinimalThreshold()
		if b := mt.ObservabilityPipelineSpikeProtectionProcessorMinimalThresholdBytes; b != nil {
			oModel.MinimalThreshold = append(oModel.MinimalThreshold, SpikeProtectionProcessorMinimalThresholdModel{
				Type:  types.StringValue(string(b.GetType())),
				Bytes: types.Int64Value(b.GetBytes()),
			})
		} else if k := mt.ObservabilityPipelineSpikeProtectionProcessorMinimalThresholdAlwaysKeep; k != nil {
			oModel.MinimalThreshold = append(oModel.MinimalThreshold, SpikeProtectionProcessorMinimalThresholdModel{
				Type:  types.StringValue(string(k.GetType())),
				Bytes: types.Int64Null(),
			})
		}
		model.MinimalThresholdOverrides = append(model.MinimalThresholdOverrides, oModel)
	}
	return model
}
