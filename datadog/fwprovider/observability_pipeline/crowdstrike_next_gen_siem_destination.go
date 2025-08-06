package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CrowdStrikeNextGenSiemDestinationModel represents the CrowdStrike NextGen SIEM destination configuration
type CrowdStrikeNextGenSiemDestinationModel struct {
	Id          types.String                   `tfsdk:"id"`
	Inputs      types.List                     `tfsdk:"inputs"`
	Encoding    types.String                   `tfsdk:"encoding"`
	Compression []*CrowdStrikeCompressionModel `tfsdk:"compression"`
	Tls         *tlsModel                      `tfsdk:"tls"`
}

// CrowdStrikeCompressionModel represents the compression configuration
type CrowdStrikeCompressionModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
	Level     types.Int64  `tfsdk:"level"`
}

// ExpandCrowdStrikeNextGenSiemDestination converts the Terraform model to the Datadog API model
func ExpandCrowdStrikeNextGenSiemDestination(ctx context.Context, src *CrowdStrikeNextGenSiemDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineCrowdStrikeNextGenSiemDestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	// Set encoding
	encoding := datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationEncoding(src.Encoding.ValueString())
	dest.SetEncoding(encoding)

	// Set compression if provided
	if len(src.Compression) > 0 {
		compression := datadogV2.NewObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompressionWithDefaults()
		algorithm := datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompressionAlgorithm(src.Compression[0].Algorithm.ValueString())
		compression.SetAlgorithm(algorithm)
		if !src.Compression[0].Level.IsNull() && !src.Compression[0].Level.IsUnknown() {
			compression.SetLevel(src.Compression[0].Level.ValueInt64())
		}
		dest.SetCompression(*compression)
	}

	// Set TLS if provided
	if src.Tls != nil {
		dest.SetTls(*ExpandTls(src.Tls))
	}

	return datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationAsObservabilityPipelineConfigDestinationItem(dest)
}

// FlattenCrowdStrikeNextGenSiemDestination converts the Datadog API model to the Terraform model
func FlattenCrowdStrikeNextGenSiemDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestination) *CrowdStrikeNextGenSiemDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())

	model := &CrowdStrikeNextGenSiemDestinationModel{
		Id:       types.StringValue(src.GetId()),
		Inputs:   inputs,
		Encoding: types.StringValue(string(src.GetEncoding())),
	}

	// Set compression if present
	if src.HasCompression() {
		compression := src.GetCompression()
		compressionModel := &CrowdStrikeCompressionModel{
			Algorithm: types.StringValue(string(compression.GetAlgorithm())),
		}
		if compression.HasLevel() {
			compressionModel.Level = types.Int64Value(compression.GetLevel())
		}
		model.Compression = []*CrowdStrikeCompressionModel{compressionModel}
	}

	// Set TLS if present
	if src.HasTls() {
		tls := src.GetTls()
		model.Tls = &tlsModel{
			CrtFile: types.StringValue(tls.GetCrtFile()),
		}
		if caFile, ok := tls.GetCaFileOk(); ok {
			model.Tls.CaFile = types.StringPointerValue(caFile)
		}
		if keyFile, ok := tls.GetKeyFileOk(); ok {
			model.Tls.KeyFile = types.StringPointerValue(keyFile)
		}
	}

	return model
}

// CrowdStrikeNextGenSiemDestinationSchema returns the schema for CrowdStrike NextGen SIEM destination
func CrowdStrikeNextGenSiemDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `crowdstrike_next_gen_siem` destination forwards logs to CrowdStrike NextGen SIEM.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:    true,
					Description: "The unique identifier for this component.",
				},
				"inputs": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
					Description: "List of input component IDs to receive logs from.",
				},
				"encoding": schema.StringAttribute{
					Required:    true,
					Description: "Encoding format for log events.",
				},
			},
			Blocks: map[string]schema.Block{
				"compression": schema.ListNestedBlock{
					Description: "Compression configuration for log events.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"algorithm": schema.StringAttribute{
								Required:    true,
								Description: "Compression algorithm for log events.",
							},
							"level": schema.Int64Attribute{
								Optional:    true,
								Description: "Compression level.",
							},
						},
					},
				},
				"tls": TlsSchema(),
			},
		},
	}
}
