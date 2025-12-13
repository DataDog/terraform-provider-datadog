package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SocketSourceModel represents the Terraform model for socket source configuration
type SocketSourceModel struct {
	Mode    types.String       `tfsdk:"mode"`
	Framing SocketFramingModel `tfsdk:"framing"`
	Tls     *tlsModel          `tfsdk:"tls"`
}

// ExpandSocketSource converts the Terraform model to the Datadog API model
func ExpandSocketSource(src *SocketSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineSocketSourceWithDefaults()
	s.SetId(id)
	s.SetMode(datadogV2.ObservabilityPipelineSocketSourceMode(src.Mode.ValueString()))

	switch src.Framing.Method.ValueString() {
	case "newline_delimited":
		s.Framing = datadogV2.ObservabilityPipelineSocketSourceFraming{
			ObservabilityPipelineSocketSourceFramingNewlineDelimited: &datadogV2.ObservabilityPipelineSocketSourceFramingNewlineDelimited{
				Method: "newline_delimited",
			},
		}
	case "bytes":
		s.Framing = datadogV2.ObservabilityPipelineSocketSourceFraming{
			ObservabilityPipelineSocketSourceFramingBytes: &datadogV2.ObservabilityPipelineSocketSourceFramingBytes{
				Method: "bytes",
			},
		}
	case "character_delimited":
		s.Framing = datadogV2.ObservabilityPipelineSocketSourceFraming{
			ObservabilityPipelineSocketSourceFramingCharacterDelimited: &datadogV2.ObservabilityPipelineSocketSourceFramingCharacterDelimited{
				Method:    "character_delimited",
				Delimiter: src.Framing.CharacterDelimited.Delimiter.ValueString(),
			},
		}
	case "octet_counting":
		s.Framing = datadogV2.ObservabilityPipelineSocketSourceFraming{
			ObservabilityPipelineSocketSourceFramingOctetCounting: &datadogV2.ObservabilityPipelineSocketSourceFramingOctetCounting{
				Method: "octet_counting",
			},
		}
	case "chunked_gelf":
		s.Framing = datadogV2.ObservabilityPipelineSocketSourceFraming{
			ObservabilityPipelineSocketSourceFramingChunkedGelf: &datadogV2.ObservabilityPipelineSocketSourceFramingChunkedGelf{
				Method: "chunked_gelf",
			},
		}
	}

	if src.Tls != nil {
		s.Tls = ExpandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSocketSource: s,
	}
}

// FlattenSocketSource converts the Datadog API model to the Terraform model
func FlattenSocketSource(src *datadogV2.ObservabilityPipelineSocketSource) *SocketSourceModel {
	if src == nil {
		return nil
	}

	out := &SocketSourceModel{
		Mode: types.StringValue(string(src.GetMode())),
	}

	if src.Tls != nil {
		tls := FlattenTls(src.Tls)
		out.Tls = &tls
	}

	switch {
	case src.Framing.ObservabilityPipelineSocketSourceFramingNewlineDelimited != nil:
		out.Framing.Method = types.StringValue("newline_delimited")
	case src.Framing.ObservabilityPipelineSocketSourceFramingBytes != nil:
		out.Framing.Method = types.StringValue("bytes")
	case src.Framing.ObservabilityPipelineSocketSourceFramingCharacterDelimited != nil:
		out.Framing.Method = types.StringValue("character_delimited")
		out.Framing.CharacterDelimited = &SocketFramingCharacterDelimitedModel{
			Delimiter: types.StringValue(src.Framing.ObservabilityPipelineSocketSourceFramingCharacterDelimited.Delimiter),
		}
	case src.Framing.ObservabilityPipelineSocketSourceFramingOctetCounting != nil:
		out.Framing.Method = types.StringValue("octet_counting")
	case src.Framing.ObservabilityPipelineSocketSourceFramingChunkedGelf != nil:
		out.Framing.Method = types.StringValue("chunked_gelf")
	}

	return out
}

// SocketSourceSchema returns the schema for socket source
func SocketSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `socket` source ingests logs over TCP or UDP.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"mode": schema.StringAttribute{
					Required:    true,
					Description: "The protocol used to receive logs.",
					Validators: []validator.String{
						stringvalidator.OneOf("tcp", "udp"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"framing": schema.SingleNestedBlock{
					Description: "Defines the framing method for incoming messages.",
					Attributes: map[string]schema.Attribute{
						"method": schema.StringAttribute{
							Optional:    true, // must be optional to make the block optional
							Description: "The framing method.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"newline_delimited",
									"bytes",
									"character_delimited",
									"octet_counting",
									"chunked_gelf",
								),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"character_delimited": schema.SingleNestedBlock{
							Description: "Used when `method` is `character_delimited`. Specifies the delimiter character.",
							Attributes: map[string]schema.Attribute{
								"delimiter": schema.StringAttribute{
									Optional:    true, // must be optional to make the block optional
									Description: "A single ASCII character used as a delimiter.",
								},
							},
						},
					},
				},
				"tls": TlsSchema(),
			},
		},
	}
}
