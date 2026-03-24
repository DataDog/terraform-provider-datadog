package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SocketDestinationModel represents the Terraform model for socket destination configuration
type SocketDestinationModel struct {
	AddressKey types.String         `tfsdk:"address_key"`
	Mode       types.String         `tfsdk:"mode"`
	Encoding   types.String         `tfsdk:"encoding"`
	Framing    []SocketFramingModel `tfsdk:"framing"`
	Tls        []TlsModel           `tfsdk:"tls"`
	Buffer     []BufferOptionsModel `tfsdk:"buffer"`
}

// ExpandSocketDestination converts the Terraform model to the Datadog API model
func ExpandSocketDestination(ctx context.Context, id string, inputs types.List, src *SocketDestinationModel) (datadogV2.ObservabilityPipelineConfigDestinationItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := datadogV2.NewObservabilityPipelineSocketDestinationWithDefaults()
	s.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	s.SetInputs(inputsList)

	if !src.AddressKey.IsNull() {
		s.SetAddressKey(src.AddressKey.ValueString())
	}
	s.SetMode(datadogV2.ObservabilityPipelineSocketDestinationMode(src.Mode.ValueString()))
	s.SetEncoding(datadogV2.ObservabilityPipelineSocketDestinationEncoding(src.Encoding.ValueString()))

	switch src.Framing[0].Method.ValueString() {
	case "newline_delimited":
		s.Framing = datadogV2.ObservabilityPipelineSocketDestinationFraming{
			ObservabilityPipelineSocketDestinationFramingNewlineDelimited: &datadogV2.ObservabilityPipelineSocketDestinationFramingNewlineDelimited{
				Method: "newline_delimited",
			},
		}
	case "bytes":
		s.Framing = datadogV2.ObservabilityPipelineSocketDestinationFraming{
			ObservabilityPipelineSocketDestinationFramingBytes: &datadogV2.ObservabilityPipelineSocketDestinationFramingBytes{
				Method: "bytes",
			},
		}
	case "character_delimited":
		charDelimited := &datadogV2.ObservabilityPipelineSocketDestinationFramingCharacterDelimited{
			Method: "character_delimited",
		}
		if len(src.Framing[0].CharacterDelimited) > 0 {
			charDelimited.Delimiter = src.Framing[0].CharacterDelimited[0].Delimiter.ValueString()
		}
		s.Framing = datadogV2.ObservabilityPipelineSocketDestinationFraming{
			ObservabilityPipelineSocketDestinationFramingCharacterDelimited: charDelimited,
		}
	}

	if len(src.Tls) > 0 {
		s.Tls = ExpandTls(src.Tls)
	}

	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			s.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSocketDestination: s,
	}, diags
}

// FlattenSocketDestination converts the Datadog API model to the Terraform model
func FlattenSocketDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSocketDestination) *SocketDestinationModel {
	if src == nil {
		return nil
	}

	out := &SocketDestinationModel{
		Mode:     types.StringValue(string(src.GetMode())),
		Encoding: types.StringValue(string(src.GetEncoding())),
	}
	if v, ok := src.GetAddressKeyOk(); ok {
		out.AddressKey = types.StringValue(*v)
	}

	if src.Tls != nil {
		out.Tls = FlattenTls(src.Tls)
	}

	outFraming := SocketFramingModel{}
	switch {
	case src.Framing.ObservabilityPipelineSocketDestinationFramingNewlineDelimited != nil:
		outFraming.Method = types.StringValue("newline_delimited")
	case src.Framing.ObservabilityPipelineSocketDestinationFramingBytes != nil:
		outFraming.Method = types.StringValue("bytes")
	case src.Framing.ObservabilityPipelineSocketDestinationFramingCharacterDelimited != nil:
		outFraming.Method = types.StringValue("character_delimited")
		outFraming.CharacterDelimited = []SocketFramingCharacterDelimitedModel{{
			Delimiter: types.StringValue(src.Framing.ObservabilityPipelineSocketDestinationFramingCharacterDelimited.Delimiter),
		}}
	}
	out.Framing = []SocketFramingModel{outFraming}

	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			out.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	return out
}

// SocketDestinationSchema returns the schema for socket destination
func SocketDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `socket` destination sends logs over TCP or UDP to a remote server.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"address_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the socket address (host:port).",
				},
				"mode": schema.StringAttribute{
					Required:    true,
					Description: "The protocol used to send logs.",
					Validators: []validator.String{
						stringvalidator.OneOf("tcp", "udp"),
					},
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
				"buffer": BufferOptionsSchema(),
				"framing": schema.ListNestedBlock{
					Description: "Defines the framing method for outgoing messages.",
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
