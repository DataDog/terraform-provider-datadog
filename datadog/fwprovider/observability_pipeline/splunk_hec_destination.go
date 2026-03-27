package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SplunkHECDestinationModel represents the Terraform model for Splunk HEC destination configuration
type SplunkHECDestinationModel struct {
	AutoExtractTimestamp types.Bool           `tfsdk:"auto_extract_timestamp"`
	Encoding             types.String         `tfsdk:"encoding"`
	EndpointUrlKey       types.String         `tfsdk:"endpoint_url_key"`
	TokenKey             types.String         `tfsdk:"token_key"`
	TokenStrategy        types.String         `tfsdk:"token_strategy"`
	Sourcetype           types.String         `tfsdk:"sourcetype"`
	Index                types.String         `tfsdk:"index"`
	IndexedFields        types.List           `tfsdk:"indexed_fields"`
	Buffer               []BufferOptionsModel `tfsdk:"buffer"`
}

// ExpandSplunkHECDestination converts the Terraform model to the Datadog API model
func ExpandSplunkHECDestination(ctx context.Context, id string, inputs types.List, src *SplunkHECDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	s := datadogV2.NewObservabilityPipelineSplunkHecDestinationWithDefaults()
	s.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	s.SetInputs(inputsList)

	if !src.AutoExtractTimestamp.IsNull() {
		s.SetAutoExtractTimestamp(src.AutoExtractTimestamp.ValueBool())
	}
	if !src.Encoding.IsNull() {
		s.SetEncoding(datadogV2.ObservabilityPipelineSplunkHecDestinationEncoding(src.Encoding.ValueString()))
	}
	if !src.Sourcetype.IsNull() {
		s.SetSourcetype(src.Sourcetype.ValueString())
	}
	if !src.Index.IsNull() {
		s.SetIndex(src.Index.ValueString())
	}
	if !src.EndpointUrlKey.IsNull() {
		s.SetEndpointUrlKey(src.EndpointUrlKey.ValueString())
	}
	if !src.TokenKey.IsNull() {
		s.SetTokenKey(src.TokenKey.ValueString())
	}
	if !src.TokenStrategy.IsNull() {
		s.SetTokenStrategy(datadogV2.ObservabilityPipelineSplunkHecDestinationTokenStrategy(src.TokenStrategy.ValueString()))
	}

	if !src.IndexedFields.IsNull() {
		var indexedFields []string
		_ = src.IndexedFields.ElementsAs(ctx, &indexedFields, false)
		s.SetIndexedFields(indexedFields)
	}

	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			s.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineSplunkHecDestination: s,
	}
}

// FlattenSplunkHECDestination converts the Datadog API model to the Terraform model
func FlattenSplunkHECDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineSplunkHecDestination) *SplunkHECDestinationModel {
	if src == nil {
		return nil
	}

	autoExtractTimestamp := types.BoolNull()
	if src.HasAutoExtractTimestamp() {
		autoExtractTimestamp = types.BoolValue(src.GetAutoExtractTimestamp())
	}

	out := &SplunkHECDestinationModel{
		AutoExtractTimestamp: autoExtractTimestamp,
		Sourcetype:           types.StringPointerValue(src.Sourcetype),
		Index:                types.StringPointerValue(src.Index),
		IndexedFields:        types.ListNull(types.StringType),
	}
	if enc, ok := src.GetEncodingOk(); ok && enc != nil {
		out.Encoding = types.StringValue(string(*enc))
	}
	if v, ok := src.GetEndpointUrlKeyOk(); ok {
		out.EndpointUrlKey = types.StringValue(*v)
	}
	if v, ok := src.GetTokenKeyOk(); ok {
		out.TokenKey = types.StringValue(*v)
	}
	if v, ok := src.GetTokenStrategyOk(); ok {
		out.TokenStrategy = types.StringValue(string(*v))
	}
	if indexedFields := src.GetIndexedFields(); len(indexedFields) > 0 {
		out.IndexedFields, _ = types.ListValueFrom(ctx, types.StringType, indexedFields)
	}
	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			out.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	return out
}

// SplunkHECDestinationSchema returns the schema for Splunk HEC destination
func SplunkHECDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `splunk_hec` destination forwards logs to Splunk using the HTTP Event Collector (HEC).",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"auto_extract_timestamp": schema.BoolAttribute{
					Optional:    true,
					Description: "If `true`, Splunk tries to extract timestamps from incoming log events.",
				},
				"encoding": schema.StringAttribute{
					Required:    true,
					Description: "Encoding format for log events.",
					Validators: []validator.String{
						stringvalidator.OneOf("json", "raw_message"),
					},
				},
				"endpoint_url_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the Splunk HEC endpoint URL.",
				},
				"token_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the Splunk HEC token.",
				},
				"token_strategy": schema.StringAttribute{
					Optional:    true,
					Description: "Determines how the token is retrieved. `custom` uses the value from `token_key`; `from_source` passes the token received by the `splunk_hec` source through to the destination.",
					Validators: []validator.String{
						stringvalidator.OneOf("custom", "from_source"),
					},
				},
				"sourcetype": schema.StringAttribute{
					Optional:    true,
					Description: "The Splunk sourcetype to assign to log events.",
				},
				"index": schema.StringAttribute{
					Optional:    true,
					Description: "Optional name of the Splunk index where logs are written.",
				},
				"indexed_fields": schema.ListAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Description: "List of log field names to send as indexed fields to Splunk HEC. Available only when `encoding` is `json`.",
				},
			},
			Blocks: map[string]schema.Block{
				"buffer": BufferOptionsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
