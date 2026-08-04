package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CloudPremDestinationModel represents the Terraform model for cloud_prem destination configuration
type CloudPremDestinationModel struct {
	EndpointUrlKey types.String         `tfsdk:"endpoint_url_key"`
	Tls            []ClientTlsModel     `tfsdk:"tls"`
	Buffer         []BufferOptionsModel `tfsdk:"buffer"`
}

// ExpandCloudPremDestination converts the Terraform model to the Datadog API model
func ExpandCloudPremDestination(ctx context.Context, id string, inputs types.List, src *CloudPremDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	d := datadogV2.NewObservabilityPipelineCloudPremDestinationWithDefaults()
	d.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	d.SetInputs(inputsList)
	if !src.EndpointUrlKey.IsNull() {
		d.SetEndpointUrlKey(src.EndpointUrlKey.ValueString())
	}

	if len(src.Tls) > 0 {
		d.Tls = ExpandClientTls(src.Tls)
	}

	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			d.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineCloudPremDestination: d,
	}
}

// FlattenCloudPremDestination converts the Datadog API model to the Terraform model
func FlattenCloudPremDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineCloudPremDestination) *CloudPremDestinationModel {
	if src == nil {
		return nil
	}

	out := &CloudPremDestinationModel{}
	if v, ok := src.GetEndpointUrlKeyOk(); ok {
		out.EndpointUrlKey = types.StringValue(*v)
	}
	if src.Tls != nil {
		out.Tls = FlattenClientTls(src.Tls)
	}
	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			out.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}
	return out
}

// CloudPremDestinationSchema returns the schema for cloud_prem destination
func CloudPremDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `cloud_prem` destination sends logs to Datadog CloudPrem.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"endpoint_url_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the endpoint URL.",
				},
			},
			Blocks: map[string]schema.Block{
				"tls":    ClientTlsSchema(),
				"buffer": BufferOptionsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
