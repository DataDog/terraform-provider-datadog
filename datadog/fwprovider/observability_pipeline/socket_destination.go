package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SocketDestinationModel represents the Terraform model for socket destination configuration
type SocketDestinationModel struct {
	Id       types.String       `tfsdk:"id"`
	Inputs   types.List         `tfsdk:"inputs"`
	Mode     types.String       `tfsdk:"mode"`
	Encoding types.String       `tfsdk:"encoding"`
	Framing  SocketFramingModel `tfsdk:"framing"`
	Tls      *tlsModel          `tfsdk:"tls"`
}

// ExpandSocketDestination converts the Terraform model to the Datadog API model
func ExpandSocketDestination(ctx context.Context, src *SocketDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	s := datadogV2.NewObservabilityPipelineSocketDestinationWithDefaults()
	s.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	s.SetInputs(inputs)

	s.SetMode(datadogV2.ObservabilityPipelineSocketDestinationMode(src.Mode.ValueString()))
	s.SetEncoding(datadogV2.ObservabilityPipelineSocketDestinationEncoding(src.Encoding.ValueString()))

	switch src.Framing.Method.ValueString() {
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
		s.Framing = datadogV2.ObservabilityPipelineSocketDestinationFraming{
			ObservabilityPipelineSocketDestinationFramingCharacterDelimited: &datadogV2.ObservabilityPipelineSocketDestinationFramingCharacterDelimited{
				Method:    "character_delimited",
				Delimiter: src.Framing.CharacterDelimited.Delimiter.ValueString(),
			},
		}
	}

	if src.Tls != nil {
		s.SetTls(*ExpandTls(src.Tls))
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSocketDestination: s,
	}
}

// FlattenSocketDestination converts the Datadog API model to the Terraform model
func FlattenSocketDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSocketDestination) *SocketDestinationModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)

	out := &SocketDestinationModel{
		Id:       types.StringValue(src.GetId()),
		Inputs:   inputs,
		Mode:     types.StringValue(string(src.GetMode())),
		Encoding: types.StringValue(string(src.GetEncoding())),
	}

	switch {
	case src.Framing.ObservabilityPipelineSocketDestinationFramingNewlineDelimited != nil:
		out.Framing.Method = types.StringValue("newline_delimited")
	case src.Framing.ObservabilityPipelineSocketDestinationFramingBytes != nil:
		out.Framing.Method = types.StringValue("bytes")
	case src.Framing.ObservabilityPipelineSocketDestinationFramingCharacterDelimited != nil:
		out.Framing.Method = types.StringValue("character_delimited")
		out.Framing.CharacterDelimited = &SocketFramingCharacterDelimitedModel{
			Delimiter: types.StringValue(src.Framing.ObservabilityPipelineSocketDestinationFramingCharacterDelimited.Delimiter),
		}
	}

	if src.Tls != nil {
		out.Tls = &tlsModel{
			CrtFile: types.StringValue(src.Tls.GetCrtFile()),
			CaFile:  types.StringValue(src.Tls.GetCaFile()),
			KeyFile: types.StringValue(src.Tls.GetKeyFile()),
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
				"id": schema.StringAttribute{
					Required:    true,
					Description: "The unique identifier for this destination.",
				},
				"inputs": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
					Description: "A list of component IDs whose output is used as the `input` for this destination.",
				},
				"mode": schema.StringAttribute{
					Required:    true,
					Description: "The protocol used to send logs. Must be either `tcp` or `udp`.",
					Validators: []validator.String{
						stringvalidator.OneOf("tcp", "udp"),
					},
				},
				"encoding": schema.StringAttribute{
					Required:    true,
					Description: "Encoding format for log events. Must be either `json` or `raw_message`.",
					Validators: []validator.String{
						stringvalidator.OneOf("json", "raw_message"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"framing": schema.SingleNestedBlock{
					Description: "Defines the framing method for outgoing messages.",
					Attributes: map[string]schema.Attribute{
						"method": schema.StringAttribute{
							Required:    true,
							Description: "The framing method. One of: `newline_delimited`, `bytes`, `character_delimited`.",
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
						"character_delimited": schema.SingleNestedBlock{
							Description: "Used when `method` is `character_delimited`. Specifies the delimiter character.",
							Attributes: map[string]schema.Attribute{
								"delimiter": schema.StringAttribute{
									Optional:    true,
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
