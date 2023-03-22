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

// Provider validator string value in ...
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

// Provider validator int64 value at least
var _ provider.ConfigValidator = &ValidateProviderInt64AtLeast{}

func NewValidateProviderInt64AtLeast(p string, v int64) *ValidateProviderInt64AtLeast {
	return &ValidateProviderInt64AtLeast{
		Path: path.Root(p),
		Min:  v,
	}
}

// ValidateProviderInt64AtLeast implements int64 at least val validator
type ValidateProviderInt64AtLeast struct {
	Path path.Path
	Min  int64
}

func (v *ValidateProviderInt64AtLeast) Description(_ context.Context) string {
	return fmt.Sprintf("value must be at least: %q", v.Min)
}

func (v *ValidateProviderInt64AtLeast) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *ValidateProviderInt64AtLeast) ValidateProvider(ctx context.Context, request provider.ValidateConfigRequest, response *provider.ValidateConfigResponse) {
	var value basetypes.Int64Value

	diags := request.Config.GetAttribute(ctx, v.Path, &value)
	if diags != nil {
		response.Diagnostics.Append(diags...)
		return
	}

	if value.IsNull() || value.IsUnknown() {
		return
	}

	if value.ValueInt64() < v.Min {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			v.Path,
			v.Description(ctx),
			fmt.Sprintf("%d", value.ValueInt64()),
		))
	}
}

// Provider validator int64 value at least
var _ provider.ConfigValidator = &ValidateProviderInt64AtLeast{}

func NewValidateProviderInt64Between(p string, min, max int64) *ValidateProviderInt64Between {
	return &ValidateProviderInt64Between{
		Path: path.Root(p),
		Min:  min,
		Max:  max,
	}
}

// ValidateProviderInt64Between implements int64 at least val validator
type ValidateProviderInt64Between struct {
	Path path.Path
	Min  int64
	Max  int64
}

func (v *ValidateProviderInt64Between) Description(_ context.Context) string {
	return fmt.Sprintf("value must be between %d and %d", v.Min, v.Max)
}

func (v *ValidateProviderInt64Between) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *ValidateProviderInt64Between) ValidateProvider(ctx context.Context, request provider.ValidateConfigRequest, response *provider.ValidateConfigResponse) {
	var value basetypes.Int64Value

	diags := request.Config.GetAttribute(ctx, v.Path, &value)
	if diags != nil {
		response.Diagnostics.Append(diags...)
		return
	}

	if value.IsNull() || value.IsUnknown() {
		return
	}

	if value.ValueInt64() < v.Min || value.ValueInt64() > v.Max {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			v.Path,
			v.Description(ctx),
			fmt.Sprintf("%d", value.ValueInt64()),
		))
	}
}
