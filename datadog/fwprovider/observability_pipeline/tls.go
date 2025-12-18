package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TlsModel represents TLS configuration
type TlsModel struct {
	CrtFile types.String `tfsdk:"crt_file"`
	CaFile  types.String `tfsdk:"ca_file"`
	KeyFile types.String `tfsdk:"key_file"`
}

// ExpandTls converts the Terraform TLS model to the Datadog API model
func ExpandTls(tlsTF []TlsModel) *datadogV2.ObservabilityPipelineTls {
	if len(tlsTF) == 0 {
		return nil
	}

	tlsItem := tlsTF[0]
	tls := &datadogV2.ObservabilityPipelineTls{
		CrtFile: tlsItem.CrtFile.ValueString(),
	}
	if !tlsItem.CaFile.IsNull() {
		tls.SetCaFile(tlsItem.CaFile.ValueString())
	}
	if !tlsItem.KeyFile.IsNull() {
		tls.SetKeyFile(tlsItem.KeyFile.ValueString())
	}
	return tls
}

// FlattenTls converts the Datadog API TLS model to the Terraform model
func FlattenTls(src *datadogV2.ObservabilityPipelineTls) []TlsModel {
	if src == nil {
		return []TlsModel{}
	}
	return []TlsModel{
		{
			CrtFile: types.StringValue(src.CrtFile),
			CaFile:  types.StringPointerValue(src.CaFile),
			KeyFile: types.StringPointerValue(src.KeyFile),
		},
	}
}

// TlsSchema returns the schema for TLS configuration
func TlsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Configuration for enabling TLS encryption between the pipeline component and external services.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"crt_file": schema.StringAttribute{
					Required:    true,
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
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
