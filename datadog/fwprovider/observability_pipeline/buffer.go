package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BufferOptionsModel struct {
	DiskBuffer   []DiskBufferOptionsModel   `tfsdk:"disk"`
	MemoryBuffer []MemoryBufferOptionsModel `tfsdk:"memory"`
}

type DiskBufferOptionsModel struct {
	MaxSize types.Int64 `tfsdk:"max_size"`
}

type MemoryBufferOptionsModel struct {
	MaxSize   types.Int64 `tfsdk:"max_size"`
	MaxEvents types.Int64 `tfsdk:"max_events"`
}

func ExpandBufferOptions(src BufferOptionsModel) *datadogV2.ObservabilityPipelineBufferOptions {
	if len(src.DiskBuffer) > 0 {
		diskBuf := src.DiskBuffer[0]
		buffer := datadogV2.NewObservabilityPipelineDiskBufferOptionsWithDefaults()

		if !diskBuf.MaxSize.IsNull() {
			buffer.SetMaxSize(diskBuf.MaxSize.ValueInt64())
		}
		buffer.SetType(datadogV2.ObservabilityPipelineBufferOptionsDiskType("disk"))

		return &datadogV2.ObservabilityPipelineBufferOptions{
			ObservabilityPipelineDiskBufferOptions: buffer,
		}
	}

	if len(src.MemoryBuffer) > 0 {
		memBuf := src.MemoryBuffer[0]

		if !memBuf.MaxEvents.IsNull() {
			buffer := datadogV2.NewObservabilityPipelineMemoryBufferSizeOptionsWithDefaults()
			buffer.SetType(datadogV2.ObservabilityPipelineBufferOptionsMemoryType("memory"))
			buffer.SetMaxEvents(memBuf.MaxEvents.ValueInt64())

			return &datadogV2.ObservabilityPipelineBufferOptions{
				ObservabilityPipelineMemoryBufferSizeOptions: buffer,
			}
		} else if !memBuf.MaxSize.IsNull() {
			buffer := datadogV2.NewObservabilityPipelineMemoryBufferOptionsWithDefaults()
			buffer.SetType(datadogV2.ObservabilityPipelineBufferOptionsMemoryType("memory"))
			buffer.SetMaxSize(memBuf.MaxSize.ValueInt64())

			return &datadogV2.ObservabilityPipelineBufferOptions{
				ObservabilityPipelineMemoryBufferOptions: buffer,
			}
		}
	}

	return nil
}

func FlattenBufferOptions(src *datadogV2.ObservabilityPipelineBufferOptions) *BufferOptionsModel {
	if src == nil {
		return nil
	}

	if diskBuf := src.ObservabilityPipelineDiskBufferOptions; diskBuf != nil {
		return &BufferOptionsModel{
			DiskBuffer: []DiskBufferOptionsModel{
				{
					MaxSize: types.Int64Value(diskBuf.GetMaxSize()),
				},
			},
		}
	}

	if memBufSize := src.ObservabilityPipelineMemoryBufferSizeOptions; memBufSize != nil {
		return &BufferOptionsModel{
			MemoryBuffer: []MemoryBufferOptionsModel{
				{
					MaxEvents: types.Int64Value(memBufSize.GetMaxEvents()),
					MaxSize:   types.Int64Null(),
				},
			},
		}
	}

	if memBuf := src.ObservabilityPipelineMemoryBufferOptions; memBuf != nil {
		return &BufferOptionsModel{
			MemoryBuffer: []MemoryBufferOptionsModel{
				{
					MaxSize:   types.Int64Value(memBuf.GetMaxSize()),
					MaxEvents: types.Int64Null(),
				},
			},
		}
	}

	return nil
}

func BufferOptionsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Configuration for buffer settings on destination components. Exactly one of `disk` or `memory` must be specified.",
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"disk": schema.ListNestedBlock{
					Description: "Options for configuring a disk buffer. Cannot be used with `memory`.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"max_size": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum size of the disk buffer (in bytes).",
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.ExactlyOneOf(frameworkPath.MatchRelative().AtParent().AtName("memory")),
					},
				},
				"memory": schema.ListNestedBlock{
					Description: "Options for configuring a memory buffer. Cannot be used with `disk`.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"max_size": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum size of the memory buffer (in bytes).",
								Validators: []validator.Int64{
									int64validator.ExactlyOneOf(frameworkPath.MatchRelative().AtParent().AtName("max_events")),
								},
							},
							"max_events": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum events for the memory buffer.",
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
	}
}
