package frameworkvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ provider.ConfigValidator = &ValidateProviderStringValIn{}

func NewValidateProviderStringValIn(p string, v ...string) *ValidateProviderStringValIn {
	return &ValidateProviderStringValIn{
		Path:   path.Root(p),
		Values: v,
	}
}

// ValidateProviderStringValIn implements case-insensitive string in validator
type ValidateProviderStringValIn struct {
	Path   path.Path
	Values []string
}

func (v *ValidateProviderStringValIn) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of: %q", v.Values)
}

func (v *ValidateProviderStringValIn) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *ValidateProviderStringValIn) ValidateProvider(ctx context.Context, request provider.ValidateConfigRequest, response *provider.ValidateConfigResponse) {
	var value attr.Value
	diags := request.Config.GetAttribute(ctx, v.Path, &value)
	if diags != nil {
		response.Diagnostics.Append(diags...)
		return
	}

	if value.IsNull() || value.IsUnknown() {
		return
	}

	for _, acceptableValue := range v.Values {
		if value.Equal(basetypes.NewStringValue(acceptableValue)) {
			return
		}
	}

	response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
		v.Path,
		v.Description(ctx),
		value.String(),
	))
}
