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
	EndpointUrlKey types.String         `tfsdk:"endpoint_url_key"`
	TokenKey       types.String         `tfsdk:"token_key"`
	Encoding       types.String         `tfsdk:"encoding"`
	Compression    []compressionModel   `tfsdk:"compression"`
	Tls            []TlsModel           `tfsdk:"tls"`
	Buffer         []BufferOptionsModel `tfsdk:"buffer"`
}

// CrowdStrikeNextGenSiemDestinationSchema returns the schema for the CrowdStrikeNextGenSiemDestination
func CrowdStrikeNextGenSiemDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `crowdstrike_next_gen_siem` destination forwards logs to CrowdStrike Next Gen SIEM.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"endpoint_url_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the endpoint URL.",
				},
				"token_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the authentication token.",
				},
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
				"buffer":      BufferOptionsSchema(),
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
	if !src.EndpointUrlKey.IsNull() {
		dest.SetEndpointUrlKey(src.EndpointUrlKey.ValueString())
	}
	if !src.TokenKey.IsNull() {
		dest.SetTokenKey(src.TokenKey.ValueString())
	}

	// Handle compression configuration
	if len(src.Compression) > 0 {
		dest.Compression = ExpandCompression(src.Compression)
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
	if v, ok := src.GetEndpointUrlKeyOk(); ok {
		out.EndpointUrlKey = types.StringValue(*v)
	}
	if v, ok := src.GetTokenKeyOk(); ok {
		out.TokenKey = types.StringValue(*v)
	}

	if src.Tls != nil {
		out.Tls = FlattenTls(src.Tls)
	}

	// Handle compression configuration
	if src.Compression != nil {
		compression := FlattenCompression(src.Compression)
		out.Compression = []compressionModel{compression}
	}

	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			out.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	return out
}
