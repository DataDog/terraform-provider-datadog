package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tlsModel represents TLS configuration
type tlsModel struct {
	CrtFile types.String `tfsdk:"crt_file"`
	CaFile  types.String `tfsdk:"ca_file"`
	KeyFile types.String `tfsdk:"key_file"`
}

// expandTls converts the Terraform TLS model to the Datadog API model
func ExpandTls(tlsTF *tlsModel) *datadogV2.ObservabilityPipelineTls {
	if tlsTF == nil {
		return nil
	}
	tls := datadogV2.NewObservabilityPipelineTlsWithDefaults()
	if !tlsTF.CrtFile.IsNull() {
		tls.SetCrtFile(tlsTF.CrtFile.ValueString())
	}
	if !tlsTF.CaFile.IsNull() {
		tls.SetCaFile(tlsTF.CaFile.ValueString())
	}
	if !tlsTF.KeyFile.IsNull() {
		tls.SetKeyFile(tlsTF.KeyFile.ValueString())
	}
	return tls
}

// flattenTls converts the Datadog API TLS model to the Terraform model
func FlattenTls(src *datadogV2.ObservabilityPipelineTls) tlsModel {
	if src == nil {
		return tlsModel{}
	}
	return tlsModel{
		CrtFile: types.StringValue(src.GetCrtFile()),
		CaFile:  types.StringValue(src.GetCaFile()),
		KeyFile: types.StringValue(src.GetKeyFile()),
	}
}

// TlsSchema returns the schema for TLS configuration
func TlsSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Configuration for enabling TLS encryption between the pipeline component and external services.",
		Attributes: map[string]schema.Attribute{
			"crt_file": schema.StringAttribute{
				Optional:    true, // must be optional to make the block optional
				Description: "Path to the TLS client certificate file used to authenticate the pipeline component with upstream or downstream services.",
			},
			"ca_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to the Certificate Authority (CA) file used to validate the server's TLS certificate.",
			},
			"key_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to the private key file associated with the TLS client certificate. Used for mutual TLS authentication.",
			},
		},
	}
}
