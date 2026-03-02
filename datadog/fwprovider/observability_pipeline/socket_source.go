package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SocketSourceModel represents the Terraform model for socket source configuration
type SocketSourceModel struct {
	AddressKey types.String         `tfsdk:"address_key"`
	Mode       types.String         `tfsdk:"mode"`
	Framing    []SocketFramingModel `tfsdk:"framing"`
	Tls        []TlsModel           `tfsdk:"tls"`
}

// ExpandSocketSource converts the Terraform model to the Datadog API model
func ExpandSocketSource(src *SocketSourceModel, id string) (datadogV2.ObservabilityPipelineConfigSourceItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := datadogV2.NewObservabilityPipelineSocketSourceWithDefaults()
	s.SetId(id)
	if !src.AddressKey.IsNull() {
		s.SetAddressKey(src.AddressKey.ValueString())
	}
	s.SetMode(datadogV2.ObservabilityPipelineSocketSourceMode(src.Mode.ValueString()))

	switch src.Framing[0].Method.ValueString() {
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
		charDelimited := &datadogV2.ObservabilityPipelineSocketSourceFramingCharacterDelimited{
			Method: "character_delimited",
		}
		if len(src.Framing[0].CharacterDelimited) > 0 {
			charDelimited.Delimiter = src.Framing[0].CharacterDelimited[0].Delimiter.ValueString()
		}
		s.Framing = datadogV2.ObservabilityPipelineSocketSourceFraming{
			ObservabilityPipelineSocketSourceFramingCharacterDelimited: charDelimited,
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
	if len(src.Tls) > 0 {
		s.Tls = ExpandTls(src.Tls)
	}
	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineSocketSource: s,
	}, diags
}

// FlattenSocketSource converts the Datadog API model to the Terraform model
func FlattenSocketSource(src *datadogV2.ObservabilityPipelineSocketSource) *SocketSourceModel {
	if src == nil {
		return nil
	}

	out := &SocketSourceModel{
		Mode: types.StringValue(string(src.GetMode())),
	}
	if v, ok := src.GetAddressKeyOk(); ok {
		out.AddressKey = types.StringValue(*v)
	}

	if src.Tls != nil {
		out.Tls = FlattenTls(src.Tls)
	}

	outFraming := SocketFramingModel{}
	switch {
	case src.Framing.ObservabilityPipelineSocketSourceFramingNewlineDelimited != nil:
		outFraming.Method = types.StringValue("newline_delimited")
	case src.Framing.ObservabilityPipelineSocketSourceFramingBytes != nil:
		outFraming.Method = types.StringValue("bytes")
	case src.Framing.ObservabilityPipelineSocketSourceFramingCharacterDelimited != nil:
		outFraming.Method = types.StringValue("character_delimited")
		outFraming.CharacterDelimited = []SocketFramingCharacterDelimitedModel{SocketFramingCharacterDelimitedModel{
			Delimiter: types.StringValue(src.Framing.ObservabilityPipelineSocketSourceFramingCharacterDelimited.Delimiter),
		}}
	case src.Framing.ObservabilityPipelineSocketSourceFramingOctetCounting != nil:
		outFraming.Method = types.StringValue("octet_counting")
	case src.Framing.ObservabilityPipelineSocketSourceFramingChunkedGelf != nil:
		outFraming.Method = types.StringValue("chunked_gelf")
	}
	out.Framing = []SocketFramingModel{outFraming}

	return out
}

// SocketSourceSchema returns the schema for socket source
func SocketSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `socket` source ingests logs over TCP or UDP.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"address_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the listen address for the socket.",
				},
				"mode": schema.StringAttribute{
					Required:    true,
					Description: "The protocol used to receive logs.",
					Validators: []validator.String{
						stringvalidator.OneOf("tcp", "udp"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"framing": schema.ListNestedBlock{
					Description: "Defines the framing method for incoming messages.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"method": schema.StringAttribute{
								Required:    true,
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
							"character_delimited": schema.ListNestedBlock{
								Description: "Used when `method` is `character_delimited`. Specifies the delimiter character.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"delimiter": schema.StringAttribute{
											Required:    true,
											Description: "A single ASCII character used as a delimiter.",
										},
									},
								},
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
							},
						},
						Validators: []validator.Object{
							SocketFramingValidator{},
						},
					},
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtMost(1),
					},
				},
				"tls": TlsSchema(),
			},
		},
	}
}
