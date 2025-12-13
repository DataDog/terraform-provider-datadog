package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CrowdStrikeNextGenSiemDestinationModel represents the Terraform model for the CrowdStrikeNextGenSiemDestination
type CrowdStrikeNextGenSiemDestinationModel struct {
	Encoding    types.String      `tfsdk:"encoding"`
	Compression *compressionModel `tfsdk:"compression"`
	Tls         *tlsModel         `tfsdk:"tls"`
}

// CrowdStrikeNextGenSiemDestinationSchema returns the schema for the CrowdStrikeNextGenSiemDestination
func CrowdStrikeNextGenSiemDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `crowdstrike_next_gen_siem` destination forwards logs to CrowdStrike Next Gen SIEM.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"encoding": schema.StringAttribute{
					Required:    true,
					Description: "Encoding format for log events.",
					Validators: []validator.String{
						stringvalidator.OneOf("json", "raw_message"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"compression": CompressionSchema(),
				"tls":         TlsSchema(),
			},
		},
	}
}

// ExpandCrowdStrikeNextGenSiemDestination converts the Terraform model to the API model
func ExpandCrowdStrikeNextGenSiemDestination(ctx context.Context, id string, inputs types.List, src *CrowdStrikeNextGenSiemDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineCrowdStrikeNextGenSiemDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	dest.SetEncoding(datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationEncoding(src.Encoding.ValueString()))

	// Handle compression configuration
	if src.Compression != nil {
		compression := ExpandCompression(src.Compression)
		if compression != nil {
			dest.SetCompression(*compression)
		}
	}

	// Handle TLS configuration
	if src.Tls != nil {
		tls := ExpandTls(src.Tls)
		if tls != nil {
			dest.SetTls(*tls)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineCrowdStrikeNextGenSiemDestination: dest,
	}
}

// FlattenCrowdStrikeNextGenSiemDestination converts the API model to the Terraform model
func FlattenCrowdStrikeNextGenSiemDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestination) *CrowdStrikeNextGenSiemDestinationModel {
	if src == nil {
		return nil
	}

	out := &CrowdStrikeNextGenSiemDestinationModel{
		Encoding: types.StringValue(string(src.GetEncoding())),
	}

	// Handle compression configuration
	if src.Compression != nil {
		compression := FlattenCompression(src.Compression)
		out.Compression = &compression
	}

	// Handle TLS configuration
	if src.Tls != nil {
		tls := FlattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}
